#!/bin/sh
set -e 

# Ensure that the first argument is did-tendermint
if [ "$1" != "did-tendermint" ]; then
  set -- "did-tendermint" "$@"
fi

# Check existence and owner of ABCI_DB_DIR_PATH and TMHOME
if [ "$1" = "did-tendermint" -a "$(id -u)" != "0" ]; then
  if [ ! -d ${ABCI_DB_DIR_PATH} ]; then 
    echo "${ABCI_DB_DIR_PATH} is not directory or missing"
    exit 1
  fi 

  if [ ! -d ${TMHOME} ]; then
    echo "${TMHOME} is not directory or missing"
    exit 1
  fi

  user="$(id -u)"
  group="$(id -g)"

  if [ ! -z "$(find ${ABCI_DB_DIR_PATH} ! -user $user ! -group $group)" ]; then
    echo "${ABCI_DB_DIR_PATH} or the files inside have incorrect owner"
    exit 1
  fi

  if [ ! -z "$(find ${TMHOME} ! -user $user ! -group $group)" ]; then
    echo "${TMHOME} the files inside have incorrect owner"
    exit 1
  fi
fi

exec "$@"
