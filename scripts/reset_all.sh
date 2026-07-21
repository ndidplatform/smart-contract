#!/bin/bash

trap killgroup SIGINT

killgroup(){
  echo killing...
  kill 0
}

BASE_DIR="$(cd "$(dirname "$0")"; pwd)"
REPO_DIR="$(dirname $BASE_DIR)"

NODE1_TENDERMINT_HOME_DIR="$REPO_DIR/config/tendermint/IdP"
NODE2_TENDERMINT_HOME_DIR="$REPO_DIR/config/tendermint/RP"
NODE3_TENDERMINT_HOME_DIR="$REPO_DIR/config/tendermint/AS"
NODE4_TENDERMINT_HOME_DIR="$REPO_DIR/config/tendermint/proxy"

TMP_DIR="$REPO_DIR/tmp"

reset_all() {
  go run ./abci --home $1 unsafe-reset-all
}

rm -rf $TMP_DIR

reset_all $NODE1_TENDERMINT_HOME_DIR
reset_all $NODE2_TENDERMINT_HOME_DIR
reset_all $NODE3_TENDERMINT_HOME_DIR
reset_all $NODE4_TENDERMINT_HOME_DIR

wait
