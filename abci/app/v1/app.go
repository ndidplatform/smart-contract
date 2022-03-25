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
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	"github.com/ndidplatform/smart-contract/v7/abci/version"
	protoTm "github.com/ndidplatform/smart-contract/v7/protos/tendermint"
)

type ABCIApplication struct {
	types.BaseApplication
	AppProtocolVersion  uint64
	CurrentChain        string
	Version             string
	checkTxNonceState   *utils.StringByteArrayMap
	deliverTxNonceState map[string][]byte
	logger              *logrus.Entry
	state               AppState
	valUpdates          map[string]types.ValidatorUpdate
	verifiedSignatures  *utils.StringMap
	lastBlockTime       time.Time
}

func NewABCIApplication(logger *logrus.Entry, db dbm.DB) *ABCIApplication {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("%s", identifyPanic())
			panic(r)
		}
	}()

	appState, err := NewAppState(db)
	if err != nil {
		panic(err)
	}

	ABCIVersion := version.Version
	ABCIProtocolVersion := version.AppProtocolVersion
	logger.Infof("Start ABCI version: %s", ABCIVersion)
	return &ABCIApplication{
		AppProtocolVersion:  ABCIProtocolVersion,
		Version:             ABCIVersion,
		checkTxNonceState:   utils.NewStringByteArrayMap(),
		deliverTxNonceState: make(map[string][]byte),
		logger:              logger,
		state:               *appState,
		valUpdates:          make(map[string]types.ValidatorUpdate),
		verifiedSignatures:  utils.NewStringMap(),
	}
}

func (app *ABCIApplication) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	var res types.ResponseInfo
	res.Version = app.Version
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	res.AppVersion = app.AppProtocolVersion
	return res
}

// Save the validators in the merkle tree
func (app *ABCIApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}

// Track the block hash and header information
func (app *ABCIApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.logger.Infof("BeginBlock: %d, Chain ID: %s", req.Header.Height, req.Header.ChainID)
	app.state.CurrentBlockHeight = req.Header.Height
	app.CurrentChain = req.Header.ChainID
	// reset valset changes
	app.valUpdates = make(map[string]types.ValidatorUpdate, 0)
	app.lastBlockTime = req.Header.Time
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *ABCIApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.logger.Infof("EndBlock: %d", req.Height)
	valUpdates := make([]types.ValidatorUpdate, 0)
	for _, newValidator := range app.valUpdates {
		valUpdates = append(valUpdates, newValidator)
	}
	return types.ResponseEndBlock{ValidatorUpdates: valUpdates}
}

func (app *ABCIApplication) DeliverTx(req types.RequestDeliverTx) (res types.ResponseDeliverTx) {
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
	publicKey, retCode, retLog := app.getNodePublicKeyForSignatureVerification(method, param, nodeID, false)
	if retCode != code.OK {
		go recordDeliverTxFailMetrics(method)
		return app.ReturnDeliverTxLog(retCode, retLog, "")
	}

	verifiedSignatureKey := string(signature) + "|" + nodeID
	verifiedSigNodePubKey, verifiedSigResultExist := app.verifiedSignatures.Load(verifiedSignatureKey)

	if verifiedSigResultExist {
		app.logger.Debugf("Found cached verified Tx signature result")
		app.verifiedSignatures.Delete(verifiedSignatureKey)
		if verifiedSigNodePubKey != publicKey {
			app.logger.Debugf("Node key updated, cached verified Tx signature result is no longer valid")
			go recordDeliverTxFailMetrics(method)
			return app.ReturnDeliverTxLog(code.VerifySignatureError, err.Error(), "")
		}
	} else {
		app.logger.Debugf("Cached verified Tx signature result could not be found")
		app.logger.Debugf("Verifying Tx signature")
		verifyResult, err := verifySignature(param, nonce, signature, publicKey, method)
		if err != nil {
			go recordDeliverTxFailMetrics(method)
			return app.ReturnDeliverTxLog(code.VerifySignatureError, err.Error(), "")
		}
		if verifyResult == false {
			go recordDeliverTxFailMetrics(method)
			return app.ReturnDeliverTxLog(code.VerifySignatureError, "Invalid Tx signature", "")
		}
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

func (app *ABCIApplication) CheckTx(req types.RequestCheckTx) (res types.ResponseCheckTx) {
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

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		go recordCheckTxDurationMetrics(duration, method)
	}()

	// ---- Check duplicate nonce ----
	nonceDup := app.isDuplicateNonce(nonce)
	if nonceDup {
		res.Code = code.DuplicateNonce
		res.Log = "Duplicate nonce"
		go recordCheckTxFailMetrics(method)
		return res
	}

	// Check duplicate nonce in checkTx state
	nonceStr := string(nonce)
	_, exist := app.checkTxNonceState.Load(nonceStr)
	if !exist {
		app.checkTxNonceState.Store(nonceStr, []byte(nil))
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
	verifiedSignatureKey := string(signature) + "|" + nodeID
	app.verifiedSignatures.Store(verifiedSignatureKey, publicKey)

	result := app.CheckTxRouter(method, param, nonce, signature, nodeID, true)
	if result.Code != code.OK {
		app.verifiedSignatures.Delete(verifiedSignatureKey)
		go recordCheckTxFailMetrics(method)
	}
	return result
}

func hash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func (app *ABCIApplication) Commit() types.ResponseCommit {
	startTime := time.Now()
	app.logger.Infof("Commit")

	app.state.Save()
	app.state.Height = app.state.Height + 1
	dbSaveDuration := time.Since(startTime)
	go recordDBSaveDurationMetrics(dbSaveDuration)

	for key := range app.deliverTxNonceState {
		app.checkTxNonceState.Delete(key)
	}
	app.deliverTxNonceState = make(map[string][]byte)

	appHashStartTime := time.Now()
	// Calculate app hash
	if len(app.state.HashData) > 0 {
		app.state.HashData = append(app.state.AppHash, app.state.HashData...)
		app.state.AppHash = hash(app.state.HashData)
	}
	appHash := app.state.AppHash
	appHashDuration := time.Since(appHashStartTime)
	go recordAppHashDurationMetrics(appHashDuration)

	app.state.HashData = make([]byte, 0)

	// Save state
	app.state.SaveMetadata()

	duration := time.Since(startTime)
	go recordCommitDurationMetrics(duration)
	return types.ResponseCommit{Data: appHash}
}

func (app *ABCIApplication) Query(reqQuery types.RequestQuery) (res types.ResponseQuery) {
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
	defer func() {
		duration := time.Since(startTime)
		go recordQueryDurationMetrics(duration, method)
	}()

	app.logger.Infof("Query: %s", method)

	height := reqQuery.Height
	if height == 0 {
		height = app.state.Height
	}

	if method == "" {
		return app.ReturnQuery(nil, "method can't be empty", app.state.Height)
	}
	return app.QueryRouter(method, param, height)
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
