package main


import (
  "../../../keybase-bot-api/kbapi"
  "fmt"
)


func main(){
    kbapi := kbapi.NewKbApi()
    teststr := `{"method":"list", "params": { "options": { "unread_only": true}}}`
    b, _ := kbapi.SendApi(teststr)
    fmt.Printf("%s\n", b)
}
