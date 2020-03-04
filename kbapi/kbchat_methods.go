package kbapi

import(
    "github.com/keybase/client/go/protocol/chat1"
    "github.com/keybase/client/go/client"
    "github.com/keybase/client/go/protocol/keybase1"
    "github.com/keybase/client/go/libkb"
    "github.com/keybase/client/go/chat/utils"
    "github.com/keybase/client/go/chat"
    "context"
    "time"
    "errors"
    "fmt"
    "strings"
)

/*
chat api methods:
list
read
listconvsoname
join
leave
send
attach
reaction
edit
advertisecommands
clearcommands
listcommands
listmembers
get

team api
list-user-memberships
list-team-memberships

wallet api
details
*/

const (
  methodList                = "list" //
  methodRead                = "read" //
  methodGet                 = "get" //
  methodSend                = "send" //
  methodEdit                = "edit" //
  methodReaction            = "reaction" //
  methodAttach              = "attach" //
  methodListConvsOnName     = "listconvsonname" //
  methodJoin                = "join" //
  methodLeave               = "leave" //
  methodAdvertiseCommands   = "advertisecommands" //
  methodClearCommands       = "clearcommands" //
  methodListCommands        = "listcommands" //
  methodListMembers         = "listmembers" //
  // team api
  listTeamMethod      = "list-team-memberships"
  listUserMethod      = "list-user-memberships"
)

/*
listOptionsV1
readOptionsV1
sendOptionsV1
getOptionsV1
editOptionsV1
reactionOptionsV1
attachOptionsV1
listConvsOnNameOptionsV1
joinOptionsV1
leaveOptionsV1
advertiseCommandsOptionsV1
listCommandsOptionsV1
listMembersOptionsV1

listTeamOptions
listUserOptions
*/
type ChatChannel chat1.ChatChannel

func (c ChatChannel) IsNil() bool {
  return c == ChatChannel{}
}

// Valid returns true if the ChatChannel has at least a Name.
func (c ChatChannel) Valid() bool {
  if len(c.Name) == 0 {
    return false
  }
  if len(c.MembersType) > 0 && !isValidMembersType(c.MembersType) {
    return false
  }
  return true
}

func (c ChatChannel) Visibility() (vis keybase1.TLFVisibility) {
  vis = keybase1.TLFVisibility_PRIVATE
  if c.Public {
    vis = keybase1.TLFVisibility_PUBLIC
  }
  return vis
}

func (c ChatChannel) GetMembersType(e *libkb.Env) chat1.ConversationMembersType {
  return client.MembersTypeFromStrDefault(c.MembersType, e)
}
// ChatMessage represents a text message to be sent.
type ChatMessage struct {
  Body string
}

// Valid returns true if the message has a body.
func (c ChatMessage) Valid() bool {
  return len(c.Body) > 0
}

type listOptionsV1 struct {
  ConversationID chat1.ConvIDStr `json:"conversation_id,omitempty"`
  UnreadOnly     bool            `json:"unread_only,omitempty"`
  TopicType      string          `json:"topic_type,omitempty"`
  ShowErrors     bool            `json:"show_errors,omitempty"`
  FailOffline    bool            `json:"fail_offline,omitempty"`
}

type readOptionsV1 struct {
  Channel        ChatChannel
  ConversationID chat1.ConvIDStr   `json:"conversation_id"`
  Pagination     *chat1.Pagination `json:"pagination,omitempty"`
  Peek           bool
  UnreadOnly     bool `json:"unread_only"`
  FailOffline    bool `json:"fail_offline"`
}

type sendOptionsV1 struct {
  Channel           ChatChannel
  ConversationID    chat1.ConvIDStr `json:"conversation_id"`
  Message           ChatMessage
  Nonblock          bool              `json:"nonblock"`
  MembersType       string            `json:"members_type"`
  EphemeralLifetime time.Duration     `json:"exploding_lifetime"`
  ConfirmLumenSend  bool              `json:"confirm_lumen_send"`
  ReplyTo           *chat1.MessageID  `json:"reply_to"`
}

type getOptionsV1 struct {
  Channel        ChatChannel
  ConversationID chat1.ConvIDStr   `json:"conversation_id"`
  MessageIDs     []chat1.MessageID `json:"message_ids"`
  Peek           bool
  FailOffline    bool `json:"fail_offline"`
}

type editOptionsV1 struct {
  Channel        ChatChannel
  ConversationID chat1.ConvIDStr `json:"conversation_id"`
  MessageID      chat1.MessageID `json:"message_id"`
  Message        ChatMessage
}

type reactionOptionsV1 struct {
  Channel        ChatChannel
  ConversationID chat1.ConvIDStr `json:"conversation_id"`
  MessageID      chat1.MessageID `json:"message_id"`
  Message        ChatMessage
}

type attachOptionsV1 struct {
  Channel           ChatChannel
  ConversationID    chat1.ConvIDStr `json:"conversation_id"`
  Filename          string
  Preview           string
  Title             string
  EphemeralLifetime time.Duration `json:"exploding_lifetime"`
}

type listConvsOnNameOptionsV1 struct {
  Name        string `json:"name,omitempty"`
  MembersType string `json:"members_type,omitempty"`
  TopicType   string `json:"topic_type,omitempty"`
}

type joinOptionsV1 struct {
  Channel        ChatChannel
  ConversationID chat1.ConvIDStr `json:"conversation_id"`
}

type leaveOptionsV1 struct {
  Channel        ChatChannel
  ConversationID chat1.ConvIDStr `json:"conversation_id"`
}

type advertiseCommandsOptionsV1 struct {
  Alias          string `json:"alias,omitempty"`
  Advertisements []chat1.AdvertiseCommandAPIParam
}

type listCommandsOptionsV1 struct {
  Channel        ChatChannel
  ConversationID chat1.ConvIDStr `json:"conversation_id"`
}

type listMembersOptionsV1 struct {
  Channel        ChatChannel
  ConversationID chat1.ConvIDStr `json:"conversation_id"`
}

func ListV1(g *libkb.GlobalContext, ctx context.Context, opts listOptionsV1) Reply {
  var cl chat1.ChatList
  var rlimits []chat1.RateLimit
  client, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  topicType, err := TopicTypeFromStrDefault(opts.TopicType)
  if err != nil {
    return errReply(err)
  }
  var convIDs []chat1.ConversationID
  if opts.ConversationID != "" {
    convID, err := chat1.MakeConvID(opts.ConversationID.String())
    if err != nil {
      return errReply(err)
    }
    convIDs = append(convIDs, convID)
  }
  res, err := client.GetInboxAndUnboxUILocal(ctx, chat1.GetInboxAndUnboxUILocalArg{
    Query: &chat1.GetInboxLocalQuery{
      ConvIDs:           convIDs,
      Status:            utils.VisibleChatConversationStatuses(),
      TopicType:         &topicType,
      UnreadOnly:        opts.UnreadOnly,
      OneChatTypePerTLF: new(bool),
    },
    IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_CLI,
  })
  if err != nil {
    return errReply(err)
  }
  rlimits = utils.AggRateLimits(res.RateLimits)
  if opts.FailOffline && res.Offline {
    return errReply(chat.OfflineError{})
  }
  cl = chat1.ChatList{
    Offline:          res.Offline,
    IdentifyFailures: res.IdentifyFailures,
  }
  for _, conv := range res.Conversations {
    cl.Conversations = append(cl.Conversations, utils.ExportToSummary(conv))
  }
  cl.RateLimits = aggRateLimits(rlimits)
  return Reply{Result: cl}
}

func ReadV1(g *libkb.GlobalContext, ctx context.Context, opts readOptionsV1) Reply {
  var rlimits []chat1.RateLimit
  client, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }

  conv, rlimits, err := findConversation(g, ctx, opts.ConversationID, opts.Channel)
  if err != nil {
    return errReply(err)
  }

  arg := chat1.GetThreadLocalArg{
    ConversationID: conv.Info.Id,
    Pagination:     opts.Pagination,
    Query: &chat1.GetThreadQuery{
      MarkAsRead: !opts.Peek,
    },
    IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_CLI,
  }
  threadView, err := client.GetThreadLocal(ctx, arg)
  if err != nil {
    return errReply(err)
  }
  rlimits = append(rlimits, threadView.RateLimits...)

  // Check to see if this was fetched offline and we should fail
  if opts.FailOffline && threadView.Offline {
    return errReply(chat.OfflineError{})
  }

  // This could be lower than the truth if any messages were
  // posted between the last two gregor rpcs.
  readMsgID := conv.ReaderInfo.ReadMsgid

  selfUID := g.Env.GetUID()
  //if selfUID.IsNil() {
  //  c.G().Log.Warning("Could not get self UID for api")
  //}

  messages, err := formatMessages(ctx, threadView.Thread.Messages, conv, selfUID, readMsgID, opts.UnreadOnly)
  if err != nil {
    return errReply(err)
  }

  thread := chat1.Thread{
    Offline:          threadView.Offline,
    IdentifyFailures: threadView.IdentifyFailures,
    Pagination:       threadView.Thread.Pagination,
    Messages:         messages,
  }

  thread.RateLimits = aggRateLimits(rlimits)
  return Reply{Result: thread}
}

func GetV1(g *libkb.GlobalContext, ctx context.Context, opts getOptionsV1) Reply {
  var rlimits []chat1.RateLimit
  client, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }

  conv, rlimits, err := findConversation(g, ctx, opts.ConversationID, opts.Channel)
  if err != nil {
    return errReply(err)
  }

  arg := chat1.GetMessagesLocalArg{
    ConversationID:   conv.Info.Id,
    MessageIDs:       opts.MessageIDs,
    IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_CLI,
  }

  res, err := client.GetMessagesLocal(ctx, arg)
  if err != nil {
    return errReply(err)
  }

  // Check to see if this was fetched offline and we should fail
  if opts.FailOffline && res.Offline {
    return errReply(chat.OfflineError{})
  }

  selfUID := g.Env.GetUID()
  //if selfUID.IsNil() {
  //  c.G().Log.Warning("Could not get self UID for api")
  //}

  messages, err := formatMessages(ctx, res.Messages, conv, selfUID, 0 /* readMsgID */, false /* unreadOnly */)
  if err != nil {
    return errReply(err)
  }

  thread := chat1.Thread{
    Offline:          res.Offline,
    IdentifyFailures: res.IdentifyFailures,
    Messages:         messages,
  }
  thread.RateLimits = aggRateLimits(rlimits)
  return Reply{Result: thread}
}






func errReply(err error) Reply {
  if rlerr, ok := err.(libkb.ChatRateLimitError); ok {
    return Reply{Error: &CallError{Message: err.Error(), Data: rlerr.RateLimit}}
  }
  return Reply{Error: &CallError{Message: err.Error()}}
}

func aggRateLimits(rlimits []chat1.RateLimit) (res []chat1.RateLimitRes) {
  m := make(map[string]chat1.RateLimit)
  for _, rl := range rlimits {
    m[rl.Name] = rl
  }
  for _, v := range m {
    res = append(res, chat1.RateLimitRes{
      Tank:     v.Name,
      Capacity: v.MaxCalls,
      Reset:    v.WindowReset,
      Gas:      v.CallsRemaining,
    })
  }
  return res
}

func findConversation(g *libkb.GlobalContext, ctx context.Context, convIDStr chat1.ConvIDStr,
  channel ChatChannel) (chat1.ConversationLocal, []chat1.RateLimit, error) {
  var conv chat1.ConversationLocal
  var rlimits []chat1.RateLimit

  if channel.IsNil() && len(convIDStr) == 0 {
    return conv, rlimits, errors.New("missing conversation specificer")
  }

  var convID chat1.ConversationID
  if channel.IsNil() {
    var err error
    convID, err = chat1.MakeConvID(convIDStr.String())
    if err != nil {
      return conv, rlimits, fmt.Errorf("invalid conversation ID: %s", convIDStr)
    }
  }

  existing, existingRl, err := getExistingConvs(g, ctx, convID, channel)
  if err != nil {
    return conv, rlimits, err
  }
  rlimits = append(rlimits, existingRl...)

  if len(existing) > 1 {
    return conv, rlimits, fmt.Errorf("multiple conversations matched %q", channel.Name)
  }
  if len(existing) == 0 {
    return conv, rlimits, fmt.Errorf("no conversations matched %q", channel.Name)
  }

  return existing[0], rlimits, nil
}

func formatMessages(ctx context.Context, messages []chat1.MessageUnboxed,
  conv chat1.ConversationLocal,
  selfUID keybase1.UID,
  readMsgID chat1.MessageID, unreadOnly bool) (ret []chat1.Message, err error) {
  for _, m := range messages {
    st, err := m.State()
    if err != nil {
      return nil, errors.New("invalid message: unknown state")
    }

    if st == chat1.MessageUnboxedState_ERROR {
      em := m.Error().ErrMsg
      ret = append(ret, chat1.Message{
        Error: &em,
      })
      continue
    }

    // skip any PLACEHOLDER or OUTBOX messages
    if st != chat1.MessageUnboxedState_VALID {
      continue
    }

    mv := m.Valid()

    if mv.ClientHeader.MessageType == chat1.MessageType_TLFNAME {
      // skip TLFNAME messages
      continue
    }

    unread := mv.ServerHeader.MessageID > readMsgID
    if unreadOnly && !unread {
      continue
    }
    if !selfUID.IsNil() {
      fromSelf := (mv.ClientHeader.Sender.String() == selfUID.String())
      unread = unread && (!fromSelf)
      if unreadOnly && fromSelf {
        continue
      }
    }

    prev := mv.ClientHeader.Prev
    // Avoid having null show up in the output JSON.
    if prev == nil {
      prev = []chat1.MessagePreviousPointer{}
    }

    msg := chat1.MsgSummary{
      Id:     mv.ServerHeader.MessageID,
      ConvID: conv.GetConvID().ConvIDStr(),
      Channel: chat1.ChatChannel{
        Name:        conv.Info.TlfName,
        Public:      mv.ClientHeader.TlfPublic,
        TopicType:   strings.ToLower(mv.ClientHeader.Conv.TopicType.String()),
        MembersType: strings.ToLower(conv.GetMembersType().String()),
        TopicName:   conv.Info.TopicName,
      },
      Sender: chat1.MsgSender{
        Uid:        keybase1.UID(mv.ClientHeader.Sender.String()),
        DeviceID:   keybase1.DeviceID(mv.ClientHeader.SenderDevice.String()),
        Username:   mv.SenderUsername,
        DeviceName: mv.SenderDeviceName,
      },
      SentAt:              mv.ServerHeader.Ctime.UnixSeconds(),
      SentAtMs:            mv.ServerHeader.Ctime.UnixMilliseconds(),
      Prev:                prev,
      Unread:              unread,
      RevokedDevice:       mv.SenderDeviceRevokedAt != nil,
      KbfsEncrypted:       mv.ClientHeader.KbfsCryptKeysUsed == nil || *mv.ClientHeader.KbfsCryptKeysUsed,
      IsEphemeral:         mv.IsEphemeral(),
      IsEphemeralExpired:  mv.IsEphemeralExpired(time.Now()),
      ETime:               mv.Etime(),
      Content:             convertMsgBody(mv.MessageBody),
      HasPairwiseMacs:     mv.HasPairwiseMacs(),
      AtMentionUsernames:  mv.AtMentionUsernames,
      ChannelMention:      strings.ToLower(mv.ChannelMention.String()),
      ChannelNameMentions: utils.PresentChannelNameMentions(ctx, mv.ChannelNameMentions),
    }
    if mv.ClientHeader.BotUID != nil {
      botUID := keybase1.UID(mv.ClientHeader.BotUID.String())
      msg.BotInfo = &chat1.MsgBotInfo{
        BotUID:      botUID,
        BotUsername: mv.BotUsername,
      }
    }
    if mv.Reactions.Reactions != nil {
      msg.Reactions = &mv.Reactions
    }

    ret = append(ret, chat1.Message{
      Msg: &msg,
    })
  }

  if ret == nil {
    // Avoid having null show up in the output JSON.
    ret = []chat1.Message{}
  }
  return ret, nil
}

func getExistingConvs(g *libkb.GlobalContext, ctx context.Context, convID chat1.ConversationID,
  channel ChatChannel) ([]chat1.ConversationLocal, []chat1.RateLimit, error) {
  client, err := client.GetChatLocalClient(g)
  if err != nil {
    return nil, nil, err
  }
  if !convID.IsNil() {
    gilres, err := client.GetInboxAndUnboxLocal(ctx, chat1.GetInboxAndUnboxLocalArg{
      Query: &chat1.GetInboxLocalQuery{
        ConvIDs: []chat1.ConversationID{convID},
      },
      IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_CLI,
    })
    if err != nil {
      //c.G().Log.Warning("GetInboxLocal error: %s", err)
      return nil, nil, err
    }
    convs := gilres.Conversations
    if len(convs) == 0 {
      // NOTE: don't change this error without also changing the managed-bots repo
      // https://github.com/keybase/managed-bots/blob/4ed0f563e6f3276a953bd33a00f98a75dc32d102/base/output.go#L75
      return nil, nil, fmt.Errorf("no conversations matched %q", convID)
    }
    return convs, gilres.RateLimits, nil
  }

  tlfName := channel.Name
  vis := keybase1.TLFVisibility_PRIVATE
  if channel.Public {
    vis = keybase1.TLFVisibility_PUBLIC
  }
  tt, err := TopicTypeFromStrDefault(channel.TopicType)
  if err != nil {
    return nil, nil, err
  }
  findRes, err := client.FindConversationsLocal(ctx, chat1.FindConversationsLocalArg{
    TlfName:          tlfName,
    MembersType:      channel.GetMembersType(g.GetEnv()),
    Visibility:       vis,
    TopicType:        tt,
    TopicName:        channel.TopicName,
    IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_CLI,
  })
  if err != nil {
    return nil, nil, err
  }

  return findRes.Conversations, findRes.RateLimits, nil
}

func convertMsgBody(mb chat1.MessageBody) chat1.MsgContent {
  return chat1.MsgContent{
    TypeName:           strings.ToLower(chat1.MessageTypeRevMap[mb.MessageType__]),
    Text:               mb.Text__,
    Attachment:         mb.Attachment__,
    Edit:               mb.Edit__,
    Reaction:           mb.Reaction__,
    Delete:             mb.Delete__,
    Metadata:           mb.Metadata__,
    Headline:           mb.Headline__,
    AttachmentUploaded: mb.Attachmentuploaded__,
    System:             mb.System__,
    SendPayment:        mb.Sendpayment__,
    RequestPayment:     mb.Requestpayment__,
    Unfurl:             mb.Unfurl__,
    Flip:               displayFlipBody(mb.Flip__),
  }
}

func displayFlipBody(flip *chat1.MessageFlip) (res *chat1.MsgFlipContent) {
  if flip == nil {
    return res
  }
  res = new(chat1.MsgFlipContent)
  res.GameID = flip.GameID.FlipGameIDStr()
  res.FlipConvID = flip.FlipConvID.ConvIDStr()
  res.TeamMentions = flip.TeamMentions
  res.UserMentions = flip.UserMentions
  res.Text = flip.Text
  return res
}

func TopicTypeFromStrDefault(str string) (chat1.TopicType, error) {
  if len(str) == 0 {
    return chat1.TopicType_CHAT, nil
  }
  tt, ok := chat1.TopicTypeMap[strings.ToUpper(str)]
  if !ok {
    return chat1.TopicType_NONE, fmt.Errorf("invalid topic type: '%v'", str)
  }
  return tt, nil
}

func isValidMembersType(mt string) bool {
  for typ := range chat1.ConversationMembersTypeMap {
    if strings.ToLower(typ) == mt {
      return true
    }
  }
  return false
}
