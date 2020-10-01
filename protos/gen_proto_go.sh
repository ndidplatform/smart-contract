#!/bin/sh

protoc -I=./data --go_out=./data ./data/data.proto
protoc -I=./tendermint --go_out=./tendermint ./tendermint/tendermint.proto