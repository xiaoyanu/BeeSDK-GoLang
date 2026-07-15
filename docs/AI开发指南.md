# AI 开发指南

本文件提供给 Claude Code、Codex、Cursor、Copilot、Hermes 等 AI 编程助手。修改本模板前必须遵守以下约束。

## 目标

开发 Bee 机器人插件，输出一个能被 32 位 Bee 框架通过 `LoadLibraryA` 加载的单 DLL：

```text
C stdcall ABI 桥接层 + Go 业务层 → PE32/i386 BeeGoPlugin.dll
```

## 绝对不能破坏的 ABI

1. 保留 `native/bridge.c`，Bee 不直接调用 Go ABI。
2. 保留 `other/BeePlugin.def` 中 11 个中文导出名；该文件是 GBK 编码，供 Zig 直接链接。
3. 所有 Bee C 入口必须是 `__stdcall`。
4. 回调参数全部是 NUL 结尾的 GBK `char*`。
5. 回调参数个数和顺序不得改变：

| Bee 回调 | 总参数 | 机器人之后的顺序 |
|---|---:|---|
| `Bee_初始化` | 1 | 机器人文本 |
| `Bee_插件被启用` | 1 | 机器人文本 |
| `Bee_插件被禁用` | 1 | 机器人文本 |
| `Bee_插件被卸载` | 1 | 机器人文本 |
| `Bee_插件设置` | 1 | 机器人文本 |
| `Bee_收到频道私信` | 6 | 频道ID、子频道ID、触发人ID、消息内容、消息ID |
| `Bee_收到频道消息` | 6 | 频道ID、子频道ID、触发人ID、消息内容、消息ID |
| `Bee_收到频道事件` | 7 | 频道ID、子频道ID、触发人ID、操作人ID、事件类型、消息内容 |
| `Bee_收到私聊消息` | 4 | 好友ID、消息内容、消息ID |
| `Bee_收到群聊消息` | 5 | 群ID、触发人ID、消息内容、消息ID |
| `Bee_收到通常事件` | 6 | 来源ID、触发人ID、操作人ID、事件类型、消息内容 |

## 编码和内存

- Bee → Go：框架传入的是 GBK `*C.char`，每个 `GoBee*` 回调在方法开头将参数统一转换一次。原始 C 参数使用 `cRobot/cGuildID/cMessage` 等名称，转换后的 UTF-8 Go 字符串使用 `robot/guildID/message` 等名称。
- 转换和业务逻辑保留在同一个回调方法中，不再额外拆分 `onBee*` 业务入口；后续业务代码直接使用转换后的字符串，不要反复调用 `goText`。
- Go → 框架 API：统一调用 `RobotContext.Call`。它会先把 GBK 无法表示的 Unicode 字符编码成 UTF-16 `\uXXXX`，例如 `哈哈😄` → `哈哈\uD83D\uDE04`，再将整条命令转为 GBK；业务代码不要重复编码。
- `Bee_初始化` 返回的 JSON 必须包含 `name/author/ver/text`。
- 初始化返回值必须使用 C 分配的内存并长期有效，禁止返回 Go `[]byte` 或 Go 字符串内部地址。
- 框架 API 返回的指针归框架所有，不要 `free`。

## 机器人上下文

每个回调的首参 `robot` 是 GBK JSON。先转换并解析：

```go
ctx, err := ParseRobotContext(goText(robot))
if err != nil {
    return MessageContinue
}
```

已知字段：

```text
api, msg, msg_id, channel_id, guild_id, form_id,
robot_id, plugin_id, event_id, raw
```

注意：框架源码实际拼写为 `form_id`，不要擅自改成 `from_id`。

### 上下文新鲜度与缓存边界（强制要求）

`robot` 不是固定的机器人对象，而是框架传给当前入口的机器人上下文 JSON。不同回调中的内容可能不同，必须先判断目标 API 是否依赖动态上下文，再决定使用当前上下文还是缓存上下文。

#### 1. 必须使用当前 `robot` 的场景

以下 API 与当前消息或当前事件绑定，必须在入口内使用本次框架传入的 `robot` 创建局部 `BeeAPI`：

- 被动回复及其他依赖 `msg_id` 的调用。
- 按钮响应及其他依赖 `event_id` 的调用。
- 明确依赖当前 `msg_id`、`event_id`、频道、群、好友或事件状态的调用。
- 无法确认是否依赖动态字段的 API。拿不准时一律使用当前 `robot`。

```go
func GoBeePrivateMessage(cRobot, cFriendID, cMessage, cMessageID *C.char) C.int {
    robot, friendID := goText(cRobot), goText(cFriendID)
    message, messageID := goText(cMessage), goText(cMessageID)

    bee, err := NewBeeAPI(robot)
    if err != nil {
        return MessageContinue
    }

    if message == "你好" {
        _, _ = bee.Friend(friendID).SendText("你好")
    }
    _ = messageID
    return MessageContinue
}
```

禁止使用旧回调的缓存上下文进行被动回复或事件响应，否则可能回复错消息、响应错事件、发送失败，或在并发场景下发生串消息。

#### 2. 允许使用缓存上下文的场景

如果某个 API 已明确确认不依赖 `msg_id`、`event_id` 等本次回调动态字段，则允许使用缓存上下文，例如：

- 输出框架日志。
- 获取机器人 ID、机器人 AppID 等机器人基本信息。
- 其他经协议确认只需要稳定字段或框架 API 地址的调用。

但缓存能力必须有清楚的用途边界：

- 不能把缓存的 `BeeAPI` 当作所有接口通用的全局 `bee`。
- 不能使用缓存上下文调用任何依赖当前消息或事件的 API。
- 缓存应表示“仅供无状态 API 使用的稳定上下文”，名称和注释必须体现限制。
- 如果框架重新传入了新的稳定字段或 API 地址，应及时刷新缓存。
- 对 API 是否依赖动态字段不确定时，禁止自行猜测为可缓存。

#### 3. 生命周期入口

启用、禁用、卸载和设置入口本身也会收到框架传入的 `robot`。入口内需要调用框架 API 时，优先基于该入口当前的 `robot` 创建局部 `bee`：

```go
func GoBeeEnable(cRobot *C.char) {
    robot := goText(cRobot)
    bee, err := NewBeeAPI(robot)
    if err != nil {
        return
    }

    _ = bee.Log("插件已启用")
}
```

#### 4. goroutine 与公共业务函数

- goroutine 如果处理某条消息或事件，必须接收该回调创建的上下文快照，不能读取会被其他回调覆盖的“最新上下文”全局变量。
- 公共业务函数应由入口把当前 `bee` 作为参数传入，不要在公共函数内部读取不明确的全局上下文。
- 只有明确属于无状态 API 的后台任务，才能使用受限的缓存上下文。

#### 5. AI 编写代码时的判断顺序

```text
这个 API 是否依赖当前 msg_id 或 event_id？
├─ 是                         → 必须使用本次回调的 robot
├─ 不确定                     → 必须使用本次回调的 robot
└─ 已确认完全不依赖动态字段   → 可以使用受限缓存，也可使用当前 robot
```

核心原则：

> 消息和事件相关 API 必须跟随框架本次传入的 `robot`；只有明确不依赖动态字段的无状态 API 才允许使用缓存，而且缓存不能越界使用。

## 框架 API 协议

基本格式：

```text
操作码%@#bee#@%plugin_id%@#bee#@%参数1...
```

必须调用现有 `RobotContext` 方法，不要在业务回调中手写命令字符串。

已知特例：

- API 31：不带 `plugin_id`。
- API 45：不带 `plugin_id`。
- API 48：不带 `plugin_id`。
- API 13：Go SDK 已修复原易语言 SDK 中“显式传假仍会拉黑”的错误。
- API 16：Go SDK 已修复原易语言 SDK 中名称后误用逗号造成的拼接错误。

参数不能包含 `%@#bee#@%`；协议没有定义转义方式。

## 消息模式

消息方法中的：

```go
active == false
```

表示被动回复，使用上下文 `msg_id`。

```go
active == true
```

表示主动发送：清空 `msg_id`，保留 `event_id`。主动消息受平台频率和次数限制。

好友接口中的 `recallInteraction` 表示互动召回参数，不是撤回已发送消息。

## 回调返回值

```go
return MessageContinue  // 0，继续给后续插件
return MessageIntercept // 1，拦截后续插件
```

`MessageIgnore` 和 `MessageContinue` 都是 0，是原 SDK 的语义别名。

## 开发位置

- 插件信息：`plugin_main.go` 的 `PluginInfo`。
- 业务逻辑：`plugin_main.go` 的 6 个消息/事件回调。
- 设置窗口：`settings.go` 只保留 Go 调用入口；全部 Win32/GDI+ 界面代码放在 `settings_window/`，不要放进 `other/`。
- 修改窗口前必须阅读并遵守同目录的 `设置窗口设计规范.md`；不得自行改回标准可见非客户区、模拟窗口阴影或把文字绘制穿插在活动的 GDI+ 绘制过程中。
- API、数据类型、操作码和辅助方法统一位于 `bee_sdk.go`。
- 插件持久数据、配置、数据库、缓存和日志必须写入 `运行目录\plugin_data\插件名称\`；“Bee 框架根目录”实现上就是取运行目录，不要写到 DLL 同目录、`plugin\` 或 `temp_plugin\`。
- 不要修改 `other/bee_bridge.c`，也不要把业务逻辑写入桥接层。

## 编码规范

1. 所有新增导出函数、方法、类型和常量必须有以标识符开头的中文注释。
2. 所有错误必须显式处理，消息回调发生错误时默认返回 `MessageContinue`。
3. 耗时任务不能阻塞 `Bee_初始化`；需要时启动 goroutine，并在禁用/卸载时停止。
4. 不要在启用回调中再次执行初始化。
5. 不要添加运行时读取同目录脚本或私有 DLL 的依赖。
6. 不确定的协议必须在注释中标记“待真实框架验证”，禁止猜测为已验证。

## 测试与验证边界

默认采用轻量验证，避免为显而易见的修改消耗过多时间、算力和上下文 Token。

AI 完成代码修改后通常只需要：

1. 检查语法结构、括号、导入、声明和明显的类型问题。
2. 核对新增函数的名称、参数、返回值和调用位置。
3. 确认 Bee ABI、GBK/UTF-8 边界、窗口生命周期等既有强约束未被破坏。
4. 检查构建命令是否包含新增源文件和所需系统库。

默认不要主动执行：

- 完整编译 DLL。
- 重复编译或反复检查同一个产物。
- `file`、`objdump`、`sha256sum` 等二进制验证。
- 为简单界面、注释或文档修改编写额外测试程序。
- 未经用户要求的真实 Bee 框架回归测试。

只有以下情况才主动编译或进行完整验证：

- 用户明确要求编译、生成 DLL 或验证。
- 修改了 Bee 导出 ABI、C/Go 桥接、链接参数或关键底层协议。
- 仅靠代码检查无法判断关键调用是否成立，而且错误可能直接导致宿主崩溃。
- 用户提供了明确的编译错误，当前任务就是解决该错误。

其余情况下保证代码结构、基本语法和调用关系合理即可。后续由用户实际编译；若出现编译或运行错误，再根据真实错误信息处理，不提前进行过度验证。

## 当前验证状态

真实 Bee 框架目前已验证：

```text
Bee_收到私聊消息
→ ParseRobotContext
→ API 52 SendFriendMessage
→ 收到 hi 后成功回复“你好”
```

除 API 52 基础纯文本好友消息外，其他操作码来自原 SDK 静态分析，尚未逐项进行真实框架验证。

## 完成任务前的默认检查

```text
1. Go 文件发生修改时执行 gofmt；仅修改 C 或文档时无需运行。
2. 静态检查修改范围内的语法、函数签名、调用关系和构建参数。
3. 不默认运行 build.bat，不默认生成或验证 DLL。
```

默认完成标准：

- 修改内容完整，文件职责和目录位置正确。
- 基本语法、函数声明及调用关系无明显错误。
- 未破坏 Bee ABI、编码边界、单 DLL 和生命周期约束。
- 不声称“编译成功”“运行成功”或“真实框架验证成功”，除非本次确实执行了对应验证。

当用户明确要求正式构建或发布验证时，再执行完整流程：生成 PE32/i386 DLL、核对 11 个 Bee 中文导出、检查导入表，并按需要进行真实 Bee 框架测试。
