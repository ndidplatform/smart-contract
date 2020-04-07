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

NODE1_ABCI_DB_DIR="$TMP_DIR/idp_abci_db"
NODE2_ABCI_DB_DIR="$TMP_DIR/rp_abci_db"
NODE3_ABCI_DB_DIR="$TMP_DIR/as_abci_db"
NODE4_ABCI_DB_DIR="$TMP_DIR/proxy_abci_db"

run_node_in_background() {
  CGO_ENABLED=1 \
  CGO_LDFLAGS="-lsnappy" \
  ABCI_DB_DIR_PATH=$2 \
  go run -tags "cleveldb" ./abci --home $1 node &
}

run_node_in_background $NODE1_TENDERMINT_HOME_DIR $NODE1_ABCI_DB_DIR
run_node_in_background $NODE2_TENDERMINT_HOME_DIR $NODE2_ABCI_DB_DIR
run_node_in_background $NODE3_TENDERMINT_HOME_DIR $NODE3_ABCI_DB_DIR
run_node_in_background $NODE4_TENDERMINT_HOME_DIR $NODE4_ABCI_DB_DIR

wait
