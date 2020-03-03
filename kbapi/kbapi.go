package kbapi

import(
    "github.com/keybase/client/go/externals"
    "github.com/keybase/client/go/libkb"
    "github.com/keybase/client/go/client"
)


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

func (kb*Kbapi) StartChatApi(){
	c := client.NewCmdChatAPIRunner(kb.g)
        c.Run()

}

func (kb*Kbapi) StopChatApi() {
}





