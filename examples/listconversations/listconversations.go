package main


import (
  "../../../keybase-bot-api/kbapi"
  "fmt"
  "io"
  "github.com/keybase/client/go/protocol/chat1"
)

type TypeHolder struct {
	Type string `json:"type"`
}

type SubscriptionMessage struct {
	Message      chat1.MsgSummary
	Conversation chat1.ConvSummary
}

func main(){

    kb := kbapi.NewKbApi()
    testList := `{"method":"list", "params": { "options": { "unread_only": true}}}`
    t, err := kb.SendTeamApi(testList)
    if err != nil {
      fmt.Printf("%v\n", err)
    } else {
      fmt.Printf("%s\n", t)
    }
    testUserMembership := `{"method": "list-user-memberships", "params": {"options": {"username": "malwareunicorn"}}}`
    t, err := kb.SendTeamApi(testUserMembership)
    if err != nil {
      fmt.Printf("%v\n", err)
    } else {
      fmt.Printf("%s\n", t)
    }
    teststr4 := `{"method": "list", "params": {"options": {"team": "nacl_miners"}}}`
    t, err = kb.SendKvstoreApi(teststr4)
    if err != nil {
      fmt.Printf("%v\n", err)
    } else {
      fmt.Printf("%s\n", t)
    }


    // Test keybase chat api-listen
    pr, pw := io.Pipe()
    go kb.StartChatApiListener(pw)
    fmt.Printf("Starting Scanner\n")
    for {
      buff, _ := kb.ReadListener(pr)
      fmt.Printf("%s\n", buff)
    }
    pw.Close()
    pr.Close()
}
