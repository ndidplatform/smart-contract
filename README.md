## Prerequire
- Go version >= 1.9.2
- [Install Go](https://golang.org/dl/) follow by [installation instructions.](https://golang.org/doc/install)
- Tendermint >= 0.16.0
- [Install Tendermint](http://tendermint.readthedocs.io/projects/tools/en/master/index.html) follow by [installation instructions.](http://tendermint.readthedocs.io/projects/tools/en/master/install.html)
- [Go dependency management tool](https://github.com/golang/dep)

## Run IdP node
1. open 4 terminal window
1. `cd $GOPATH/src/github.com/ndidplatform/ndid` and then `go run abci/server.go tcp://127.0.0.1:46000`
1. `tendermint --home ./config/tendermint/IDP unsafe_reset_all && tendermint --home ./config/tendermint/IDP node --consensus.create_empty_blocks=false`

## Run RP node
1. open 4 terminal window
1. `cd $GOPATH/src/github.com/ndidplatform/ndid` and then `go run abci/server.go tcp://127.0.0.1:46001`
1. `tendermint --home ./config/tendermint/RP unsafe_reset_all && tendermint --home ./config/tendermint/RP node --consensus.create_empty_blocks=false`