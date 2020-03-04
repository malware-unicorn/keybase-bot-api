package kbapi

import(
    "github.com/keybase/client/go/externals"
    "github.com/keybase/client/go/libkb"
    "github.com/keybase/client/go/client"
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

      return nil, nil
    case methodSend:
      return nil, nil
    case methodEdit:
      return nil, nil
    case methodReaction:
      return nil, nil
    case methodAttach:
      return nil, nil
    case methodListConvsOnName:
      return nil, nil
    case methodJoin:
      return nil, nil
    case methodLeave:
      return nil, nil
    case methodAdvertiseCommands:
      return nil, nil
    case methodClearCommands:
      return nil, nil
    case methodListCommands:
      return nil, nil
    case methodListMembers:
      return nil, nil
    default:
      return nil, nil
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
