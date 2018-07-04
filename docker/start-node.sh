#!/bin/sh

TMHOME=${TMHOME:-/tendermint}
TM_RPC_PORT=${TM_RPC_PORT:-45000}
TM_P2P_PORT=${TM_P2P_PORT:-47000}
ABCI_PORT=${ABCI_PORT:-46000}

if [ -z "${SEED_RPC_PORT}" ]; then SEED_RPC_PORT=$TM_RPC_PORT; fi

usage() {
  echo "Usage: $(basename ${0}) <mode>"
  echo "where mode can be :"
  echo "genesis = run this node as genesis node"
  echo "secondary = run this node as secondary node"
  echo "reset = call unsafe_reset_all"
}

tendermint_init() {
  echo "Initializing tendermint..."
  tendermint init --home=${TMHOME}
}

tendermint_reset() {
  echo "Resetting tendermint..."
  tendermint --home=${TMHOME} unsafe_reset_all
}

tendermint_get_genesis_from_seed() {
  curl -s http://${SEED_HOSTNAME}:${SEED_RPC_PORT}/genesis | jq -r .result.genesis > ${TMHOME}/config/genesis.json
}

tendermint_get_id_from_seed() {
  if [ ! -f ${TMHOME}/config/seed.host ]; then
    curl -s http://${SEED_HOSTNAME}:${SEED_RPC_PORT}/status | jq -r .result.node_info.id | tee ${TMHOME}/config/seed.host
  else
    cat ${TMHOME}/config/seed.host
  fi
}

tendermint_wait_for_sync_complete() {
  echo "Waiting for tendermint at ${1}:${2} to be ready..."
  while true; do 
    [ ! "$(curl -s http://${1}:${2}/status | jq -r .result.sync_info.catching_up)" = "false" ] || break  
    sleep 1
  done
}

tendermint_set_addr_book_strict() {
  sed -iE "s/addr_book_strict = (true|false)/addr_book_strict = ${1}/" ${TMHOME}/config/config.toml
}

TYPE=${1}
shift

if [ ! -f ${TMHOME}/config/genesis.json ]; then
  case ${TYPE} in
    genesis) 
      tendermint_init
      tendermint_set_addr_book_strict false
      tendermint node --consensus.create_empty_blocks=false --moniker=${HOSTNAME} $@
      ;;
    secondary) 
      if [ -z ${SEED_HOSTNAME} ]; then echo "Error: env SEED_HOSTNAME is not set"; exit 1; fi

      tendermint_init
      tendermint_set_addr_book_strict false
      until tendermint_wait_for_sync_complete ${SEED_HOSTNAME} ${SEED_RPC_PORT}; do sleep 1; done
      until SEED_ID=$(tendermint_get_id_from_seed) && [ ! "${SEED_ID}" = "" ]; do sleep 1; done
      until tendermint_get_genesis_from_seed; do sleep 1; done
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
      until SEED_ID=$(tendermint_get_id_from_seed); do sleep 1; done
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
