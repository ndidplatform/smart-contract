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

	appV1 "github.com/ndidplatform/smart-contract/v4/abci/app/v1"
	// appV2 "github.com/ndidplatform/smart-contract/v4/abci/app2/v2"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var _ types.Application = (*ABCIApplicationInterface)(nil)

type ABCIApplicationInterface struct {
	appV1 *appV1.ABCIApplication
	// appV2        *appV2.ABCIApplication
	CurrentBlock int64
}

func NewABCIApplicationInterface() *ABCIApplicationInterface {
	logger := logrus.WithFields(logrus.Fields{"module": "abci-app"})

	var dbType = getEnv("ABCI_DB_TYPE", "goleveldb")
	var dbDir = getEnv("ABCI_DB_DIR_PATH", "./DID")

	if err := cmn.EnsureDir(dbDir, 0700); err != nil {
		panic(fmt.Errorf("Could not create DB directory: %v", err.Error()))
	}
	name := "didDB"
	db := dbm.NewDB(name, dbm.DBBackendType(dbType), dbDir)
	// tree := iavl.NewMutableTree(db, 0)
	// tree.Load()

	return &ABCIApplicationInterface{
		appV1: appV1.NewABCIApplication(logger, db),
		// appV2: appV2.NewABCIApplication(logger, tree),
	}
}

func (app *ABCIApplicationInterface) Info(req types.RequestInfo) types.ResponseInfo {
	return app.appV1.Info(req)
}

func (app *ABCIApplicationInterface) SetOption(req types.RequestSetOption) types.ResponseSetOption {
	return app.appV1.SetOption(req)
}

func (app *ABCIApplicationInterface) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	switch {
	case app.CurrentBlock >= 0:
		return app.appV1.CheckTx(req)
	default:
		return app.appV1.CheckTx(req)
	}
}

func (app *ABCIApplicationInterface) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	switch {
	case app.CurrentBlock >= 0:
		return app.appV1.DeliverTx(req)
	default:
		return app.appV1.DeliverTx(req)
	}
}

func (app *ABCIApplicationInterface) Commit() types.ResponseCommit {
	return app.appV1.Commit()
}

func (app *ABCIApplicationInterface) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	return app.appV1.Query(reqQuery)
}

func (app *ABCIApplicationInterface) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	return app.appV1.InitChain(req)
}

func (app *ABCIApplicationInterface) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.CurrentBlock = req.Header.Height
	return app.appV1.BeginBlock(req)
}

func (app *ABCIApplicationInterface) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return app.appV1.EndBlock(req)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
