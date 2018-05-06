#!/bin/bash
/usr/bin/tendermint unsafe_reset_all &&
/usr/bin/tendermint node --consensus.create_empty_blocks=false \
