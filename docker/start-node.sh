#!/bin/sh

if [ -z ${TMHOME} ]; then TMHOME=/tendermint; fi

usage() {
  echo "Usage: $(basename ${0}) <mode>"
  echo "where mode can be :"
  echo "genesis = run this node as genesis node"
  echo "secondary = run this node as secondary node"
  echo "reset = call unsafe_reset_all"
}

tendermint_init() {
  tendermint init --home=${TMHOME}
}

tendermint_reset() {
  tendermint --home=${TMHOME} unsafe_reset_all
}

tendermint_get_genesis_from_seed() {
  wget -qO - http://${SEED_HOSTNAME}:${TM_RPC_PORT}/genesis | jq -r .result.genesis > ${TMHOME}/config/genesis.json
}

tendermint_new_priv_validator() {
  tendermint gen_validator > ${TMHOME}/config/priv_validator.json
}

tendermint_wait_for_sync_complete() {
  local HOSTNAME=$1
  local PORT=$2
  while true; do
    [ ! "$(wget -qO - http://${HOSTNAME}:${PORT}/status | jq -r .result.syncing)" = "false" ] || break
    sleep 1
  done;
}

tendermint_add_validator() {
  local PUBKEY=$(cat ${TMHOME}/config/priv_validator.json | jq -r .pub_key.data)
  wget -qO - http://${SEED_HOSTNAME}:${TM_RPC_PORT}/broadcast_tx_commit?tx=\"val:${PUBKEY}\"
}

TYPE=${1}

if [ ! -f ${TMHOME}/config/genesis.json ]; then
  case ${TYPE} in
    genesis) 
      tendermint_init
      shift
      tendermint node --consensus.create_empty_blocks=false --moniker=${HOSTNAME} $@
      ;;
    secondary) 
      if [ -z ${SEED_HOSTNAME} ]; then echo "Error: env SEED_HOSTNAME is not set"; exit 1; fi
      tendermint_init
      tendermint_wait_for_sync_complete ${SEED_HOSTNAME} ${TM_RPC_PORT}
      tendermint_get_genesis_from_seed
      shift
      tendermint node --consensus.create_empty_blocks=false --moniker=${HOSTNAME} $@
      tendermint_wait_for_sync_complete localhost ${TM_RPC_PORT}
      tendermint_add_validator
      ;;
    reset)
      tendermint_reset
      exit 0
      ;;
    *)
      usage
      exit 1
      ;;
  esac
else
  shift
  tendermint node --consensus.create_empty_blocks=false --moniker=${HOSTNAME} $@
fi
