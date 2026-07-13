package main

/*
#include "settings_window/settings_window.h"
*/
import "C"

// showSettingsWindow 打开独立的 Win32 + GDI+ 设置窗口；重复打开时聚焦已有窗口。
func showSettingsWindow() { C.BeeSettingsOpen() }

// closeSettingsWindow 关闭设置窗口并等待 UI 线程退出，确保插件可以安全卸载。
func closeSettingsWindow() { C.BeeSettingsClose() }
