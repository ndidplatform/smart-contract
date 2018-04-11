# Prerequisites
- Go version >= 1.9.2
  - [Install Go](https://golang.org/dl/) by following [installation instructions.](https://golang.org/doc/install)
  - Set GOPATH environment variable (https://github.com/golang/go/wiki/SettingGOPATH)
  
- Tendermint 0.16.0
  - [Install Tendermint](http://tendermint.readthedocs.io/projects/tools/en/v0.16.0/) by following [installation instructions.](http://tendermint.readthedocs.io/projects/tools/en/v0.16.0/install.html)  
  **Important**: After running `go get github.com/tendermint/tendermint/cmd/tendermint`, you need to change tendermint cloned source to version 0.16.0 before continuing the installation)
  
    ```
    cd $GOPATH/src/github.com/tendermint/tendermint
    git checkout v0.16.0
    ```

# Setup
1. Get dependency (tendermint ABCI)
    ```
    go get -u github.com/tendermint/abci/cmd/abci-cli
    ```

1. Create an directory for the project
    ```
    mkdir -p $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic
    ```

1. Clone the project
    ```
    git clone git@repo.blockfint.com:digital-id/ndid-node-logic.git $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic
    ```

## Run IdP node
1. Run ABCI server
    ```
    cd $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic

    go run abci/server.go tcp://127.0.0.1:46000
    ```

1. Run tendermint
    ```
    cd $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic

    tendermint --home ./config/tendermint/IdP unsafe_reset_all && tendermint --home ./config/tendermint/IdP node --consensus.create_empty_blocks=false
    ```

## Run RP node
1. Run ABCI server
    ```
    cd $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic

    CALLBACK_URI=http://localhost:3001/callback go run abci/server.go tcp://127.0.0.1:46001
    ```

1. Run tendermint
    ```
    cd $GOPATH/src/repo.blockfint.com/digital-id/ndid-node-logic

    tendermint --home ./config/tendermint/RP unsafe_reset_all && tendermint --home ./config/tendermint/RP node --consensus.create_empty_blocks=false
    ```

# Example
### Receive input in BASE64
## AddNodePublicKey
### Input
 ```sh
AddNodePublicKey|{
  "node_id": "IdP_f924-5069-4c6a-a4e4-134cd1a3d3d0",
  "public_key": "AAAAB3NzaC1yc2EAAAADAQABAAABAQC+RP+svJPfeâ€¦"
}|nonce1
 ```
 
### Expected Output
 ```sh
log: "success"
 ```

 ## AddNodePublicKey
### Input
 ```sh
AddNodePublicKey|{
  "node_id": "IdP_f924-5069-4c6a-a4e4-134cd1a3d3d0",
  "public_key": "AAAAB3NzaC1yc2EAAAADAQABAAABAQC+RP+svJPfeâ€¦"
}|nonce1
 ```
 
### Expected Output
 ```sh
log: "success"
 ```
