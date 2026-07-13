@echo off
setlocal EnableExtensions DisableDelayedExpansion
chcp 936 >nul
pushd "%~dp0" >nul 2>nul
if errorlevel 1 goto :path_error

set "EXIT_CODE=1"
set "TEMP_DIR=%CD%\temp"
set "BUILD_DIR=%CD%\build"
set "CC_WRAPPER=%TEMP_DIR%\i686-w64-mingw32-gcc.bat"
set "AR_WRAPPER=%TEMP_DIR%\ar.bat"

echo ========================================
echo       Bee Go 插件一键编译
echo ========================================
echo.

where go >nul 2>nul
if errorlevel 1 goto :no_go
where zig >nul 2>nul
if errorlevel 1 goto :no_zig

if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"
mkdir "%TEMP_DIR%"
if errorlevel 1 goto :failed
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"
if errorlevel 1 goto :failed

set "OUTPUT_NAME="
set /p "OUTPUT_NAME=请输入编译后的 DLL 文件名（直接回车使用 BeeGoPlugin）："
if not defined OUTPUT_NAME set "OUTPUT_NAME=BeeGoPlugin"

if /i "%OUTPUT_NAME:~-4%"==".dll" set "OUTPUT_NAME=%OUTPUT_NAME:~0,-4%"
if not defined OUTPUT_NAME goto :empty_name

echo.
echo 开始编译...

>"%CC_WRAPPER%" echo @echo off
>>"%CC_WRAPPER%" echo zig cc -target x86-windows-gnu %%*
>"%AR_WRAPPER%" echo @echo off
>>"%AR_WRAPPER%" echo zig ar %%*

set "PATH=%TEMP_DIR%;%PATH%"
set "GOOS=windows"
set "GOARCH=386"
set "CGO_ENABLED=1"
set "CC=%CC_WRAPPER%"

echo.
echo [1/3] 编译 Go 静态库...
go build -buildvcs=false -buildmode=c-archive -trimpath -ldflags="-s -w" -o "%TEMP_DIR%\go_plugin.a" .
if errorlevel 1 goto :failed

echo [2/3] 链接 32 位 DLL...
zig cc -target x86-windows-gnu -O2 -s -shared other\bee_bridge.c "%TEMP_DIR%\go_plugin.a" other\BeePlugin.def -I"%TEMP_DIR%" -luser32 -lkernel32 -lws2_32 -lntdll -o "%BUILD_DIR%\%OUTPUT_NAME%.dll"
if errorlevel 1 goto :failed
if exist "%BUILD_DIR%\%OUTPUT_NAME%.pdb" del /q "%BUILD_DIR%\%OUTPUT_NAME%.pdb"

echo [3/3] 清理临时文件...
if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"

set "EXIT_CODE=0"
echo.
echo [成功] 已生成：%BUILD_DIR%\%OUTPUT_NAME%.dll
goto :finish

:no_go
echo [错误] 未找到 Go。
echo 下载地址：https://go.dev/dl/
echo 安装后请重新打开命令行窗口，再运行本脚本。
goto :finish

:no_zig
echo [错误] 未找到 Zig。
echo 下载地址：https://ziglang.org/download/
echo 安装后请将 zig.exe 所在目录加入系统 PATH。
goto :finish

:empty_name
echo [错误] DLL 文件名不能为空。
goto :finish

:path_error
echo [错误] 无法进入模板所在目录。
goto :finish_no_popd

:failed
echo.
echo [失败] 编译未完成，请查看上方错误信息。
if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"

:finish
echo.
echo 按任意键关闭窗口...
pause >nul
popd >nul 2>nul
exit /b %EXIT_CODE%

:finish_no_popd
echo.
echo 按任意键关闭窗口...
pause >nul
exit /b 1
