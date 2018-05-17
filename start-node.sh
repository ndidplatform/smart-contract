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
  wget -qO - ${SEED_HOSTNAME}:${TM_RPC_PORT}/genesis | jq -r .result.genesis > ${TMHOME}/config/genesis.json
}

tendermint_new_priv_validator() {
  tendermint gen_validator > ${TMHOME}/config/priv_validator.json
}

tendermint_wait_for_sync_complete() {
  HOSTNAME=$1
  PORT=$2
  while [ ! $(wget -qO - ${HOSTNAME}:${PORT}/status | jq -r .result.syncing) = "false" ]; do 
    sleep 1; 
  done;
}

tendermint_add_validator() {
  PUBKEY=$(cat ${TMHOME}/config/priv_validator.json | jq -r .pub_key.data)
  wget -qO - ${SEED_HOSTNAME}:${TM_RPC_PORT}/broadcast_tx_commit?tx=\"val:${PUBKEY}\"
}

TYPE=${1}

if [ ! -d ${TMHOME}/config/genesis.json ]; then
  case ${TYPE} in
    genesis) 
      tendermint_init
      ;;
    secondary) 
      if [ -z ${SEED_HOSTNAME} ]; then echo "Error: env SEED_HOSTNAME is not set"; exit 1; fi
      tendermint_init
      tendermint_get_genesis_from_seed
      tendermint_wait_for_sync_complete ${SEED_HOSTNAME} ${TM_RPC_PORT}
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
fi

shift
tendermint node --consensus.create_empty_blocks=false $@
