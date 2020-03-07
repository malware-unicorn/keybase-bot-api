GOBIN="$PWD/../go/bin/go"
export GOPATH="$PWD"

$GOBIN get -v github.com/araddon/dateparse
$GOBIN get -v github.com/keybase/go-framed-msgpack-rpc/rpc
$GOBIN get -v github.com/keybase/backoff
$GOBIN get -v github.com/keybase/msgpackzip
$GOBIN get -v github.com/keybase/go-codec
# Remove the usage of vendors to force our latest version:
rm -rf ./src/github.com/keybase/client/go/vendor/github.com/keybase/go-framed-msgpack-rpc
rm -rf ./src/github.com/keybase/client/go/vendor/github.com/keybase/backoff

