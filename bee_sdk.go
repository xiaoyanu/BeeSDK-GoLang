package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef char* (__stdcall *BeeAPIFn)(char*);
static char* bee_call_api(uintptr_t address, char* command) {
    if (address == 0) return NULL;
    return ((BeeAPIFn)address)(command);
}
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

// ==================== opcodes.go ====================
const (
	OpLog                        = 1
	OpSendChannelMessage         = 2
	OpSendChannelDM              = 3
	OpListGuilds                 = 4
	OpGetGuild                   = 5
	OpListChannels               = 6
	OpGetChannel                 = 7
	OpCreateChannel              = 8
	OpGetChannelOnlineCount      = 9
	OpGetGuildMember             = 10
	OpUpdateChannel              = 11
	OpDeleteChannel              = 12
	OpDeleteGuildMember          = 13
	OpListGuildRoles             = 14
	OpCreateGuildRole            = 15
	OpUpdateGuildRole            = 16
	OpIsGuildOwner               = 17
	OpIsGuildAdmin               = 18
	OpIsChannelAdmin             = 19
	OpHasGuildRole               = 20
	OpDeleteGuildRole            = 21
	OpAddGuildMemberRole         = 22
	OpRemoveGuildMemberRole      = 23
	OpRecallChannelMessage       = 24
	OpSendChannelReply           = 25
	OpSendChannelTextCard        = 26
	OpSendChannelCustom          = 27
	OpSendChannelLargeCard       = 28
	OpMuteGuildMember            = 29
	OpMuteGuild                  = 30
	OpGetRobotID                 = 31
	OpGetRobotInfo               = 32
	OpSendAdaptiveMessage        = 33
	OpSendGroupMessage           = 34
	OpSendGroupVideo             = 35
	OpSendGroupAudio             = 36
	OpGetFrameworkInfo           = 37
	OpSendGroupMarkdown          = 38
	OpSendGroupTextCard          = 39
	OpGetQQNickname              = 40
	OpSendGroupLargeCard         = 41
	OpSendAdaptiveLargeCard      = 42
	OpSendGroupThumbnailCard     = 43
	OpSendChannelThumbnailCard   = 44
	OpUploadImage                = 45
	OpRespondButton              = 46
	OpSendChannelMarkdown        = 47
	OpGetRobotAppID              = 48
	OpGetAvatar                  = 49
	OpGetQQAvatar                = 50
	OpRecallGroupMessage         = 51
	OpSendFriendMessage          = 52
	OpSendFriendVideo            = 53
	OpSendFriendAudio            = 54
	OpSendFriendMarkdown         = 55
	OpSendFriendTextCard         = 56
	OpSendFriendLargeCard        = 57
	OpSendFriendThumbnailCard    = 58
	OpRecallFriendMessage        = 59
	OpSendAdaptivePrivateMessage = 60
	OpAddChannelReaction         = 61
	OpDeleteChannelReaction      = 62
	OpListChannelReactionUsers   = 63
	OpGetRobotStats              = 64
	OpSendGroupButton            = 65
	OpSendFriendButton           = 66
	OpGetRobotToken              = 67
	OpGetRobotSecret             = 68
	OpSendGroupFile              = 69
	OpSendFriendFile             = 70
	OpSendGroupReply             = 71
	OpSendFriendReply            = 72
)

// ==================== types.go ====================
// EventType 表示 Bee 事件类型字符串，并允许未来新增事件。
type EventType string

const (
	EventGuildCreate           EventType = "GUILD_CREATE"
	EventGuildUpdate           EventType = "GUILD_UPDATE"
	EventGuildDelete           EventType = "GUILD_DELETE"
	EventChannelCreate         EventType = "CHANNEL_CREATE"
	EventChannelUpdate         EventType = "CHANNEL_UPDATE"
	EventChannelDelete         EventType = "CHANNEL_DELETE"
	EventGuildMemberAdd        EventType = "GUILD_MEMBER_ADD"
	EventGuildMemberUpdate     EventType = "GUILD_MEMBER_UPDATE"
	EventGuildMemberRemove     EventType = "GUILD_MEMBER_REMOVE"
	EventMessageDelete         EventType = "MESSAGE_DELETE"
	EventMessageReactionAdd    EventType = "MESSAGE_REACTION_ADD"
	EventMessageReactionRemove EventType = "MESSAGE_REACTION_REMOVE"
	EventForumThreadCreate     EventType = "FORUM_THREAD_CREATE"
	EventForumThreadUpdate     EventType = "FORUM_THREAD_UPDATE"
	EventForumThreadDelete     EventType = "FORUM_THREAD_DELETE"
	EventForumPostCreate       EventType = "FORUM_POST_CREATE"
	EventForumPostDelete       EventType = "FORUM_POST_DELETE"
	EventForumReplyCreate      EventType = "FORUM_REPLY_CREATE"
	EventForumReplyDelete      EventType = "FORUM_REPLY_DELETE"
	EventAudioLiveEnter        EventType = "AUDIO_OR_LIVE_CHANNEL_MEMBER_ENTER"
	EventAudioLiveExit         EventType = "AUDIO_OR_LIVE_CHANNEL_MEMBER_EXIT"
	EventGroupAddRobot         EventType = "GROUP_ADD_ROBOT"
	EventGroupDelRobot         EventType = "GROUP_DEL_ROBOT"
	EventGroupMessageReject    EventType = "GROUP_MSG_REJECT"
	EventGroupMessageReceive   EventType = "GROUP_MSG_RECEIVE"
	EventFriendAdd             EventType = "FRIEND_ADD"
	EventFriendDelete          EventType = "FRIEND_DEL"
	EventC2CMessageReject      EventType = "C2C_MSG_REJECT"
	EventC2CMessageReceive     EventType = "C2C_MSG_RECEIVE"
	EventGroupMemberAdd        EventType = "GROUP_MEMBER_ADD"
	EventGroupMemberRemove     EventType = "GROUP_MEMBER_REMOVE"
	EventInteractionCreate     EventType = "INTERACTION_CREATE"
)

// GuildInfo 表示频道详细信息。
type GuildInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IconURL     string `json:"icon"`
	OwnerID     string `json:"owner_id"`
	JoinedAt    string `json:"joined_at"`
	MemberCount int    `json:"member_count"`
	MaxMembers  int    `json:"max_members"`
	Description string `json:"description"`
}

// ChannelInfo 表示子频道详细信息。
type ChannelInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            int    `json:"type"`
	SubType         int    `json:"sub_type"`
	ParentID        string `json:"parent_id"`
	OwnerID         string `json:"owner_id"`
	PrivateType     int    `json:"private_type"`
	SpeakPermission int    `json:"speak_permission"`
	ApplicationID   int    `json:"application_id"`
}

// GuildMemberInfo 表示频道成员详细信息。
type GuildMemberInfo struct {
	UserName  string
	NickName  string
	AvatarURL string
	JoinedAt  string
	Roles     []string
}

// UnmarshalJSON 解析框架返回的频道成员 JSON。
func (v *GuildMemberInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		User struct {
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		} `json:"user"`
		Nick     string   `json:"nick"`
		JoinedAt string   `json:"joined_at"`
		Roles    []string `json:"roles"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	v.UserName, v.NickName, v.AvatarURL = raw.User.Name, raw.Nick, raw.User.Avatar
	v.JoinedAt, v.Roles = raw.JoinedAt, raw.Roles
	return nil
}

// FrameworkInfo 表示 Bee 框架信息。
type FrameworkInfo struct {
	Name        string `json:"name"`
	Version     string `json:"ver"`
	Sent        string `json:"send"`
	Received    string `json:"receive"`
	AccountType string `json:"type"`
	Uptime      string `json:"time"`
}

// RobotInfo 表示机器人账号信息。
type RobotInfo struct {
	ID        string `json:"id"`
	Name      string `json:"username"`
	AvatarURL string `json:"avatar"`
}

// MarkdownParam 表示 Markdown 模板的一组键和值。
type MarkdownParam struct{ Key, Value string }

// MarkdownMessage 表示原生或模板 Markdown 消息参数。
type MarkdownMessage struct {
	Native        string
	TemplateIndex int
	TemplateID    string
	Params        []MarkdownParam
	KeyboardJSON  string
	KeyboardID    string
}

// Button 表示 Markdown 自定义键盘中的一个按钮。
type Button struct {
	Label                string
	VisitedLabel         string
	Style                int
	Type                 int
	Permission           int
	UserIDs              []string
	RoleIDs              []string
	Data                 string
	Reply                bool
	Enter                bool
	Anchor               int
	ClickLimit           int
	AtBotShowChannelList bool
	UnsupportTips        string
}

// ==================== api.go ====================
func decodeCall[T any](ctx *RobotContext, op int, args ...string) (T, error) {
	var result T
	out, err := ctx.Call(op, args...)
	if err != nil {
		return result, err
	}
	if out == "" {
		return result, errors.New("框架返回空数据")
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return result, err
	}
	return result, nil
}

// Log 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) Log(content string) error { _, err := ctx.Call(OpLog, content); return err }

// ListGuilds 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) ListGuilds() (string, error) { return ctx.Call(OpListGuilds) }

// GetGuild 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetGuild(id string) (GuildInfo, error) {
	return decodeCall[GuildInfo](ctx, OpGetGuild, id)
}

// ListChannels 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) ListChannels(guildID string) (string, error) {
	return ctx.Call(OpListChannels, guildID)
}

// GetChannel 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetChannel(id string) (ChannelInfo, error) {
	return decodeCall[ChannelInfo](ctx, OpGetChannel, id)
}

// CreateChannel 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) CreateChannel(guildID, name string, channelType, subType, privateType, speakPermission, applicationID int) (bool, error) {
	return ctx.CallBool(OpCreateChannel, guildID, name, intText(channelType), intText(subType), intText(privateType), intText(speakPermission), intText(applicationID))
}

// GetChannelOnlineCount 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetChannelOnlineCount(id string) (string, error) {
	return ctx.Call(OpGetChannelOnlineCount, id)
}

// GetGuildMember 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetGuildMember(guildID, userID string) (GuildMemberInfo, error) {
	return decodeCall[GuildMemberInfo](ctx, OpGetGuildMember, guildID, userID)
}

// UpdateChannel 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) UpdateChannel(channelID, name string, privateType, speakPermission int) (bool, error) {
	return ctx.CallBool(OpUpdateChannel, channelID, name, intText(privateType), intText(speakPermission))
}

// DeleteChannel 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) DeleteChannel(id string) (bool, error) {
	return ctx.CallBool(OpDeleteChannel, id)
}

// DeleteGuildMember 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) DeleteGuildMember(guildID, userID string, recallDays int, blacklist bool) (bool, error) {
	// 修复原 SDK：显式传 false 不应被改成 true。
	return ctx.CallBool(OpDeleteGuildMember, guildID, userID, intText(recallDays), boolText(blacklist))
}

// ListGuildRoles 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) ListGuildRoles(guildID string) (string, error) {
	return ctx.Call(OpListGuildRoles, guildID)
}

// CreateGuildRole 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) CreateGuildRole(guildID, name string, color int, hoist bool) (string, error) {
	return ctx.Call(OpCreateGuildRole, guildID, name, intText(color), boolText(hoist))
}

// UpdateGuildRole 修改频道身份组。name 为空、color 为 -1 时表示对应项不修改；hoist 为 nil 时不修改展示状态。
func (ctx *RobotContext) UpdateGuildRole(guildID, roleID, name string, color int, hoist *bool) error {
	// 修复原 SDK：名称、颜色、展示状态必须属于同一个协议字符串。
	hoistText := ""
	if hoist != nil {
		hoistText = boolText(*hoist)
	}
	_, err := ctx.Call(OpUpdateGuildRole, guildID, roleID, name, intText(color), hoistText)
	return err
}

// IsGuildOwner 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) IsGuildOwner(guildID, userID string) (bool, error) {
	return ctx.CallBool(OpIsGuildOwner, guildID, userID)
}

// IsGuildAdmin 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) IsGuildAdmin(guildID, userID string) (bool, error) {
	return ctx.CallBool(OpIsGuildAdmin, guildID, userID)
}

// IsChannelAdmin 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) IsChannelAdmin(guildID, userID string) (bool, error) {
	return ctx.CallBool(OpIsChannelAdmin, guildID, userID)
}

// HasGuildRole 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) HasGuildRole(guildID, userID, roleID string) (bool, error) {
	return ctx.CallBool(OpHasGuildRole, guildID, userID, roleID)
}

// DeleteGuildRole 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) DeleteGuildRole(guildID, roleID string) error {
	_, err := ctx.Call(OpDeleteGuildRole, guildID, roleID)
	return err
}

// AddGuildMemberRole 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) AddGuildMemberRole(guildID, userID, roleID string) error {
	_, err := ctx.Call(OpAddGuildMemberRole, guildID, userID, roleID)
	return err
}

// RemoveGuildMemberRole 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) RemoveGuildMemberRole(guildID, userID, roleID string) error {
	_, err := ctx.Call(OpRemoveGuildMemberRole, guildID, userID, roleID)
	return err
}

// RecallChannelMessage 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) RecallChannelMessage(channelID, messageID string, hideTip bool) error {
	_, err := ctx.Call(OpRecallChannelMessage, channelID, messageID, boolText(hideTip))
	return err
}

// MuteGuildMember 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) MuteGuildMember(guildID, userID string, seconds int) error {
	_, err := ctx.Call(OpMuteGuildMember, guildID, userID, intText(seconds))
	return err
}

// MuteGuild 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) MuteGuild(guildID string, seconds int) error {
	_, err := ctx.Call(OpMuteGuild, guildID, intText(seconds))
	return err
}

// GetRobotID 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetRobotID() (string, error) { return ctx.Call(OpGetRobotID) }

// GetRobotInfo 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetRobotInfo() (RobotInfo, error) {
	return decodeCall[RobotInfo](ctx, OpGetRobotInfo)
}

// GetFrameworkInfo 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetFrameworkInfo() (FrameworkInfo, error) {
	return decodeCall[FrameworkInfo](ctx, OpGetFrameworkInfo)
}

// GetQQNickname 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetQQNickname(qq string) (string, error) {
	return ctx.Call(OpGetQQNickname, qq)
}

// UploadImage 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) UploadImage(pathOrURL string) (string, error) {
	return ctx.Call(OpUploadImage, pathOrURL)
}

// RespondButton 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) RespondButton(eventID string, responseType int) (string, error) {
	return ctx.Call(OpRespondButton, eventID, intText(responseType))
}

// GetRobotAppID 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetRobotAppID() (string, error) { return ctx.Call(OpGetRobotAppID) }

// GetAvatar 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetAvatar(userID string, size int) (string, error) {
	return ctx.Call(OpGetAvatar, userID, intText(size))
}

// GetQQAvatar 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetQQAvatar(qq string, size int) (string, error) {
	return ctx.Call(OpGetQQAvatar, qq, intText(size))
}

// RecallGroupMessage 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) RecallGroupMessage(groupID, messageID string) (string, error) {
	return ctx.Call(OpRecallGroupMessage, groupID, messageID)
}

// RecallFriendMessage 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) RecallFriendMessage(friendID, messageID string) (string, error) {
	return ctx.Call(OpRecallFriendMessage, friendID, messageID)
}

// AddChannelReaction 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) AddChannelReaction(channelID, messageID string, reactionType, reactionID int) (string, error) {
	return ctx.Call(OpAddChannelReaction, channelID, messageID, intText(reactionType), intText(reactionID))
}

// DeleteChannelReaction 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) DeleteChannelReaction(channelID, messageID string, reactionType, reactionID int) (string, error) {
	return ctx.Call(OpDeleteChannelReaction, channelID, messageID, intText(reactionType), intText(reactionID))
}

// ListChannelReactionUsers 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) ListChannelReactionUsers(channelID, messageID string, reactionType, reactionID int) (string, error) {
	return ctx.Call(OpListChannelReactionUsers, channelID, messageID, intText(reactionType), intText(reactionID))
}

// GetRobotStats 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetRobotStats(limit *int) (string, error) {
	value := ""
	if limit != nil {
		value = strconv.Itoa(*limit)
	}
	return ctx.Call(OpGetRobotStats, value)
}

// GetRobotToken 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetRobotToken() (string, error) { return ctx.Call(OpGetRobotToken) }

// GetRobotSecret 封装对应的 Bee 框架 API；参数和返回值与方法签名一致。
func (ctx *RobotContext) GetRobotSecret() (string, error) { return ctx.Call(OpGetRobotSecret) }

// ==================== messages.go ====================
func sendMessage(ctx *RobotContext, op int, target, content, media string, deleteMedia, active bool, recallInteraction *bool) (string, error) {
	messageID, eventID := activeIDs(ctx, active)
	args := []string{target, content, media, boolText(deleteMedia), messageID, eventID}
	if recallInteraction != nil {
		args = append(args, boolText(*recallInteraction))
	}
	return ctx.Call(op, args...)
}

func sendMedia(ctx *RobotContext, op int, target, file string, deleteFile, active bool, recallInteraction *bool) (string, error) {
	messageID, eventID := activeIDs(ctx, active)
	args := []string{target, file, boolText(deleteFile), messageID, eventID}
	if recallInteraction != nil {
		args = append(args, boolText(*recallInteraction))
	}
	return ctx.Call(op, args...)
}

func sendReply(ctx *RobotContext, op int, target, quotedMessageID, content, image string, deleteImage, active bool, recallInteraction *bool) (string, error) {
	messageID, eventID := activeIDs(ctx, active)
	args := []string{target, quotedMessageID, content, image, boolText(deleteImage), messageID, eventID}
	if recallInteraction != nil {
		args = append(args, boolText(*recallInteraction))
	}
	return ctx.Call(op, args...)
}

func sendCard(ctx *RobotContext, op int, target string, fields []string, active bool, recallInteraction *bool) (string, error) {
	messageID, eventID := activeIDs(ctx, active)
	args := append([]string{target}, fields...)
	args = append(args, messageID, eventID)
	if recallInteraction != nil {
		args = append(args, boolText(*recallInteraction))
	}
	return ctx.Call(op, args...)
}

func markdownArgs(ctx *RobotContext, target string, message MarkdownMessage, active bool, recallInteraction *bool) []string {
	args := []string{target, message.Native, intText(message.TemplateIndex), message.TemplateID}
	for i := 0; i < 10; i++ {
		if i < len(message.Params) {
			args = append(args, message.Params[i].Key, message.Params[i].Value)
		} else {
			args = append(args, "", "")
		}
	}
	messageID, eventID := activeIDs(ctx, active)
	args = append(args, messageID, message.KeyboardJSON, eventID, message.KeyboardID)
	if recallInteraction != nil {
		args = append(args, boolText(*recallInteraction))
	}
	return args
}

// SendChannelMessage 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendChannelMessage(channelID, content, image string, deleteImage, active bool) (string, error) {
	return sendMessage(ctx, OpSendChannelMessage, channelID, content, image, deleteImage, active, nil)
}

// SendChannelDM 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendChannelDM(guildID, content, image string, deleteImage, active bool) (string, error) {
	return sendMessage(ctx, OpSendChannelDM, guildID, content, image, deleteImage, active, nil)
}

// SendAdaptiveMessage 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendAdaptiveMessage(sourceID, content, image string, deleteImage, active bool) (string, error) {
	return sendMessage(ctx, OpSendAdaptiveMessage, sourceID, content, image, deleteImage, active, nil)
}

// SendGroupMessage 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupMessage(groupID, content, image string, deleteImage, active bool) (string, error) {
	return sendMessage(ctx, OpSendGroupMessage, groupID, content, image, deleteImage, active, nil)
}

// SendFriendMessage 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendMessage(friendID, content, image string, deleteImage, active, recallInteraction bool) (string, error) {
	return sendMessage(ctx, OpSendFriendMessage, friendID, content, image, deleteImage, active, &recallInteraction)
}

// SendAdaptivePrivateMessage 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendAdaptivePrivateMessage(sourceID, content, image string, deleteImage, active bool) (string, error) {
	return sendMessage(ctx, OpSendAdaptivePrivateMessage, sourceID, content, image, deleteImage, active, nil)
}

// SendGroupVideo 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupVideo(groupID, file string, deleteFile, active bool) (string, error) {
	return sendMedia(ctx, OpSendGroupVideo, groupID, file, deleteFile, active, nil)
}

// SendGroupAudio 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupAudio(groupID, file string, deleteFile, active bool) (string, error) {
	return sendMedia(ctx, OpSendGroupAudio, groupID, file, deleteFile, active, nil)
}

// SendFriendVideo 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendVideo(friendID, file string, deleteFile, active, recallInteraction bool) (string, error) {
	return sendMedia(ctx, OpSendFriendVideo, friendID, file, deleteFile, active, &recallInteraction)
}

// SendFriendAudio 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendAudio(friendID, file string, deleteFile, active, recallInteraction bool) (string, error) {
	return sendMedia(ctx, OpSendFriendAudio, friendID, file, deleteFile, active, &recallInteraction)
}

// SendGroupFile 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupFile(groupID, file string, deleteFile, active bool) (string, error) {
	return sendMedia(ctx, OpSendGroupFile, groupID, file, deleteFile, active, nil)
}

// SendFriendFile 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendFile(friendID, file string, deleteFile, active, recallInteraction bool) (string, error) {
	return sendMedia(ctx, OpSendFriendFile, friendID, file, deleteFile, active, &recallInteraction)
}

// SendChannelReply 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendChannelReply(channelID, quotedMessageID, content, image string, active bool) (string, error) {
	messageID, eventID := activeIDs(ctx, active)
	return ctx.Call(OpSendChannelReply, channelID, quotedMessageID, content, image, messageID, eventID)
}

// SendGroupReply 发送群引用消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupReply(groupID, quotedMessageID, content, image string, deleteImage, active bool) (string, error) {
	return sendReply(ctx, OpSendGroupReply, groupID, quotedMessageID, content, image, deleteImage, active, nil)
}

// SendFriendReply 发送好友引用消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendReply(friendID, quotedMessageID, content, image string, deleteImage, active, recallInteraction bool) (string, error) {
	return sendReply(ctx, OpSendFriendReply, friendID, quotedMessageID, content, image, deleteImage, active, &recallInteraction)
}

// SendChannelCustom 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendChannelCustom(channelID, rawJSON string, active bool) (string, error) {
	messageID, eventID := activeIDs(ctx, active)
	return ctx.Call(OpSendChannelCustom, channelID, rawJSON, messageID, eventID)
}

// SendChannelTextCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendChannelTextCard(channelID, title, preview, content, imageURL string, active bool) (string, error) {
	return sendCard(ctx, OpSendChannelTextCard, channelID, []string{title, preview, content, imageURL}, active, nil)
}

// SendGroupTextCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupTextCard(groupID, preview, content string, active bool) (string, error) {
	return sendCard(ctx, OpSendGroupTextCard, groupID, []string{preview, content}, active, nil)
}

// SendFriendTextCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendTextCard(friendID, preview, content string, active, recallInteraction bool) (string, error) {
	return sendCard(ctx, OpSendFriendTextCard, friendID, []string{preview, content}, active, &recallInteraction)
}

// SendChannelLargeCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendChannelLargeCard(channelID, title, subtitle, preview, imageURL, jumpURL string, active bool) (string, error) {
	return sendCard(ctx, OpSendChannelLargeCard, channelID, []string{title, subtitle, preview, imageURL, jumpURL}, active, nil)
}

// SendGroupLargeCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupLargeCard(groupID, title, subtitle, preview, imageURL, jumpURL string, active bool) (string, error) {
	return sendCard(ctx, OpSendGroupLargeCard, groupID, []string{title, subtitle, preview, imageURL, jumpURL}, active, nil)
}

// SendAdaptiveLargeCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendAdaptiveLargeCard(sourceID, title, subtitle, preview, imageURL, jumpURL string, active bool) (string, error) {
	return sendCard(ctx, OpSendAdaptiveLargeCard, sourceID, []string{title, subtitle, preview, imageURL, jumpURL}, active, nil)
}

// SendFriendLargeCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendLargeCard(friendID, title, subtitle, preview, imageURL, jumpURL string, active, recallInteraction bool) (string, error) {
	return sendCard(ctx, OpSendFriendLargeCard, friendID, []string{title, subtitle, preview, imageURL, jumpURL}, active, &recallInteraction)
}

// SendGroupThumbnailCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupThumbnailCard(groupID, title, subtitle, preview, imageURL, jumpURL string, active bool) (string, error) {
	return sendCard(ctx, OpSendGroupThumbnailCard, groupID, []string{title, subtitle, preview, imageURL, jumpURL}, active, nil)
}

// SendChannelThumbnailCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendChannelThumbnailCard(channelID, title, subtitle, preview, imageURL, jumpURL string, active bool) (string, error) {
	return sendCard(ctx, OpSendChannelThumbnailCard, channelID, []string{title, subtitle, preview, imageURL, jumpURL}, active, nil)
}

// SendFriendThumbnailCard 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendThumbnailCard(friendID, title, subtitle, preview, imageURL, jumpURL string, active, recallInteraction bool) (string, error) {
	return sendCard(ctx, OpSendFriendThumbnailCard, friendID, []string{title, subtitle, preview, imageURL, jumpURL}, active, &recallInteraction)
}

// SendGroupMarkdown 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupMarkdown(groupID string, message MarkdownMessage, active bool) (string, error) {
	return ctx.Call(OpSendGroupMarkdown, markdownArgs(ctx, groupID, message, active, nil)...)
}

// SendChannelMarkdown 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendChannelMarkdown(channelID string, message MarkdownMessage, active bool) (string, error) {
	return ctx.Call(OpSendChannelMarkdown, markdownArgs(ctx, channelID, message, active, nil)...)
}

// SendFriendMarkdown 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendMarkdown(friendID string, message MarkdownMessage, active, recallInteraction bool) (string, error) {
	return ctx.Call(OpSendFriendMarkdown, markdownArgs(ctx, friendID, message, active, &recallInteraction)...)
}

// SendGroupButton 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendGroupButton(groupID, keyboardID string, active bool) (string, error) {
	return sendCard(ctx, OpSendGroupButton, groupID, []string{keyboardID}, active, nil)
}

// SendFriendButton 发送对应类型的 Bee 消息；active 为 true 时使用主动消息模式。
func (ctx *RobotContext) SendFriendButton(friendID, keyboardID string, active, recallInteraction bool) (string, error) {
	return sendCard(ctx, OpSendFriendButton, friendID, []string{keyboardID}, active, &recallInteraction)
}

// ==================== helpers.go ====================
// At 生成艾特指定用户的文本代码。
func At(userID string) string { return "<@!" + userID + ">" }

// AtEveryone 返回艾特全体成员的文本代码。
func AtEveryone() string { return "@everyone" }

// MentionedUserID 从 <@!用户ID> 中提取用户 ID。
func MentionedUserID(text string) string {
	start := strings.Index(text, "<@!")
	if start < 0 {
		return ""
	}
	rest := text[start+3:]
	end := strings.IndexByte(rest, '>')
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// ImageDownloadURL 从 Bee 图片消息代码中提取下载地址。
func ImageDownloadURL(message string) string {
	start := strings.Index(message, ",url=")
	if start < 0 {
		return ""
	}
	rest := message[start+5:]
	end := strings.IndexByte(rest, ']')
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// InlineCommand 生成可嵌入 Markdown 的 QQ 指令链接。
func InlineCommand(label, command string, send bool) string {
	enter := "false"
	if send {
		enter = "true"
	}
	return fmt.Sprintf("[%s](mqqapi://aio/inlinecmd?command=%s&reply=false&enter=%s)", label, url.QueryEscape(command), enter)
}

// ResolveRedirect 请求网址并返回重定向后的地址。
func ResolveRedirect(rawURL string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	location := resp.Header.Get("Location")
	if location == "" {
		return rawURL, nil
	}
	base, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	target, err := base.Parse(location)
	if err != nil {
		return "", err
	}
	return target.String(), nil
}

type keyboard struct {
	Content keyboardContent `json:"content"`
}
type keyboardContent struct {
	BotAppID string        `json:"bot_appid"`
	Rows     []keyboardRow `json:"rows"`
}
type keyboardRow struct {
	Buttons []keyboardButton `json:"buttons"`
}
type keyboardButton struct {
	ID         string         `json:"id"`
	RenderData keyboardRender `json:"render_data"`
	Action     keyboardAction `json:"action"`
}
type keyboardRender struct {
	Label        string `json:"label"`
	VisitedLabel string `json:"visited_label"`
	Style        int    `json:"style"`
}
type keyboardAction struct {
	Type                 int                `json:"type"`
	Permission           keyboardPermission `json:"permission"`
	Data                 string             `json:"data"`
	Reply                bool               `json:"reply,omitempty"`
	Enter                bool               `json:"enter,omitempty"`
	Anchor               int                `json:"anchor,omitempty"`
	ClickLimit           int                `json:"click_limit,omitempty"`
	AtBotShowChannelList bool               `json:"at_bot_show_channel_list,omitempty"`
	UnsupportTips        string             `json:"unsupport_tips,omitempty"`
}
type keyboardPermission struct {
	Type           int      `json:"type"`
	SpecifyUserIDs []string `json:"specify_user_ids,omitempty"`
	SpecifyRoleIDs []string `json:"specify_role_ids,omitempty"`
}

// BuildKeyboard 使用传入的机器人 AppID，避免原 SDK 将 AppID 固定为 102069021。
func BuildKeyboard(botAppID string, rows [][]Button) (string, error) {
	if botAppID == "" {
		return "", errors.New("botAppID 不能为空")
	}
	if len(rows) > 5 {
		return "", errors.New("按钮最多 5 行")
	}
	result := keyboard{Content: keyboardContent{BotAppID: botAppID}}
	for rowIndex, row := range rows {
		if len(row) > 10 {
			return "", fmt.Errorf("第 %d 行按钮超过 10 个", rowIndex+1)
		}
		outRow := keyboardRow{}
		for columnIndex, button := range row {
			label := button.Label
			if label == "" {
				label = "按钮"
			}
			visited := button.VisitedLabel
			if visited == "" {
				visited = "按钮"
			}
			data := button.Data
			if data == "" {
				data = "你好"
			}
			outRow.Buttons = append(outRow.Buttons, keyboardButton{
				ID:         fmt.Sprintf("%d_%d", rowIndex, columnIndex),
				RenderData: keyboardRender{Label: label, VisitedLabel: visited, Style: button.Style},
				Action: keyboardAction{
					Type:                 button.Type,
					Permission:           keyboardPermission{Type: button.Permission, SpecifyUserIDs: button.UserIDs, SpecifyRoleIDs: button.RoleIDs},
					Data:                 data,
					Reply:                button.Reply,
					Enter:                button.Enter,
					Anchor:               button.Anchor,
					ClickLimit:           button.ClickLimit,
					AtBotShowChannelList: button.AtBotShowChannelList,
					UnsupportTips:        button.UnsupportTips,
				},
			})
		}
		result.Content.Rows = append(result.Content.Rows, outRow)
	}
	data, err := json.Marshal(result)
	return string(data), err
}

// ==================== windows.go ====================
var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	user32                  = syscall.NewLazyDLL("user32.dll")
	procMultiByteToWideChar = kernel32.NewProc("MultiByteToWideChar")
	procWideCharToMultiByte = kernel32.NewProc("WideCharToMultiByte")
	procMessageBoxW         = user32.NewProc("MessageBoxW")
)

const cpGBK = 936

func gbkToUTF8(src []byte) string {
	if len(src) == 0 {
		return ""
	}
	n, _, _ := procMultiByteToWideChar.Call(cpGBK, 0, uintptr(unsafe.Pointer(&src[0])), uintptr(len(src)), 0, 0)
	if n == 0 {
		return ""
	}
	wide := make([]uint16, n)
	procMultiByteToWideChar.Call(cpGBK, 0, uintptr(unsafe.Pointer(&src[0])), uintptr(len(src)), uintptr(unsafe.Pointer(&wide[0])), n)
	return string(utf16.Decode(wide))
}

func utf8ToGBK(src []byte) []byte {
	wide := utf16.Encode([]rune(string(src)))
	if len(wide) == 0 {
		return nil
	}
	n, _, _ := procWideCharToMultiByte.Call(cpGBK, 0, uintptr(unsafe.Pointer(&wide[0])), uintptr(len(wide)), 0, 0, 0, 0)
	if n == 0 {
		return nil
	}
	out := make([]byte, n)
	procWideCharToMultiByte.Call(cpGBK, 0, uintptr(unsafe.Pointer(&wide[0])), uintptr(len(wide)), uintptr(unsafe.Pointer(&out[0])), n, 0, 0)
	return out
}

func messageBox(text string) {
	t, _ := syscall.UTF16PtrFromString(text)
	caption, _ := syscall.UTF16PtrFromString("Bee Go SDK")
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(caption)), 0x40)
}

// ==================== sdk.go ====================
// BeeSeparator 是 Bee 框架命令协议使用的字段分隔符。
const BeeSeparator = "%@#bee#@%"

const (
	// MessageIgnore 与 MessageContinue 都是 0，是原 Bee SDK 中的语义别名。
	MessageIgnore = 0
	// MessageContinue 表示消息处理完成后继续投递给后续插件。
	MessageContinue = 0
	// MessageIntercept 表示拦截消息，不再投递给后续插件。
	MessageIntercept = 1
)

// RobotContext 保存框架随事件传入的机器人上下文。
type RobotContext struct {
	API       uintptr         `json:"-"`
	APIText   string          `json:"api"`
	Message   string          `json:"msg"`
	MessageID string          `json:"msg_id"`
	ChannelID string          `json:"channel_id"`
	GuildID   string          `json:"guild_id"`
	FromID    string          `json:"form_id"`
	RobotID   string          `json:"robot_id"`
	PluginID  string          `json:"plugin_id"`
	EventID   string          `json:"event_id"`
	Raw       json.RawMessage `json:"raw"`
}

// BeeAPI 是面向插件开发者的简化 SDK 入口。
// 它保存的是单次回调上下文快照，禁止跨消息保存为全局变量复用。
type BeeAPI struct {
	ctx *RobotContext
}

// MessageTarget 是已经绑定消息目标的快捷发送器。
type MessageTarget struct {
	ctx      *RobotContext
	targetID string
	kind     int
}

const (
	targetFriend = iota
	targetGroup
	targetChannel
	targetChannelDM
)

// NewBeeAPI 根据当前 Bee 回调传入的最新机器人 JSON 创建简化 SDK 入口。
// 每次收到消息或事件时都必须重新调用，确保 msg_id、event_id、api 和 plugin_id 属于当前回调。
func NewBeeAPI(robotJSON string) (*BeeAPI, error) {
	ctx, err := ParseRobotContext(robotJSON)
	if err != nil {
		return nil, err
	}
	return &BeeAPI{ctx: ctx}, nil
}

// Log 向 Bee 框架输出日志。
func (api *BeeAPI) Log(content string) error {
	return api.ctx.Log(content)
}

// Friend 绑定好友 ID，之后可直接调用 SendText、SendImage 等方法。
func (api *BeeAPI) Friend(friendID string) *MessageTarget {
	return &MessageTarget{ctx: api.ctx, targetID: friendID, kind: targetFriend}
}

// Group 绑定群 ID，之后可直接调用 SendText、SendImage 等方法。
func (api *BeeAPI) Group(groupID string) *MessageTarget {
	return &MessageTarget{ctx: api.ctx, targetID: groupID, kind: targetGroup}
}

// Channel 绑定子频道 ID，之后可直接调用 SendText、SendImage 等方法。
func (api *BeeAPI) Channel(channelID string) *MessageTarget {
	return &MessageTarget{ctx: api.ctx, targetID: channelID, kind: targetChannel}
}

// ChannelDM 绑定频道 ID，用于发送频道私信。
func (api *BeeAPI) ChannelDM(guildID string) *MessageTarget {
	return &MessageTarget{ctx: api.ctx, targetID: guildID, kind: targetChannelDM}
}

// IsGuildOwner 判断指定用户是否为频道主。
func (api *BeeAPI) IsGuildOwner(guildID, userID string) (bool, error) {
	return api.ctx.IsGuildOwner(guildID, userID)
}

// MuteMember 设置频道成员禁言秒数，传 0 表示解除禁言。
func (api *BeeAPI) MuteMember(guildID, userID string, seconds int) error {
	return api.ctx.MuteGuildMember(guildID, userID, seconds)
}

// MuteAll 设置频道全员禁言秒数，传 0 表示关闭全员禁言。
func (api *BeeAPI) MuteAll(guildID string, seconds int) error {
	return api.ctx.MuteGuild(guildID, seconds)
}

// ParseMention 判断消息是否艾特当前机器人，并返回移除艾特代码后的消息内容。
func (api *BeeAPI) ParseMention(content string) (bool, string, error) {
	robotID, err := api.ctx.GetRobotID()
	if err != nil {
		return false, content, err
	}
	mention := At(robotID)
	if !strings.Contains(content, mention) {
		return false, strings.TrimSpace(content), nil
	}
	return true, strings.TrimSpace(strings.ReplaceAll(content, mention, "")), nil
}

// SendText 发送纯文本消息，默认使用当前消息进行被动回复。
func (target *MessageTarget) SendText(content string) (string, error) {
	switch target.kind {
	case targetFriend:
		return target.ctx.SendFriendMessage(target.targetID, content, "", false, false, false)
	case targetGroup:
		return target.ctx.SendGroupMessage(target.targetID, content, "", false, false)
	case targetChannelDM:
		return target.ctx.SendChannelDM(target.targetID, content, "", false, false)
	default:
		return target.ctx.SendChannelMessage(target.targetID, content, "", false, false)
	}
}

// SendImage 发送图片消息，可同时携带文字。
func (target *MessageTarget) SendImage(content, image string) (string, error) {
	switch target.kind {
	case targetFriend:
		return target.ctx.SendFriendMessage(target.targetID, content, image, false, false, false)
	case targetGroup:
		return target.ctx.SendGroupMessage(target.targetID, content, image, false, false)
	case targetChannelDM:
		return target.ctx.SendChannelDM(target.targetID, content, image, false, false)
	default:
		return target.ctx.SendChannelMessage(target.targetID, content, image, false, false)
	}
}

// SendActiveText 主动发送纯文本消息；主动消息受平台频率和次数限制。
func (target *MessageTarget) SendActiveText(content string) (string, error) {
	switch target.kind {
	case targetFriend:
		return target.ctx.SendFriendMessage(target.targetID, content, "", false, true, false)
	case targetGroup:
		return target.ctx.SendGroupMessage(target.targetID, content, "", false, true)
	case targetChannelDM:
		return target.ctx.SendChannelDM(target.targetID, content, "", false, true)
	default:
		return target.ctx.SendChannelMessage(target.targetID, content, "", false, true)
	}
}

// SendTextCard 发送频道文字卡片消息；仅用于 Channel 目标。
func (target *MessageTarget) SendTextCard(title, preview, content, imageURL string) (string, error) {
	return target.ctx.SendChannelTextCard(target.targetID, title, preview, content, imageURL, false)
}

// SendCustom 发送频道自定义 JSON 消息；仅用于 Channel 目标。
func (target *MessageTarget) SendCustom(rawJSON string) (string, error) {
	return target.ctx.SendChannelCustom(target.targetID, rawJSON, false)
}

// SendLargeCard 发送频道大图卡片消息；仅用于 Channel 目标。
func (target *MessageTarget) SendLargeCard(title, subtitle, preview, imageURL, jumpURL string) (string, error) {
	return target.ctx.SendChannelLargeCard(target.targetID, title, subtitle, preview, imageURL, jumpURL, false)
}

// Reply 发送引用消息，按当前目标自动选择频道、群或好友引用接口。
func (target *MessageTarget) Reply(messageID, content, image string) (string, error) {
	switch target.kind {
	case targetFriend:
		return target.ctx.SendFriendReply(target.targetID, messageID, content, image, false, false, false)
	case targetGroup:
		return target.ctx.SendGroupReply(target.targetID, messageID, content, image, false, false)
	default:
		return target.ctx.SendChannelReply(target.targetID, messageID, content, image, false)
	}
}

// ParseRobotContext 从 JSON 解析机器人上下文和框架 API 地址。
func ParseRobotContext(text string) (*RobotContext, error) {
	var ctx RobotContext
	if err := json.Unmarshal([]byte(text), &ctx); err != nil {
		return nil, fmt.Errorf("解析机器人上下文: %w", err)
	}
	if ctx.APIText == "" {
		return nil, errors.New("机器人上下文缺少 api")
	}
	address, err := strconv.ParseUint(ctx.APIText, 10, 32)
	if err != nil || address == 0 {
		return nil, fmt.Errorf("无效 api 地址 %q", ctx.APIText)
	}
	ctx.API = uintptr(address)
	return &ctx, nil
}

// MustRobotContext 解析机器人上下文，失败时触发 panic。
func MustRobotContext(text string) *RobotContext {
	ctx, err := ParseRobotContext(text)
	if err != nil {
		panic(err)
	}
	return ctx
}

func (ctx *RobotContext) command(op int, args ...string) string {
	fields := make([]string, 0, len(args)+2)
	fields = append(fields, strconv.Itoa(op))
	// 操作码 31、45、48 按原 SDK 不带 plugin_id。
	if op != OpGetRobotID && op != OpUploadImage && op != OpGetRobotAppID {
		fields = append(fields, ctx.PluginID)
	}
	fields = append(fields, args...)
	return strings.Join(fields, BeeSeparator)
}

func encodeFrameworkText(value string) string {
	var out strings.Builder
	for _, r := range value {
		text := string(r)
		// Bee 的易语言边界只能稳定处理 GBK。GBK 可表示的文字原样保留；
		// 其他 Unicode 字符转成 UTF-16 \uXXXX，非 BMP 字符会自动变成代理对。
		if gbkToUTF8(utf8ToGBK([]byte(text))) == text {
			out.WriteString(text)
			continue
		}
		for _, unit := range utf16.Encode([]rune{r}) {
			fmt.Fprintf(&out, `\u%04X`, unit)
		}
	}
	return out.String()
}

// Call 按 Bee 协议拼接命令，调用机器人上下文中的 32 位 stdcall 框架函数。
// 所有传给框架的内容会先把 GBK 无法表示的 Unicode 字符编码为 UTF-16 \uXXXX；
// 例如“哈哈😄”会转换为“哈哈\uD83D\uDE04”，再将整条命令转为 GBK。
// 返回值按 GBK解码为 UTF-8；返回内存归框架所有。
func (ctx *RobotContext) Call(op int, args ...string) (string, error) {
	if ctx == nil || ctx.API == 0 {
		return "", errors.New("无有效框架 API 地址")
	}
	command := encodeFrameworkText(ctx.command(op, args...))
	commandGBK := utf8ToGBK([]byte(command))
	cCommand := C.CString(string(commandGBK))
	defer C.free(unsafe.Pointer(cCommand))
	result := C.bee_call_api(C.uintptr_t(ctx.API), cCommand)
	if result == nil {
		return "", nil
	}
	return goText(result), nil
}

// CallBool 调用框架 API，并将返回值“1”转换为 true。
func (ctx *RobotContext) CallBool(op int, args ...string) (bool, error) {
	out, err := ctx.Call(op, args...)
	return out == "1", err
}

func boolText(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
func intText(v int) string { return strconv.Itoa(v) }
func activeIDs(ctx *RobotContext, active bool) (string, string) {
	if active {
		return "", ctx.EventID
	}
	return ctx.MessageID, ctx.EventID
}
