package main


import (
	"../../../keybase-bot-api/kbapi"
)


func main(){
    kbapi := kbapi.NewKbApi()
    kbapi.StartChatApi()
}
