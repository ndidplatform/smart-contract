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
NODE4_TENDERMINT_HOME_DIR="$REPO_DIR/config/tendermint/AS2"

TMP_DIR="$REPO_DIR/tmp"

NODE1_ABCI_DB_DIR="$TMP_DIR/idp_abci_db"
NODE2_ABCI_DB_DIR="$TMP_DIR/rp_abci_db"
NODE3_ABCI_DB_DIR="$TMP_DIR/as_abci_db"
NODE4_ABCI_DB_DIR="$TMP_DIR/proxy_abci_db"

reset_all_and_run_node_in_background() {
  go run ./abci --home $1 unsafe_reset_all && \
  CGO_ENABLED=1 \
  CGO_LDFLAGS="-lsnappy" \
  ABCI_DB_DIR_PATH=$2 \
  go run -tags "cleveldb" ./abci --home $1 node &
}

rm -rf $TMP_DIR

reset_all_and_run_node_in_background $NODE1_TENDERMINT_HOME_DIR $NODE1_ABCI_DB_DIR

# Wait a bit for the first node (seed node) to start
sleep 2

reset_all_and_run_node_in_background $NODE2_TENDERMINT_HOME_DIR $NODE2_ABCI_DB_DIR
reset_all_and_run_node_in_background $NODE3_TENDERMINT_HOME_DIR $NODE3_ABCI_DB_DIR
reset_all_and_run_node_in_background $NODE4_TENDERMINT_HOME_DIR $NODE4_ABCI_DB_DIR

wait
