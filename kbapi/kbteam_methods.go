package kbapi

import(
  "github.com/keybase/client/go/client"
  "github.com/keybase/client/go/libkb"
  "github.com/keybase/client/go/protocol/keybase1"
  "context"
)

const (
  // team api
  listTeamMethod      = "list-team-memberships"
  listUserMethod      = "list-user-memberships"
)

type listTeamOptions struct {
  Team      string `json:"team"`
  ForcePoll bool   `json:"force-poll"`
}

type listUserOptions struct {
  UserAssertion        string `json:"username"`
  IncludeImplicitTeams bool   `json:"include-implicit-teams"`
}

func listTeamMemberships(g *libkb.GlobalContext, ctx context.Context, opts listTeamOptions) Reply {
  cli, err := client.GetTeamsClient(g)
  arg := keybase1.TeamGetArg{
    Name: opts.Team,
  }
  details, err := cli.TeamGet(ctx, arg)
  if err != nil {
    return errReply(err)
  }
  return Reply{
    Result: details,
  }

}

func listUserMemberships(g *libkb.GlobalContext, ctx context.Context, opts listUserOptions) Reply {
  cli, err := client.GetTeamsClient(g)
  arg := keybase1.TeamListUnverifiedArg{
    UserAssertion:        opts.UserAssertion,
    IncludeImplicitTeams: opts.IncludeImplicitTeams,
  }
  list, err := cli.TeamListUnverified(context.Background(), arg)
  if err != nil {
    return errReply(err)
  }
  return Reply{
    Result: list,
  }
}
