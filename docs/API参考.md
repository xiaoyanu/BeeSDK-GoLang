# API 参考

所有框架 API 都通过 `*RobotContext` 调用：

```go
ctx, err := ParseRobotContext(goText(robot))
if err != nil {
    return MessageContinue
}
```

方法通常返回：

```go
(result, error)
```

其中 `result` 是框架返回的消息 ID、JSON、布尔值或错误原因。

> 当前只有 API 52 的好友纯文本发送经过真实 Bee 框架验证。其余 API 已按原易语言 SDK 完整封装，但仍需逐项实测。

## 通用参数

| 参数 | 含义 |
|---|---|
| `active` | `false` 被动回复；`true` 主动发送，清空 `msg_id`，保留 `event_id` |
| `deleteImage/deleteFile` | 发送完成后删除本地文件；网络 URL 不受影响 |
| `recallInteraction` | 好友互动召回参数，不等于撤回消息 |
| `image/file` | 按原接口可传网络地址或本地路径，具体支持范围需按接口验证 |
| `color` | 频道身份组颜色值；修改时 `-1` 表示不修改 |

## 日志、频道和成员

| 操作码 | Go 方法 | 作用 |
|---:|---|---|
| 1 | `Log` | 输出框架日志 |
| 4 | `ListGuilds` | 获取频道列表 |
| 5 | `GetGuild` | 获取频道详细信息 |
| 6 | `ListChannels` | 获取子频道列表 |
| 7 | `GetChannel` | 获取子频道详细信息 |
| 8 | `CreateChannel` | 创建子频道 |
| 9 | `GetChannelOnlineCount` | 获取直播/音视频子频道在线人数 |
| 10 | `GetGuildMember` | 获取频道成员详细信息 |
| 11 | `UpdateChannel` | 修改子频道；整数项传 `-1` 表示不修改 |
| 12 | `DeleteChannel` | 删除子频道 |
| 13 | `DeleteGuildMember` | 移除频道成员，可撤回历史消息及拉黑 |
| 14 | `ListGuildRoles` | 获取频道身份组列表 JSON |
| 15 | `CreateGuildRole` | 创建频道身份组 |
| 16 | `UpdateGuildRole` | 修改频道身份组；`hoist=nil` 表示不修改 |
| 17 | `IsGuildOwner` | 判断是否频道主 |
| 18 | `IsGuildAdmin` | 判断是否频道管理员 |
| 19 | `IsChannelAdmin` | 判断是否子频道管理员 |
| 20 | `HasGuildRole` | 判断是否属于指定身份组 |
| 21 | `DeleteGuildRole` | 删除身份组 |
| 22 | `AddGuildMemberRole` | 给成员添加身份组 |
| 23 | `RemoveGuildMemberRole` | 从成员移除身份组 |
| 29 | `MuteGuildMember` | 设置成员禁言，0 秒表示取消 |
| 30 | `MuteGuild` | 设置全员禁言，0 秒表示取消 |

## 普通消息和媒体

| 操作码 | Go 方法 | 作用 |
|---:|---|---|
| 2 | `SendChannelMessage` | 发送频道文字/图片消息 |
| 3 | `SendChannelDM` | 发送频道私信 |
| 25 | `SendChannelReply` | 发送频道引用消息 |
| 27 | `SendChannelCustom` | 发送频道自定义 JSON 消息 |
| 33 | `SendAdaptiveMessage` | 在当前群/频道场景自适应发送 |
| 34 | `SendGroupMessage` | 发送群文字/图片消息 |
| 35 | `SendGroupVideo` | 发送群视频 |
| 36 | `SendGroupAudio` | 发送群语音 |
| 52 | `SendFriendMessage` | 发送好友文字/图片消息；纯文本链路已验证 |
| 53 | `SendFriendVideo` | 发送好友视频 |
| 54 | `SendFriendAudio` | 发送好友语音 |
| 60 | `SendAdaptivePrivateMessage` | 在好友/频道私信场景自适应发送 |
| 69 | `SendGroupFile` | 发送群文件 |
| 70 | `SendFriendFile` | 发送好友文件 |
| 71 | `SendGroupReply` | 发送群引用消息 |
| 72 | `SendFriendReply` | 发送好友引用消息 |

## 卡片、Markdown 和按钮

| 操作码 | Go 方法 | 作用 |
|---:|---|---|
| 26 | `SendChannelTextCard` | 频道文字卡片 |
| 28 | `SendChannelLargeCard` | 频道大图卡片 |
| 38 | `SendGroupMarkdown` | 群 Markdown |
| 39 | `SendGroupTextCard` | 群文字卡片 |
| 41 | `SendGroupLargeCard` | 群大图卡片 |
| 42 | `SendAdaptiveLargeCard` | 群/频道自适应大图卡片 |
| 43 | `SendGroupThumbnailCard` | 群缩略图卡片 |
| 44 | `SendChannelThumbnailCard` | 频道缩略图卡片 |
| 47 | `SendChannelMarkdown` | 频道 Markdown |
| 55 | `SendFriendMarkdown` | 好友 Markdown |
| 56 | `SendFriendTextCard` | 好友文字卡片 |
| 57 | `SendFriendLargeCard` | 好友大图卡片 |
| 58 | `SendFriendThumbnailCard` | 好友缩略图卡片 |
| 65 | `SendGroupButton` | 群按钮模板消息 |
| 66 | `SendFriendButton` | 好友按钮模板消息 |

`MarkdownMessage.Params` 最多使用前 10 组，多余参数不会进入协议。建议调用前自行限制长度。

## 查询、撤回、表态和凭据

| 操作码 | Go 方法 | 作用 |
|---:|---|---|
| 24 | `RecallChannelMessage` | 撤回频道消息 |
| 31 | `GetRobotID` | 获取机器人长 ID |
| 32 | `GetRobotInfo` | 获取机器人信息 |
| 37 | `GetFrameworkInfo` | 获取框架信息 |
| 40 | `GetQQNickname` | 获取 QQ 昵称 |
| 45 | `UploadImage` | 上传图片到图床 |
| 46 | `RespondButton` | 响应按钮事件 |
| 48 | `GetRobotAppID` | 获取机器人 AppID |
| 49 | `GetAvatar` | 获取用户头像 |
| 50 | `GetQQAvatar` | 获取 QQ 头像 |
| 51 | `RecallGroupMessage` | 撤回群消息 |
| 59 | `RecallFriendMessage` | 撤回好友消息，超过两分钟可能失败 |
| 61 | `AddChannelReaction` | 添加频道表情表态 |
| 62 | `DeleteChannelReaction` | 删除频道表情表态 |
| 63 | `ListChannelReactionUsers` | 获取表态用户列表 JSON |
| 64 | `GetRobotStats` | 获取机器人统计 JSON；`limit=nil` 表示全部 |
| 67 | `GetRobotToken` | 获取短期 Token，禁止记录到日志 |
| 68 | `GetRobotSecret` | 获取 AppSecret，禁止发送或记录到日志 |

## 本地辅助方法

| 方法 | 作用 |
|---|---|
| `At` | 生成 `<@!用户ID>` |
| `AtEveryone` | 生成 `@everyone` |
| `MentionedUserID` | 从艾特代码提取用户 ID |
| `ImageDownloadURL` | 从图片消息代码中提取 URL |
| `InlineCommand` | 生成 QQ Markdown 内嵌指令 |
| `ResolveRedirect` | 获取网址重定向地址 |
| `BuildKeyboard` | 生成 Markdown 自定义按钮 JSON |

## 按钮限制

`BuildKeyboard`：

- 最多 5 行。
- 每行最多 10 个按钮。
- 必须传正确的机器人 AppID。
- 空显示文字默认为“按钮”。
- 空点击后文字默认为“按钮”。
- 空 `Data` 默认使用“你好”，调用者应主动填写业务内容。
