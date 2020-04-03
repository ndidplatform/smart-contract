#!/bin/sh

BASE_DIR="$(cd "$(dirname "$0")"; pwd)"
REPO_DIR="$(dirname $BASE_DIR)"

TENDERMINT_HOME_DIR="$REPO_DIR/config/tendermint/proxy"

ABCI_DB_DIR="$REPO_DIR/tmp/proxy_abci_db"

rm -rf $ABCI_DB_DIR

go run ./abci --home $TENDERMINT_HOME_DIR unsafe_reset_all && \
CGO_ENABLED=1 \
CGO_LDFLAGS="-lsnappy" \
ABCI_DB_DIR_PATH=$ABCI_DB_DIR \
go run -tags "cleveldb" ./abci --home $TENDERMINT_HOME_DIR node