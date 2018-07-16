[![CircleCI](https://circleci.com/gh/ndidplatform/smart-contract.svg?style=svg)](https://circleci.com/gh/ndidplatform/smart-contract)

# NDID Smart Contract

tendermint ABCI app

## Note

Test this app with command below

```sh
TENDERMINT_ADDRESS=http://localhost:45000 go test -v
```

## Add new validator (For testing)
get PubKey from pub_key.data in priv_validator.json 
```sh
curl -s 'localhost:45000/broadcast_tx_commit?tx="val:PubKey"'
```

## Prerequisites

* Go version >= 1.9.2

  * [Install Go](https://golang.org/dl/) by following [installation instructions.](https://golang.org/doc/install)
  * Set GOPATH environment variable (https://github.com/golang/go/wiki/SettingGOPATH)

* Tendermint 0.21.0

    ```sh
    mkdir -p $GOPATH/src/github.com/tendermint
    cd $GOPATH/src/github.com/tendermint
    git clone https://github.com/tendermint/tendermint.git
    cd tendermint
    git checkout v0.21.0
    make get_tools
    make get_vendor_deps
    make install
    ```

## Setup

1.  Create a directory for the project

    ```sh
    mkdir -p $GOPATH/src/github.com/ndidplatform/smart-contract
    ```

2.  Clone the project
    ```sh
    git clone https://github.com/ndidplatform/smart-contract.git $GOPATH/src/github.com/ndidplatform/smart-contract
    ```

3.  Get dependency (tendermint ABCI)

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract/abci
    dep ensure
    ```

### Run IdP node

1.  Run ABCI server

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract

    DB_NAME=IdP_DB go run abci/server.go tcp://127.0.0.1:46000
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

    DB_NAME=RP_DB go run abci/server.go tcp://127.0.0.1:46001
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

    DB_NAME=AS_DB go run abci/server.go tcp://127.0.0.1:46002
    ```

2.  Run tendermint

    ```sh
    cd $GOPATH/src/github.com/ndidplatform/smart-contract

    tendermint --home ./config/tendermint/AS unsafe_reset_all && tendermint --home ./config/tendermint/AS node --consensus.create_empty_blocks=false
    ```

## Run in Docker
Required
- Docker CE 17.06+ [Install docker](https://docs.docker.com/install/)
- docker-compose 1.14.0+ [Install docker-compose](https://docs.docker.com/compose/install/)

### Run

```
docker-compose -f docker/docker-compose.yml up
```

### Build

```
./docker/build.sh
```

### Note about docker

* To run docker container without building image, run command in **Run** section (no building required). It will run docker container with images from Dockerhub (https://hub.docker.com/r/ndidplatform/abci/ and https://hub.docker.com/r/ndidplatform/tendermint/). 
* To pull latest image from Dockerhub, run `docker pull ndidplatform/abci` and ``docker pull ndidplatform/tendermint``
    
## IMPORTANT NOTE

1.  You must start IDP, RP and AS nodes in order to run the platform.
2.  After starting BOTH nodes, please wait for

    ```
    Commit
    Commit
    ```

    to show in the first terminal (`go run abci ...`) of both processes before starting `api` processes.

3.  When IDP node and RP node run on separate machines, please edit `seeds` in `config/tendermint/{RP or IdP}/config/config.toml` to match address of another machines.

## Technical details to connect with `api`

# Broadcast tx format
```sh
functionName|base64(parameterJSON)|nonce|base64(sign(param+nonce))|base64(nodeID)
```

# Query format
```sh
functionName|base64(parameterJSON)
```

# Create transaction function

## InitNDID
### Parameter
```sh
{
  "node_id": "NDID",
  "public_key": "-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEA30i6deo6vqxPdoxA9pUpuBag/cVwEVWO8dds5QDfu/z957zxXUCY\nRxaiRWGAbOta4K5/7cxlsqI8fCvoSyAa/B7GTSc3vivK/GWUFP+sQ/Mj6C/fgw5p\nxK/+olBzfzLMDEOwFRbnYtPtbWozfvceq77fEReTUdBGRLak7twxLrRPNzIu/Gqv\nn5AR8urXyF4r143CgReGkXTTmOvHpHu98kCQSINFuwBB98RLFuWdVwkrHyzaGnym\nQu+0OR1Z+1MDIQ9WlViD1iaJhYKA6a0G0O4Nns6ISPYSh7W7fI31gWTgHUZN5iTk\nLb9t27DpW9G+DXryq+Pnl5c+z7es/7T34QIDAQAB\n-----END RSA PUBLIC KEY-----\n",
  "master_public_key": "-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEA30i6deo6vqxPdoxA9pUpuBag/cVwEVWO8dds5QDfu/z957zxXUCY\nRxaiRWGAbOta4K5/7cxlsqI8fCvoSyAa/B7GTSc3vivK/GWUFP+sQ/Mj6C/fgw5p\nxK/+olBzfzLMDEOwFRbnYtPtbWozfvceq77fEReTUdBGRLak7twxLrRPNzIu/Gqv\nn5AR8urXyF4r143CgReGkXTTmOvHpHu98kCQSINFuwBB98RLFuWdVwkrHyzaGnym\nQu+0OR1Z+1MDIQ9WlViD1iaJhYKA6a0G0O4Nns6ISPYSh7W7fI31gWTgHUZN5iTk\nLb9t27DpW9G+DXryq+Pnl5c+z7es/7T34QIDAQAB\n-----END RSA PUBLIC KEY-----\n"
}
```
### Expected Output
```sh
log: "success"
```

## RegisterNode
### Parameter
Posible role is RP,IdP and AS
```sh
{
  "node_id": "IdP1",
  "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n",
  "master_public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n",
  "node_name": "IdP Number 1 from ...",
  "role": "IdP",
  "max_ial": 3,
  "max_aal": 3
}
```
### Expected Output
```sh
log: "success"
```

## RegisterMsqDestination
### Parameter
```sh
{
  "users": [
    {
      "hash_id": "ece8921066562be07ba4ec44449646fc3b48d6b8a660a2e1e6a4bc7117edebba",
      "ial": 3,
      "first": true
    }
  ]
}
```
### Expected Output
```sh
log: "success"
```

## AddService
### Parameter
```sh
{
  "service_id": "statement",
  "service_name": "Bank statement"
}
```
### Expected Output
```sh
log: "success"
```

## DisableService
### Parameter
```sh
{
  "service_id": "statement"
}
```
### Expected Output
```sh
log: "success"
```

## RegisterServiceDestination
### Parameter
```sh
{
  "service_id": "statement",
  "min_ial": 1.1,
  "min_aal": 1.2
}
```
### Expected Output
```sh
log: "success"
```

## RegisterMsqAddress
### Parameter
```sh
{
  "node_id": "IdP1",
  "ip": "192.168.3.99",
  "port": 8000
}
```
### Expected Output
```sh
log: "success"
```

## CreateRequest
### Parameter
```sh
{
  "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
  "min_idp": 1,
  "min_aal": 3,
  "min_ial": 3,
  "timeout": 259200,
  "data_request_list": [
    {
      "service_id": "statement",
      "as_id_list": [
        "AS1",
        "AS2"
      ],
      "min_as": 1,
      "request_params_hash": "hash"
    }
  ],
  "request_message_hash": "hash('Please allow...')",
  "mode": 3
}
```
### Expected Output
```sh
log: "success"
```

## CreateIdpResponse
### Parameter
```sh
{
  "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
  "aal": 3,
  "ial": 3,
  "status": "accept",
  "signature": "signature",
  "identity_proof": "Magic"
}
```
### Expected Output
```sh
log: "success"
```

## SignData
### Parameter
```sh
{
  "service_id": "statement",
  "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
  "signature": "sign(data,asKey)"
}
```
### Expected Output
```sh
log: "success"
```

## SetNodeToken
### Parameter
```sh
{
  "node_id": "RP1",
  "amount": 100
}
```
### Expected Output
```sh
log: "success"
```

## AddNodeToken
### Parameter
```sh
{
  "node_id": "RP1",
  "amount": 111.11
}
```
### Expected Output
```sh
log: "success"
```

## ReduceNodeToken
### Parameter
```sh
{
  "node_id": "RP1",
  "amount": 61.11
}
```
### Expected Output
```sh
log: "success"
```

## SetPriceFunc
### Parameter
```sh
{
  "func": "CreateRequest",
  "price": 99.99
}
```
### Expected Output
```sh
log: "success"
```

## CloseRequest
### Parameter
```sh
{
  "requestId": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
  "response_valid_list": [
    {
      "idp_id": "IdP1",
      "valid_proof": true,
      "valid_ial": true
    }
  ]
}
```
### Expected Output
```sh
log: "success"
```

## TimeOutRequest
### Parameter
```sh
{
  "requestId": "ef6f4c9c-818b-42b8-8904-3d97c4c11111",
  "response_valid_list": [
    {
      "idp_id": "IdP1",
      "valid_proof": false,
      "valid_ial": false
    }
  ]
}
```
### Expected Output
```sh
log: "success"
```

## AddNamespace
### Parameter
```sh
{
  "namespace": "CID",
  "description": "Citizen ID"
}
```
### Expected Output
```sh
log: "success"
```

## DisableNamespace
### Parameter
```sh
{
  "namespace": "Tel"
}
```
### Expected Output
```sh
log: "success"
```

## UpdateNode
### Parameter
```sh
{
  "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n",
  "master_public_key": ""
}
```
### Expected Output
```sh
log: "success"
```

## CreateIdentity
### Parameter
```sh
{
  "accessor_id": "accessor_id",
  "accessor_type": "accessor_type",
  "accessor_public_key": "accessor_public_key",
  "accessor_group_id": "accessor_group_id"
}
```
### Expected Output
```sh
log: "success"
```

## AddAccessorMethod
### Parameter
```sh
{
  "accessor_id": "accessor_id_2",
  "accessor_type": "accessor_type_2",
  "accessor_public_key": "accessor_public_key_2",
  "accessor_group_id": "accessor_group_id",
  "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
}
```

## SetValidator
### Parameter
```sh
{
  "public_key": "5/6rEo7aQYq31J32higcxi3i8xp9MG/r5Ho5NemwZ+g=",
  "power": 0
}
```
### Expected Output
```sh
log: "success"
```

## SetDataReceived
### Parameter
```sh
{
  "requestId": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
  "service_id": "statement",
  "as_id": "AS1"
}
```
### Expected Output
```sh
log: "success"
```

## UpdateNodeByNDID
### Parameter
```sh
{
  "node_id": "IdP1",
  "max_ial": 2.3,
  "max_aal": 2.4
}
```
### Expected Output
```sh
log: "success"
```

## UpdateIdentity
### Parameter
```sh
{
  "hash_id": "ece8921066562be07ba4ec44449646fc3b48d6b8a660a2e1e6a4bc7117edebba",
  "ial": 2.2
}
```
### Expected Output
```sh
log: "success"
```

## DeclareIdentityProof
### Parameter
```sh
{
  "identity_proof": "Magic",
  "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
}
```
### Expected Output
```sh
log: "success"
```

## UpdateServiceDestination
### Parameter
```sh
{
  "service_id": "statement",
  "min_ial": 1.4,
  "min_aal": 1.5
}
```
### Expected Output
```sh
log: "success"
```

## UpdateService
### Parameter
```sh
{
  "service_id": "statement",
  "service_name": "Bank statement (ย้อนหลัง 3 เดือน)"
}
```
### Expected Output
```sh
log: "success"
```

## DisableMsqDestination
### Parameter
```sh
{
  "hash_id": "ece8921066562be07ba4ec44449646fc3b48d6b8a660a2e1e6a4bc7117edebba"
}
```
### Expected Output
```sh
log: "success"
```

## DisableAccessorMethod
### Parameter
```sh
{
  "accessor_id": "accessor_id"
}
```
### Expected Output
```sh
log: "success"
```

## RegisterServiceDestinationByNDID
### Parameter
```sh
{
  "service_id": "statement",
  "node_id": "AS2"
}
```
### Expected Output
```sh
log: "success"
```

## DisableNode
### Parameter
```sh
{
  "node_id": "IdP1"
}
```
### Expected Output
```sh
log: "success"
```

## DisableServiceDestinationByNDID
### Parameter
```sh
{
  "service_id": "BankStatement2",
  "node_id": "AS1"
}
```
### Expected Output
```sh
log: "success"
```

## EnableMsqDestination
### Parameter
```sh
{
  "hash_id": "ece8921066562be07ba4ec44449646fc3b48d6b8a660a2e1e6a4bc7117edebba"
}
```
### Expected Output
```sh
log: "success"
```

## EnableAccessorMethod
### Parameter
```sh
{
  "accessor_id": "accessor_id"
}
```
### Expected Output
```sh
log: "success"
```

## EnableNode
### Parameter
```sh
{
  "node_id": "IdP1"
}
```
### Expected Output
```sh
log: "success"
```

## EnableServiceDestinationByNDID
### Parameter
```sh
{
  "service_id": "BankStatement2",
  "node_id": "AS1"
}
```
### Expected Output
```sh
log: "success"
```

## EnableNamespace
### Parameter
```sh
{
  "namespace": "Tel"
}
```
### Expected Output
```sh
log: "success"
```

## EnableService
### Parameter
```sh
{
  "service_id": "statement"
}
```
### Expected Output
```sh
log: "success"
```

## DisableService
### Parameter
```sh
{
  "service_id": "statement"
}
```
### Expected Output
```sh
log: "success"
```

## DisableServiceDestination
### Parameter
```sh
{
  "service_id": "statement"
}
```
### Expected Output
```sh
log: "success"
```

## EnableServiceDestination
### Parameter
```sh
{
  "service_id": "statement"
}
```
### Expected Output
```sh
log: "success"
```

# Query function

## GetNodeMasterPublicKey
### Parameter
```sh
{
  "node_id": "RP1"
}
```
### Expected Output
```sh
{
  "master_public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1QXXrV7X1b8uFL1PW7+F\nimlAwxwbEMG5hFru1CN8WsRt8ZVQIkXRpiwNNXh1GS0Qmshnv8pKaNCZ5q5wFdUe\nlYspZHVRbIkHiQAaEU5yG9SyavHsDntUOd50PQ3nC71feW+ff8tvQcJ7+gqf8nZ6\nUAWpG4bvakPtrJ81h4/Qc23vhtbcouP0adgdw6UA0kcdGhTESYMBU0dx/NNysvJh\nNx36z2UU6kbQ3a2/bINEZAgLfJ7/Y+/647+tc7bUYdqj3dNkbnk1xiXh5dTLsiow\n5Xvukpy2uA44M/r2Q5VRfbH2ZrBZlgf/XEOZs7zppySgaTWRB5eDTm+YxxyOyykn\n8wIDAQAB\n-----END PUBLIC KEY-----\n"
}
```

## GetNodePublicKey
### Parameter
```sh
{
  "node_id": "RP1"
}
```
### Expected Output
```sh
{
  "public_key": "-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEAwCB4UBzQcnd6GAzPgbt9j2idW23qKZrsvldPNifmOPLfLlMusv4E\ncyJf4L42/aQbTn1rVSu1blGkuCK+oRlKWmZEWh3xv9qrwCwov9Jme/KOE98zOMB1\n0/xwnYotPadV0de80wGvKT7OlBlGulQRRhhgENNCPSxdUlozrPhrzGstXDr9zTYQ\noR3UD/7Ntmew3mnXvKj/8+U48hw913Xn6btBP3Uqg2OurXDGdrWciWgIMDEGyk65\nNOc8FOGa4AjYXzyi9TqOIfmysWhzKzU+fLysZQo10DfznnQN3w9+pI+20j2zB6gg\npL75RjZKYgHU49pbvjF/eOSTOg9o5HwX0wIDAQAB\n-----END RSA PUBLIC KEY-----\n"
}
```

## GetIdpNodes
### Parameter
```sh
{
  "hash_id": "ece8921066562be07ba4ec44449646fc3b48d6b8a660a2e1e6a4bc7117edebba",
  "min_ial": 3,
  "min_aal": 3,
}
```
### Expected Output
```sh
{
  "node": [
    {
      "node_id": "IdP1",
      "name": "IdP Number 1 from ...",
      "max_ial": 3,
      "max_aal": 3
    },
    {
      "node_id": "IdP2",
      "name": "",
      "max_ial": 3,
      "max_aal": 3
    }
  ]
}
```

## GetAsNodesByServiceId
### Parameter
```sh
{
  "service_id": "statement"
}
```
### Expected Output
```sh
{
  "node": [
    {
      "node_id": "AS1",
      "name": "AS1",
      "min_ial": 1.1,
      "min_aal": 1.2
    }
  ]
}
```

## GetMsqAddress
### Parameter
```sh
{
  "node_id": "IdP1"
}
```
### Expected Output
```sh
{
  "ip": "192.168.3.99",
  "port": 8000
}
```

## GetRequest
### Parameter
```sh
{
  "requestId": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
}
```
### Expected Output
```sh
{
  "closed": true,
  "timed_out": false,
  "request_message_hash": "hash('Please allow...')",
  "mode": 3
}
```

## GetRequestDetail
### Parameter
```sh
{
  "requestId": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
}
```
### Expected Output
```sh
{
  "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
  "min_idp": 1,
  "min_aal": 3,
  "min_ial": 3,
  "request_timeout": 259200,
  "data_request_list": [
    {
      "service_id": "statement",
      "as_id_list": [],
      "min_as": 1,
      "request_params_hash": "hash",
      "answered_as_id_list": [
        "AS1"
      ],
      "received_data_from_list": [
        "AS1"
      ]
    }
  ],
  "request_message_hash": "hash('Please allow...')",
  "response_list": [
    {
      "ial": 3,
      "aal": 3,
      "status": "accept",
      "signature": "signature",
      "identity_proof": "Magic",
      "private_proof_hash": "Magic",
      "idp_id": "IdP1",
      "valid_proof": true,
      "valid_ial": true
    }
  ],
  "closed": true,
  "timed_out": false,
  "special": false,
  "mode": 3
}
```

## GetNodeToken
### Parameter
```sh
{
  "node_id": "RP1"
}
```
### Expected Output
```sh
{
  "amount": 100
}
```

## GetPriceFunc
### Parameter
```sh
{
  "func": "CreateRequest"
}
```
### Expected Output
```sh
{
  "price": 99.99
}
```

## GetUsedTokenReport
### Parameter
```sh
{
  "node_id": "AS1"
}
```
### Expected Output
```sh
[
  {
    "method": "RegisterServiceDestination",
    "price": 1,
    "data": ""
  },
  {
    "method": "SignData",
    "price": 1,
    "data": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
  }
]
```

## GetServiceDetail
### Parameter
```sh
{
  "service_id": "statement"
}
```
### Expected Output
```sh
{
  "service_name": "Bank statement",
  "active": true
}
```

## GetNamespaceList
### Parameter
```sh
```
### Expected Output
```sh
[
  {
    "namespace": "CID",
    "description": "Citizen ID",
    "active": true
  }
]
```

## CheckExistingIdentity
### Parameter
```sh
{
   "hash_id": "hash(ns+id)"
}
```
### Expected Output
```sh
{
  "exist": true
}
```

## GetAccessorGroupID
### Parameter
```sh
{
  "accessor_id": "accessor_id_001"
}
```
### Expected Output
```sh
{
  "accessor_group_id": "accessor_group_id"
}
```

## GetAccessorKey
### Parameter
```sh
{
  "accessor_id": "accessor_id_001"
}
```
### Expected Output
```sh
{
  "accessor_public_key": "accessor_public_key",
  "active": true
}
```

## GetServiceList
### Parameter
```sh
```
### Expected Output
```sh
[
  {
    "service_id": "statement",
    "service_name": "Bank statement",
    "active": true
  }
]
```

## GetNodeInfo
### Parameter
```sh
{
  "node_id": "IdP1"
}
```
### Expected Output
```sh
{
  "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n",
  "master_public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n",
  "node_name": "IdP Number 1 from ...",
  "role": "IdP",
  "max_ial": 3,
  "max_aal": 3
}
```

## CheckExistingAccessorID
### Parameter
```sh
{
  "accessor_id": "accessor_id"
}
```
### Expected Output
```sh
{
  "exist": true
}
```

## CheckExistingAccessorGroupID
### Parameter
```sh
{
  "accessor_group_id": "accessor_group_id"
}
```
### Expected Output
```sh
{
  "exist": true
}
```

## GetIdentityInfo
### Parameter
```sh
{
  "hash_id": "ece8921066562be07ba4ec44449646fc3b48d6b8a660a2e1e6a4bc7117edebba",
  "node_id": "IdP1"
}
```
### Expected Output
```sh
{
  "ial": 3
}
```

## GetDataSignature
### Parameter
```sh
{
  "node_id": "AS1",
  "service_id": "statement",
  "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
}
```
### Expected Output
```sh
{
  "signature": "sign(data,asKey)"
}
```

## GetIdentityProof
### Parameter
```sh
{
  "idp_id": "IdP1",
  "request_id": "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
}
```
### Expected Output
```sh
{
  "identity_proof": "Magic"
}
```

## GetServicesByAsID
### Parameter
```sh
{
  "as_id": "AS1"
}
```
### Expected Output
```sh
{
  "services": [
    {
      "service_id": "BankStatement2",
      "min_ial": 2.2,
      "min_aal": 2.2,
      "active": true,
      "suspended": false
    },
    {
      "service_id": "BankStatement3",
      "min_ial": 3.3,
      "min_aal": 3.3,
      "active": true,
      "suspended": true
    }
  ]
}
```

