package main

/*
#include <windows.h>

static LRESULT CALLBACK BeeSettingsProc(HWND hwnd, UINT msg, WPARAM wp, LPARAM lp) {
    switch (msg) {
        case WM_CREATE:
            CreateWindowExW(0, L"STATIC", L"使用Go编写", WS_CHILD | WS_VISIBLE | SS_CENTER,
                            30, 35, 300, 28, hwnd, NULL, GetModuleHandleW(NULL), NULL);
            CreateWindowExW(0, L"BUTTON", L"关闭", WS_CHILD | WS_VISIBLE | BS_PUSHBUTTON,
                            135, 90, 90, 30, hwnd, (HMENU)1, GetModuleHandleW(NULL), NULL);
            return 0;
        case WM_COMMAND:
            if (LOWORD(wp) == 1) DestroyWindow(hwnd);
            return 0;
    }
    return DefWindowProcW(hwnd, msg, wp, lp);
}

static void BeeShowSettings(void) {
    HINSTANCE instance = GetModuleHandleW(NULL);
    WNDCLASSEXW wc = {0};
    wc.cbSize = sizeof(wc);
    wc.lpfnWndProc = BeeSettingsProc;
    wc.hInstance = instance;
    wc.hCursor = LoadCursorW(NULL, (LPCWSTR)MAKEINTRESOURCEW(32512));
    wc.hbrBackground = (HBRUSH)(COLOR_WINDOW + 1);
    wc.lpszClassName = L"BeeGoSettingsWindow";
    RegisterClassExW(&wc);

    const int width = 375, height = 180;
    RECT work;
    SystemParametersInfoW(SPI_GETWORKAREA, 0, &work, 0);
    int x = work.left + ((work.right - work.left) - width) / 2;
    int y = work.top + ((work.bottom - work.top) - height) / 2;

    HWND existing = FindWindowW(wc.lpszClassName, L"Bee插件设置");
    if (existing) {
        ShowWindow(existing, SW_RESTORE);
        SetForegroundWindow(existing);
        return;
    }

    HWND hwnd = CreateWindowExW(WS_EX_DLGMODALFRAME, wc.lpszClassName, L"Bee插件设置",
        WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU,
        x, y, width, height, NULL, NULL, instance, NULL);
    if (hwnd) {
        ShowWindow(hwnd, SW_SHOW);
        UpdateWindow(hwnd);
    }
}

static void BeeCloseSettings(void) {
    HWND existing = FindWindowW(L"BeeGoSettingsWindow", L"Bee插件设置");
    if (existing) {
        DestroyWindow(existing);
    }
}
*/
import "C"

// showSettingsWindow 打开设置窗口；重复打开时聚焦已有窗口。
func showSettingsWindow() { C.BeeShowSettings() }

// closeSettingsWindow 关闭已经创建的设置窗口；窗口不存在时不执行操作。
func closeSettingsWindow() { C.BeeCloseSettings() }
