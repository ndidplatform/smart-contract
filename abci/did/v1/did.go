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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/ndidplatform/smart-contract/v4/abci/code"
	"github.com/ndidplatform/smart-contract/v4/abci/version"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"

	protoTm "github.com/ndidplatform/smart-contract/v4/protos/tendermint"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	stateKey = []byte("stateKey")
	// nonceKeyPrefix  = []byte("nonce:")
)

type State struct {
	db      dbm.DB
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

// func prefixNonceKey(nonceKey []byte) []byte {
// 	return append(nonceKeyPrefix, nonceKey...)
// }

var _ types.Application = (*DIDApplication)(nil)

type DIDApplication struct {
	types.BaseApplication
	state                    State
	checkTxNonceState        map[string][]byte
	deliverTxNonceState      map[string][]byte
	ValUpdates               map[string]types.ValidatorUpdate
	logger                   *logrus.Entry
	Version                  string
	AppProtocolVersion       uint64
	CurrentBlock             int64
	CurrentChain             string
	HashData                 []byte
	UncommittedState         map[string][]byte
	UncommittedVersionsState map[string][]int64
	verifiedSignatures       map[string]string
	nodeKeyUpdate            map[string]bool
}

func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

func NewDIDApplication(logger *logrus.Entry, db dbm.DB) *DIDApplication {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("%s", identifyPanic())
			panic(r)
		}
	}()
	// var state State
	// state.db = db
	state := loadState(db)

	ABCIVersion := version.Version
	ABCIProtocolVersion := version.AppProtocolVersion
	logger.Infof("Start ABCI version: %s", ABCIVersion)
	return &DIDApplication{
		state:                    state,
		checkTxNonceState:        make(map[string][]byte),
		deliverTxNonceState:      make(map[string][]byte),
		logger:                   logger,
		Version:                  ABCIVersion,
		AppProtocolVersion:       ABCIProtocolVersion,
		UncommittedState:         make(map[string][]byte),
		UncommittedVersionsState: make(map[string][]int64),
		ValUpdates:               make(map[string]types.ValidatorUpdate),
		verifiedSignatures:       make(map[string]string),
		nodeKeyUpdate:            make(map[string]bool),
	}
}

func (app *DIDApplication) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	var res types.ResponseInfo
	res.Version = app.Version
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	res.AppVersion = app.AppProtocolVersion
	app.CurrentBlock = app.state.Height
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
	app.logger.Infof("BeginBlock: %d, Chain ID: %s", req.Header.Height, req.Header.ChainID)
	app.CurrentBlock = req.Header.Height
	app.CurrentChain = req.Header.ChainID
	// reset valset changes
	app.ValUpdates = make(map[string]types.ValidatorUpdate, 0)
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *DIDApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.logger.Infof("EndBlock: %d", req.Height)
	valUpdates := make([]types.ValidatorUpdate, 0)
	for _, newValidator := range app.ValUpdates {
		valUpdates = append(valUpdates, newValidator)
	}
	return types.ResponseEndBlock{ValidatorUpdates: valUpdates}
}

func (app *DIDApplication) DeliverTx(req types.RequestDeliverTx) (res types.ResponseDeliverTx) {

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = app.ReturnDeliverTxLog(code.UnknownError, "Unknown error", "")
		}
	}()

	var txObj protoTm.Tx
	err := proto.Unmarshal(req.Tx, &txObj)
	if err != nil {
		app.logger.Error(err.Error())
	}

	method := txObj.Method
	param := txObj.Params
	nonce := txObj.Nonce
	signature := txObj.Signature
	nodeID := txObj.NodeId

	go recordDeliverTxMetrics(method)

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		go recordDeliverTxDurationMetrics(duration, method)
	}()
	// ---- Check duplicate nonce ----
	nonceDup := app.isDuplicateNonce(nonce)
	if nonceDup {
		go recordDeliverTxFailMetrics(method)
		return app.ReturnDeliverTxLog(code.DuplicateNonce, "Duplicate nonce", "")
	}

	app.logger.Infof("DeliverTx: %s, NodeID: %s", method, nodeID)

	if method == "" {
		go recordDeliverTxFailMetrics(method)
		return app.ReturnDeliverTxLog(code.MethodCanNotBeEmpty, "method can not be empty", "")
	}

	// Check signature
	_, nodeKeyUpdateExist := app.nodeKeyUpdate[nodeID]

	signatureStr := string(signature)
	verifiedSigNodeID, verifiedSigResultExist := app.verifiedSignatures[signatureStr]

	var signatureReverificationNeeded bool
	if nodeKeyUpdateExist {
		app.logger.Debugf("Node key updated, cached verified Tx signature result will not be used")
		signatureReverificationNeeded = true
	} else if !verifiedSigResultExist {
		app.logger.Debugf("Cached verified Tx signature result could not be found")
		signatureReverificationNeeded = true
	} else {
		signatureReverificationNeeded = false
	}

	if signatureReverificationNeeded {
		app.logger.Debugf("Verifying Tx signature")
		publicKey, retCode, retLog := app.getNodePublicKeyForSignatureVerification(method, param, nodeID, false)
		if retCode != code.OK {
			go recordDeliverTxFailMetrics(method)
			return app.ReturnDeliverTxLog(retCode, retLog, "")
		}
		verifyResult, err := verifySignature(param, nonce, signature, publicKey, method)
		if err != nil {
			go recordDeliverTxFailMetrics(method)
			return app.ReturnDeliverTxLog(code.VerifySignatureError, err.Error(), "")
		}
		if verifyResult == false {
			go recordDeliverTxFailMetrics(method)
			return app.ReturnDeliverTxLog(code.VerifySignatureError, "Invalid Tx signature", "")
		}
	} else {
		app.logger.Debugf("Found verified Tx signature result")
		if verifiedSigNodeID != nodeID {
			go recordDeliverTxFailMetrics(method)
			return app.ReturnDeliverTxLog(code.VerifySignatureError, err.Error(), "")
		}
	}

	if verifiedSigResultExist {
		delete(app.verifiedSignatures, signatureStr)
	}

	result := app.DeliverTxRouter(method, param, nonce, signature, nodeID)
	app.logger.Infof(
		`DeliverTx response: {"code":%d,"log":"%s","attributes":[{"key":"%s","value":"%s"}]}`,
		result.Code,
		result.Log,
		string(result.Events[0].Attributes[0].Key), string(result.Events[0].Attributes[0].Value),
	)
	if result.Code != code.OK {
		go recordDeliverTxFailMetrics(method)
	}
	return result
}

func (app *DIDApplication) CheckTx(req types.RequestCheckTx) (res types.ResponseCheckTx) {
	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = ReturnCheckTx(code.UnknownError, "Unknown error")
		}
	}()

	var txObj protoTm.Tx
	err := proto.Unmarshal(req.Tx, &txObj)
	if err != nil {
		app.logger.Error(err.Error())
	}

	method := txObj.Method
	param := txObj.Params
	nonce := txObj.Nonce
	signature := txObj.Signature
	nodeID := txObj.NodeId

	go recordCheckTxMetrics(method)

	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		go recordCheckTxDurationMetrics(duration, method)
	}()

	// TODO: Check for not enough token here as well to exclude those Txs from going into DeliverTx
	// Set checkTx state for each node's available token or token difference
	// Deduct used token if passed
	// Error response if not enough
	// Adjust difference on Commit()

	// ---- Check duplicate nonce ----
	nonceDup := app.isDuplicateNonce(nonce)
	if nonceDup {
		res.Code = code.DuplicateNonce
		res.Log = "Duplicate nonce"
		go recordCheckTxFailMetrics(method)
		return res
	}

	// Check duplicate nonce in checkTx stateDB
	_, exist := app.checkTxNonceState[nonceBase64]
	if !exist {
		app.checkTxNonceState[nonceBase64] = []byte(nil)
	} else {
		res.Code = code.DuplicateNonce
		res.Log = "Duplicate nonce"
		go recordCheckTxFailMetrics(method)
		return res
	}

	app.logger.Infof("CheckTx: %s, NodeID: %s", method, nodeID)

	if method == "" || param == "" || nonce == nil || signature == nil || nodeID == "" {
		res.Code = code.InvalidTransactionFormat
		res.Log = "Invalid transaction format"
		go recordCheckTxFailMetrics(method)
		return res
	}

	// Check has function in system
	if !IsMethod[method] {
		res.Code = code.UnknownMethod
		res.Log = "Unknown method name"
		go recordCheckTxFailMetrics(method)
		return res
	}

	// Check signature
	publicKey, retCode, retLog := app.getNodePublicKeyForSignatureVerification(method, param, nodeID, true)
	if retCode != code.OK {
		return ReturnCheckTx(retCode, retLog)
	}

	verifyResult, err := verifySignature(param, nonce, signature, publicKey, method)
	if err != nil {
		go recordCheckTxFailMetrics(method)
		return ReturnCheckTx(code.VerifySignatureError, err.Error())
	}
	if verifyResult == false {
		go recordCheckTxFailMetrics(method)
		return ReturnCheckTx(code.VerifySignatureError, "Invalid Tx signature")
	}
	app.verifiedSignatures[string(signature)] = nodeID

	result := app.CheckTxRouter(method, param, nonce, signature, nodeID)
	if result.Code != code.OK {
		delete(app.verifiedSignatures, string(signature))
		go recordCheckTxFailMetrics(method)
	}
	return result
}

func hash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func (app *DIDApplication) Commit() types.ResponseCommit {
	startTime := time.Now()
	app.logger.Infof("Commit")

	app.SaveDBState()
	app.state.Height = app.state.Height + 1
	go recordDBSaveDurationMetrics(startTime)

	for key := range app.deliverTxNonceState {
		delete(app.checkTxNonceState, key)
	}
	app.deliverTxNonceState = make(map[string][]byte)
	app.nodeKeyUpdate = make(map[string]bool)

	appHashStartTime := time.Now()
	// Calculate app hash
	if len(app.HashData) > 0 {
		app.HashData = append(app.state.AppHash, app.HashData...)
		app.state.AppHash = hash(app.HashData)
	}
	appHash := app.state.AppHash
	go recordAppHashDurationMetrics(appHashStartTime)

	app.HashData = make([]byte, 0)

	// Save state
	saveState(app.state)

	go recordCommitDurationMetrics(startTime)
	return types.ResponseCommit{Data: appHash}
}

func (app *DIDApplication) Query(reqQuery types.RequestQuery) (res types.ResponseQuery) {

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = app.ReturnQuery(nil, "Unknown error", app.state.Height)
		}
	}()

	var query protoTm.Query
	err := proto.Unmarshal(reqQuery.Data, &query)
	if err != nil {
		app.logger.Error(err.Error())
	}

	method := query.Method
	param := query.Params

	startTime := time.Now()
	go recordQueryMetrics(method)

	app.logger.Infof("Query: %s", method)

	height := reqQuery.Height
	if height == 0 {
		height = app.state.Height
	}

	if method != "" {
		go recordQueryDurationMetrics(startTime, method)
		return app.QueryRouter(method, param, height)
	}
	go recordQueryDurationMetrics(startTime, method)
	return app.ReturnQuery(nil, "method can't empty", app.state.Height)
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
