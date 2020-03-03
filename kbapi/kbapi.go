package kbapi

import(
    "os"
    "os/signal"
    "github.com/keybase/client/go/externals"
    "github.com/keybase/client/go/libkb"
    "github.com/keybase/client/go/client"
    "github.com/keybase/client/go/logger"
    "github.com/keybase/client/go/libcmdline"
    keybase1 "github.com/keybase/client/go/protocol/keybase1"
    "syscall"
    "fmt"
)


type Kbapi struct {
    GlobalContext *libkb.GlobalContext
}

func (kb*Kbapi) Init() {
}

func (kb*Kbapi) StartChatApi(stdin *os.Pipe, stdout *os.Pipe){
	g := externals.NewGlobalContextInit()
	usage := libkb.Usage{
		API:       true,
		KbKeyring: true,
		Config:    true,
		//Socket:    true,
	}
        g.Env.Test.UseProductionRunMode = true
	g.ConfigureUsage(usage);
	c := client.NewCmdChatAPIRunner(g)
        c.Run()

}

func (kb*Kbapi) StopChatApi() {
}





