package kbapi

import(
    "github.com/keybase/client/go/protocol/chat1"
    "github.com/keybase/client/go/client"
    "github.com/keybase/client/go/protocol/keybase1"
    "github.com/keybase/client/go/libkb"
    "github.com/keybase/client/go/chat/utils"
    "github.com/keybase/client/go/chat"
    "github.com/keybase/go-framed-msgpack-rpc/rpc"
    gregor1 "github.com/keybase/client/go/protocol/gregor1"
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

type sendArgV1 struct {
  // convQuery  chat1.GetInboxLocalQuery
  conversationID    chat1.ConversationID
  channel           ChatChannel
  body              chat1.MessageBody
  mtype             chat1.MessageType
  supersedes        chat1.MessageID
  deletes           []chat1.MessageID
  response          string
  nonblock          bool
  ephemeralLifetime time.Duration
  replyTo           *chat1.MessageID
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

type postHeader struct {
  conversationID chat1.ConversationID
  clientHeader   chat1.MessageClientHeader
  rateLimits     []chat1.RateLimit
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

type ChatAPIUI struct {
  utils.DummyChatUI
  sessionID            int
  allowStellarPayments bool
}

var _ chat1.ChatUiInterface = (*ChatAPIUI)(nil)

func AllowStellarPayments(enabled bool) func(*ChatAPIUI) {
  return func(c *ChatAPIUI) {
    c.SetAllowStellarPayments(enabled)
  }
}

func NewChatAPIUI(opts ...func(*ChatAPIUI)) *ChatAPIUI {
  c := &ChatAPIUI{
    DummyChatUI: utils.DummyChatUI{},
    sessionID:   randSessionID(),
  }
  for _, o := range opts {
    o(c)
  }
  return c
}

func (u *ChatAPIUI) ChatStellarDataConfirm(ctx context.Context, arg chat1.ChatStellarDataConfirmArg) (bool, error) {
  return u.allowStellarPayments, nil
}

func (u *ChatAPIUI) SetAllowStellarPayments(enabled bool) {
  u.allowStellarPayments = enabled
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

// SendV1 implements ChatServiceHandler.SendV1.
func SendV1(g *libkb.GlobalContext, ctx context.Context, opts sendOptionsV1, chatUI chat1.ChatUiInterface) Reply {
  convID, err := chat1.MakeConvID(opts.ConversationID.String())
  if err != nil {
    return errReply(fmt.Errorf("invalid conv ID: %s", opts.ConversationID))
  }
  arg := sendArgV1{
    conversationID:    convID,
    channel:           opts.Channel,
    body:              chat1.NewMessageBodyWithText(chat1.MessageText{Body: opts.Message.Body}),
    mtype:             chat1.MessageType_TEXT,
    response:          "message sent",
    nonblock:          opts.Nonblock,
    ephemeralLifetime: opts.EphemeralLifetime,
    replyTo:           opts.ReplyTo,
  }
  return sendV1(g, ctx, arg, chatUI)
}

func sendV1(g *libkb.GlobalContext, ctx context.Context, arg sendArgV1, chatUI chat1.ChatUiInterface) Reply {
  kbchatUI := newDelegateChatUI()
  kbchatUI.RegisterChatUI(chatUI)
  defer kbchatUI.DeregisterChatUI(chatUI)

  cclient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  protocols := []rpc.Protocol{
    chat1.ChatUiProtocol(kbchatUI),
  }
  if err := client.RegisterProtocolsWithContext(protocols, g); err != nil {
    return errReply(err)
  }

  var rl []chat1.RateLimit
  existing, existingRl, err := getExistingConvs(g, ctx, arg.conversationID, arg.channel)
  if err != nil {
    return errReply(err)
  }
  rl = append(rl, existingRl...)

  header, err := makePostHeader(g, ctx, arg, existing)
  if err != nil {
    return errReply(err)
  }
  rl = append(rl, header.rateLimits...)

  postArg := chat1.PostLocalArg{
      SessionID:      getSessionID(chatUI),
      ConversationID: header.conversationID,
      Msg: chat1.MessagePlaintext{
      ClientHeader: header.clientHeader,
      MessageBody:  arg.body,
    },
    ReplyTo:          arg.replyTo,
    IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_CLI,
  }
  var idFails []keybase1.TLFIdentifyFailure
  var msgID *chat1.MessageID
  var obid *chat1.OutboxID
  if arg.nonblock {
    var nbarg chat1.PostLocalNonblockArg
    nbarg.ConversationID = postArg.ConversationID
    nbarg.Msg = postArg.Msg
    nbarg.IdentifyBehavior = postArg.IdentifyBehavior
    plres, err := cclient.PostLocalNonblock(ctx, nbarg)
    if err != nil {
      return errReply(err)
    }
    obid = &plres.OutboxID
    rl = append(rl, plres.RateLimits...)
    idFails = plres.IdentifyFailures
  } else {
    plres, err := cclient.PostLocal(ctx, postArg)
    if err != nil {
      return errReply(err)
    }
    msgID = &plres.MessageID
    rl = append(rl, plres.RateLimits...)
    idFails = plres.IdentifyFailures
  }

  res := chat1.SendRes{
    Message:          arg.response,
    MessageID:        msgID,
    OutboxID:         obid,
    RateLimits:       aggRateLimits(rl),
    IdentifyFailures: idFails,
  }

  return Reply{Result: res}
}

func EditV1(g *libkb.GlobalContext, ctx context.Context, opts editOptionsV1) Reply {
  convID, err := chat1.MakeConvID(opts.ConversationID.String())
  if err != nil {
    return errReply(fmt.Errorf("invalid conv ID: %s", opts.ConversationID))
  }
  arg := sendArgV1{
    conversationID: convID,
    channel:        opts.Channel,
    body:           chat1.NewMessageBodyWithEdit(chat1.MessageEdit{MessageID: opts.MessageID, Body: opts.Message.Body}),
    mtype:          chat1.MessageType_EDIT,
    supersedes:     opts.MessageID,
    response:       "message edited",
  }
  return sendV1(g, ctx, arg, utils.DummyChatUI{})
}

func ReactionV1(g *libkb.GlobalContext, ctx context.Context, opts reactionOptionsV1) Reply {
  convID, err := chat1.MakeConvID(opts.ConversationID.String())
  if err != nil {
    return errReply(fmt.Errorf("invalid conv ID: %s", opts.ConversationID))
  }
  arg := sendArgV1{
    conversationID: convID,
    channel:        opts.Channel,
    body:           chat1.NewMessageBodyWithReaction(chat1.MessageReaction{MessageID: opts.MessageID, Body: opts.Message.Body}),
    mtype:          chat1.MessageType_REACTION,
    supersedes:     opts.MessageID,
    response:       "message reacted to",
  }
  return sendV1(g, ctx, arg, utils.DummyChatUI{})
}

func AttachV1(g *libkb.GlobalContext, ctx context.Context, opts attachOptionsV1,
  chatUI chat1.ChatUiInterface, notifyUI chat1.NotifyChatInterface) Reply {
  var rl []chat1.RateLimit
  convID, err := chat1.MakeConvID(opts.ConversationID.String())
  if err != nil {
    return errReply(fmt.Errorf("invalid conv ID: %s", opts.ConversationID))
  }
  sarg := sendArgV1{
    conversationID:    convID,
    channel:           opts.Channel,
    mtype:             chat1.MessageType_ATTACHMENT,
    ephemeralLifetime: opts.EphemeralLifetime,
  }
  existing, existingRl, err := getExistingConvs(g, ctx, sarg.conversationID, sarg.channel)
  if err != nil {
    return errReply(err)
  }
  rl = append(rl, existingRl...)

  header, err := makePostHeader(g, ctx, sarg, existing)
  if err != nil {
    return errReply(err)
  }
  rl = append(rl, header.rateLimits...)

  vis := keybase1.TLFVisibility_PRIVATE
  if header.clientHeader.TlfPublic {
    vis = keybase1.TLFVisibility_PUBLIC
  }
  var ephemeralLifetime *gregor1.DurationSec
  if header.clientHeader.EphemeralMetadata != nil {
    ephemeralLifetime = &header.clientHeader.EphemeralMetadata.Lifetime
  }
  arg := chat1.PostFileAttachmentArg{
    ConversationID:    header.conversationID,
    TlfName:           header.clientHeader.TlfName,
    Visibility:        vis,
    Filename:          opts.Filename,
    Title:             opts.Title,
    EphemeralLifetime: ephemeralLifetime,
  }
  // check for preview
  if len(opts.Preview) > 0 {
    loc := chat1.NewPreviewLocationWithFile(opts.Preview)
    arg.CallerPreview = &chat1.MakePreviewRes{
      Location: &loc,
    }
  }
  kbchatUI := newDelegateChatUI()
  kbchatUI.RegisterChatUI(chatUI)
  defer kbchatUI.DeregisterChatUI(chatUI)
  cclient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  protocols := []rpc.Protocol{
    client.NewStreamUIProtocol(g),
    chat1.ChatUiProtocol(kbchatUI),
    chat1.NotifyChatProtocol(notifyUI),
  }
  if err := client.RegisterProtocolsWithContext(protocols,g); err != nil {
    return errReply(err)
  }
  cli, err := client.GetNotifyCtlClient(g)
  if err != nil {
    return errReply(err)
  }
  channels := keybase1.NotificationChannels{
    Chatattachments: true,
  }
  if err := cli.SetNotifications(context.TODO(), channels); err != nil {
    return errReply(err)
  }

  var pres chat1.PostLocalRes
  pres, err = cclient.PostFileAttachmentLocal(ctx, chat1.PostFileAttachmentLocalArg{
    SessionID: getSessionID(chatUI),
    Arg:       arg,
  })
  rl = append(rl, pres.RateLimits...)
  if err != nil {
    return errReply(err)
  }

  res := chat1.SendRes{
    Message:    "attachment sent",
    MessageID:  &pres.MessageID,
    RateLimits: aggRateLimits(rl),
  }

  return Reply{Result: res}
}

func ListConvsOnNameV1(g *libkb.GlobalContext, ctx context.Context, opts listConvsOnNameOptionsV1) Reply {
  cclient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  topicType, err := TopicTypeFromStrDefault(opts.TopicType)
  if err != nil {
    return errReply(err)
  }
  mt := client.MembersTypeFromStrDefault(opts.MembersType, g.GetEnv())

  listRes, err := cclient.GetTLFConversationsLocal(ctx, chat1.GetTLFConversationsLocalArg{
    TlfName:     opts.Name,
    TopicType:   topicType,
    MembersType: mt,
  })
  if err != nil {
    return errReply(err)
  }
  var cl chat1.ChatList
  cl.RateLimits = aggRateLimits(listRes.RateLimits)
  for _, conv := range listRes.Convs {
    cl.Conversations = append(cl.Conversations, utils.ExportToSummary(conv))
  }
  return Reply{Result: cl}
}

func JoinV1(g *libkb.GlobalContext, ctx context.Context, opts joinOptionsV1) Reply {
  cclient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  convID, rl, err := resolveAPIConvID(g, ctx, opts.ConversationID, opts.Channel)
  if err != nil {
    return errReply(err)
  }
  res, err := cclient.JoinConversationByIDLocal(ctx, convID)
  if err != nil {
    return errReply(err)
  }
  allLimits := append(rl, res.RateLimits...)
  cres := chat1.EmptyRes{
    RateLimits: aggRateLimits(allLimits),
  }
  return Reply{Result: cres}
}

func LeaveV1(g *libkb.GlobalContext, ctx context.Context, opts leaveOptionsV1) Reply {
  cclient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  convID, rl, err := resolveAPIConvID(g, ctx, opts.ConversationID, opts.Channel)
  if err != nil {
    return errReply(err)
  }
  res, err := cclient.LeaveConversationLocal(ctx, convID)
  if err != nil {
    return errReply(err)
  }
  allLimits := append(rl, res.RateLimits...)
  cres := chat1.EmptyRes{
    RateLimits: aggRateLimits(allLimits),
  }
  return Reply{Result: cres}
}

func AdvertiseCommandsV1(g *libkb.GlobalContext, ctx context.Context, opts advertiseCommandsOptionsV1) Reply {
  cclient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  var alias *string
  if opts.Alias != "" {
    alias = new(string)
    *alias = opts.Alias
  }
  var ads []chat1.AdvertiseCommandsParam
  for _, ad := range opts.Advertisements {
    typ, err := getAdvertTyp(ad.Typ)
    if err != nil {
      return errReply(err)
    }
    var teamName *string
    if ad.TeamName != "" {
      adTeamName := ad.TeamName
      teamName = &adTeamName
    }
    ads = append(ads, chat1.AdvertiseCommandsParam{
      Typ:      typ,
      Commands: ad.Commands,
      TeamName: teamName,
    })
  }
  res, err := cclient.AdvertiseBotCommandsLocal(ctx, chat1.AdvertiseBotCommandsLocalArg{
    Alias:          alias,
    Advertisements: ads,
  })
  if err != nil {
    return errReply(err)
  }
  return Reply{Result: res}
}

func ClearCommandsV1(g *libkb.GlobalContext, ctx context.Context) Reply {
  cclient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  res, err := cclient.ClearBotCommandsLocal(ctx)
  if err != nil {
    return errReply(err)
  }
  return Reply{Result: res}
}

func ListCommandsV1(g *libkb.GlobalContext, ctx context.Context, opts listCommandsOptionsV1) Reply {
  cclient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  convID, rl, err := resolveAPIConvID(g, ctx, opts.ConversationID, opts.Channel)
  if err != nil {
    return errReply(err)
  }
  lres, err := cclient.ListBotCommandsLocal(ctx, convID)
  if err != nil {
    return errReply(err)
  }
  res := chat1.ListCommandsRes{
    Commands: lres.Commands,
  }
  res.RateLimits = aggRateLimits(append(rl, lres.RateLimits...))
  return Reply{Result: res}
}

func ListMembersV1(g *libkb.GlobalContext, ctx context.Context, opts listMembersOptionsV1) Reply {
  conv, _, err := findConversation(g, ctx, opts.ConversationID, opts.Channel)
  if err != nil {
    return errReply(err)
  }

  chatClient, err := client.GetChatLocalClient(g)
  if err != nil {
    return errReply(err)
  }
  teamID, err := chatClient.TeamIDFromTLFName(ctx, chat1.TeamIDFromTLFNameArg{
    TlfName:     conv.Info.TlfName,
    MembersType: conv.Info.MembersType,
    TlfPublic:   conv.Info.Visibility == keybase1.TLFVisibility_PUBLIC,
  })
  if err != nil {
    return errReply(err)
  }

  cli, err := client.GetTeamsClient(g)
  if err != nil {
    return errReply(err)
  }
  details, err := cli.TeamGetByID(context.Background(), keybase1.TeamGetByIDArg{Id: teamID})
  if err != nil {
    return errReply(err)
  }

  // filter the member list down to the specific conversation members based on the server-trust list
  if conv.Info.TopicName != "" && opts.Channel.TopicName != "general" {
    details = keybase1.FilterTeamDetailsForMembers(conv.AllNames(), details)
  }

  return Reply{Result: details}
}





func getAdvertTyp(typ string) (chat1.BotCommandsAdvertisementTyp, error) {
  switch typ {
  case "public":
    return chat1.BotCommandsAdvertisementTyp_PUBLIC, nil
  case "teamconvs":
    return chat1.BotCommandsAdvertisementTyp_TLFID_CONVS, nil
  case "teammembers":
    return chat1.BotCommandsAdvertisementTyp_TLFID_MEMBERS, nil
  default:
    return chat1.BotCommandsAdvertisementTyp_PUBLIC, fmt.Errorf("unknown advertisement type %q", typ)
  }
}

func resolveAPIConvID(g *libkb.GlobalContext, ctx context.Context, convID chat1.ConvIDStr,
  channel ChatChannel) (chat1.ConversationID, []chat1.RateLimit, error) {
  conv, limits, err := findConversation(g, ctx, convID, channel)
  if err != nil {
    return chat1.ConversationID{}, nil, err
  }
  return conv.Info.Id, limits, nil
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

func makePostHeader(g *libkb.GlobalContext, ctx context.Context, arg sendArgV1, existing []chat1.ConversationLocal) (*postHeader, error) {
  client, err := client.GetChatLocalClient(g)
  if err != nil {
    return nil, err
  }

  membersType := arg.channel.GetMembersType(g.GetEnv())
  var header postHeader
  var convTriple chat1.ConversationIDTriple
  var tlfName string
  var visibility keybase1.TLFVisibility
  switch len(existing) {
  case 0:
    visibility = keybase1.TLFVisibility_PRIVATE
    if arg.channel.Public {
      visibility = keybase1.TLFVisibility_PUBLIC
    }
    tt, err := TopicTypeFromStrDefault(arg.channel.TopicType)
    if err != nil {
      return nil, err
    }

    var topicName *string
    if arg.channel.TopicName != "" {
      topicName = &arg.channel.TopicName
    }
    channelName := arg.channel.Name
    ncres, err := client.NewConversationLocal(ctx, chat1.NewConversationLocalArg{
      TlfName:          channelName,
      TlfVisibility:    visibility,
      TopicName:        topicName,
      TopicType:        tt,
      IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_CLI,
      MembersType:      membersType,
    })
    if err != nil {
      return nil, err
    }
    header.rateLimits = append(header.rateLimits, ncres.RateLimits...)
    convTriple = ncres.Conv.Info.Triple
    tlfName = ncres.Conv.Info.TlfName
    visibility = ncres.Conv.Info.Visibility
    header.conversationID = ncres.Conv.Info.Id
  case 1:
    convTriple = existing[0].Info.Triple
    tlfName = existing[0].Info.TlfName
    visibility = existing[0].Info.Visibility
    header.conversationID = existing[0].Info.Id
  default:
    return nil, fmt.Errorf("multiple conversations matched")
  }
  var ephemeralMetadata *chat1.MsgEphemeralMetadata
  if arg.ephemeralLifetime != 0 && membersType != chat1.ConversationMembersType_KBFS {
    ephemeralLifetime := gregor1.ToDurationSec(arg.ephemeralLifetime)
    ephemeralMetadata = &chat1.MsgEphemeralMetadata{Lifetime: ephemeralLifetime}
  }

  header.clientHeader = chat1.MessageClientHeader{
    Conv:              convTriple,
    TlfName:           tlfName,
    TlfPublic:         visibility == keybase1.TLFVisibility_PUBLIC,
    MessageType:       arg.mtype,
    Supersedes:        arg.supersedes,
    Deletes:           arg.deletes,
    EphemeralMetadata: ephemeralMetadata,
  }

  return &header, nil
}
