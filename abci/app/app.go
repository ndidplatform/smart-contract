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
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	appV1 "github.com/ndidplatform/smart-contract/v7/abci/app/v1"
	// appV2 "github.com/ndidplatform/smart-contract/v7/abci/app2/v2"
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
		panic(fmt.Errorf("Could not create DB directory: %v", err.Error()))
	}
	name := "didDB"
	db, err := dbm.NewDB(name, dbm.BackendType(dbType), dbDir)
	if err != nil {
		panic(fmt.Errorf("Could not create DB instance: %v", err.Error()))
	}

	return &ABCIApplicationInterface{
		appV1: appV1.NewABCIApplication(logger, db),
		// appV2: appV2.NewABCIApplication(logger, db),
	}
}

func (app *ABCIApplicationInterface) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return app.appV1.Info(req)
}

func (app *ABCIApplicationInterface) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	// IMPORTANT: Need to move app state load to this struct level if using multiple ABCI app versions
	// otherwise app.CurrentBlockHeight will always be 0 on process start
	switch {
	case app.CurrentBlockHeight >= 0:
		return app.appV1.CheckTx(req)
	default:
		return app.appV1.CheckTx(req)
	}
}

func (app *ABCIApplicationInterface) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	switch {
	case app.CurrentBlockHeight >= 0:
		return app.appV1.DeliverTx(req)
	default:
		return app.appV1.DeliverTx(req)
	}
}

func (app *ABCIApplicationInterface) Commit() abcitypes.ResponseCommit {
	return app.appV1.Commit()
}

func (app *ABCIApplicationInterface) Query(reqQuery abcitypes.RequestQuery) abcitypes.ResponseQuery {
	return app.appV1.Query(reqQuery)
}

func (app *ABCIApplicationInterface) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return app.appV1.InitChain(req)
}

func (app *ABCIApplicationInterface) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.CurrentBlockHeight = req.Header.Height
	return app.appV1.BeginBlock(req)
}

func (app *ABCIApplicationInterface) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return app.appV1.EndBlock(req)
}

func (app *ABCIApplicationInterface) ListSnapshots(req abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return app.appV1.ListSnapshots(req)
}

func (app *ABCIApplicationInterface) OfferSnapshot(req abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return app.appV1.OfferSnapshot(req)
}

func (app *ABCIApplicationInterface) LoadSnapshotChunk(req abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return app.appV1.LoadSnapshotChunk(req)
}

func (app *ABCIApplicationInterface) ApplySnapshotChunk(req abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return app.appV1.ApplySnapshotChunk(req)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
