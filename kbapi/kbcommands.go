package kbapi

import(
    "io"
    "github.com/keybase/client/go/client"
    "golang.org/x/net/context"
    "github.com/keybase/client/go/protocol/keybase1"
    "github.com/keybase/client/go/libkb"
    "github.com/keybase/client/go/install"
    "github.com/keybase/go-framed-msgpack-rpc/rpc"
    "github.com/keybase/client/go/protocol/chat1"
    "runtime"
    "strconv"
    "time"
    "fmt"

)

func Logout(g *libkb.GlobalContext, force bool) (err error) {
  cli, err := client.GetLoginClient(g)
  if err != nil {
    return err
  }
  ctx := context.TODO()
  return cli.Logout(ctx, keybase1.LogoutArg{Force: force})
}

func LoginOneshot(g *libkb.GlobalContext, username string, paperkey string) (err error) {
  protocols := []rpc.Protocol{}
  if err := client.RegisterProtocolsWithContext(protocols, g); err != nil {
    return err
  }
  cclient, err := client.GetLoginClient(g)
  if err != nil {
    return err
  }

  err = cclient.LoginOneshot(context.Background(), keybase1.LoginOneshotArg{SessionID: 0, Username: username, PaperKey: paperkey})
  return err
}

func CtlStop(g *libkb.GlobalContext, shutdown bool) (err error) {
  mctx := libkb.NewMetaContextTODO(g)
  switch runtime.GOOS {
  case "windows":
    if !shutdown {
      install.StopAllButService(mctx, keybase1.ExitCode_OK)
    }
    cli, err := client.GetCtlClient(g)
    if err != nil {
      return err
    }
    return cli.StopService(mctx.Ctx(), keybase1.StopServiceArg{ExitCode: keybase1.ExitCode_OK})
  default:
    // On Linux, StopAllButService depends on a running service to tell it
    // what clients to shut down, so we can't call it directly from here,
    // but need to go through the RPC first.
    cli, err := client.GetCtlClient(g)
    if err != nil {
      return err
    }
    if shutdown {
      return cli.StopService(mctx.Ctx(), keybase1.StopServiceArg{ExitCode: keybase1.ExitCode_OK})
    }
    return cli.Stop(mctx.Ctx(), keybase1.StopArg{ExitCode: keybase1.ExitCode_OK})
  }
}

type ListenOptions struct {
  Wallet bool
  Convs  bool
}

func ChatApiListen(g *libkb.GlobalContext, opts ListenOptions, output io.Writer) (err error) {
  sessionClient, err := client.GetSessionClient(g)
  if err != nil {
    return err
  }
  chatConfig := chatNotificationConfig{
    showNewConvs:  opts.Convs,
    showLocal:     false,
    hideExploding: false,
  }
  // TODO: Set Up Wallet & Filters
  chatDisplay := newChatNotificationDisplay(g, chatConfig, output)
  //if err := chatDisplay.setupFilters(context.TODO(), c.channelFilters); err != nil {
	//	return err
	//}
  protocols := []rpc.Protocol{
    chat1.NotifyChatProtocol(chatDisplay),
  }
  if err := client.RegisterProtocolsWithContext(protocols, g); err != nil {
  return err
  }
  cli, err := client.GetNotifyCtlClient(g)
  if err != nil {
    return err
  }
  channels := keybase1.NotificationChannels{
    Chat:    true,
    Chatdev: false,
    Wallet:  false,
  }
  if err := cli.SetNotifications(context.TODO(), channels); err != nil {
    return err
  }
  for {
    if err := sendPing(sessionClient); err != nil {
      return fmt.Errorf("connection to service lost: error during ping: %v", err)
  }
  time.Sleep(time.Second)
 }
  return nil
}

func SetChatSettings(g *libkb.GlobalContext) (err error) {
  var settings chat1.GlobalAppNotificationSettings
  lcli, err := client.GetChatLocalClient(g)
  if err != nil {
    return err
  }
  settings.Settings = make(map[chat1.GlobalAppNotificationSetting]bool)
  settings.Settings[chat1.GlobalAppNotificationSetting_DISABLETYPING] = true
  strSettings := map[string]bool{}
  for setting, enabled := range settings.Settings {
    strSettings[strconv.FormatInt(int64(setting), 10)] = enabled
  }
  return lcli.SetGlobalAppNotificationSettingsLocal(context.TODO(), strSettings)
}
