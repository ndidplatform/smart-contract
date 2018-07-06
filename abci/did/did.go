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
	"encoding/base64"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	stateKey        = []byte("stateKey")
	kvPairPrefixKey = []byte("kvPairKey:")
)

type State struct {
	db *iavl.VersionedTree
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

var _ types.Application = (*DIDApplication)(nil)

type DIDApplication struct {
	types.BaseApplication
	state      State
	ValUpdates []types.Validator
	logger     *logrus.Entry
	Version    string
}

func NewDIDApplication() *DIDApplication {
	logger := logrus.WithFields(logrus.Fields{"module": "abci-app"})
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("%s", identifyPanic())
			panic(r)
		}
	}()
	logger.Infoln("NewDIDApplication")
	var dbDir = getEnv("DB_NAME", "DID")
	name := "didDB"
	db := dbm.NewDB(name, "leveldb", dbDir)
	tree := iavl.NewVersionedTree(db, 0)
	tree.Load()
	var state State
	state.db = tree
	return &DIDApplication{state: state,
		logger:  logger,
		Version: "0.2.0", // Hard code set version
	}
}

func (app *DIDApplication) SetStateDB(key, value []byte) {
	app.state.db.Set(prefixKey(key), value)
}

func (app *DIDApplication) DeleteStateDB(key []byte) {
	app.state.db.Remove(prefixKey(key))
}

func (app *DIDApplication) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	var res types.ResponseInfo
	res.Version = app.Version
	res.LastBlockHeight = app.state.db.Version64()
	res.LastBlockAppHash = app.state.db.Hash()
	return res
}

// Save the validators in the merkle tree
func (app *DIDApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}

// Track the block hash and header information
func (app *DIDApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.logger.Infof("BeginBlock: %d", req.Header.Height)
	// reset valset changes
	app.ValUpdates = make([]types.Validator, 0)
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *DIDApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.logger.Infof("EndBlock: %d", req.Height)
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

func (app *DIDApplication) DeliverTx(tx []byte) (res types.ResponseDeliverTx) {
	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = ReturnDeliverTxLog(code.WrongTransactionFormat, "wrong transaction format", "")
		}
	}()

	txString := string(tx)
	parts := strings.Split(string(txString), "|")

	paramByte, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		app.logger.Error(err.Error())
		return ReturnDeliverTxLog(code.DecodingError, err.Error(), "")
	}
	nodeIDByte, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		app.logger.Error(err.Error())
		return ReturnDeliverTxLog(code.DecodingError, err.Error(), "")
	}

	method := string(parts[0])
	param := string(paramByte)
	nonce := string(parts[2])
	signature := string(parts[3])
	nodeID := string(nodeIDByte)

	app.logger.Infof("DeliverTx: %s, NodeID: %s", method, nodeID)

	if method != "" {
		return DeliverTxRouter(method, param, nonce, signature, nodeID, app)
	}
	return ReturnDeliverTxLog(code.MethodCanNotBeEmpty, "method can not be empty", "")
}

func (app *DIDApplication) CheckTx(tx []byte) (res types.ResponseCheckTx) {
	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = ReturnCheckTx(false)
		}
	}()

	parts := strings.Split(string(tx), "|")
	paramByte, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		app.logger.Error(err.Error())
		return ReturnCheckTx(false)
	}
	nodeIDByte, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		app.logger.Error(err.Error())
		return ReturnCheckTx(false)
	}

	method := string(parts[0])
	param := string(paramByte)
	nonce := string(parts[2])
	signature := string(parts[3])
	nodeID := string(nodeIDByte)

	app.logger.Infof("CheckTx: %s, NodeID: %s", method, nodeID)

	if method != "" && param != "" && nonce != "" && signature != "" && nodeID != "" {
		// Check has function in system
		if IsMethod[method] {
			return ReturnCheckTx(true)
		}
		res.Code = code.Unauthorized
		res.Log = "Invalid method name"
		return res
	}
	res.Code = code.Unauthorized
	res.Log = "Invalid transaction format"
	return res
}

func (app *DIDApplication) Commit() types.ResponseCommit {
	app.logger.Infof("Commit")
	app.state.db.SaveVersion()
	return types.ResponseCommit{Data: app.state.db.Hash()}
}

func (app *DIDApplication) Query(reqQuery types.RequestQuery) (res types.ResponseQuery) {

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = ReturnQuery(nil, "wrong query format", app.state.db.Version64(), app)
		}
	}()

	txString, err := base64.StdEncoding.DecodeString(string(reqQuery.Data))
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]

	app.logger.Infof("Query: %s", method)

	height := reqQuery.Height
	if height == 0 {
		height = app.state.db.Version64()
	}

	if method != "" {
		return QueryRouter(method, param, app, height)
	}
	return ReturnQuery(nil, "method can't empty", app.state.db.Version64(), app)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}

	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}

	return fmt.Sprintf("pc:%x", pc)
}
