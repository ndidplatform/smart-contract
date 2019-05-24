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

package did

import (
	"fmt"
	"os"

	didV1 "github.com/ndidplatform/smart-contract/abci/did/v1"
	// didV2 "github.com/ndidplatform/smart-contract/abci/did2/v2"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var _ types.Application = (*DIDApplicationInterface)(nil)

type DIDApplicationInterface struct {
	appV1 *didV1.DIDApplication
	// appV2        *didV2.DIDApplication
	CurrentBlock int64
}

func NewDIDApplicationInterface() *DIDApplicationInterface {
	logger := logrus.WithFields(logrus.Fields{"module": "abci-app"})

	var dbType = getEnv("ABCI_DB_TYPE", "leveldb")
	var dbDir = getEnv("ABCI_DB_DIR_PATH", "./DID")

	if err := cmn.EnsureDir(dbDir, 0700); err != nil {
		panic(fmt.Errorf("Could not create DB directory: %v", err.Error()))
	}
	name := "didDB"
	db := dbm.NewDB(name, dbm.DBBackendType(dbType), dbDir)
	// tree := iavl.NewMutableTree(db, 0)
	// tree.Load()

	return &DIDApplicationInterface{
		appV1: didV1.NewDIDApplication(logger, db),
		// appV2: didV2.NewDIDApplication(logger, tree),
	}
}

func (app *DIDApplicationInterface) Info(req types.RequestInfo) types.ResponseInfo {
	return app.appV1.Info(req)
}

func (app *DIDApplicationInterface) SetOption(req types.RequestSetOption) types.ResponseSetOption {
	return app.appV1.SetOption(req)
}

func (app *DIDApplicationInterface) CheckTx(tx []byte) types.ResponseCheckTx {
	switch {
	case app.CurrentBlock >= 0:
		return app.appV1.CheckTx(tx)
	default:
		return app.appV1.CheckTx(tx)
	}
}

func (app *DIDApplicationInterface) DeliverTx(tx []byte) types.ResponseDeliverTx {
	switch {
	case app.CurrentBlock >= 0:
		return app.appV1.DeliverTx(tx)
	default:
		return app.appV1.DeliverTx(tx)
	}
}

func (app *DIDApplicationInterface) Commit() types.ResponseCommit {
	return app.appV1.Commit()
}

func (app *DIDApplicationInterface) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	return app.appV1.Query(reqQuery)
}

func (app *DIDApplicationInterface) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	return app.appV1.InitChain(req)
}

func (app *DIDApplicationInterface) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.CurrentBlock = req.Header.Height
	return app.appV1.BeginBlock(req)
}

func (app *DIDApplicationInterface) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return app.appV1.EndBlock(req)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
