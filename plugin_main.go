package main

/*
#include <stdlib.h>

*/
import "C"

import (
	"encoding/json"
	"sync"
)

// 插件基本信息统一在这里配置。
// PluginName 同时作为 build.bat 未指定输出名时使用的默认 DLL 文件名。
const (
	PluginName        = "Go测试插件"
	PluginAuthor      = "周星星"
	PluginVersion     = "1.0"
	PluginDescription = "使用 Go 编写的 Bee 测试插件\n欢迎使用"
)

// PluginInfo 表示 Bee 加载插件时读取的基本信息。
type PluginInfo struct {
	Name   string `json:"name"`
	Author string `json:"author"`
	Ver    string `json:"ver"`
	Text   string `json:"text"`
}

var (
	initOnce sync.Once
	infoGBK  *C.char
)

// GoBeeInit 初始化插件并返回长期有效的 GBK JSON。
//
//export GoBeeInit
func GoBeeInit(_ *C.char) *C.char {
	initOnce.Do(func() {
		info := PluginInfo{
			Name:   PluginName,
			Author: PluginAuthor,
			Ver:    PluginVersion,
			Text:   PluginDescription,
		}
		utf8JSON, _ := json.Marshal(info)
		gbkJSON := utf8ToGBK(utf8JSON)
		infoGBK = C.CString(string(gbkJSON))
	})
	return infoGBK
}

// GoBeeEnable 在插件被启用时执行。
// 可以在这里执行插件初始化内容；耗时操作必须另开 goroutine。
// 此回调在框架首次启动时也会调用一次。
//
//export GoBeeEnable
func GoBeeEnable(cRobot *C.char) {
	robot := goText(cRobot)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return
	}

	// 在这里启动插件任务、定时器或后台服务。
	// 例如：bee.Log("插件已启用")
	_ = bee
}

// GoBeeDisable 在插件被禁用时执行。
// 在这里停止后台任务、释放资源并关闭设置窗口。
//
//export GoBeeDisable
func GoBeeDisable(cRobot *C.char) {
	robot := goText(cRobot)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return
	}

	// 在这里销毁插件创建的 goroutine、定时器、连接等资源。
	// 例如：bee.Log("插件已禁用")
	_ = bee
	closeSettingsWindow()
}

// GoBeeUnload 在插件被卸载时执行最终清理。
// 与易语言模板一致，卸载时会先调用一次插件禁用逻辑。
//
//export GoBeeUnload
func GoBeeUnload(cRobot *C.char) {
	robot := goText(cRobot)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return
	}

	GoBeeDisable(cRobot)
	// 在这里执行仅卸载时需要的最终清理操作。
	// 例如：bee.Log("插件已卸载")
	_ = bee
}

// GoBeeSettings 在用户点击插件设置时打开设置窗口。
// 等同于易语言中的：载入 (窗口_主窗口, , 假)。
//
//export GoBeeSettings
func GoBeeSettings(cRobot *C.char) {
	robot := goText(cRobot)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return
	}

	// 可以在这里调用 Bee API，例如记录用户打开设置窗口的日志。
	_ = bee
	showSettingsWindow()
}

// GoBeeChannelDM 处理频道私信消息。
//
//export GoBeeChannelDM
func GoBeeChannelDM(cRobot, cGuildID, cChannelID, cUserID, cMessage, cMessageID *C.char) C.int {
	robot, guildID, channelID := goText(cRobot), goText(cGuildID), goText(cChannelID)
	userID, message, messageID := goText(cUserID), goText(cMessage), goText(cMessageID)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return MessageContinue
	}

	// 以下是示例代码，可以删除或修改。
	// if message == "测试" {
	// 	_, _ = bee.ChannelDM(guildID).SendText("你好哦")
	// }
	_, _, _, _, _, _ = bee, guildID, channelID, userID, message, messageID
	return MessageContinue
}

// GoBeeChannelMessage 处理频道消息。
// 公域机器人收到的消息通常没有艾特符号；私域机器人可能包含艾特符号，需要自行判断和清理。
//
//export GoBeeChannelMessage
func GoBeeChannelMessage(cRobot, cGuildID, cChannelID, cUserID, cMessage, cMessageID *C.char) C.int {
	robot, guildID, channelID := goText(cRobot), goText(cGuildID), goText(cChannelID)
	userID, message, messageID := goText(cUserID), goText(cMessage), goText(cMessageID)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return MessageContinue
	}

	// 频道和群消息的公共业务入口；同时开发两个场景时，建议把通用逻辑写在这里。
	handleChannelAndGroupMessage(bee, guildID, channelID, userID, message, messageID)

	// 以下都是示例代码，可以删除或修改。需要使用时去掉对应代码前的 //。
	// content := message
	// currentGuildID := guildID
	// currentChannelID := channelID
	// currentUserID := userID
	// currentMessageID := messageID
	//
	// if content == "hi" || content == "/hi" {
	// 	_, _ = bee.Channel(currentChannelID).SendImage(At(currentUserID)+"\nHi~Bee", "https://q.qq.com/bot/static/images/f3648b8001dfa331020c096a85057715.png")
	// 	_, _ = bee.Channel(currentChannelID).SendText("芜湖~欢迎使用！")
	// }
	//
	// if content == "重启框架" || content == "/重启框架" {
	// 	// Go 模板默认不提供直接重启宿主进程的方法，避免插件误杀框架。
	// 	// 如确实需要，应先使用 bee.IsGuildOwner(currentGuildID, currentUserID) 校验频道主身份，再实现受控重启。
	// }
	//
	// mentioned, cleanContent, _ := bee.ParseMention(content)
	// if mentioned {
	// 	_, _ = bee.Channel(currentChannelID).SendText("我被艾特了！")
	// 	if cleanContent == "/你好" {
	// 		_, _ = bee.Channel(currentChannelID).SendText("你好哦~")
	// 	}
	// }
	//
	// if content == "解除禁言" {
	// 	if owner, _ := bee.IsGuildOwner(currentGuildID, currentUserID); owner {
	// 		_ = bee.MuteMember(currentGuildID, MentionedUserID(content), 0)
	// 	}
	// }
	// if strings.HasPrefix(content, "禁言") {
	// 	if owner, _ := bee.IsGuildOwner(currentGuildID, currentUserID); owner {
	// 		_ = bee.MuteMember(currentGuildID, MentionedUserID(content), 120)
	// 	}
	// }
	//
	// if content == "全体禁言开" {
	// 	if owner, _ := bee.IsGuildOwner(currentGuildID, currentUserID); owner {
	// 		_ = bee.MuteAll(currentGuildID, 360)
	// 	}
	// }
	// if content == "全体禁言关" {
	// 	if owner, _ := bee.IsGuildOwner(currentGuildID, currentUserID); owner {
	// 		_ = bee.MuteAll(currentGuildID, 0)
	// 	}
	// }
	//
	// if content == "测试消息" {
	// 	_, _ = bee.Channel(currentChannelID).SendTextCard("大标题", "小标题", "测试\n你真帅", "")
	// 	_, _ = bee.Channel(currentChannelID).SendText("测试\n你真帅")
	// 	_, _ = bee.Channel(currentChannelID).SendCustom(`{"content":"自定义消息"}`)
	// 	_, _ = bee.Channel(currentChannelID).SendLargeCard("大图", "123", "666", "", "")
	// 	_, _ = bee.Channel(currentChannelID).Reply(currentMessageID, "233", "")
	// }
	//
	// if content == "引用我" {
	// 	_, _ = bee.Channel(currentChannelID).Reply(currentMessageID, "这是普通引用", "")
	// 	_, _ = bee.Channel(currentChannelID).Reply(currentMessageID, "这是图片引用", "https://fanyi.youdao.com/img/logo.50fdfa99.png")
	// }

	_ = bee
	return MessageContinue
}

// GoBeeChannelEvent 处理频道事件。
// eventType 对应 bee_sdk.go 中定义的 EventType 常量。
//
//export GoBeeChannelEvent
func GoBeeChannelEvent(cRobot, cGuildID, cChannelID, cUserID, cOperatorID, cEventType, cMessage *C.char) C.int {
	robot, guildID, channelID := goText(cRobot), goText(cGuildID), goText(cChannelID)
	userID, operatorID := goText(cUserID), goText(cOperatorID)
	eventTypeText, message := goText(cEventType), goText(cMessage)
	eventType := EventType(eventTypeText)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return MessageContinue
	}

	currentGuildID := guildID
	currentChannelID := channelID
	currentUserID := userID
	currentOperatorID := operatorID
	currentEventType := eventType
	currentMessage := message

	// 按事件类型编写对应业务。暂时不需要的分支可以保留为空，或直接删除。
	switch currentEventType {
	case EventGuildMemberRemove:
		// 当成员被移除时。
	case EventMessageDelete:
		// 有人撤回消息时。
	case EventChannelCreate:
		// 当子频道被创建时。
	case EventChannelUpdate:
		// 当子频道被更新时。
	case EventChannelDelete:
		// 当子频道被删除时。
	case EventGuildMemberAdd:
		// 当有人加入频道时。
	case EventGuildMemberUpdate:
		// 当成员资料变更时。
	case EventForumThreadCreate:
		// 当有人创建主题时。
	case EventForumThreadUpdate:
		// 当有人更新主题时。
	case EventForumThreadDelete:
		// 当有人删除主题时。
	case EventGuildDelete:
		// 当机器人退出频道时。
	case EventMessageReactionAdd:
		// 为消息添加表情表态。
	case EventMessageReactionRemove:
		// 为消息删除表情表态。
	case EventGuildCreate:
		// 当机器人加入新频道时。
	case EventGuildUpdate:
		// 当频道资料发生变更时。
	case EventAudioLiveEnter:
		// 当有人进入音视频或直播子频道时。
	case EventAudioLiveExit:
		// 当有人离开音视频或直播子频道时。
	case EventForumPostCreate:
		// 当有人创建帖子时。
	case EventForumReplyDelete:
		// 当有人删除评论时。
	case EventForumPostDelete:
		// 当有人删除帖子时。
	case EventForumReplyCreate:
		// 当有人回复帖子时。
	case EventInteractionCreate:
		// 按钮事件。使用 bee.ctx.RespondButton(bee.ctx.EventID, responseType) 响应按钮。
	default:
		// 未知或后续新增的频道事件。
	}

	// 删除未使用标记后，即可在对应分支直接使用这些当前事件参数。
	_, _, _, _, _, _ = bee, currentGuildID, currentChannelID, currentUserID, currentOperatorID, currentMessage
	return MessageContinue
}

// GoBeePrivateMessage 处理好友私聊消息。
//
//export GoBeePrivateMessage
func GoBeePrivateMessage(cRobot, cFriendID, cMessage, cMessageID *C.char) C.int {
	robot, friendID := goText(cRobot), goText(cFriendID)
	message, messageID := goText(cMessage), goText(cMessageID)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return MessageContinue
	}

	// 在这里编写好友私聊消息业务。
	_, _, _, _ = bee, friendID, message, messageID
	return MessageContinue
}

// GoBeeGroupMessage 处理群聊消息。
// 群聊的触发人 ID 与频道的触发人 ID 格式不同，群 ID 中会包含字母。
//
//export GoBeeGroupMessage
func GoBeeGroupMessage(cRobot, cGroupID, cUserID, cMessage, cMessageID *C.char) C.int {
	robot, groupID, userID := goText(cRobot), goText(cGroupID), goText(cUserID)
	message, messageID := goText(cMessage), goText(cMessageID)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return MessageContinue
	}

	// 群消息也交给频道和群消息的统一业务入口处理。
	// guildID 传空，sourceID 传群 ID。
	handleChannelAndGroupMessage(bee, "", groupID, userID, message, messageID)

	// 以下是示例代码，可以删除或修改。
	// if message == "你好" {
	// 	_, _ = bee.Group(groupID).SendText("你好哦")
	// }
	return MessageContinue
}

// handleChannelAndGroupMessage 统一处理频道消息和群聊消息。
// 这里会同时收到频道消息和群消息；同时开发两个场景时，建议把公共业务写在这里。
// guildID 不为空表示频道消息，sourceID 是子频道 ID；guildID 为空表示群消息，sourceID 是群 ID。
// 按原 Bee 模板的格式特征，群 ID 和群触发人 ID 中包含字母，子频道 ID 和频道触发人 ID 不包含字母。
// 公域机器人收到的频道消息通常不含艾特符号；私域机器人被艾特时可能包含，需要自行判断和清理。
func handleChannelAndGroupMessage(bee *BeeAPI, guildID, sourceID, userID, message, messageID string) {
	isChannelMessage := guildID != ""

	// 以下是示例代码，可以删除或修改。
	// if message == "2023" {
	// 	_, _ = bee.ctx.SendAdaptiveMessage(sourceID, "2023年！！！", "", false, false)
	// }
	//
	// if isChannelMessage {
	// 	isOwner, err := bee.ctx.IsGuildOwner(guildID, userID)
	// 	if err == nil {
	// 		if isOwner {
	// 			_ = bee.ctx.Log("是")
	// 		} else {
	// 			_ = bee.ctx.Log("否")
	// 		}
	// 	}
	// }

	_, _, _, _, _, _ = bee, isChannelMessage, sourceID, userID, message, messageID
}

// GoBeeCommonEvent 处理好友和群聊的通常事件。
// eventType 对应 bee_sdk.go 中定义的 EventType 常量。
//
//export GoBeeCommonEvent
func GoBeeCommonEvent(cRobot, cSourceID, cUserID, cOperatorID, cEventType, cMessage *C.char) C.int {
	robot, sourceID, userID := goText(cRobot), goText(cSourceID), goText(cUserID)
	operatorID := goText(cOperatorID)
	eventTypeText, message := goText(cEventType), goText(cMessage)
	eventType := EventType(eventTypeText)
	bee, err := NewBeeAPI(robot)
	if err != nil {
		return MessageContinue
	}

	currentSourceID := sourceID
	currentUserID := userID
	currentOperatorID := operatorID
	currentEventType := eventType
	currentMessage := message

	switch currentEventType {
	case EventInteractionCreate:
		// 按钮事件。使用 bee.ctx.RespondButton(bee.ctx.EventID, responseType) 响应按钮。
	case EventFriendAdd:
		// 添加机器人为好友。
	case EventFriendDelete:
		// 删除机器人好友。
	case EventGroupAddRobot:
		// 机器人被添加到群聊。
	case EventGroupDelRobot:
		// 机器人被移除群聊。
	case EventGroupMessageReceive:
		// 群聊接受机器人主动消息。
	case EventGroupMessageReject:
		// 群聊拒绝机器人主动消息。
	case EventC2CMessageReceive:
		// 私聊接受机器人主动消息。
	case EventC2CMessageReject:
		// 私聊拒绝机器人主动消息。
	case EventGroupMemberAdd:
		// 有人加入群聊。
	case EventGroupMemberRemove:
		// 成员被移除群聊。
	default:
		// 未知或后续新增的通常事件。
	}

	_, _, _, _, _ = bee, currentSourceID, currentUserID, currentOperatorID, currentMessage
	return MessageContinue
}

func goText(value *C.char) string {
	if value == nil {
		return ""
	}
	return gbkToUTF8([]byte(C.GoString(value)))
}

func main() {}
