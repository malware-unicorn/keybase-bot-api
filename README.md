# keybase-bot-api

This library uses the RPC system from keybase to communicate to the local keybase server for bots to use
```
git submodule init
git submodule update
```

dependencies in ./src:
```
go get -v github.com/araddon/dateparse
go get -v github.com/keybase/go-framed-msgpack-rpc/rpc
go get -v github.com/keybase/backoff
go get -v github.com/keybase/msgpackzip
go get -v github.com/keybase/go-codec
```

Example:
```
GOPATH=/home/rtvm/keybase-bot-api go build examples/listconversations/listconversations.go && ./listconversations
```

### Supported Keybase Chat Api methods
* list
* read
* get
* join
* leave
* send
* attach
* reaction
* edit
* listconvsoname
* advertisecommands
* clearcommands
* listcommands
* listmembers

### Supported Keybase Team Api methods
* list-user-memberships
* list-team-memberships

### Supported Keybase Kvstore Api methods
* put
* del
* get
* list
