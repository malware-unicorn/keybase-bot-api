package kbapi

import(
    "errors"
    "github.com/keybase/client/go/client"
    "github.com/keybase/client/go/protocol/keybase1"
    "github.com/keybase/client/go/libkb"
    "golang.org/x/net/context"
    "fmt"
)

const (
  getEntryMethod = "get"
  putEntryMethod = "put"
  listMethod     = "list"
  delEntryMethod = "del"
)

type getEntryOptions struct {
  Team      *string `json:"team,omitempty"`
  Namespace string  `json:"namespace"`
  EntryKey  string  `json:"entryKey"`
}

func (a *getEntryOptions) Check() error {
  if len(a.Namespace) == 0 {
    return errors.New("`namespace` field required")
  }
  if len(a.EntryKey) == 0 {
    return errors.New("`entryKey` field required")
  }
  return nil
}

type putEntryOptions struct {
  Team       *string `json:"team,omitempty"`
  Namespace  string  `json:"namespace"`
  EntryKey   string  `json:"entryKey"`
  Revision   *int    `json:"revision"`
  EntryValue string  `json:"entryValue"`
}

func (a *putEntryOptions) Check() error {
  if len(a.Namespace) == 0 {
    return errors.New("`namespace` field required")
  }
  if len(a.EntryKey) == 0 {
    return errors.New("`entryKey` field required")
  }
  if len(a.EntryValue) == 0 {
    return errors.New("`entryValue` field required")
  }
  if a.Revision != nil && *a.Revision <= 0 {
    return errors.New("if setting optional `revision` field, it needs to be a positive integer")
  }
  return nil
}

type deleteEntryOptions struct {
  Team      *string `json:"team,omitempty"`
  Namespace string  `json:"namespace"`
  EntryKey  string  `json:"entryKey"`
  Revision  *int    `json:"revision"`
}

func (a *deleteEntryOptions) Check() error {
  if len(a.Namespace) == 0 {
    return errors.New("`namespace` field required")
  }
  if len(a.EntryKey) == 0 {
    return errors.New("`entryKey` field required")
  }
  if a.Revision != nil && *a.Revision <= 0 {
    return errors.New("if setting optional `revision` field, it needs to be a positive integer")
  }
  return nil
}

type listOptions struct {
  Team      *string `json:"team,omitempty"`
  Namespace string  `json:"namespace"`
}

func (a *listOptions) Check() error {
  return nil
}


func getEntry(g *libkb.GlobalContext, ctx context.Context, opts getEntryOptions) Reply {
  config, err := client.GetConfigClient(g)
  if err != nil {
    return errReply(err)
  }
  status, err := config.GetCurrentStatus(context.Background(), 0)
  if err != nil {
    return errReply(err)
  }
  username := status.User.Username
  selfTeam := fmt.Sprintf("%s,%s", username, username)

  if opts.Team == nil {
      opts.Team = &selfTeam
  }

  kvstore, err := client.GetKVStoreClient(g)
  if err != nil {
    return errReply(err)
  }
  arg := keybase1.GetKVEntryArg{
    SessionID: 0,
    TeamName:  *opts.Team,
    Namespace: opts.Namespace,
    EntryKey:  opts.EntryKey,
  }
  res, err := kvstore.GetKVEntry(ctx, arg)
  if err != nil {
    return errReply(err)
  }
  return Reply{
    Result: res,
  }
}

func putEntry(g *libkb.GlobalContext, ctx context.Context, opts putEntryOptions) Reply {
  config, err := client.GetConfigClient(g)
  if err != nil {
    return errReply(err)
  }
  status, err := config.GetCurrentStatus(context.Background(), 0)
  if err != nil {
    return errReply(err)
  }
  username := status.User.Username
  selfTeam := fmt.Sprintf("%s,%s", username, username)

  if opts.Team == nil {
      opts.Team = &selfTeam
  }
  var revision int
  if opts.Revision != nil {
    revision = *opts.Revision
  }

  kvstore, err := client.GetKVStoreClient(g)
  if err != nil {
    return errReply(err)
  }
  arg := keybase1.PutKVEntryArg{
    SessionID:  0,
    TeamName:   *opts.Team,
    Namespace:  opts.Namespace,
    EntryKey:   opts.EntryKey,
    Revision:   revision,
    EntryValue: opts.EntryValue,
  }
  res, err := kvstore.PutKVEntry(ctx, arg)
  if err != nil {
    return errReply(err)
  }
  return Reply{
    Result: res,
  }
}

func deleteEntry(g *libkb.GlobalContext, ctx context.Context, opts deleteEntryOptions) Reply {
  config, err := client.GetConfigClient(g)
  if err != nil {
    return errReply(err)
  }
  status, err := config.GetCurrentStatus(context.Background(), 0)
  if err != nil {
    return errReply(err)
  }
  username := status.User.Username
  selfTeam := fmt.Sprintf("%s,%s", username, username)

  if opts.Team == nil {
      opts.Team = &selfTeam
  }
  var revision int
  if opts.Revision != nil {
    revision = *opts.Revision
  }

  kvstore, err := client.GetKVStoreClient(g)
  if err != nil {
    return errReply(err)
  }
  arg := keybase1.DelKVEntryArg{
    SessionID: 0,
    TeamName:  *opts.Team,
    Namespace: opts.Namespace,
    EntryKey:  opts.EntryKey,
    Revision:  revision,
  }
  res, err := kvstore.DelKVEntry(ctx, arg)
  if err != nil {
    return errReply(err)
  }
  return Reply{
    Result: res,
  }
}

func listEntries(g *libkb.GlobalContext, ctx context.Context, opts listOptions) Reply {
  config, err := client.GetConfigClient(g)
  if err != nil {
    return errReply(err)
  }
  status, err := config.GetCurrentStatus(context.Background(), 0)
  if err != nil {
    return errReply(err)
  }
  username := status.User.Username
  selfTeam := fmt.Sprintf("%s,%s", username, username)

  if opts.Team == nil {
      opts.Team = &selfTeam
  }

  kvstore, err := client.GetKVStoreClient(g)
  if err != nil {
    return errReply(err)
  }
  if len(opts.Namespace) == 0 {
    // listing namespaces
    arg := keybase1.ListKVNamespacesArg{
      SessionID: 0,
      TeamName:  *opts.Team,
    }
    res, err := kvstore.ListKVNamespaces(ctx, arg)
    if err != nil {
      return errReply(err)
    }
    return Reply{
      Result: res,
    }
  }
  arg := keybase1.ListKVEntriesArg{
    SessionID: 0,
    TeamName:  *opts.Team,
    Namespace: opts.Namespace,
  }
  res, err := kvstore.ListKVEntries(ctx, arg)
  if err != nil {
    return errReply(err)
  }
  return Reply{
    Result: res,
  }
}
