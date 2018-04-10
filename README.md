## Prerequire
- Go version >= 1.9.2
- [Install Go](https://golang.org/dl/) follow by [installation instructions.](https://golang.org/doc/install)
- Tendermint >= 0.16.0
- [Install Tendermint](http://tendermint.readthedocs.io/projects/tools/en/master/index.html) follow by [installation instructions.](http://tendermint.readthedocs.io/projects/tools/en/master/install.html)

## Setup
1. `go get -u github.com/tendermint/abci/cmd/abci-cli`
1. `mkdir -p $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic`
1. `git clone git@repo.blockfint.com:digital-id/ndid-node-logic.git $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic`

## Run IdP node
1. open 2 terminal window
1. `cd $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic` and then `go run abci/server.go tcp://127.0.0.1:46000`
1. `cd $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic` and then `tendermint --home ./config/tendermint/IDP unsafe_reset_all && tendermint --home ./config/tendermint/IDP node --consensus.create_empty_blocks=false`

## Run RP node
1. open 2 terminal window
1. `cd $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic` and then `go run abci/server.go tcp://127.0.0.1:46001`
1. `cd $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic` and then `tendermint --home ./config/tendermint/RP unsafe_reset_all && tendermint --home ./config/tendermint/RP node --consensus.create_empty_blocks=false`

# Example
## RegisterMsqDestination
### Data
 ```sh
RegisterMsqDestination|{
  "users": [
    {
      "namespace": "cid",
      "id": "0123456789123"
    }
  ],
  "ip": "127.0.0.1",
  "port": "5000"
}|nonce1
 ```
 
### URI
 ```sh
curl -s 'localhost:45000/broadcast_tx_commit?tx="UmVnaXN0ZXJNc3FEZXN0aW5hdGlvbnx7DQogICJ1c2VycyI6IFsNCiAgICB7DQogICAgICAibmFtZXNwYWNlIjogImNpZCIsDQogICAgICAiaWQiOiAiMDEyMzQ1Njc4OTEyMyINCiAgICB9DQogIF0sDQogICJpcCI6ICIxMjcuMC4wLjEiLA0KICAicG9ydCI6ICI1MDAwIg0KfXxub25jZTE="'
 ```
 
### Result
 ```sh
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "check_tx": {
      "fee": {}
    },
    "deliver_tx": {
      "log": "success"
    },
    "hash": "450820781326A7CB7A8F5B05D10DE53212C6CDE7",
    "height": 3
  }
}
 ```
 
## GetMsqDestination
### Data
```sh
GetMsqDestination|{
  "namespace": "cid",
  "id": "0123456789123"
}
```
### URI
```sh
curl -s 'localhost:45000/abci_query?data="R2V0TXNxRGVzdGluYXRpb258ew0KICAibmFtZXNwYWNlIjogImNpZCIsDQogICJpZCI6ICIwMTIzNDU2Nzg5MTIzIg0KfQ=="'
```
### Result
```sh
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "response": {
      "value": "W3siaXAiOiIxMjcuMC4wLjEiLCJwb3J0IjoiNTAwMCJ9XQ=="
    }
  }
}
```

## CreateRequest
### Data
```sh
CreateRequest|{
  "requestId": "e3cb44c9-8848-4dec-98c8-8083f373b1f7",
  "minIdp": 1,
  "messageHash": "5727be55ab962ac4adcb3fb97c95ebad796132d95b2ed79b771fcbbe76dbfed5374713f71bfdbbf62f6815f119680b7e2355248fd67acfd5bf714ef17110a8c4"
}|nonce1
```
### URI
```sh
curl -s 'localhost:45000/broadcast_tx_commit?tx="Q3JlYXRlUmVxdWVzdHx7DQogICJyZXF1ZXN0SWQiOiAiZTNjYjQ0YzktODg0OC00ZGVjLTk4YzgtODA4M2YzNzNiMWY3IiwNCiAgIm1pbklkcCI6IDEsDQogICJtZXNzYWdlSGFzaCI6ICI1NzI3YmU1NWFiOTYyYWM0YWRjYjNmYjk3Yzk1ZWJhZDc5NjEzMmQ5NWIyZWQ3OWI3NzFmY2JiZTc2ZGJmZWQ1Mzc0NzEzZjcxYmZkYmJmNjJmNjgxNWYxMTk2ODBiN2UyMzU1MjQ4ZmQ2N2FjZmQ1YmY3MTRlZjE3MTEwYThjNCINCn18bm9uY2Ux"'
```
### Result
```sh
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "check_tx": {
      "fee": {}
    },
    "deliver_tx": {
      "log": "success"
    },
    "hash": "7DC457373CFD9FEB4D11AAEEAD7A0CBDCC2B0AF0",
    "height": 5
  }
}
```

## GetRequest
### Data
```sh
GetRequest|{
  "requestId": "e3cb44c9-8848-4dec-98c8-8083f373b1f7"
}
```
### URI
```sh
curl -s 'localhost:45000/abci_query?data="R2V0UmVxdWVzdHx7DQogICJyZXF1ZXN0SWQiOiAiZTNjYjQ0YzktODg0OC00ZGVjLTk4YzgtODA4M2YzNzNiMWY3Ig0KfQ=="'
```
### Result
```sh
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "response": {
      "value": "eyJzdGF0dXMiOiJwZW5kaW5nIiwibWVzc2FnZUhhc2giOiI1NzI3YmU1NWFiOTYyYWM0YWRjYjNmYjk3Yzk1ZWJhZDc5NjEzMmQ5NWIyZWQ3OWI3NzFmY2JiZTc2ZGJmZWQ1Mzc0NzEzZjcxYmZkYmJmNjJmNjgxNWYxMTk2ODBiN2UyMzU1MjQ4ZmQ2N2FjZmQ1YmY3MTRlZjE3MTEwYThjNCJ9"
    }
  }
}
```

## CreateIdpResponse
### Data
```sh
CreateIdpResponse|{
  "requestId": "e3cb44c9-8848-4dec-98c8-8083f373b1f7",
  "status": "accept",
  "signature": "TEyMyINCiAgICB9DQogIF0sDQogICJpcCI6ICIxM"
}|nonce1
```
### URI
```sh
curl -s 'localhost:45000/broadcast_tx_commit?tx="Q3JlYXRlSWRwUmVzcG9uc2V8ew0KICAicmVxdWVzdElkIjogImUzY2I0NGM5LTg4NDgtNGRlYy05OGM4LTgwODNmMzczYjFmNyIsDQogICJzdGF0dXMiOiAiYWNjZXB0IiwNCiAgInNpZ25hdHVyZSI6ICJURXlNeUlOQ2lBZ0lDQjlEUW9nSUYwc0RRb2dJQ0pwY0NJNklDSXhNIg0KfXxub25jZTE="'
```
### Result
```sh
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "check_tx": {
      "fee": {}
    },
    "deliver_tx": {
      "log": "success"
    },
    "hash": "9FB471CE265DB42BADA1FC6D3857A775C0877927",
    "height": 4
  }
}
```