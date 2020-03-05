package main


import (
  "../../../keybase-bot-api/kbapi"
  "fmt"
)


func main(){
    kbapi := kbapi.NewKbApi()
    //teststr := `{"method":"list", "params": { "options": { "unread_only": true}}}`
    //teststr := `{"method": "send", "params":{"version": 1, "options": {"channel": {"name": "giteabottest,malwareunicorn"}}}}`
    //teststr := `{"method": "send", "params":{"version": 1, "options": {"conversation_id": "0000a37ea2048dbd087a13cb784631d63a46fb87f2ce13beadde229a7a0e85d5", "message": {"body": "hi this is Amanda's keybase-bot-api"}}}}`
    //teststr1 := `{"method": "read", "params":{"version": 1, "options": {"channel": {"name": "giteabottest,malwareunicorn"}}}}`
    //b, _ := kbapi.SendApi(teststr1)
    //fmt.Printf("%s\n", b)
    //teststr2 := `{"id": 30, "method": "edit", "params":{"version": 1, "options": {"channel": {"name": "giteabottest,malwareunicorn"}, "message_id": 148, "message": {"body": "BLAH"}}}}`
    //teststr2 := `{"id": 30, "method": "reaction", "params":{"version": 1, "options": {"channel": {"name": "giteabottest,malwareunicorn"}, "message_id": 148, "message": {"body": ":+1:"}}}}`
    //teststr2 := `{"method": "attach", "params":{"options": {"channel": {"name": "giteabottest,malwareunicorn"}, "filename": "/home/rtvm/vladIssue/local/keybase-bot-api/photo.png"}}}`
    //teststr2 := `{"method": "listconvsonname", "params":{"version": 1, "options": {"name": "giteabottest,malwareunicorn"}}}`

    //teststr2 := `{"method": "join", "params":{"version": 1, "options": {"conversation_id": "0000dcefee2030f004245c7fa401d7b31f9c1eb135070923110bf0549a1bb79f"}}}`
    //teststr2 := `{"method": "leave", "params":{"version": 1, "options": {"conversation_id": "0000dcefee2030f004245c7fa401d7b31f9c1eb135070923110bf0549a1bb79f"}}}`
    //teststr2 := `{"method": "advertisecommands", "params": {"options":{"alias": "giteabottest", "advertisements":[{"type": "public", "commands": [{"name": "malwareunicorn", "description": "This is just a test"}]}]}}}`
    //teststr2 := `{"method": "clearcommands"}`
    //teststr2 := `{"method": "listcommands", "params": {"options": {"channel": {"name": "giteabottest,malwareunicorn"}}}}`
    //teststr2 := `{"method": "listmembers", "params": {"options": {"conversation_id":"0000a37ea2048dbd087a13cb784631d63a46fb87f2ce13beadde229a7a0e85d5"}}}`
    //b, _ := kbapi.SendChatApi(teststr2)
    //fmt.Printf("%s\n", b)
    //teststr3 := `{"method": "list-team-memberships", "params": {"options": {"team": "nacl_miners"}}}`
    teststr3 := `{"method": "list-user-memberships", "params": {"options": {"username": "malwareunicorn"}}}`
    t, err := kbapi.SendTeamApi(teststr3)
    if err != nil {
      fmt.Printf("%v\n", err)
    } else {
      fmt.Printf("%s\n", t)
    }
    teststr4 := `{"method": "list", "params": {"options": {"team": "nacl_miners"}}}`
    t, err = kbapi.SendKvstoreApi(teststr4)
    if err != nil {
      fmt.Printf("%v\n", err)
    } else {
      fmt.Printf("%s\n", t)
    }
}
