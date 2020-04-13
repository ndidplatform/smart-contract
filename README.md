[![CircleCI](https://circleci.com/gh/ndidplatform/smart-contract.svg?style=svg)](https://circleci.com/gh/ndidplatform/smart-contract)

# NDID Smart Contract

Tendermint bundled with ABCI app

## Prerequisites

- Go version >= 1.13.0

  - [Install Go](https://golang.org/dl/) by following [installation instructions.](https://golang.org/doc/install)
  - Set GOPATH environment variable (https://github.com/golang/go/wiki/SettingGOPATH)

- (Optional) LevelDB version >= 1.7 and snappy

  - Ubuntu (Ref: https://tendermint.com/docs/introduction/install.html#compile-with-cleveldb-support)

    ```sh
    sudo apt-get update
    sudo apt install build-essential

    sudo apt-get install libsnappy-dev

    wget https://github.com/google/leveldb/archive/v1.20.tar.gz && \
      tar -zxvf v1.20.tar.gz && \
      cd leveldb-1.20/ && \
      make && \
      sudo cp -r out-static/lib* out-shared/lib* /usr/local/lib/ && \
      cd include/ && \
      sudo cp -r leveldb /usr/local/include/ && \
      sudo ldconfig && \
      rm -f v1.20.tar.gz
    ```

  - macOS (Homebrew)

    ```sh
    brew install snappy
    brew install leveldb
    ```

## Setup

1.  Clone the project

    ```sh
    git clone https://github.com/ndidplatform/smart-contract.git
    ```

2.  Build or run

    ```sh
    cd smart-contract

    # Build (See Build section in README.md)

    # Run
    ./did-tendermint
    ```

<!-- Cannot be done with Go module? -->
<!-- 3.  (Optional) Patch Tendermint LevelDB adapters

    ```sh
    git apply $GOPATH/src/github.com/ndidplatform/smart-contract/patches/tm_goleveldb_bloom_filter.patch && \
    git apply $GOPATH/src/github.com/ndidplatform/smart-contract/patches/tm_cleveldb_cache_and_bloom_filter.patch
    ``` -->

**Environment variable options**

- `ABCI_DB_DIR_PATH`: Directory path for ABCI app persistence data files [Default: `./DID`]
- `ABCI_DB_TYPE`: Database type (same options as Tendermint's `db_backend`) [Default: `cleveldb`]
- `ABCI_LOG_LEVEL`: Log level. Allowed values are `error`, `warn`, `info` and `debug` [Default: `debug`]
- `ABCI_LOG_TARGET`: Where should logger writes logs to. Allowed values are `console` or `file` (eg. `ABCI.log`) [Default: `console`]
- `ABCI_LOG_FILE_PATH`: File path for log file (use when `ABCI_LOG_TARGET` is set to `file`) [Default: `./abci-<PID>-<CURRENT_DATETIME>.log`]

## Build

```sh
CGO_ENABLED=1 go build -ldflags "-X github.com/ndidplatform/smart-contract/v4/abci/version.GitCommit=`git rev-parse --short=8 HEAD`" -tags "cleveldb" -o ./did-tendermint ./abci
```

or with snappy lib used by cleveldb

```sh
CGO_ENABLED=1 CGO_LDFLAGS="-lsnappy" go build -ldflags "-X github.com/ndidplatform/smart-contract/v4/abci/version.GitCommit=`git rev-parse --short=8 HEAD`" -tags "cleveldb" -o ./did-tendermint ./abci
```

## Usage

```sh
./did-tendermint --home $TENDERMINT_HOME_DIR node
```

### Examples

- Run IdP node

  ```sh
  ./did-tendermint --home ./config/tendermint/IdP unsafe_reset_all && ABCI_DB_DIR_PATH=IdP_DB ./did-tendermint --home ./config/tendermint/IdP node
  ```

  or

  ```sh
  go run ./abci --home ./config/tendermint/IdP unsafe_reset_all && CGO_ENABLED=1 CGO_LDFLAGS="-lsnappy" ABCI_DB_DIR_PATH=IdP_DB go run -tags "cleveldb" ./abci --home ./config/tendermint/IdP node
  ```

- Run RP node

  ```sh
  ./did-tendermint --home ./config/tendermint/RP unsafe_reset_all && ABCI_DB_DIR_PATH=RP_DB ./did-tendermint --home ./config/tendermint/RP node
  ```

  or

  ```sh
  go run ./abci --home ./config/tendermint/RP unsafe_reset_all && CGO_ENABLED=1 CGO_LDFLAGS="-lsnappy" ABCI_DB_DIR_PATH=RP_DB go run -tags "cleveldb" ./abci --home ./config/tendermint/RP node
  ```

- Run AS node

  ```sh
  ./did-tendermint --home ./config/tendermint/AS unsafe_reset_all && ABCI_DB_DIR_PATH=AS_DB ./did-tendermint --home ./config/tendermint/AS node
  ```

  or

  ```sh
  go run ./abci --home ./config/tendermint/AS unsafe_reset_all && CGO_ENABLED=1 CGO_LDFLAGS="-lsnappy" ABCI_DB_DIR_PATH=AS_DB go run -tags "cleveldb" ./abci --home ./config/tendermint/AS node
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

- To run docker container without building an image, run command in **Run** section (no building required). It will run docker container with images from Dockerhub (https://hub.docker.com/r/ndidplatform/abci/ and https://hub.docker.com/r/ndidplatform/tendermint/).
- To pull latest image from Dockerhub, run `docker pull ndidplatform/abci` and `docker pull ndidplatform/tendermint`
- Docker container can be run with `-u` or `--user` flag (e.g. `-u 65534:65534`). In case you are using docker-compose, `user` can be specified in docker-compose file (e.g. `user: 65534:65534`) (see [Compose file reference](https://docs.docker.com/compose/compose-file/#domainname-hostname-ipc-mac_address-privileged-read_only-shm_size-stdin_open-tty-user-working_dir) for more detail).
- When running docker container with non-root user, source directories that will be mounted into the container as `ABCI_DB_DIR_PATH` and `TMHOME` must be created beforehand with the non-root user as owner.

## IMPORTANT NOTE

1. You must start IdP, RP and AS nodes in order to run the platform.
2. When running nodes on separate machines, please edit `seeds` in `config/tendermint/{RP or IdP or AS}/config/config.toml` to match address of other machines.

## Tests

Test this app with the command below

```sh
cd test

TENDERMINT_ADDRESS=http://localhost:45000 go test -v
```

# Technical details to connect with `api`

# Broadcast tx format (Protobuf)

```
message Tx {
  string method = 1;
  string params = 2;
  bytes nonce = 3;
  bytes signature = 4;
  string node_id = 5;
}
```

# Query format (Protobuf)

```
message Query {
  string method = 1;
  string params = 2;
}
```

# Create transaction function

## AddAccessor

### Parameter

```sh
{
  "reference_group_code": "0821573f-6aff-49ad-b17b-5586774e07c5",
  "identity_namespace": "cid",
  "identity_identifier_hash": "766f1de3e522f22a16fe520312051e68d9cbf1b177e702cee827f0107752a418",
  "accessor_id": "dd1158c8-5833-48f9-95d5-b28193bb993a",
  "accessor_public_key": "-----BEGIN PUBLIC KEY-----nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhdKdvawPO8XXroiAGkxFnfLRCqvk4X2iAMStq1ADjmPPWhKgF/ssU9LBdHKHPPX1+NMOX29gOL3ZCxfZamKO6nAbODt1e0bVfblWWMq5uMwzNrFo4nKas74SLJwiMg0vtn1NnHU4QTTrMYmGqRf2WZnIN9Iro4LytUTLEBCpimWM2hodO8I60bANAO0gI96BzAWMleoioOzWlq6JKkiDsj7n8EjCI/bY1T/v4F7rg2FxrIH/BH4TUDy88pIvAYy4nNEyGyr8KzMm1cKxOgnJI8OnnwT8HrAJQ58T3HCCiCrKAohkYBWITPk3cmqGfOKrqZ2DI+a6URofMVvQFlwfYvqU6n5QIDAQABn-----END PUBLIC KEY-----",
  "accessor_type": "RSA2048",
  "request_id": "099ffff6-60b0-431f-b72a-167848febe1b"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## AddNamespace

### Parameter

```sh
{
  "namespace": "cid",
  "description": "Citizen ID",
  "allowed_identifier_count_in_reference_group": 1,
  "allowed_active_identifier_count_in_reference_group": 1
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## AddNodeToken

### Parameter

```sh
{
  "amount": 222.22,
  "node_id": "CuQfyyhjGcCAzKREzHmL"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## AddNodeToProxyNode

### Parameter

```sh
{
  "config": "KEY_ON_PROXY",
  "node_id": "BLUbbuoywxSirpxDIPgW",
  "proxy_node_id": "KWipXqVCIprtsbBptmtB"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## AddService

### Parameter

```sh
{
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "service_name": "Bank statement",
  "data_schema": "string",
  "data_schema_version": "string"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## CloseRequest

### Parameter

```sh
{
  "request_id": "ca3c69a9-d67c-4e85-878a-ab5d8e0df32b",
  "response_valid_list": [
    {
      "idp_id": "fJjCNCYyDxHrfMrbZGCu",
      "valid_ial": true,
      "valid_signature": true
    }
  ]
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## CreateIdpResponse

### Parameter

```sh
{
  "request_id": "46a08787-6014-42d1-a6cf-1094e4cf2cb8",
  "aal": 3,
  "ial": 3,
  "signature": "signature",
  "status": "accept"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## CreateRequest

### Parameter

```sh
{
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74",
  "idp_id_list": [
    "lvEzsuTcZvIRvZyrdEsi",
    "njHtYuHHxCvzzofcpwon"
  ],
  "data_request_list": [
    {
      "answered_as_id_list": null,
      "as_id_list": null,
      "min_as": 1,
      "received_data_from_list": null,
      "request_params_hash": "hash",
      "service_id": "LlUXaAYeAoVDiQziKPMc"
    }
  ],
  "min_aal": 3,
  "min_ial": 3,
  "min_idp": 1,
  "mode": 3,
  "request_message_hash": "hash('Please allow...')",
  "request_timeout": 259200,
  "purpose": "AddAccessor"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## DisableNamespace

### Parameter

```sh
{
  "namespace": "SJsMIeJcerfZpBfXkJgU"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## DisableNode

### Parameter

```sh
{
  "node_id": "sqldejLfsEObEFKmRfwz"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## DisableService

### Parameter

```sh
{
  "service_id": "PAfvPhGmrzILDePeXsMm"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## DisableServiceDestination

### Parameter

```sh
{
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## DisableServiceDestinationByNDID

### Parameter

```sh
{
  "node_id": "XckRuCmVliLThncSTnfG",
  "service_id": "qvyfrfJRsfaesnDsYHbH"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## EnableNamespace

### Parameter

```sh
{
  "namespace": "SJsMIeJcerfZpBfXkJgU"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## EnableNode

### Parameter

```sh
{
  "node_id": "CuQfyyhjGcCAzKREzHmL"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## EnableService

### Parameter

```sh
{
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## EnableServiceDestination

### Parameter

```sh
{
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## EnableServiceDestinationByNDID

### Parameter

```sh
{
  "node_id": "XckRuCmVliLThncSTnfG",
  "service_id": "qvyfrfJRsfaesnDsYHbH"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## InitNDID

### Parameter

```sh
{
  "node_id": "NDID",
  "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA30i6deo6vqxPdoxA9pUp\nuBag/cVwEVWO8dds5QDfu/z957zxXUCYRxaiRWGAbOta4K5/7cxlsqI8fCvoSyAa\n/B7GTSc3vivK/GWUFP+sQ/Mj6C/fgw5pxK/+olBzfzLMDEOwFRbnYtPtbWozfvce\nq77fEReTUdBGRLak7twxLrRPNzIu/Gqvn5AR8urXyF4r143CgReGkXTTmOvHpHu9\n8kCQSINFuwBB98RLFuWdVwkrHyzaGnymQu+0OR1Z+1MDIQ9WlViD1iaJhYKA6a0G\n0O4Nns6ISPYSh7W7fI31gWTgHUZN5iTkLb9t27DpW9G+DXryq+Pnl5c+z7es/7T3\n4QIDAQAB\n-----END PUBLIC KEY-----\n",
  "master_public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA30i6deo6vqxPdoxA9pUp\nuBag/cVwEVWO8dds5QDfu/z957zxXUCYRxaiRWGAbOta4K5/7cxlsqI8fCvoSyAa\n/B7GTSc3vivK/GWUFP+sQ/Mj6C/fgw5pxK/+olBzfzLMDEOwFRbnYtPtbWozfvce\nq77fEReTUdBGRLak7twxLrRPNzIu/Gqvn5AR8urXyF4r143CgReGkXTTmOvHpHu9\n8kCQSINFuwBB98RLFuWdVwkrHyzaGnymQu+0OR1Z+1MDIQ9WlViD1iaJhYKA6a0G\n0O4Nns6ISPYSh7W7fI31gWTgHUZN5iTkLb9t27DpW9G+DXryq+Pnl5c+z7es/7T3\n4QIDAQAB\n-----END PUBLIC KEY-----\n",
  "chain_history_info": "{\"chains\":[{\"chain_id\":\"test-chain-NDID\",\"latest_block_hash\":\"39DAE266185B54C62A6932445021FEB641E5D5DB\",\"latest_app_hash\":\"588C2F4A1B236281565C301EEA9BA863CF5F3E28\",\"latest_block_height\":\"164\"},{\"chain_id\":\"test-chain-NDID\",\"latest_block_hash\":\"E25104B14BD7BB48734BDB1E7EAF5E494318C4C3\",\"latest_app_hash\":\"632990575BFC06B7CE5C57D0D0AD9AEA3DBBB230\",\"latest_block_height\":\"174\"}]}\r\n"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## ReduceNodeToken

### Parameter

```sh
{
  "amount": 61.11,
  "node_id": "nfhwDGTTeRdMeXzAgLij"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## RegisterIdentity

### Parameter

```sh
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "new_identity_list": [{
    "identity_namespace": "citizenId",
    "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  }],
  "ial": 3,
  "mode_list": [2, 3], // allow only 2, 3
  "accessor_id": "11267a29-2196-4400-8b67-7424519b87ec",
  "accessor_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7BjIuleY9/5ObFl0w+U2\\nfID4cC8v3yIaOjsImXYNon04TZ6lHs8gNvrR1Q0MRtGTugL8XJPj3tw1AbHj01L8\\nW0HwKpFQxhwvGzi0Sesb9Lhn9aA4MCmfMG7PwLGzgdeHR7TVl7VhKx7gedyYIdju\\nEFzAtsJYO1plhUfFv6gdg/05VOjFTtVdWtwKgjUesmuv1ieZDj64krDS84Hka0gM\\njNKm4+mX8HGUPEkHUziyBpD3MwAzyA+I+Z90khDBox/+p+DmlXuzMNTHKE6bwesD\\n9ro1+LVKqjR/GjSZDoxL13c+Va2a9Dvd2zUoSVcDwNJzSJtBrxMT/yoNhlUjqlU0\\nYQIDAQAB\\n-----END PUBLIC KEY-----",
  "accessor_type": "accessor_type",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## RegisterNode

### Parameter

```sh
{
  "master_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\\nDQIDAQAB\\n-----END PUBLIC KEY-----\\n",
  "max_aal": 3,
  "max_ial": 3,
  "node_id": "CuQfyyhjGcCAzKREzHmL",
  "node_name": "IdP Number 1 from ...",
  "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\\njwIDAQAB\\n-----END PUBLIC KEY-----\\n",
  "role": "IdP"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## RegisterServiceDestination

### Parameter

```sh
{
  "min_aal": 1.2,
  "min_ial": 1.1,
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "supported_namespace_list": [
    "citizenId"
  ]
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## RegisterServiceDestinationByNDID

### Parameter

```sh
{
  "node_id": "XckRuCmVliLThncSTnfG",
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## RemoveNodeFromProxyNode

### Parameter

```sh
{
  "node_id": "BLUbbuoywxSirpxDIPgW"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## SetDataReceived

### Parameter

```sh
{
  "as_id": "XckRuCmVliLThncSTnfG",
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74",
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## SetMqAddresses

### Parameter

```sh
{
  "addresses": [
    {
      "ip": "192.168.3.99",
      "port": 8000
    }
  ]
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## SetNodeToken

### Parameter

```sh
{
  "amount": 100,
  "node_id": "nfhwDGTTeRdMeXzAgLij"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## SetPriceFunc

### Parameter

```sh
{
  "func": "CreateRequest",
  "price": 1
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## SetValidator

### Parameter

```sh
{
  "power": 100,
  "public_key": "qJ0HsJvzHz/CAEBMCpvqfIpMIktfOsN0kh5O3+d0bks="
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## CreateAsResponse

### Parameter

```sh
{
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74",
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "signature": "sign(data,asKey)",
  "error_code": 0
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## TimeOutRequest

### Parameter

```sh
{
  "request_id": "04db0ddf-4d3f-4b40-93b0-af418ad8a2d7",
  "response_valid_list": [
    {
      "idp_id": "CuQfyyhjGcCAzKREzHmL",
      "valid_ial": false,
      "valid_signature": false
    }
  ]
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## UpdateIdentity

### Parameter

```sh
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "ial": 2.2
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## UpdateNode

### Parameter

```sh
{
  "master_public_key": "",
  "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\\nPwIDAQAB\\n-----END PUBLIC KEY-----\\n",
  "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"] // For IdP node only, "text/plain" must be supported always
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## UpdateNodeByNDID

### Parameter

```sh
{
  "max_aal": 2.4,
  "max_ial": 2.3,
  "node_id": "CuQfyyhjGcCAzKREzHmL",
  "node_name": ""
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## UpdateNodeProxyNode

### Parameter

```sh
{
  "config": "KEY_ON_PROXY",
  "node_id": "BLUbbuoywxSirpxDIPgW",
  "proxy_node_id": "LvFjFNAPnfEwPFGEEbdx"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## UpdateService

### Parameter

```sh
{
  "min_aal": 1.5,
  "min_ial": 1.4,
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "supported_namespace_list": [
    "citizenId"
  ]
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## UpdateServiceDestination

### Parameter

```sh
{
  "min_aal": 1.5,
  "min_ial": 1.4,
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "supported_namespace_list": [
    "citizenId"
  ]
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## RevokeAccessor

### Parameter

```sh
{
  "accessor_id_list": [
    "11d10976-aede-4ba0-9f44-fc0c96db1f32"
  ],
  "request_id": "e7dcf1c2-eea7-4dc8-af75-724cf86454ef"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## SetLastBlock

### Parameter

```sh
{
  "block_height": 0
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## AddIdentity

### Parameter

```sh
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "new_identity_list": [{
    "identity_namespace": "citizenId",
    "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  }],
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## UpdateIdentityModeList

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "mode_list": [2, 3], // allow only 2,3
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## RevokeIdentityAssociation

### Parameter

```json
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## SetAllowedModeList

### Parameter

```json
{
  "purpose": "",
  "allowed_mode_list": [1, 2, 3]
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## UpdateNamespace

### Parameter

```json
{
  "description": "Citizen ID",
  "namespace": "citizenId",
  "allowed_identifier_count_in_reference_group": 1,
  "allowed_active_identifier_count_in_reference_group": 1
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## SetAllowedMinIalForRegisterIdentityAtFirstIdp

### Parameter

```json
{
  "min_ial": 2.3
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

## RevokeAndAddAccessor

### Parameter

```json
{
  "revoking_accessor_id": "11d10976-aede-4ba0-9f44-fc0c96db1f32",
  "accessor_id": "07938aa2-2aaf-4bb5-9ccd-33700581e870",
  "accessor_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhdKdvawPO8XXroiAGkxF\\nfLRCqvk4X2iAMStq1ADjmPPWhKgF/ssU9LBdHKHPPX1+NMOX29gOL3ZCxfZamKO6\\nAbODt1e0bVfblWWMq5uMwzNrFo4nKas74SLJwiMg0vtn1NnHU4QTTrMYmGqRf2WZ\\nIN9Iro4LytUTLEBCpimWM2hodO8I60bANAO0gI96BzAWMleoioOzWlq6JKkiDsj7\\n8EjCI/bY1T/v4F7rg2FxrIH/BH4TUDy88pIvAYy4nNEyGyr8KzMm1cKxOgnJI8On\\nwT8HrAJQ58T3HCCiCrKAohkYBWITPk3cmqGfOKrqZ2DI+a6URofMVvQFlwfYvqU6\\n5QIDAQAB\\n-----END PUBLIC KEY-----",
  "accessor_type": "accessor_type_2",
  "request_id": "edaec8df-7865-4473-8707-054dd0cffe2d"
}
```

### Expected Output

```sh
{
  "code": 0,
  "log": "success",
  "tags": [
    {
      "key": "success",
      "value": "true"
    }
  ]
}
```

# Query function

## CheckExistingAccessorGroupID

### Parameter

```sh
{
  "accessor_group_id": "0d855490-0723-4e0d-b39b-3f230c68f815"
}
```

### Expected Output

```sh
{
  "exist": true
}
```

## CheckExistingAccessorID

### Parameter

```sh
{
  "accessor_id": "11267a29-2196-4400-8b67-7424519b87ec"
}
```

### Expected Output

```sh
{
  "exist": true
}
```

## CheckExistingIdentity

### Parameter

```sh
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
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
  "accessor_id": "07938aa2-2aaf-4bb5-9ccd-33700581e870"
}
```

### Expected Output

```sh
{
  "accessor_group_id": "0d855490-0723-4e0d-b39b-3f230c68f815"
}
```

## GetAccessorKey

### Parameter

```sh
{
  "accessor_id": "11267a29-2196-4400-8b67-7424519b87ec"
}
```

### Expected Output

```sh
{
  "accessor_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7BjIuleY9/5ObFl0w+U2\\nfID4cC8v3yIaOjsImXYNon04TZ6lHs8gNvrR1Q0MRtGTugL8XJPj3tw1AbHj01L8\\nW0HwKpFQxhwvGzi0Sesb9Lhn9aA4MCmfMG7PwLGzgdeHR7TVl7VhKx7gedyYIdju\\nEFzAtsJYO1plhUfFv6gdg/05VOjFTtVdWtwKgjUesmuv1ieZDj64krDS84Hka0gM\\njNKm4+mX8HGUPEkHUziyBpD3MwAzyA+I+Z90khDBox/+p+DmlXuzMNTHKE6bwesD\\n9ro1+LVKqjR/GjSZDoxL13c+Va2a9Dvd2zUoSVcDwNJzSJtBrxMT/yoNhlUjqlU0\\nYQIDAQAB\\n-----END PUBLIC KEY-----",
  "active": true
}
```

## GetAsNodesByServiceId

### Parameter

```sh
{
  "node_id_list": null,
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "node": [
    {
      "min_aal": 1.5,
      "min_ial": 1.4,
      "node_id": "XckRuCmVliLThncSTnfG",
      "node_name": "AS1",
      "supported_namespace_list": [
        "citizenId"
      ]
    }
  ]
}
```

## GetAsNodesInfoByServiceId

### Parameter

```sh
{
  "node_id_list": null,
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "node": [
    {
      "min_aal": 1.5,
      "min_ial": 1.4,
      "mq": [
        {
          "ip": "192.168.3.102",
          "port": 8000
        }
      ],
      "name": "AS1",
      "node_id": "XckRuCmVliLThncSTnfG",
      "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\\ndQIDAQAB\\n-----END PUBLIC KEY-----\\n",
      "supported_namespace_list": [
        "citizenId"
      ]
    }
  ]
}
```

## GetDataSignature

### Parameter

```sh
{
  "node_id": "XckRuCmVliLThncSTnfG",
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74",
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "signature": "sign(data,asKey)"
}
```

## GetIdentityInfo

### Parameter

```sh
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "node_id": "CuQfyyhjGcCAzKREzHmL"
}
```

### Expected Output

```sh
{
  "ial": 2.2,
  "mode_list": [2, 3]
}
```

## GetIdpNodes

### Parameter

```sh
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "min_aal": 3,
  "min_ial": 3,
  "node_id_list": [],
  "supported_request_message_data_url_type_list": [],
  "mode_list": [3]
}
```

### Expected Output

```sh
{
  "node": [
    {
      "max_aal": 3,
      "max_ial": 3,
      "node_id": "CuQfyyhjGcCAzKREzHmL",
      "node_name": "IdP Number 1 from ...",
      "ial": 3,
      "mode_list": [2, 3],
      "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"]
    }
  ]
}
```

## GetIdpNodesInfo

### Parameter

```sh
{
  "reference_group_code": "aaaaa-bbbbb-ccccc-ddddd",
  "identity_namespace": "citizenId",
  "identity_identifier_hash": "c765a80f1ee71299c361c1b4cb4d9c36b44061a526348a71287ea0a97cea80f6",
  "min_aal": 3,
  "min_ial": 3,
  "node_id_list": [], //array of string
  "supported_request_message_data_url_type_list": [], //array of string
  "ial": 3,
  "mode_list": [3]
}
```

### Expected Output

```sh
{
  "node": [
    {
      "max_aal": 3,
      "max_ial": 3,
      "mode_list": [2, 3], //array of available mode
      "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"],
      "mq": [
        {
          "ip": "192.168.3.99",
          "port": 8000
        }
      ],
      "name": "IdP Number 1 from ...",
      "node_id": "CuQfyyhjGcCAzKREzHmL",
      "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\\njwIDAQAB\\n-----END PUBLIC KEY-----\\n"
    }
  ]
}
```

## GetMqAddresses

### Parameter

```sh
{
  "node_id": "CuQfyyhjGcCAzKREzHmL"
}
```

### Expected Output

```sh
[
  {
    "ip": "192.168.3.99",
    "port": 8000
  }
]
```

## GetNamespaceList

### Parameter

```sh

```

### Expected Output

```sh
[
  {
    "namespace": "WsvGOEjoFqvXsvcfFVWm",
    "description": "Citizen ID",
    "active": true,
    "allowed_identifier_count_in_reference_group": 1,
    "allowed_active_identifier_count_in_reference_group": 1
  },
  {
    "namespace": "SJsMIeJcerfZpBfXkJgU",
    "description": "Tel number",
    "active": true,
  }
]
```

## GetNodeIDList

### Parameter

```sh
{
  "role": ""
}
```

### Expected Output

```sh
{
  "node_id_list": [
    "nfhwDGTTeRdMeXzAgLij",
    "CuQfyyhjGcCAzKREzHmL",
    "XckRuCmVliLThncSTnfG",
    "QTspWhbMDeVXIIJfXcBa",
    "daEmAcxLsUcucuRWeYbK",
    "KWipXqVCIprtsbBptmtB",
    "BLUbbuoywxSirpxDIPgW",
    "xRvyWoEGrOmPVYXdyWbw",
    "LvFjFNAPnfEwPFGEEbdx"
  ]
}
```

## GetNodeInfo

### Parameter

```sh
{
  "node_id": "CuQfyyhjGcCAzKREzHmL"
}
```

### Expected Output

```sh
{
  "master_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\\nDQIDAQAB\\n-----END PUBLIC KEY-----\\n",
  "max_aal": 2.4,
  "max_ial": 2.3,
  "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"],
  "mq": [
    {
      "ip": "192.168.3.99",
      "port": 8000
    }
  ],
  "node_name": "IdP Number 1 from ...",
  "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\\nPwIDAQAB\\n-----END PUBLIC KEY-----\\n",
  "role": "IdP",
  "active": true
}
```

## GetNodeMasterPublicKey

### Parameter

```sh
{
  "node_id": "nfhwDGTTeRdMeXzAgLij"
}
```

### Expected Output

```sh
{
  "master_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\\nDQIDAQAB\\n-----END PUBLIC KEY-----\\n"
}
```

## GetNodePublicKey

### Parameter

```sh
{
  "node_id": "nfhwDGTTeRdMeXzAgLij"
}
```

### Expected Output

```sh
{
  "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwCB4UBzQcnd6GAzPgbt9\\nj2idW23qKZrsvldPNifmOPLfLlMusv4EcyJf4L42/aQbTn1rVSu1blGkuCK+oRlK\\nWmZEWh3xv9qrwCwov9Jme/KOE98zOMB10/xwnYotPadV0de80wGvKT7OlBlGulQR\\nRhhgENNCPSxdUlozrPhrzGstXDr9zTYQoR3UD/7Ntmew3mnXvKj/8+U48hw913Xn\\n6btBP3Uqg2OurXDGdrWciWgIMDEGyk65NOc8FOGa4AjYXzyi9TqOIfmysWhzKzU+\\nfLysZQo10DfznnQN3w9+pI+20j2zB6ggpL75RjZKYgHU49pbvjF/eOSTOg9o5HwX\\n0wIDAQAB\\n-----END PUBLIC KEY-----\\n"
}
```

## GetNodesBehindProxyNode

### Parameter

```sh
{
  "proxy_node_id": "KWipXqVCIprtsbBptmtB"
}
```

### Expected Output

```sh
{
  "nodes": [
    {
      "config": "KEY_ON_PROXY",
      "master_public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\\njwIDAQAB\\n-----END PUBLIC KEY-----\\n",
      "max_aal": 3,
      "max_ial": 3,
      "supported_request_message_data_url_type_list": ["text/plain", "application/pdf"],
      "node_id": "BLUbbuoywxSirpxDIPgW",
      "node_name": "IdP6BehindProxy1",
      "public_key": "-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\\njwIDAQAB\\n-----END PUBLIC KEY-----\\n",
      "role": "IdP"
    }
  ]
}
```

## GetNodeToken

### Parameter

```sh
{
  "node_id": "nfhwDGTTeRdMeXzAgLij"
}
```

### Expected Output

```sh
{
  "amount": 50
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
  "price": 9.99
}
```

## GetRequest

### Parameter

```sh
{
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74"
}
```

### Expected Output

```sh
{
  "closed": true,
  "mode": 3,
  "request_message_hash": "hash('Please allow...')",
  "timed_out": false
}
```

## GetRequestDetail

### Parameter

```sh
{
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74"
}
```

### Expected Output

```sh
{
  "closed": false,
  "data_request_list": [
    {
      "answered_as_id_list": [
        "XckRuCmVliLThncSTnfG"
      ],
      "as_id_list": [],
      "min_as": 1,
      "received_data_from_list": [
        "XckRuCmVliLThncSTnfG"
      ],
      "request_params_hash": "hash",
      "service_id": "LlUXaAYeAoVDiQziKPMc"
    }
  ],
  "min_aal": 3,
  "min_ial": 3,
  "min_idp": 1,
  "mode": 3,
  "request_id": "16dc0550-a6e4-4e1f-8338-37c2ac85af74",
  "request_message_hash": "hash('Please allow...')",
  "request_timeout": 259200,
  "idp_id_list": [
    "lvEzsuTcZvIRvZyrdEsi",
    "njHtYuHHxCvzzofcpwon"
  ],
  "requester_node_id": "nfhwDGTTeRdMeXzAgLij",
  "response_list": [
    {
      "aal": 3,
      "ial": 3,
      "idp_id": "CuQfyyhjGcCAzKREzHmL",
      "signature": "signature",
      "status": "accept",
      "valid_ial": null,
      "valid_signature": null
    }
  ],
  "purpose": "",
  "timed_out": false,
  "creation_block_height": 50,
  "creation_chain_id": "test-chain-NDID"
}
```

## GetServiceDetail

### Parameter

```sh
{
  "service_id": "LlUXaAYeAoVDiQziKPMc"
}
```

### Expected Output

```sh
{
  "active": true,
  "service_id": "LlUXaAYeAoVDiQziKPMc",
  "service_name": "Bank statement (ย้อนหลัง 3 เดือน)",
  "data_schema": "string",
  "data_schema_version": "string"
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
    "active": true,
    "service_id": "LlUXaAYeAoVDiQziKPMc",
    "service_name": "Bank statement (ย้อนหลัง 3 เดือน)"
  }
]
```

## GetServicesByAsID

### Parameter

```sh
{
  "as_id": "XckRuCmVliLThncSTnfG"
}
```

### Expected Output

```sh
{
  "services": [
    {
      "active": true,
      "min_aal": 1.1,
      "min_ial": 1.1,
      "service_id": "AFLHeKQVLNQOkIOxoNid",
      "supported_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    },
    {
      "active": true,
      "min_aal": 2.2,
      "min_ial": 2.2,
      "service_id": "qvyfrfJRsfaesnDsYHbH",
      "supported_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    },
    {
      "active": true,
      "min_aal": 3.3,
      "min_ial": 3.3,
      "service_id": "JTFHqDoJRccWcikcJqnL",
      "supported_namespace_list": [
        "citizenId"
      ],
      "suspended": false
    }
  ]
}
```

## GetAccessorsInAccessorGroup

### Parameter

```sh
{
  "accessor_group_id": "b0dbc48f-9b72-42fa-904e-22c00c30d5e5",
  "idp_id": "xTkDRjpgwuIazfaCHAAM"
}
```

### Expected Output

```sh
{
  "accessor_list": [
    "c719e6aa-16ab-4ecb-9063-eff1a2e75fd3"
  ]
}
```

## GetAccessorOwner

### Parameter

```sh
{
  "accessor_id": "11d10976-aede-4ba0-9f44-fc0c96db1f32"
}
```

### Expected Output

```sh
{
  "node_id": "NsutHiOdeiAGSODKTNOF"
}
```

## IsInitEnded

### Parameter

```sh

```

### Expected Output

```sh
{
  "init_ended": true
}
```

## GetChainHistory

### Parameter

```sh

```

### Expected Output

```sh
{
  "chains": [
    {
      "chain_id": "test-chain-NDID",
      "latest_block_hash": "39DAE266185B54C62A6932445021FEB641E5D5DB",
      "latest_app_hash": "588C2F4A1B236281565C301EEA9BA863CF5F3E28",
      "latest_block_height": "164"
    },
    {
      "chain_id": "test-chain-NDID",
      "latest_block_hash": "E25104B14BD7BB48734BDB1E7EAF5E494318C4C3",
      "latest_app_hash": "632990575BFC06B7CE5C57D0D0AD9AEA3DBBB230",
      "latest_block_height": "174"
    }
  ]
}
```
