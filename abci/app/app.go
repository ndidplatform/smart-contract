/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package app

import (
	"context"
	"fmt"
	"os"
	"strconv"

	dbm "github.com/cometbft/cometbft-db"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/sirupsen/logrus"

	appV1 "github.com/ndidplatform/smart-contract/v9/abci/app/v1"
	// appV2 "github.com/ndidplatform/smart-contract/v9/abci/app2/v2"
)

type ABCIApplicationInterface struct {
	appV1 *appV1.ABCIApplication
	// appV2        *appV2.ABCIApplication
	CurrentBlockHeight int64
}

func NewABCIApplicationInterface() *ABCIApplicationInterface {
	logger := logrus.WithFields(logrus.Fields{"module": "abci-app"})

	var dbType = getEnv("ABCI_DB_TYPE", "goleveldb")
	var dbDir = getEnv("ABCI_DB_DIR_PATH", "./DID")

	if err := tmos.EnsureDir(dbDir, 0700); err != nil {
		panic(fmt.Errorf("could not create DB directory: %v", err.Error()))
	}
	name := "didDB"
	db, err := dbm.NewDB(name, dbm.BackendType(dbType), dbDir)
	if err != nil {
		panic(fmt.Errorf("could not create DB instance: %v", err.Error()))
	}

	var initialStateDir = getEnv("ABCI_INITIAL_STATE_DIR_PATH", "")

	var retainBlockCountStr = getEnv("TENDERMINT_RETAIN_BLOCK_COUNT", "")
	var retainBlockCount int64
	if retainBlockCountStr == "" {
		retainBlockCount = 0
	} else {
		retainBlockCount, err = strconv.ParseInt(retainBlockCountStr, 10, 64)
		if err != nil {
			panic(fmt.Errorf("could not parse TENDERMINT_RETAIN_BLOCK_COUNT: %v", err.Error()))
		}
	}

	return &ABCIApplicationInterface{
		appV1: appV1.NewABCIApplication(logger, db, initialStateDir, retainBlockCount),
		// appV2: appV2.NewABCIApplication(logger, db, initialStateDir, retainBlockCount),
	}
}

func (app *ABCIApplicationInterface) Info(_ context.Context, info *abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	return app.appV1.Info(info)
}

func (app *ABCIApplicationInterface) CheckTx(_ context.Context, check *abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	// IMPORTANT: Need to move app state load to this struct level if using multiple ABCI app versions
	// otherwise app.CurrentBlockHeight will always be 0 on process start
	switch {
	case app.CurrentBlockHeight >= 0:
		return app.appV1.CheckTx(check)
	default:
		return app.appV1.CheckTx(check)
	}
}

func (app *ABCIApplicationInterface) FinalizeBlock(_ context.Context, req *abcitypes.RequestFinalizeBlock) (*abcitypes.ResponseFinalizeBlock, error) {
	app.CurrentBlockHeight = req.Height
	switch {
	case app.CurrentBlockHeight >= 0:
		return app.appV1.FinalizeBlock(req)
	default:
		return app.appV1.FinalizeBlock(req)
	}
}

func (app *ABCIApplicationInterface) Commit(_ context.Context, commit *abcitypes.RequestCommit) (*abcitypes.ResponseCommit, error) {
	return app.appV1.Commit(commit)
}

func (app *ABCIApplicationInterface) Query(_ context.Context, req *abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	return app.appV1.Query(req)
}

func (app *ABCIApplicationInterface) InitChain(_ context.Context, chain *abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	return app.appV1.InitChain(chain)
}

func (app *ABCIApplicationInterface) PrepareProposal(ctx context.Context, proposal *abcitypes.RequestPrepareProposal) (*abcitypes.ResponsePrepareProposal, error) {
	return app.appV1.PrepareProposal(ctx, proposal)
}

func (app *ABCIApplicationInterface) ProcessProposal(ctx context.Context, proposal *abcitypes.RequestProcessProposal) (*abcitypes.ResponseProcessProposal, error) {
	return app.appV1.ProcessProposal(ctx, proposal)
}

func (app *ABCIApplicationInterface) ListSnapshots(ctx context.Context, snapshots *abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	return app.appV1.ListSnapshots(ctx, snapshots)
}

func (app *ABCIApplicationInterface) OfferSnapshot(ctx context.Context, snapshot *abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	return app.appV1.OfferSnapshot(ctx, snapshot)
}

func (app *ABCIApplicationInterface) LoadSnapshotChunk(ctx context.Context, chunk *abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	return app.appV1.LoadSnapshotChunk(ctx, chunk)
}

func (app *ABCIApplicationInterface) ApplySnapshotChunk(ctx context.Context, chunk *abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	return app.appV1.ApplySnapshotChunk(ctx, chunk)
}

func (app ABCIApplicationInterface) ExtendVote(ctx context.Context, extend *abcitypes.RequestExtendVote) (*abcitypes.ResponseExtendVote, error) {
	return app.appV1.ExtendVote(ctx, extend)
}

func (app *ABCIApplicationInterface) VerifyVoteExtension(ctx context.Context, verify *abcitypes.RequestVerifyVoteExtension) (*abcitypes.ResponseVerifyVoteExtension, error) {
	return app.appV1.VerifyVoteExtension(ctx, verify)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
