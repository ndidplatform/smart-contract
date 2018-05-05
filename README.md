# NDID Smart Contract

tendermint ABCI app

## Note

The repo was renamed at 25 Apr 2018 and local go-path was changed from digital-id to ndidplatform.
For anyone who cloned this project before 25 Apr 2018, do as follows:

```sh
mkdir -p $GOPATH/src/github.com/ndidplatform
mv $GOPATH/src/github.com/digital-id/ndid-smart-contract $GOPATH/src/github.com/ndidplatform/smart-contract
```

## Prerequisites

* Go version >= 1.9.2

  * [Install Go](https://golang.org/dl/) by following [installation instructions.](https://golang.org/doc/install)
  * Set GOPATH environment variable (https://github.com/golang/go/wiki/SettingGOPATH)

* Tendermint 0.16.0

  * [Install Tendermint](http://tendermint.readthedocs.io/projects/tools/en/v0.16.0/) by following [installation instructions.](http://tendermint.readthedocs.io/projects/tools/en/v0.16.0/install.html)  
    **Important**: After running `go get github.com/tendermint/tendermint/cmd/tendermint`, you need to change tendermint cloned source to version 0.16.0 before continuing the installation)

    ```sh
    cd $GOPATH/src/github.com/tendermint/tendermint
    git checkout v0.16.0
    ```

## Setup

1.  Get dependency (tendermint ABCI)

    ```sh
    go get -u github.com/tendermint/abci/cmd/abci-cli
    ```

2.  Create a directory for the project

    ```sh
    mkdir -p $GOPATH/src/github.com/ndidplatform/smart-contract
    ```

3.  Clone the project
    ```sh
    git clone https://github.com/ndidplatform/smart-contract.git $GOPATH/src/github.com/ndidplatform/smart-contract
    ```

### Run IdP node

1.  Run ABCI server

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract

    CALLBACK_URI=http://{$ndid-api-address}:{ndid-api-callback-port}{ndid-api-callback-path} go run abci/server.go tcp://127.0.0.1:46000
    ```

    Example

    ```sh
    CALLBACK_URI=http://localhost:3001/callback go run abci/server.go tcp://127.0.0.1:46000
    ```

2.  Run tendermint

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract

    tendermint --home ./config/tendermint/IdP unsafe_reset_all && tendermint --home ./config/tendermint/IdP node --consensus.create_empty_blocks=false
    ```

### Run RP node

1.  Run ABCI server

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract

    CALLBACK_URI=http://{$ndid-api-address}:{ndid-api-callback-port}{ndid-api-callback-path} go run abci/server.go tcp://127.0.0.1:46001
    ```

    Example

    ```sh
    CALLBACK_URI=http://localhost:3001/callback go run abci/server.go tcp://127.0.0.1:46001
    ```

2.  Run tendermint

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract

    tendermint --home ./config/tendermint/RP unsafe_reset_all && tendermint --home ./config/tendermint/RP node --consensus.create_empty_blocks=false
    ```
    
### Run AS node

1.  Run ABCI server

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract

    CALLBACK_URI=http://{$ndid-api-address}:{ndid-api-callback-port}{ndid-api-callback-path} go run abci/server.go tcp://127.0.0.1:46001
    ```

    Example

    ```sh
    CALLBACK_URI=http://localhost:3001/callback go run abci/server.go tcp://127.0.0.1:46002
    ```

2.  Run tendermint

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract

    tendermint --home ./config/tendermint/AS unsafe_reset_all && tendermint --home ./config/tendermint/AS node --consensus.create_empty_blocks=false
    ```

## Run in Docker
Required
- Docker CE [Install docker](https://docs.docker.com/install/)
- docker-compose [Install docker-compose](https://docs.docker.com/compose/install/)

```
docker network create ndidplatform
docker-compose up
```
    
## IMPORTANT NOTE

1.  You must start IDP, RP and AS nodes in order to run the platform.
2.  After starting BOTH nodes, please wait for

    ```
    Commit
    Commit
    ```

    to show in the first terminal (`go run abci ...`) of both processes before starting `ndid-api` processes.

3.  When IDP node and RP node run on separate machines, please edit `seeds` in `config/tendermint/{RP or IdP}/config/config.toml` to match address of another machines.

## Technical details to connect with `ndid-api`

Interact with `ndid-api` in BASE64 format data.

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

## GetNodePublicKey

### Input

```sh
GetNodePublicKey|{
 "node_id": "IdP_f924-5069-4c6a-a4e4-134cd1a3d3d0"
}
```

### Expected Output

```sh
{
 "public_key": "AAAAB3NzaC1yc2EAAAADAQABAAABAQC+RP+svJPfeâ€¦"
}
```

## RegisterMsqDestination

### Input

```sh
RegisterMsqDestination|{
 "users": [
   {
      "hash_id": "fc7ba91796fa25bad3c94aa9782266cabee3a933edbdfe2d46cb393ace89de1f",
      "ial": 1
   }
 ],
 "node_id": "IdP_f924-5069-4c6a-a4e4-134cd1a3d3d0"
}|nonce1
```

### Expected Output

```sh
log: "success"
```

## GetMsqDestination

### Input

```sh
GetMsqDestination|{
 "hash_id": "fc7ba91796fa25bad3c94aa9782266cabee3a933edbdfe2d46cb393ace89de1f",
 "min_ial": 1
}
```

### Expected Output

```sh
{
 "node_id": ["IdP_f924-5069-4c6a-a4e4-134cd1a3d3d0",...]
}
```

## AddAccessorMethod

### Input

```sh
AddAccessorMethod|{
 "accessor_id":"acc_f328-53da-4d51-a927-3cc6d3ed3feb",
 "accessor_type":"RSA-2048",
 "accessor_key":"AAAAB3NzaC1yc2EAAAADAQABAAABâ€¦",
 "commitment":"(magic)"
}|nonce1
```

### Expected Output

```sh
log: "success"
```

## GetAccessorMethod

### Input

```sh
GetAccessorMethod|{
 "accessor_id":"acc_f328-53da-4d51-a927-3cc6d3ed3feb"
}
```

### Expected Output

```sh
{
 "accessor_type":"RSA-2048",
 "accessor_key":"AAAAB3NzaC1yc2EAAAADAQABAAABâ€¦",
 "commitment":"(magic)"
}
```

## CreateRequest

### Input

```sh
CreateRequest|{
 "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
 "min_idp": 1,
 "min_aal": 1,
 "min_ial": 2,
 "timeout": 259200,
 "data_request_list": [
   {
     "service_id": "bank_statement",
     "as": [
       "AS1",
       "AS2"
     ],
     "count": "1",
     "request_params": {
       "format": "pdf",
       "language": "en"
     }
   }
 ],
 "message_hash": "hash('Please allow...')"
}
```

### Expected Output

```sh
log: "success"
```

## GetRequest

### Input

```sh
GetRequest|{
 "requestId": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
}
```

### Expected Output

```sh
{
 "status": "complete",
 "messageHash" : "hash('Please allow...')"
}
```

## CreatIdpResponse

### Input

```sh
CreatIdpResponse|{
 "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
 "aal": 3,
 "ial": 2,
 "status": "accept",
 "signature": "(signature)",
 "accessor_id": "12a8f328-53da-4d51-a927-3cc6d3ed3feb",
 "identity_proof": "(identity_proof)"
}|nonce1
```

### Expected Output

```sh
log: "success"
```

## GetRequestDetail

### Input

```sh
GetRequestDetail|{
 "requestId": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
}
```

### Expected Output

```sh
{
 "status": "complete",
 "messageHash" : "hash('Please allow...')"
}
```
