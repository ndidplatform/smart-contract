#!/bin/sh

TMHOME=${TMHOME:-/tendermint}
TM_RPC_PORT=${TM_RPC_PORT:-45000}
TM_P2P_PORT=${TM_P2P_PORT:-47000}
ABCI_PORT=${ABCI_PORT:-46000}

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

tendermint_get_id_from_seed() {
  if [ ! -f ${TMHOME}/config/seed.host ]; then
    wget -qO - http://${SEED_HOSTNAME}:${TM_RPC_PORT}/status | jq -r .result.node_info.id > ${TMHOME}/config/seed.host
  fi
  cat ${TMHOME}/config/seed.host
}

tendermint_new_priv_validator() {
  tendermint gen_validator > ${TMHOME}/config/priv_validator.json
}

tendermint_wait_for_sync_complete() {
  local HOSTNAME=$1
  local PORT=$2
  while true; do
    [ ! "$(wget -qO - http://${HOSTNAME}:${PORT}/status | jq -r .result.sync_info.syncing)" = "false" ] || break
    sleep 1
  done;
}

tendermint_add_validator() {
  tendermint_wait_for_sync_complete localhost ${TM_RPC_PORT}
  # need to escape "/" and "+" with % encoding as pub_key.value is base64 in tendermint 0.19.5
  local PUBKEY=$(cat ${TMHOME}/config/priv_validator.json | jq -r .pub_key.value | sed 's/\//%2F/g;s/+/%2B/g')
  wget -qO - http://${SEED_HOSTNAME}:${TM_RPC_PORT}/broadcast_tx_commit?tx=\"val:${PUBKEY}\"
}

TYPE=${1}
shift

if [ ! -f ${TMHOME}/config/genesis.json ]; then
  case ${TYPE} in
    genesis) 
      tendermint_init
      sed -i 's/addr_book_strict = true/addr_book_strict = false/' ${TMHOME}/config/config.toml
      tendermint node --consensus.create_empty_blocks=false --moniker=${HOSTNAME} $@
      ;;
    secondary) 
      if [ -z ${SEED_HOSTNAME} ]; then echo "Error: env SEED_HOSTNAME is not set"; exit 1; fi
      tendermint_init
      sed -i 's/addr_book_strict = true/addr_book_strict = false/' ${TMHOME}/config/config.toml
      tendermint_wait_for_sync_complete ${SEED_HOSTNAME} ${TM_RPC_PORT}
      SEED_ID=$(tendermint_get_id_from_seed)
      tendermint_get_genesis_from_seed
      tendermint_add_validator &
      tendermint node --consensus.create_empty_blocks=false --moniker=${HOSTNAME} --p2p.seeds=${SEED_ID}@${SEED_HOSTNAME}:${TM_P2P_PORT} $@
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
  case ${TYPE} in
    genesis) 
      tendermint node --consensus.create_empty_blocks=false --moniker=${HOSTNAME} $@
      ;;
    secondary)
      SEED_ID=$(tendermint_get_id_from_seed)
      tendermint node --consensus.create_empty_blocks=false --moniker=${HOSTNAME} --p2p.seeds=${SEED_ID}@${SEED_HOSTNAME}:${TM_P2P_PORT} $@
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
