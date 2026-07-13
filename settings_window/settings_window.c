//go:build ignore
// +build ignore

#define UNICODE
#define _UNICODE
#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <windowsx.h>
#include <dwmapi.h>
#include <wincodec.h>
#include <gdiplus.h>
#include <math.h>
#include "settings_window.h"

#define SETTINGS_CLASS_NAME L"BeeModernSettingsWindow"
#define SETTINGS_TITLE      L"Bee 插件设置"
#define WM_BEE_ACTIVATE     (WM_APP + 1)
#define WM_BEE_CLOSE        (WM_APP + 2)



#define COLOR_BG            0xFFF6F7FBu
#define COLOR_SURFACE       0xFFFFFFFFu
#define COLOR_TEXT          0xFF182230u
#define COLOR_MUTED         0xFF667085u
#define COLOR_BORDER        0xFFE4E7ECu
#define COLOR_PRIMARY       0xFF6366F1u
#define COLOR_PRIMARY_HOVER 0xFF5558E8u
#define COLOR_PRIMARY_DOWN  0xFF484BD6u
#define COLOR_SUCCESS       0xFF12B76Au
#define COLOR_CLOSE_HOVER   0xFFF2F4F7u

static volatile LONG g_starting = 0;
static HWND g_window = NULL;
static HANDLE g_thread = NULL;
static DWORD g_thread_id = 0;

static RECT g_close_rect;
static RECT g_save_rect;
static BOOL g_close_hot = FALSE;
static BOOL g_save_hot = FALSE;
static BOOL g_save_down = FALSE;
static BOOL g_tracking_mouse = FALSE;
static BOOL g_saved_notice = FALSE;
static UINT_PTR g_notice_timer = 0;


static ULONG_PTR g_gdiplus_token = 0;


static UINT BeeDpi(HWND hwnd) {
    HMODULE user32 = GetModuleHandleW(L"user32.dll");
    if (user32) {
        typedef UINT (WINAPI *GetDpiForWindowFn)(HWND);
        GetDpiForWindowFn get_dpi = (GetDpiForWindowFn)GetProcAddress(user32, "GetDpiForWindow");
        if (get_dpi && hwnd) return get_dpi(hwnd);
    }
    HDC dc = GetDC(hwnd);
    UINT dpi = dc ? (UINT)GetDeviceCaps(dc, LOGPIXELSX) : 96;
    if (dc) ReleaseDC(hwnd, dc);
    return dpi ? dpi : 96;
}

static INT S(HWND hwnd, INT value) {
    return MulDiv(value, (INT)BeeDpi(hwnd), 96);
}

static void ExtendDwmShadow(HWND hwnd) {
    // 保留可组合的系统窗口框架，但把非客户区压缩为 0；再向客户区扩展
    // 1px DWM frame，让系统继续为窗口生成原生阴影。
    HMODULE dwmapi = LoadLibraryW(L"dwmapi.dll");
    if (!dwmapi) return;

    typedef HRESULT (WINAPI *DwmExtendFrameIntoClientAreaFn)(HWND, const MARGINS *);
    DwmExtendFrameIntoClientAreaFn extend_frame =
        (DwmExtendFrameIntoClientAreaFn)GetProcAddress(dwmapi, "DwmExtendFrameIntoClientArea");
    if (extend_frame) {
        MARGINS margins = {1, 1, 1, 1};
        extend_frame(hwnd, &margins);
    }
    FreeLibrary(dwmapi);
}


static GpRectF RF(REAL x, REAL y, REAL width, REAL height) {
    GpRectF rect;
    rect.X = x; rect.Y = y; rect.Width = width; rect.Height = height;
    return rect;
}

static void FillRoundRect(GpGraphics *graphics, GpBrush *brush, GpRectF rect, REAL radius) {
    GpPath *path = NULL;
    REAL diameter = radius * 2.0f;
    if (GdipCreatePath(FillModeAlternate, &path) != Ok || !path) return;
    GdipAddPathArc(path, rect.X, rect.Y, diameter, diameter, 180.0f, 90.0f);
    GdipAddPathArc(path, rect.X + rect.Width - diameter, rect.Y, diameter, diameter, 270.0f, 90.0f);
    GdipAddPathArc(path, rect.X + rect.Width - diameter, rect.Y + rect.Height - diameter, diameter, diameter, 0.0f, 90.0f);
    GdipAddPathArc(path, rect.X, rect.Y + rect.Height - diameter, diameter, diameter, 90.0f, 90.0f);
    GdipClosePathFigure(path);
    GdipFillPath(graphics, brush, path);
    GdipDeletePath(path);
}

static void DrawRoundRect(GpGraphics *graphics, GpPen *pen, GpRectF rect, REAL radius) {
    GpPath *path = NULL;
    REAL diameter = radius * 2.0f;
    if (GdipCreatePath(FillModeAlternate, &path) != Ok || !path) return;
    GdipAddPathArc(path, rect.X, rect.Y, diameter, diameter, 180.0f, 90.0f);
    GdipAddPathArc(path, rect.X + rect.Width - diameter, rect.Y, diameter, diameter, 270.0f, 90.0f);
    GdipAddPathArc(path, rect.X + rect.Width - diameter, rect.Y + rect.Height - diameter, diameter, diameter, 0.0f, 90.0f);
    GdipAddPathArc(path, rect.X, rect.Y + rect.Height - diameter, diameter, diameter, 90.0f, 90.0f);
    GdipClosePathFigure(path);
    GdipDrawPath(graphics, pen, path);
    GdipDeletePath(path);
}

static void DrawTextAt(GpGraphics *graphics, const WCHAR *text, REAL x, REAL y, REAL width, REAL height,
                       REAL size, INT style, ARGB color, StringAlignment alignment) {
    (void)graphics; (void)text; (void)x; (void)y; (void)width; (void)height;
    (void)size; (void)style; (void)color; (void)alignment;
}

static void DrawNativeText(HDC dc, const WCHAR *text, RECT rect, INT pixel_height,
                           INT weight, COLORREF color, UINT align) {
    HFONT font = CreateFontW(-pixel_height, 0, 0, 0, weight, FALSE, FALSE, FALSE,
                             DEFAULT_CHARSET, OUT_TT_PRECIS, CLIP_DEFAULT_PRECIS,
                             CLEARTYPE_QUALITY, DEFAULT_PITCH | FF_DONTCARE,
                             L"Microsoft YaHei UI");
    if (!font) font = (HFONT)GetStockObject(DEFAULT_GUI_FONT);
    HFONT old_font = (HFONT)SelectObject(dc, font);
    SetBkMode(dc, TRANSPARENT);
    SetTextColor(dc, color);
    DrawTextW(dc, text, -1, &rect,
              DT_SINGLELINE | DT_VCENTER | DT_END_ELLIPSIS | DT_NOPREFIX | align);
    SelectObject(dc, old_font);
    if (font != GetStockObject(DEFAULT_GUI_FONT)) DeleteObject(font);
}

static void DrawAllText(HWND hwnd, HDC dc, INT width, INT height) {
    INT margin = S(hwnd, 28);
    INT top = S(hwnd, 82);
    INT card_width = width - margin * 2;
    INT button_width = S(hwnd, 112);
    RECT rect;

    rect = (RECT){margin, S(hwnd, 17), width - margin, S(hwnd, 49)};
    DrawNativeText(dc, L"插件设置", rect, S(hwnd, 22), FW_BOLD, RGB(24, 34, 48), DT_LEFT);
    rect = (RECT){margin, S(hwnd, 48), width - margin, S(hwnd, 70)};
    DrawNativeText(dc, L"简洁、安全地管理插件配置", rect, S(hwnd, 13), FW_NORMAL, RGB(102, 112, 133), DT_LEFT);

    rect = (RECT){margin + S(hwnd, 20), top + S(hwnd, 22), margin + S(hwnd, 62), top + S(hwnd, 64)};
    DrawNativeText(dc, L"B", rect, S(hwnd, 18), FW_BOLD, RGB(99, 102, 241), DT_CENTER);
    rect = (RECT){margin + S(hwnd, 76), top + S(hwnd, 20), margin + card_width - S(hwnd, 22), top + S(hwnd, 48)};
    DrawNativeText(dc, L"Bee Go 插件", rect, S(hwnd, 17), FW_BOLD, RGB(24, 34, 48), DT_LEFT);
    rect = (RECT){margin + S(hwnd, 76), top + S(hwnd, 49), margin + card_width - S(hwnd, 22), top + S(hwnd, 71)};
    DrawNativeText(dc, L"当前模板已启用原生现代化设置窗口", rect, S(hwnd, 13), FW_NORMAL, RGB(102, 112, 133), DT_LEFT);

    rect = (RECT){margin + S(hwnd, 42), top + S(hwnd, 110), margin + S(hwnd, 90), top + S(hwnd, 140)};
    DrawNativeText(dc, L"运行正常", rect, S(hwnd, 12), FW_BOLD, RGB(18, 183, 106), DT_CENTER);
    rect = (RECT){margin + S(hwnd, 116), top + S(hwnd, 110), margin + card_width - S(hwnd, 20), top + S(hwnd, 140)};
    DrawNativeText(dc, L"Win32 + GDI+ · 单 DLL", rect, S(hwnd, 12), FW_NORMAL, RGB(102, 112, 133), DT_LEFT);

    rect = g_save_rect;
    DrawNativeText(dc, L"保存设置", rect, S(hwnd, 13), FW_BOLD, RGB(255, 255, 255), DT_CENTER);
    rect = (RECT){margin, g_save_rect.top, width - margin - button_width - S(hwnd, 18), g_save_rect.bottom};
    DrawNativeText(dc, g_saved_notice ? L"设置已保存" : L"修改后点击保存", rect,
                   S(hwnd, 12), FW_NORMAL,
                   g_saved_notice ? RGB(18, 183, 106) : RGB(102, 112, 133), DT_LEFT);
}

static void DrawWindow(HWND hwnd, HDC target) {
    RECT client;
    GetClientRect(hwnd, &client);
    INT width = client.right;
    INT height = client.bottom;
    HDC memory = CreateCompatibleDC(target);
    HBITMAP bitmap = CreateCompatibleBitmap(target, width, height);
    HBITMAP old_bitmap = (HBITMAP)SelectObject(memory, bitmap);
    GpGraphics *graphics = NULL;
    GdipCreateFromHDC(memory, &graphics);
    if (!graphics) goto cleanup;

    GdipSetSmoothingMode(graphics, SmoothingModeAntiAlias);
    GdipSetTextRenderingHint(graphics, TextRenderingHintClearTypeGridFit);

    GpSolidFill *fill = NULL;
    GpPen *pen = NULL;
    GdipCreateSolidFill(COLOR_BG, &fill);
    GdipFillRectangleI(graphics, fill, 0, 0, width, height);
    GdipDeleteBrush(fill);

    REAL scale = (REAL)BeeDpi(hwnd) / 96.0f;
    REAL margin = 28.0f * scale;
    REAL top = 82.0f * scale;
    REAL card_h = 170.0f * scale;
    REAL radius = 14.0f * scale;

    DrawTextAt(graphics, L"插件设置", margin, 17.0f * scale, width - margin * 2, 32.0f * scale,
               22.0f * scale, FontStyleBold, COLOR_TEXT, StringAlignmentNear);
    DrawTextAt(graphics, L"简洁、安全地管理插件配置", margin, 48.0f * scale, width - margin * 2, 22.0f * scale,
               13.0f * scale, FontStyleRegular, COLOR_MUTED, StringAlignmentNear);

    g_close_rect.left = width - S(hwnd, 52); g_close_rect.top = S(hwnd, 18);
    g_close_rect.right = width - S(hwnd, 18); g_close_rect.bottom = S(hwnd, 52);
    if (g_close_hot) {
        GdipCreateSolidFill(COLOR_CLOSE_HOVER, &fill);
        FillRoundRect(graphics, (GpBrush *)fill, RF((REAL)g_close_rect.left, (REAL)g_close_rect.top,
                      (REAL)(g_close_rect.right-g_close_rect.left), (REAL)(g_close_rect.bottom-g_close_rect.top)), 8.0f * scale);
        GdipDeleteBrush(fill);
    }
    GdipCreatePen1(COLOR_MUTED, 1.7f * scale, UnitPixel, &pen);
    GdipDrawLine(graphics, pen, (REAL)g_close_rect.left + 11*scale, (REAL)g_close_rect.top + 11*scale,
                 (REAL)g_close_rect.right - 11*scale, (REAL)g_close_rect.bottom - 11*scale);
    GdipDrawLine(graphics, pen, (REAL)g_close_rect.right - 11*scale, (REAL)g_close_rect.top + 11*scale,
                 (REAL)g_close_rect.left + 11*scale, (REAL)g_close_rect.bottom - 11*scale);
    GdipDeletePen(pen);


    GpRectF card = RF(margin, top, width - margin * 2, card_h);
    GdipCreateSolidFill(COLOR_SURFACE, &fill);
    FillRoundRect(graphics, (GpBrush *)fill, card, radius);
    GdipDeleteBrush(fill);

    GdipCreateSolidFill(0xFFEFF0FFu, &fill);
    FillRoundRect(graphics, (GpBrush *)fill, RF(margin + 20*scale, top + 22*scale, 42*scale, 42*scale), 11*scale);
    GdipDeleteBrush(fill);
    DrawTextAt(graphics, L"B", margin + 20*scale, top + 22*scale, 42*scale, 42*scale,
               18*scale, FontStyleBold, COLOR_PRIMARY, StringAlignmentCenter);
    DrawTextAt(graphics, L"Bee Go 插件", margin + 76*scale, top + 20*scale, card.Width - 98*scale, 28*scale,
               17*scale, FontStyleBold, COLOR_TEXT, StringAlignmentNear);
    DrawTextAt(graphics, L"当前模板已启用原生现代化设置窗口", margin + 76*scale, top + 49*scale, card.Width - 98*scale, 22*scale,
               12.5f*scale, FontStyleRegular, COLOR_MUTED, StringAlignmentNear);

    GdipCreatePen1(COLOR_BORDER, 1.0f * scale, UnitPixel, &pen);
    GdipDrawLine(graphics, pen, margin + 20*scale, top + 88*scale, margin + card.Width - 20*scale, top + 88*scale);
    GdipDeletePen(pen);

    GdipCreateSolidFill(0xFFE7F8F0u, &fill);
    FillRoundRect(graphics, (GpBrush *)fill, RF(margin + 20*scale, top + 110*scale, 78*scale, 30*scale), 15*scale);
    GdipDeleteBrush(fill);
    GdipCreateSolidFill(COLOR_SUCCESS, &fill);
    GdipFillEllipse(graphics, (GpBrush *)fill, margin + 32*scale, top + 122*scale, 6*scale, 6*scale);
    GdipDeleteBrush(fill);
    DrawTextAt(graphics, L"运行正常", margin + 42*scale, top + 110*scale, 48*scale, 30*scale,
               11.5f*scale, FontStyleBold, COLOR_SUCCESS, StringAlignmentCenter);
    DrawTextAt(graphics, L"Win32 + GDI+ · 单 DLL", margin + 116*scale, top + 110*scale, card.Width - 136*scale, 30*scale,
               12*scale, FontStyleRegular, COLOR_MUTED, StringAlignmentNear);

    INT button_w = S(hwnd, 112), button_h = S(hwnd, 42);
    g_save_rect.left = width - margin - button_w;
    g_save_rect.top = height - S(hwnd, 66);
    g_save_rect.right = width - margin;
    g_save_rect.bottom = g_save_rect.top + button_h;
    ARGB button_color = g_save_down ? COLOR_PRIMARY_DOWN : (g_save_hot ? COLOR_PRIMARY_HOVER : COLOR_PRIMARY);
    GdipCreateSolidFill(button_color, &fill);
    FillRoundRect(graphics, (GpBrush *)fill, RF((REAL)g_save_rect.left, (REAL)g_save_rect.top,
                  (REAL)button_w, (REAL)button_h), 10*scale);
    GdipDeleteBrush(fill);
    DrawTextAt(graphics, L"保存设置", (REAL)g_save_rect.left, (REAL)g_save_rect.top, (REAL)button_w, (REAL)button_h,
               13*scale, FontStyleBold, 0xFFFFFFFFu, StringAlignmentCenter);

    DrawTextAt(graphics, g_saved_notice ? L"设置已保存" : L"修改后点击保存", margin, (REAL)g_save_rect.top,
               width - margin*2 - button_w - S(hwnd, 18), (REAL)button_h,
               12*scale, FontStyleRegular, g_saved_notice ? COLOR_SUCCESS : COLOR_MUTED, StringAlignmentNear);

    GdipDeleteGraphics(graphics);
    graphics = NULL;

    // 所有 GDI+ 图形完成后，最后用 Win32 GDI 统一绘制文字。
    // 不能把 DrawTextW 穿插在 GDI+ 绘制中，否则后续 GDI+ 刷新会覆盖文字。
    DrawAllText(hwnd, memory, width, height);
    BitBlt(target, 0, 0, width, height, memory, 0, 0, SRCCOPY);
cleanup:
    if (graphics) GdipDeleteGraphics(graphics);
    SelectObject(memory, old_bitmap);
    DeleteObject(bitmap);
    DeleteDC(memory);
}

static void UpdateHover(HWND hwnd, POINT point) {
    BOOL close_hot = PtInRect(&g_close_rect, point);
    BOOL save_hot = PtInRect(&g_save_rect, point);
    if (close_hot != g_close_hot || save_hot != g_save_hot) {
        g_close_hot = close_hot;
        g_save_hot = save_hot;
        InvalidateRect(hwnd, NULL, FALSE);
    }
    if (!g_tracking_mouse) {
        TRACKMOUSEEVENT tracking = {sizeof(TRACKMOUSEEVENT), TME_LEAVE, hwnd, 0};
        TrackMouseEvent(&tracking);
        g_tracking_mouse = TRUE;
    }
}

static LRESULT CALLBACK SettingsProc(HWND hwnd, UINT message, WPARAM wparam, LPARAM lparam) {
    switch (message) {
        case WM_NCCREATE:
            g_window = hwnd;
            return TRUE;
        case WM_NCCALCSIZE:
            // 去除标题栏和边框占用的非客户区，客户区覆盖整个窗口。
            return 0;
        case WM_NCHITTEST: {
            POINT point = {GET_X_LPARAM(lparam), GET_Y_LPARAM(lparam)};
            ScreenToClient(hwnd, &point);
            if (point.y >= 0 && point.y < S(hwnd, 72) && !PtInRect(&g_close_rect, point)) return HTCAPTION;
            return HTCLIENT;
        }
        case WM_ERASEBKGND:
            return 1;
        case WM_PAINT: {
            PAINTSTRUCT paint;
            HDC dc = BeginPaint(hwnd, &paint);
            DrawWindow(hwnd, dc);
            EndPaint(hwnd, &paint);
            return 0;
        }
        case WM_MOUSEMOVE: {
            POINT point = {GET_X_LPARAM(lparam), GET_Y_LPARAM(lparam)};
            UpdateHover(hwnd, point);
            return 0;
        }
        case WM_MOUSELEAVE:
            g_tracking_mouse = FALSE;
            g_close_hot = g_save_hot = FALSE;
            InvalidateRect(hwnd, NULL, FALSE);
            return 0;
        case WM_LBUTTONDOWN: {
            POINT point = {GET_X_LPARAM(lparam), GET_Y_LPARAM(lparam)};
            if (PtInRect(&g_save_rect, point)) {
                g_save_down = TRUE;
                SetCapture(hwnd);
                InvalidateRect(hwnd, &g_save_rect, FALSE);
            }
            return 0;
        }
        case WM_LBUTTONUP: {
            POINT point = {GET_X_LPARAM(lparam), GET_Y_LPARAM(lparam)};
            BOOL was_down = g_save_down;
            g_save_down = FALSE;
            if (GetCapture() == hwnd) ReleaseCapture();
            if (PtInRect(&g_close_rect, point)) {
                PostMessageW(hwnd, WM_CLOSE, 0, 0);
            } else if (was_down && PtInRect(&g_save_rect, point)) {
                g_saved_notice = TRUE;
                if (g_notice_timer) KillTimer(hwnd, g_notice_timer);
                g_notice_timer = SetTimer(hwnd, 1, 1800, NULL);
                InvalidateRect(hwnd, NULL, FALSE);
            } else {
                InvalidateRect(hwnd, &g_save_rect, FALSE);
            }
            return 0;
        }
        case WM_TIMER:
            if (wparam == 1) {
                KillTimer(hwnd, 1);
                g_notice_timer = 0;
                g_saved_notice = FALSE;
                InvalidateRect(hwnd, NULL, FALSE);
            }
            return 0;
        case WM_SETCURSOR: {
            POINT point;
            GetCursorPos(&point);
            ScreenToClient(hwnd, &point);
            if (PtInRect(&g_close_rect, point) || PtInRect(&g_save_rect, point)) {
                SetCursor(LoadCursorW(NULL, IDC_HAND));
                return TRUE;
            }
            break;
        }
        case WM_KEYDOWN:
            if (wparam == VK_ESCAPE) PostMessageW(hwnd, WM_CLOSE, 0, 0);
            return 0;
        case WM_BEE_ACTIVATE:
            ShowWindow(hwnd, SW_RESTORE);
            SetForegroundWindow(hwnd);
            return 0;
        case WM_BEE_CLOSE:
            PostMessageW(hwnd, WM_CLOSE, 0, 0);
            return 0;
        case WM_CLOSE:
            DestroyWindow(hwnd);
            return 0;
        case WM_DESTROY:
            if (g_notice_timer) KillTimer(hwnd, g_notice_timer);
            g_notice_timer = 0;
            g_window = NULL;
            PostQuitMessage(0);
            return 0;
    }
    return DefWindowProcW(hwnd, message, wparam, lparam);
}

static DWORD WINAPI SettingsThread(LPVOID unused) {
    (void)unused;
    GdiplusStartupInput input;
    input.GdiplusVersion = 1;
    input.DebugEventCallback = NULL;
    input.SuppressBackgroundThread = FALSE;
    input.SuppressExternalCodecs = FALSE;
    if (GdiplusStartup(&g_gdiplus_token, &input, NULL) != Ok) goto done;

    HINSTANCE instance = GetModuleHandleW(NULL);
    WNDCLASSEXW window_class;
    ZeroMemory(&window_class, sizeof(window_class));
    window_class.cbSize = sizeof(window_class);
    window_class.style = CS_HREDRAW | CS_VREDRAW;
    window_class.lpfnWndProc = SettingsProc;
    window_class.hInstance = instance;
    window_class.hCursor = LoadCursorW(NULL, IDC_ARROW);
    window_class.hIcon = LoadIconW(NULL, IDI_APPLICATION);
    window_class.hbrBackground = NULL;
    window_class.lpszClassName = SETTINGS_CLASS_NAME;
    RegisterClassExW(&window_class);

    UINT dpi = 96;
    HDC screen_dc = GetDC(NULL);
    if (screen_dc) { dpi = (UINT)GetDeviceCaps(screen_dc, LOGPIXELSX); ReleaseDC(NULL, screen_dc); }
    INT client_w = MulDiv(520, dpi, 96);
    INT client_h = MulDiv(350, dpi, 96);
    RECT outer = {0, 0, client_w, client_h};
    // 样式中保留系统窗口框架，让 DWM 继续识别并提供阴影、贴靠排列等能力；
    // WM_NCCALCSIZE 会把可见非客户区移除。没有 WS_THICKFRAME，因此不能缩放。
    DWORD window_style = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_MINIMIZEBOX;
    AdjustWindowRectEx(&outer, window_style, FALSE, WS_EX_APPWINDOW | WS_EX_CONTROLPARENT);
    INT width = outer.right - outer.left;
    INT height = outer.bottom - outer.top;
    RECT work;
    SystemParametersInfoW(SPI_GETWORKAREA, 0, &work, 0);
    INT x = work.left + ((work.right - work.left) - width) / 2;
    INT y = work.top + ((work.bottom - work.top) - height) / 2;

    HWND hwnd = CreateWindowExW(WS_EX_APPWINDOW | WS_EX_CONTROLPARENT,
                                SETTINGS_CLASS_NAME, SETTINGS_TITLE,
                                window_style,
                                x, y, width, height, NULL, NULL, instance, NULL);
    if (!hwnd) goto shutdown;
    ExtendDwmShadow(hwnd);
    ShowWindow(hwnd, SW_SHOW);
    UpdateWindow(hwnd);
    SetForegroundWindow(hwnd);

    MSG message;
    while (GetMessageW(&message, NULL, 0, 0) > 0) {
        TranslateMessage(&message);
        DispatchMessageW(&message);
    }

shutdown:
    GdiplusShutdown(g_gdiplus_token);
    g_gdiplus_token = 0;
done:
    g_window = NULL;
    InterlockedExchange(&g_starting, 0);
    return 0;
}

void BeeSettingsOpen(void) {
    HWND hwnd = g_window;
    if (hwnd && IsWindow(hwnd)) {
        PostMessageW(hwnd, WM_BEE_ACTIVATE, 0, 0);
        return;
    }
    if (InterlockedCompareExchange(&g_starting, 1, 0) != 0) return;
    if (g_thread) { CloseHandle(g_thread); g_thread = NULL; }
    g_thread = CreateThread(NULL, 0, SettingsThread, NULL, 0, &g_thread_id);
    if (!g_thread) InterlockedExchange(&g_starting, 0);
}

void BeeSettingsClose(void) {
    HWND hwnd = g_window;
    if (hwnd && IsWindow(hwnd)) PostMessageW(hwnd, WM_BEE_CLOSE, 0, 0);
    if (g_thread) {
        if (GetCurrentThreadId() != g_thread_id) WaitForSingleObject(g_thread, 3000);
        CloseHandle(g_thread);
        g_thread = NULL;
    }
    g_thread_id = 0;
    InterlockedExchange(&g_starting, 0);
}
