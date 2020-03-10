package kbapi


import (
 "fmt"
 "strings"
 "io"
 "github.com/keybase/client/go/chat/utils"
 "github.com/keybase/client/go/libkb"
 "github.com/keybase/client/go/protocol/chat1"
 keybase1 "github.com/keybase/client/go/protocol/keybase1"
 "golang.org/x/net/context"
 "encoding/json"
 "time"
 //"bytes"
)



type baseNotificationDisplay struct {
 libkb.Contextified
}

func newBaseNotificationDisplay(g *libkb.GlobalContext) *baseNotificationDisplay {
 return &baseNotificationDisplay{
  Contextified: libkb.NewContextified(g),
 }
}

func (d *baseNotificationDisplay) printf(fmt string, args ...interface{}) error {
 _, err := d.G().UI.GetTerminalUI().Printf(fmt, args...)
 return err
}

func (d *baseNotificationDisplay) errorf(format string, args ...interface{}) error {
 _, err := d.G().UI.GetTerminalUI().ErrorWriter().Write([]byte(fmt.Sprintf(format, args...)))
 return err
}

func (d *baseNotificationDisplay) printJSON(data interface{}) {
 if jsonStr, err := json.Marshal(data); err != nil {
  _ = d.errorf("Error while marshaling JSON: %s\n", err)
 } else {
  _ = d.printf("%s\n", string(jsonStr))
 }
}

type chatNotificationConfig struct {
 showLocal     bool
 showNewConvs  bool
 hideExploding bool
}

type chatNotificationDisplay struct {
 g *libkb.GlobalContext
 config            chatNotificationConfig
 filtersNormalized []chat1.ConversationID
 output io.Writer
}

func newChatNotificationDisplay(g *libkb.GlobalContext, config chatNotificationConfig, output io.Writer) *chatNotificationDisplay {
 return &chatNotificationDisplay{ g: g, config: config, output: output}
}

const notifTypeChat = "chat"

func newMsgNotification(source string) *chat1.MsgNotification {
 return &chat1.MsgNotification{
  Type:   notifTypeChat,
  Source: source,
 }
}

const notifTypeChatConv = "chat_conv"

func newConvNotification() *chat1.ConvNotification {
 return &chat1.ConvNotification{
  Type: notifTypeChatConv,
 }
}

func (d *chatNotificationDisplay) setupFilters(ctx context.Context, channelFilters []ChatChannel) error {
 for _, v := range channelFilters {
  if MembersTypeFromStrDefault(v.MembersType, d.g.Env) == chat1.ConversationMembersType_TEAM &&
   len(v.TopicName) == 0 {
   // treat this formulation of a channel as listing all team convs the users is in
   topicType, err := TopicTypeFromStrDefault(v.TopicType)
   if err != nil {
    return err
   }
   convs, _, err := getAllTeamConvs(d.g, ctx, v.Name, &topicType)
   if err != nil {
    return err
   }
   for _, conv := range convs {
    d.filtersNormalized = append(d.filtersNormalized, conv.GetConvID())
   }
  } else {
   conv, _, err := findConversation(d.g, ctx, "", v)
   if err != nil {
    return err
   }
   d.filtersNormalized = append(d.filtersNormalized, conv.GetConvID())
  }
 }
 return nil
}

func (d *chatNotificationDisplay) formatMessage(inMsg chat1.IncomingMessage) *chat1.Message {
 state, err := inMsg.Message.State()
 if err != nil {
  errStr := err.Error()
  return &chat1.Message{Error: &errStr}
 }

 switch state {
 case chat1.MessageUnboxedState_ERROR:
  errStr := inMsg.Message.Error().ErrMsg
  return &chat1.Message{Error: &errStr}
 case chat1.MessageUnboxedState_VALID:
  // if we weren't able to get an inbox item here, then just return an error
  if inMsg.Conv == nil {
   msg := "unable to get chat channel"
   return &chat1.Message{Error: &msg}
  }
  mv := inMsg.Message.Valid()
  summary := &chat1.MsgSummary{
   Id:     mv.MessageID,
   ConvID: inMsg.ConvID.ConvIDStr(),
   Channel: chat1.ChatChannel{
    Name:        inMsg.Conv.Name,
    MembersType: strings.ToLower(inMsg.Conv.MembersType.String()),
    TopicType:   strings.ToLower(inMsg.Conv.TopicType.String()),
    TopicName:   inMsg.Conv.Channel,
    Public:      inMsg.Conv.Visibility == keybase1.TLFVisibility_PUBLIC,
   },
   Sender: chat1.MsgSender{
    Uid:        keybase1.UID(mv.SenderUID.String()),
    DeviceID:   keybase1.DeviceID(mv.SenderDeviceID.String()),
    Username:   mv.SenderUsername,
    DeviceName: mv.SenderDeviceName,
   },
   SentAt:              mv.Ctime.UnixSeconds(),
   SentAtMs:            mv.Ctime.UnixMilliseconds(),
   RevokedDevice:       mv.SenderDeviceRevokedAt != nil,
   IsEphemeral:         mv.IsEphemeral,
   IsEphemeralExpired:  mv.IsEphemeralExpired,
   ETime:               mv.Etime,
   HasPairwiseMacs:     mv.HasPairwiseMacs,
   Content:             convertMsgBody(mv.MessageBody),
   AtMentionUsernames:  mv.AtMentions,
   ChannelMention:      strings.ToLower(mv.ChannelMention.String()),
   ChannelNameMentions: mv.ChannelNameMentions,
  }
  if mv.Reactions.Reactions != nil {
   summary.Reactions = &mv.Reactions
  }
  return &chat1.Message{Msg: summary}
 default:
  return nil
 }
}

func (d *chatNotificationDisplay) matchFilters(convID chat1.ConversationID) bool {
 if len(d.filtersNormalized) == 0 {
  // No filters - every message is relayed.
  return true
 }
 for _, v := range d.filtersNormalized {
  if convID.Eq(v) {
   return true
  }
 }
 // None of our filters matched.
 return false
}

func (d *chatNotificationDisplay) NewChatActivity(ctx context.Context, arg chat1.NewChatActivityArg) error {
 if !d.config.showLocal && arg.Source == chat1.ChatActivitySource_LOCAL {
  // Skip local message
  return nil
 }

 activity := arg.Activity
 typ, err := activity.ActivityType()
 if err != nil {
  return err
 }
 if typ == chat1.ChatActivityType_INCOMING_MESSAGE {
  inMsg := activity.IncomingMessage()
  if d.config.hideExploding && inMsg.Message.IsEphemeral() {
   // Skip exploding message
   return nil
  }
  if !d.matchFilters(inMsg.ConvID) {
   // Skip filtered out message.
   return nil
  }
  msg := d.formatMessage(inMsg)
  if msg == nil {
   return nil
  }
  source := strings.ToLower(arg.Source.String())
  notif := newMsgNotification(source)
  notif.Msg = msg.Msg
  notif.Error = msg.Error
  notif.Pagination = inMsg.Pagination

  // send output to pipe
  jsonStr, err := json.Marshal(notif)
  if err != nil {
    //fmt.Printf("ERROR JSON MARSHAL\n")
    return err
  }
  d.output.Write(jsonStr)
 } else if d.config.showNewConvs && typ == chat1.ChatActivityType_NEW_CONVERSATION {
  convInfo := activity.NewConversation()
  notif := newConvNotification()
  if convInfo.Conv == nil {
   err := fmt.Sprintf("No conversation info found: %v", convInfo.ConvID.String())
   notif.Error = &err
  } else {
   conv := utils.ExportToSummary(*convInfo.Conv)
   notif.Conv = &conv
  }
  jsonStr, err := json.Marshal(notif)
  if err != nil {
    //fmt.Printf("ERROR JSON MARSHAL\n")
    return err
  }
  d.output.Write(jsonStr)
 }
 return nil
}

func (d *chatNotificationDisplay) ChatJoinedConversation(ctx context.Context, arg chat1.ChatJoinedConversationArg) error {
 if !d.config.showNewConvs {
  return nil
 }

 notif := newConvNotification()
 if arg.Conv == nil {
  err := fmt.Sprintf("No conversation info found: %v", arg.ConvID.String())
  notif.Error = &err
 } else {
  conv := utils.ExportToSummary(*arg.Conv)
  notif.Conv = &conv
 }
 jsonStr, err := json.Marshal(notif)
 if err != nil {
   //fmt.Printf("ERROR JSON MARSHAL\n")
   return err
 }
 d.output.Write(jsonStr)
 return nil
}

func (d *chatNotificationDisplay) ChatIdentifyUpdate(context.Context, keybase1.CanonicalTLFNameAndIDWithBreaks) error {
 return nil
}
func (d *chatNotificationDisplay) ChatTLFFinalize(context.Context, chat1.ChatTLFFinalizeArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatTLFResolve(context.Context, chat1.ChatTLFResolveArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatInboxStale(context.Context, keybase1.UID) error { return nil }
func (d *chatNotificationDisplay) ChatThreadsStale(context.Context, chat1.ChatThreadsStaleArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatTypingUpdate(context.Context, []chat1.ConvTypingUpdate) error {
 return nil
}
func (d *chatNotificationDisplay) ChatLeftConversation(context.Context, chat1.ChatLeftConversationArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatResetConversation(context.Context, chat1.ChatResetConversationArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatInboxSyncStarted(context.Context, keybase1.UID) error {
 return nil
}
func (d *chatNotificationDisplay) ChatInboxSynced(context.Context, chat1.ChatInboxSyncedArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatSetConvRetention(context.Context, chat1.ChatSetConvRetentionArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatSetTeamRetention(context.Context, chat1.ChatSetTeamRetentionArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatSetConvSettings(context.Context, chat1.ChatSetConvSettingsArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatSubteamRename(context.Context, chat1.ChatSubteamRenameArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatKBFSToImpteamUpgrade(context.Context, chat1.ChatKBFSToImpteamUpgradeArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatAttachmentUploadStart(context.Context, chat1.ChatAttachmentUploadStartArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatAttachmentUploadProgress(context.Context, chat1.ChatAttachmentUploadProgressArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatPaymentInfo(context.Context, chat1.ChatPaymentInfoArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatRequestInfo(context.Context, chat1.ChatRequestInfoArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatPromptUnfurl(context.Context, chat1.ChatPromptUnfurlArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatConvUpdate(context.Context, chat1.ChatConvUpdateArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatWelcomeMessageLoaded(context.Context, chat1.ChatWelcomeMessageLoadedArg) error {
 return nil
}
func (d *chatNotificationDisplay) ChatParticipantsInfo(context.Context,
 map[chat1.ConvIDStr][]chat1.UIParticipant) error {
 return nil
}


func sendPing(cli keybase1.SessionClient) error {
 ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
 defer cancel()
 return cli.SessionPing(ctx)
}
