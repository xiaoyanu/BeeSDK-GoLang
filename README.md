<div align="center">

# Bee Go 插件开发模板

使用 Go 编写 Bee 机器人插件，构建可由 32 位 Bee 框架直接加载的单 DLL。

[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Zig](https://img.shields.io/badge/Zig-0.13%2B-F7A41D?logo=zig&logoColor=white)](https://ziglang.org/)
[![Platform](https://img.shields.io/badge/Platform-Windows-0078D4?logo=windows&logoColor=white)](#环境要求)
[![Architecture](https://img.shields.io/badge/Target-PE32%20%2F%20i386-555555)](#技术架构)
[![Build](https://img.shields.io/badge/Build-build.bat-4EAA25)](#编译)

**C `stdcall` ABI 桥接层 + Go 业务层 → PE32/i386 单 DLL**

[快速开始](#快速开始) · [编写回调](#编写回调) · [调用框架-api](#调用框架-api) · [编译](#编译) · [文档](#相关文档)

</div>

---

## 项目简介

这是一个面向 Bee 机器人框架的 Go 插件开发模板。

模板已经处理 Bee 与 Go 之间最容易出错的部分，包括 32 位 `stdcall` 调用约定、中文导出名称、GBK/UTF-8 编码转换、框架 API 函数指针调用和单 DLL 构建。开发插件时主要修改 `plugin_main.go`，不需要重新实现底层协议。

最终发布物只有一个 DLL：

```text
build/你的插件.dll
```

Go 运行时会链接进 DLL。使用插件的人不需要安装 Go、Zig，也不需要复制源码或其他私有 DLL。

## 主要特性

- 输出 Bee 可加载的 **32 位 PE32/i386 DLL**。
- 保留 Bee 所需的 **11 个中文导出入口**。
- 所有 Bee C 入口使用 `__stdcall`。
- 自动处理 Bee GBK 文本与 Go UTF-8 字符串之间的转换。
- 自动将 GBK 无法表示的 Unicode 字符编码为 UTF-16 `\uXXXX`。
- 封装 Bee 操作码 1～70、机器人上下文、事件类型和常用数据结构。
- 提供简洁的链式消息 API，例如：

```go
bee.Friend(friendID).SendText("你好")
bee.Group(groupID).SendText("群消息")
bee.Channel(channelID).SendImage("图片消息", imageURL)
```

- 提供好友、群聊、频道、频道私信、频道事件和通常事件回调骨架。
- 提供 Windows 原生 `build.bat`，不依赖 PowerShell、Python 或 Node.js。
- 编译中间文件统一放入 `temp/`，完成后自动清理。
- `build.bat` 固定使用 GBK/CP936 编码和 CRLF 换行，匹配 `chcp 936`，避免 Windows `cmd.exe` 中文乱码或误解析。
- 最终 DLL 统一输出到 `build/`，自动剥离符号并清理 PDB。
- 支持 Go 1.25 在 Windows构建 `c-archive` 时所需的 `zig ar` 包装。
- 内置 `C + Win32 + GDI+` 简约现代化设置窗口；保留 `WS_CAPTION` 等系统框架样式供 DWM 识别，通过 `WM_NCCALCSIZE` 将可见非客户区压缩为 0，并用 `DwmExtendFrameIntoClientArea(1px)` 保留系统阴影；客户区自绘标题栏、关闭按钮和现代 UI，顶部可拖动，窗口不可缩放。

## 验证状态

当前已经完成以下真实链路验证：

```text
Bee_收到私聊消息
→ 解析当前 robot 上下文
→ 调用 API 52 SendFriendMessage
→ 收到好友消息 hi
→ 成功回复“你好”
```

这条测试验证了：

- Bee 32 位回调进入 C `stdcall` 桥接层。
- GBK 参数能够正确转换为 Go UTF-8 字符串。
- `robot.api` 中的框架函数地址可以按 32 位 `stdcall` 调用。
- API 52 的基础纯文本好友消息字段顺序正确。
- Go 业务层可以通过单 DLL 在 Bee 中运行。

> [!IMPORTANT]
> 当前只有 API 52 的基础好友纯文本发送经过真实 Bee 框架验证。其他 API 根据原 Bee SDK 完成封装，但仍需结合实际机器人权限和平台行为逐项测试。

## 技术架构

```text
Bee 32 位框架
    │
    │ LoadLibraryA / GetProcAddress
    ▼
other/bee_bridge.c
    │  11 个中文 __stdcall 导出入口
    │  只负责 ABI 桥接
    ▼
plugin_main.go
    │  生命周期、消息和事件回调
    │  GBK → UTF-8
    ▼
bee_sdk.go
    │  RobotContext / BeeAPI / 操作码 1～70
    │  命令编码与框架函数指针调用
    ▼
Bee 框架 API
```

### ABI 约定

| 项目 | 约定 |
|---|---|
| 目标平台 | Windows |
| 目标架构 | PE32 / Intel i386 |
| Bee 调用约定 | `stdcall` |
| Bee 输入文本 | NUL 结尾 GBK `char*` |
| Go 内部文本 | UTF-8 `string` |
| 初始化返回值 | 长期有效的 GBK JSON `char*` |
| 消息继续投递 | `MessageContinue`，值为 `0` |
| 消息拦截 | `MessageIntercept`，值为 `1` |

## 项目结构

```text
Bee开发模板-GOLANG/
├── plugin_main.go       # 插件信息、生命周期、消息和事件回调
├── bee_sdk.go           # 完整 Bee Go SDK、操作码、类型和辅助方法
├── settings.go          # Go 与原生设置窗口的调用入口
├── build.bat            # Windows 一键编译
├── go.mod               # Go 模块配置
├── README.md            # 项目说明
├── docs/                # 开发文档
│   ├── API参考.md       # API 分类、操作码和参数参考
│   ├── AI开发指南.md    # 给 AI 编程助手使用的强约束文档
│   └── 设置窗口设计规范.md # Win32/GDI+/DWM 窗口设计和验收规范
├── settings_window/     # 设置窗口相关代码（不放入 other）
│   ├── settings_window.c # C + Win32 + GDI+ 窗口实现
│   └── settings_window.h # 窗口公开接口
└── other/
    ├── bee_bridge.c     # Bee stdcall 到 Go 的 C ABI 桥接
    └── BeePlugin.def    # 11 个中文导出名称，GBK 编码
```

编译时还会使用两个目录：

```text
temp/                    # 临时目录，成功或失败后自动删除
build/                   # 最终 DLL 输出目录，编译后保留
```

`settings_window/` 专门存放界面代码。`other/` 仍只存放 Bee ABI 桥接和导出定义等低频固定文件。

插件运行时产生的持久数据统一写入运行目录下的插件数据目录。这里的“Bee 框架根目录”实现上就是取运行目录：

```text
运行目录\plugin_data\插件名称
```

例如配置、数据库、缓存、日志都放在该目录下。不要把持久数据写到 DLL 同目录、`plugin\` 或 `temp_plugin\`，因为 Bee 加载器会把 DLL 复制到 `temp_plugin` 后再加载。

根目录只保留 `README.md`。其余说明统一放入 `docs/`；修改设置窗口前必须阅读 [设置窗口设计规范](./docs/设置窗口设计规范.md)。

## 环境要求

在 Windows 上安装并配置：

| 工具 | 最低版本 | 用途 |
|---|---:|---|
| [Go](https://go.dev/dl/) | 1.24 | 编译 Go 业务层和 C Archive |
| [Zig](https://ziglang.org/download/) | 0.13 | 32 位 Windows C 编译和 DLL 链接 |

安装后确认 `go.exe` 和 `zig.exe` 已加入 `PATH`：

```bat
go version
zig version
```

已知可用环境示例：

```text
go version go1.25.6 windows/amd64
zig version 0.17.0-dev.1387+01b60634c
```

> [!NOTE]
> Go 和 Zig 本身可以是 64 位程序。`build.bat` 会通过 `GOARCH=386` 和 `-target x86-windows-gnu` 生成 Bee 所需的 32 位 DLL。

## 快速开始

### 1. 修改插件信息

打开 `plugin_main.go`，修改文件顶部的常量：

```go
const (
    PluginName        = "我的插件"
    PluginAuthor      = "作者名称"
    PluginVersion     = "1.0"
    PluginDescription = "插件说明\n第二行说明"
)
```

字段用途：

| 常量 | 用途 |
|---|---|
| `PluginName` | Bee 插件初始化名称；也是插件重复判断的重要标识 |
| `PluginAuthor` | 插件作者 |
| `PluginVersion` | 插件版本 |
| `PluginDescription` | 插件说明；Go 字符串中的 `\n` 表示换行 |

### 2. 编写回调

所有生命周期、消息和事件回调都在 `plugin_main.go`。

框架传入的是 GBK `*C.char`。模板会在回调开头把参数转换为 UTF-8 Go 字符串，后续业务直接使用转换后的变量。

好友私聊示例：

```go
//export GoBeePrivateMessage
func GoBeePrivateMessage(cRobot, cFriendID, cMessage, cMessageID *C.char) C.int {
    robot, friendID := goText(cRobot), goText(cFriendID)
    message, messageID := goText(cMessage), goText(cMessageID)

    bee, err := NewBeeAPI(robot)
    if err != nil {
        return MessageContinue
    }

    if message == "你好" {
        _, _ = bee.Friend(friendID).SendText("你好哦")
    }

    _ = messageID
    return MessageContinue
}
```

常用回调：

| Go 入口 | Bee 场景 |
|---|---|
| `GoBeeEnable` | 插件被启用 |
| `GoBeeDisable` | 插件被禁用 |
| `GoBeeUnload` | 插件被卸载 |
| `GoBeeSettings` | 用户点击插件设置 |
| `GoBeeChannelDM` | 收到频道私信 |
| `GoBeeChannelMessage` | 收到频道消息 |
| `GoBeeChannelEvent` | 收到频道事件 |
| `GoBeePrivateMessage` | 收到好友私聊消息 |
| `GoBeeGroupMessage` | 收到群聊消息 |
| `GoBeeCommonEvent` | 收到好友或群聊通常事件 |

> [!CAUTION]
> 不要修改导出回调的参数个数、顺序和类型。Bee 按固定 ABI 调用这些入口，签名不匹配可能导致插件加载失败或宿主进程崩溃。

## 调用框架 API

### 创建当前回调的 BeeAPI

每个回调都应使用本次框架传入的 `robot` 创建局部 `BeeAPI`：

```go
bee, err := NewBeeAPI(robot)
if err != nil {
    return MessageContinue
}
```

`robot` 中可能包含当前回调的：

```text
api, plugin_id, msg_id, event_id, robot_id, guild_id, channel_id ...
```

### 快捷消息 API

```go
// 被动回复当前好友消息
_, err := bee.Friend(friendID).SendText("你好")

// 主动发送好友消息
_, err = bee.Friend(friendID).SendActiveText("主动消息")

// 回复群消息
_, err = bee.Group(groupID).SendText("群回复")

// 发送频道消息
_, err = bee.Channel(channelID).SendText("频道回复")

// 发送频道图片
_, err = bee.Channel(channelID).SendImage("图片说明", imageURL)

// 发送频道私信
_, err = bee.ChannelDM(guildID).SendText("频道私信")
```

### 日志和频道管理

```go
_ = bee.Log("插件开始处理消息")

owner, err := bee.IsGuildOwner(guildID, userID)
if err == nil && owner {
    _ = bee.MuteMember(guildID, userID, 120)
}
```

完整 API 分类、操作码和参数说明见 [API参考](./docs/API参考.md)。

## robot 上下文与缓存规则

`robot` 不是固定的机器人对象，而是框架传给当前入口的上下文 JSON。部分字段会随消息或事件变化。

### 必须使用当前 robot

以下调用必须使用当前回调创建的 `bee`：

- 依赖 `msg_id` 的被动回复。
- 依赖 `event_id` 的按钮响应。
- 与当前消息、群、频道、好友或事件绑定的 API。
- 无法确认是否依赖动态字段的 API。

### 可以使用受限缓存

明确不依赖 `msg_id`、`event_id` 的接口可以使用受限缓存，例如：

- 输出框架日志。
- 获取机器人 ID、AppID 等基本信息。
- 经协议确认只依赖稳定字段的查询。

缓存的上下文不能作为通用全局 `bee`，也不能用于被动回复或事件响应。具体规则见 [AI开发指南](./docs/AI开发指南.md#机器人上下文)。

## 编码处理

### Bee 到 Go

Bee 传入 GBK `char*`，模板通过 `goText` 转成 UTF-8：

```go
message := goText(cMessage)
```

同一个参数在一次回调中只转换一次。后续业务直接使用 `message`，不要重复调用 `goText`。

### Go 到 Bee

所有框架 API 最终通过 `RobotContext.Call` 发送命令。底层会：

1. 将 GBK 无法表示的 Unicode 字符转换为 UTF-16 `\uXXXX`。
2. 将完整命令编码为 GBK。
3. 调用当前 `robot.api` 指向的 32 位 `stdcall` 框架函数。

例如：

```text
哈哈😄
```

会在发送边界转换为：

```text
哈哈\uD83D\uDE04
```

业务代码不需要手动处理 emoji 或其他 Unicode 字符。

## 编译

双击运行：

```text
build.bat
```

也可以在命令行执行：

```bat
build.bat
```

脚本会询问 DLL 文件名：

```text
请输入编译后的 DLL 文件名（直接回车使用 BeeGoPlugin）：
```

可以输入：

```text
我的插件
```

或者：

```text
我的插件.dll
```

最终都会生成：

```text
build/我的插件.dll
```

### 构建流程

```text
1. 创建 temp/ 和 build/
2. 使用 Go 构建 Windows 386 C Archive
3. 使用 Zig 链接 Bee C 桥接层、Go Archive 和导出定义
4. 剥离调试符号并删除可能生成的 PDB
5. 删除 temp/
6. 将最终 DLL 保留在 build/
```

Go 构建命令使用 `-buildvcs=false`，不会读取或写入 Git 提交信息。这样即使模板位于损坏的仓库、受 Git `safe.directory` 限制的目录、从压缩包解压的目录，或机器上没有可用 Git，也不会因 VCS 状态获取失败而中断编译。

编译失败时窗口不会立即关闭，会保留错误信息并等待按键。

## 发布和安装

### 发布

正式发布只需要 `build/` 中生成的 DLL：

```text
我的插件.dll
```

不要附带：

- `.pdb`
- `.lib`
- `.a`
- `.h`
- `temp/`
- Go 源码
- 构建脚本
- 项目私有 DLL

### 安装

将 DLL 放入 Bee 的插件目录，然后由 Bee 加载插件。具体目录名称以正在使用的 Bee 版本为准。

Bee 的加载流程要求 DLL 提供 11 个固定中文导出，因此不要修改 `other/bee_bridge.c` 和 `other/BeePlugin.def` 中的导出约定。

## 回调返回值

消息和事件回调必须返回以下值之一：

```go
return MessageContinue  // 0：继续向后续插件投递
return MessageIntercept // 1：拦截，后续插件不再收到
```

`MessageIgnore` 与 `MessageContinue` 都是 `0`，它们是原 SDK 中不同语义的别名。

## 常见问题

### 为什么必须构建 32 位 DLL？

Bee 宿主是 32 位程序。32 位进程不能加载 64 位 DLL，因此目标必须是 PE32/i386。

### 为什么需要 C 桥接层？

Bee 按 32 位 `stdcall` ABI 调用插件。Go 导出函数不能直接作为 Bee 的稳定 ABI 入口，所以模板使用 `other/bee_bridge.c` 接收 Bee 调用，再转发到 Go。

### 为什么需要 Zig？

Zig 提供 Windows 386 C 编译和链接能力。模板用它编译 C 桥接层、链接 Go C Archive，并生成最终 DLL。

### 为什么编译时需要 ar？

Go 1.25 在 Windows 构建 `c-archive` 时会调用名为 `ar` 的归档工具。`build.bat` 会在 `temp/` 自动创建一个转发到 `zig ar` 的包装器，无需额外安装 MinGW。

### 为什么 DLL 旁边曾经出现 PDB？

PDB 是 Windows 调试符号文件，不是 Bee 运行插件所必需的文件。当前脚本会在链接时剥离符号，并删除 Zig 可能生成的同名 PDB。

### 为什么消息回调不能复用全局 bee？

被动回复和按钮响应依赖当前 `msg_id` 或 `event_id`。复用旧上下文可能回复错误消息、响应错误事件，或者在并发消息中发生串话。

### 为什么不能直接返回 Go 字符串指针？

Go 内存由垃圾回收器管理，不能把 Go 字符串内部地址作为长期有效的 C 指针返回。初始化 JSON 使用 `C.CString` 分配 C 内存，保证返回后仍然有效。

### 可以发送 emoji 吗？

可以。底层会把 GBK 无法表示的字符转换为 UTF-16 `\uXXXX`，非 BMP 字符会使用代理对。

## 重要限制

1. Bee 是 32 位程序，最终 DLL 必须是 PE32/i386。
2. 11 个 Bee 回调的参数数量、顺序和调用约定不能修改。
3. Bee 输入文本使用 GBK，Go 内部业务使用 UTF-8。
4. 初始化返回值必须位于长期有效的 C 内存中。
5. 框架 API 返回的内存归框架所有，不要自行释放。
6. API 参数不能包含协议分隔符 `%@#bee#@%`，当前协议没有定义转义方式。
7. 操作码 31、45、48 按原 SDK 不携带 `plugin_id`。
8. Token、AppSecret 等敏感信息不能写入日志或发送给无关用户。
9. 未经过真实 Bee 测试的接口不能声明为已验证兼容。
10. 正式发布物只保留 DLL。

## 相关文档

| 文档 | 适合阅读的人 | 内容 |
|---|---|---|
| [README.md](./README.md) | 所有开发者 | 项目介绍、快速开始、编译和发布 |
| [API 参考](./docs/API参考.md) | 插件开发者 | 操作码、Go 方法、参数和辅助函数 |
| [AI 开发指南](./docs/AI开发指南.md) | AI 编程助手 | ABI、上下文、编码、内存、轻量验证和修改约束 |
| [设置窗口设计规范](./docs/设置窗口设计规范.md) | UI 开发者与 AI | Win32/GDI+/DWM 架构、视觉令牌、拖动、阴影、绘制顺序和验收要求 |

## 开发检查清单

提交或发布前建议逐项检查：

- [ ] 插件名称、作者、版本和说明已经修改。
- [ ] 回调使用本次 `robot` 创建 `BeeAPI`。
- [ ] 没有缓存用于被动回复的旧 `msg_id` 或 `event_id`。
- [ ] 没有修改 11 个 Bee 回调的 ABI。
- [ ] Go 文件已经执行 `gofmt`。
- [ ] `build.bat` 编译成功。
- [ ] 最终文件位于 `build/插件名称.dll`。
- [ ] `build/` 中没有 PDB、LIB、静态归档或临时文件。
- [ ] DLL 是 PE32/i386，而不是 PE32+ x64。
- [ ] 11 个 Bee 中文导出名称完整存在。
- [ ] DLL 没有依赖项目私有 DLL。
- [ ] 已在真实 Bee 框架中完成加载和关键功能测试。

## 安全说明

- 不要在日志中输出机器人 Token、AppSecret 或完整敏感上下文。
- 不要把用户输入直接拼接进系统命令、文件路径或未经验证的 JSON。
- 对禁言、移除成员、撤回消息等管理操作进行权限校验。
- 后台 goroutine 应在插件禁用或卸载时停止，避免资源泄漏。
- 不要在 `Bee_初始化` 中执行耗时网络请求。

## 参与贡献

提交修改前请保持以下约定：

- `plugin_main.go` 保持为插件开发主入口。
- 完整 SDK 保持集中在 `bee_sdk.go`，不要随意拆成大量小文件。
- `other/bee_bridge.c` 只处理 ABI，不加入业务逻辑。
- 新增导出函数、类型、常量和方法时补充中文注释。
- 不确定的协议行为应标记为“待真实框架验证”。
- 修复或新增框架 API 后，同步更新 `docs/API参考.md` 和 `docs/AI开发指南.md`。
- 提交前清理 DLL、PDB、LIB、EXE、静态归档和临时构建目录。

---

<div align="center">

如果你准备使用 AI 辅助开发，请先让它完整阅读 [AI开发指南](./docs/AI开发指南.md)。

</div>
