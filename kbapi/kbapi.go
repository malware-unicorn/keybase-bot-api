package kbapi

import(
    "github.com/keybase/client/go/externals"
    "github.com/keybase/client/go/libkb"
    "github.com/keybase/client/go/client"
    "github.com/keybase/client/go/chat/utils"
    "context"
    "fmt"
    "encoding/json"
    "strings"
    "io"
    "bytes"
    "errors"
)

// CallError is the result when there is an error.
type CallError struct {
  Code    int         `json:"code"`
  Message string      `json:"message"`
  Data    interface{} `json:"data,omitempty"`
}

// Reply is returned with the results of processing a Call.
type Reply struct {
  Jsonrpc string      `json:"jsonrpc,omitempty"`
  ID      int         `json:"id,omitempty"`
  Error   *CallError  `json:"error,omitempty"`
  Result  interface{} `json:"result,omitempty"`
}

type Call struct {
  Jsonrpc string
  ID      int
  Method  string
  Params  Params
}
// Params represents the `params` portion of the JSON api call.
type Params struct {
  Version int
  Options json.RawMessage
}

type ErrInvalidOptions struct {
  method  string
  version int
  err     error
}

func (e ErrInvalidOptions) Error() string {
  return fmt.Sprintf("invalid %s v%d options: %s", e.method, e.version, e.err)
}

type ErrInvalidMethod struct {
	name    string
	version int
}

func (e ErrInvalidMethod) Error() string {
	return fmt.Sprintf("invalid v%d method %q", e.version, e.name)
}

type Kbapi struct {
  g *libkb.GlobalContext
}

func NewKbApi() *Kbapi {
  g := externals.NewGlobalContextInit()
  kb := Kbapi{ g: g }
  kb.g.Env.Test.UseProductionRunMode = true
  usage := libkb.Usage{
          API:       true,
          KbKeyring: true,
          Config:    true,
          //Socket:    true,
  }
  kb.g.ConfigureUsage(usage);
  return &kb
}

// Get Current Keybase Username.
func (kb*Kbapi) GetUsername() string {
  return kb.g.Env.GetUsername().String()
}

func (kb*Kbapi) StartChatApi(){
  c := client.NewCmdChatAPIRunner(kb.g)
        c.Run()
}

func (kb*Kbapi) SendApi(apiInput string) (b []byte, err error) {
  var call Call
  dec := json.NewDecoder(strings.NewReader(apiInput))
  for {
    if err := dec.Decode(&call); err == io.EOF {
      break
    } else if err != nil {
      if err == io.ErrUnexpectedEOF {
        fmt.Printf("expected more JSON in input\n")
        return nil, err
      }
      return nil, err
    }
    fmt.Printf("Method: %s\n", call.Method)
    switch call.Method {
    case methodList:
      var opts listOptionsV1
      if len(call.Params.Options) != 0 {
        if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
          return nil, err
        }
      }
      reply := ListV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodRead:
      if len(call.Params.Options) == 0 {
          return nil, ErrInvalidOptions{version: 1, method: methodRead, err: errors.New("empty options")}
      }
      var opts readOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
          return nil, err
      }
      reply := ReadV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodGet:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodRead, err: errors.New("empty options")}
      }
      var opts getOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := GetV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodSend:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodSend, err: errors.New("empty options")}
      }
      var opts sendOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      // opts are valid for send v1
      chatUI := NewChatAPIUI(AllowStellarPayments(opts.ConfirmLumenSend))
      reply := SendV1(kb.g, context.Background(), opts, chatUI)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodEdit:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodEdit, err: errors.New("empty options")}
      }
      var opts editOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := EditV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodReaction:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodReaction, err: errors.New("empty options")}
      }
      var opts reactionOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := ReactionV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodAttach:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodAttach, err: errors.New("empty options")}
      }
      var opts attachOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      chatUI := NewChatAPIUI()
      reply := AttachV1(kb.g, context.Background(), opts, chatUI, utils.DummyChatNotifications{})
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodListConvsOnName:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodListConvsOnName, err: errors.New("empty options")}
      }
      var opts listConvsOnNameOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := ListConvsOnNameV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodJoin:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodJoin, err: errors.New("empty options")}
      }
      var opts joinOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := JoinV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodLeave:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodLeave, err: errors.New("empty options")}
      }
      var opts leaveOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := LeaveV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodAdvertiseCommands:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodAdvertiseCommands,
          err: errors.New("empty options")}
      }
      var opts advertiseCommandsOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := AdvertiseCommandsV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodClearCommands:
      reply := ClearCommandsV1(kb.g, context.Background())
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodListCommands:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodListCommands,
          err: errors.New("empty options")}
      }
      var opts listCommandsOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := ListCommandsV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    case methodListMembers:
      if len(call.Params.Options) == 0 {
        return nil, ErrInvalidOptions{version: 1, method: methodListMembers, err: errors.New("empty options")}
      }
      var opts listMembersOptionsV1
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return nil, err
      }
      reply := ListMembersV1(kb.g, context.Background(), opts)
      reply.Jsonrpc = call.Jsonrpc
      reply.ID = call.ID
      b := new(bytes.Buffer)
      enc := json.NewEncoder(b)
      enc.Encode(reply)
      return b.Bytes(), nil
    default:
      return nil, ErrInvalidMethod{name: call.Method, version: 1}
    }
  }
  return nil, nil
}

func (kb*Kbapi) Test() {
  var err error
  teststr := `{"method":"list", "params": { "options": { "unread_only": true}}}`
  //teststr := `{"method": "listconvsonname", "params": {"options": {"topic_type": "CHAT", "members_type": "team", "name": "nacl_miners"}}}`
  dec := json.NewDecoder(strings.NewReader(teststr))
  var call Call
  defer func() {
    if err != nil {
      fmt.Printf("%v\n", err)
      //err = encodeErr(call, err, w, false)
    }
  }()
  for {
    if err := dec.Decode(&call); err == io.EOF {
      break
    } else if err != nil {
      if err == io.ErrUnexpectedEOF {
        //return ErrInvalidJSON{message: "expected more JSON in input"}
        fmt.Printf("expected more JSON in input\n")
      }
      return
    }
    fmt.Printf("Method: %s\n", call.Method)
    var opts listOptionsV1
    if len(call.Params.Options) != 0 {
      if err := json.Unmarshal(call.Params.Options, &opts); err != nil {
        return
      }
    }
    reply := ListV1(kb.g, context.Background(), opts)
    reply.Jsonrpc = call.Jsonrpc
    reply.ID = call.ID
    b := new(bytes.Buffer)
    enc := json.NewEncoder(b)
    enc.SetIndent("", "    ")
    enc.Encode(reply)
    fmt.Printf("%v\n", b)
  }
}
