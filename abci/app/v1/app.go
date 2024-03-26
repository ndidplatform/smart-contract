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
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	"github.com/ndidplatform/smart-contract/v9/abci/version"
	protoTm "github.com/ndidplatform/smart-contract/v9/protos/tendermint"
)

type ABCIApplication struct {
	abcitypes.BaseApplication
	AppProtocolVersion  uint64
	CurrentChain        string
	Version             string
	checkTxNonceState   *utils.StringByteArrayMap
	deliverTxNonceState map[string][]byte
	logger              *logrus.Entry
	state               AppState
	valUpdates          map[string]abcitypes.ValidatorUpdate
	verifiedSignatures  *utils.StringMap
	lastBlockTime       time.Time
	initialStateDir     string
	retainBlockCount    int64
}

func NewABCIApplication(logger *logrus.Entry, db dbm.DB, initialStateDir string, retainBlockCount int64) *ABCIApplication {
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
		CurrentChain:        appState.ChainID,
		checkTxNonceState:   utils.NewStringByteArrayMap(),
		deliverTxNonceState: make(map[string][]byte),
		logger:              logger,
		state:               *appState,
		valUpdates:          make(map[string]abcitypes.ValidatorUpdate),
		verifiedSignatures:  utils.NewStringMap(),
		initialStateDir:     initialStateDir,
		retainBlockCount:    retainBlockCount,
	}
}

type InfoData struct {
	InitialStateDataLoaded bool `json:"initial_state_data_loaded"`
}

func (app *ABCIApplication) Info(info *abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	var res *abcitypes.ResponseInfo = &abcitypes.ResponseInfo{}
	res.Version = app.Version
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash

	infoData := &InfoData{
		InitialStateDataLoaded: app.state.InitialStateDataLoaded,
	}
	infoDataBytes, err := json.Marshal(infoData)
	if err != nil {
		app.logger.Warnf("ABCI Info: JSON marshal err: %+v", err)
		res.Data = "{}"
	} else {
		res.Data = string(infoDataBytes)
	}

	res.AppVersion = app.AppProtocolVersion

	return res, nil
}

// Save the validators in the merkle tree
func (app *ABCIApplication) InitChain(chain *abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	app.logger.Infof("InitChain: %s", chain.ChainId)

	// load initial state data from file if provided
	if app.initialStateDir != "" {
		app.logger.Infof("Loading initial state data from directory: %s", app.initialStateDir)

		hash, err := app.state.LoadInitialState(app.logger, app.initialStateDir)
		if err != nil {
			panic(err)
		}

		app.state.HasHashData = true
		app.state.HashDigest.Write(hash)

		app.state.InitialStateDataLoaded = true
	} else {
		app.logger.Infof("No initial state data provided")
	}

	app.CurrentChain = chain.ChainId
	app.state.ChainID = chain.ChainId

	for _, v := range chain.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}

	if app.state.HasHashData {
		app.state.AppHash = app.state.HashDigest.Sum(nil)
	}
	// Save state
	app.state.Save()

	return &abcitypes.ResponseInitChain{
		AppHash: app.state.AppHash,
	}, nil
}

func (app *ABCIApplication) FinalizeBlock(req *abcitypes.RequestFinalizeBlock) (*abcitypes.ResponseFinalizeBlock, error) {
	app.logger.Infof("FinalizeBlock: %d, Txs count: %d", req.Height, len(req.Txs))

	app.state.CurrentBlockHeight = req.Height

	app.state.HasHashData = false
	app.state.HashDigest = sha256.New()
	app.state.HashDigest.Write(app.state.AppHash)

	// FIXME: other ways to get chain ID? or there's no need to check?
	// if app.state.ChainID != req.Header.ChainID {
	// 	panic(errors.New("chain ID mismatch (ABCI state != Tendermint)"))
	// }

	app.lastBlockTime = req.Time

	/*
	 * execute transactions
	 */

	var txs = make([]*abcitypes.ExecTxResult, len(req.Txs))

	for i, tx := range req.Txs {
		execTxResult, err := app.execTx(tx)
		if err != nil {
			app.logger.Errorf("execTx err: %+v", err)
			execTxResult = app.NewExecTxResult(code.UnknownError, "Unknown error", "")
		}

		txs[i] = execTxResult
	}

	/*
	 * app hash
	 */

	appHashCalcStartTime := time.Now()
	// Calculate app hash
	if app.state.HasHashData {
		app.state.AppHash = app.state.HashDigest.Sum(nil)
	}
	appHash := app.state.AppHash
	appHashCalcDuration := time.Since(appHashCalcStartTime)
	go recordAppHashDurationMetrics(appHashCalcDuration)

	/*
	 * validator updates
	 */

	valUpdates := make([]abcitypes.ValidatorUpdate, 0)
	for _, newValidator := range app.valUpdates {
		valUpdates = append(valUpdates, newValidator)
	}
	// reset valset changes
	app.valUpdates = make(map[string]abcitypes.ValidatorUpdate, 0)

	return &abcitypes.ResponseFinalizeBlock{
		AppHash:          appHash,
		TxResults:        txs,
		ValidatorUpdates: valUpdates,
	}, nil
}

func (app *ABCIApplication) execTx(tx []byte) (*abcitypes.ExecTxResult, error) {
	var res *abcitypes.ExecTxResult

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = app.NewExecTxResult(code.UnknownError, "Unknown error", "")
		}
	}()

	var txObj protoTm.Tx
	err := proto.Unmarshal(tx, &txObj)
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

	if mustCheckNodeSignature(method) {
		// ---- Check duplicate nonce ----
		nonceDup := app.isDuplicateNonce(nonce, false)
		if nonceDup {
			go recordDeliverTxFailMetrics(method)
			return app.NewExecTxResult(code.DuplicateNonce, "Duplicate nonce", ""), nil
		}
	}

	app.logger.Infof("DeliverTx: %s, NodeID: %s", method, nodeID)

	if method == "" {
		go recordDeliverTxFailMetrics(method)
		return app.NewExecTxResult(code.MethodCanNotBeEmpty, "method can not be empty", ""), nil
	}

	if mustCheckNodeSignature(method) {
		// Check signature
		publicKey, signingAlgorithm, retCode, retLog := app.getNodePublicKeyForSignatureVerification(method, param, nodeID, false)
		if retCode != code.OK {
			go recordDeliverTxFailMetrics(method)
			return app.NewExecTxResult(retCode, retLog, ""), nil
		}

		verifiedSignatureKey := string(signature) + "|" + nodeID
		verifiedSigNodePubKey, verifiedSigResultExist := app.verifiedSignatures.Load(verifiedSignatureKey)

		if verifiedSigResultExist {
			app.logger.Debugf("Found cached verified Tx signature result")
			app.verifiedSignatures.Delete(verifiedSignatureKey)
			if verifiedSigNodePubKey != publicKey {
				app.logger.Debugf("Node key updated, cached verified Tx signature result is no longer valid")
				go recordDeliverTxFailMetrics(method)
				return app.NewExecTxResult(code.VerifySignatureError, err.Error(), ""), nil
			}
		} else {
			app.logger.Debugf("Cached verified Tx signature result could not be found")
			app.logger.Debugf("Verifying Tx signature")
			verifyResult, err := verifySignature(method, param, app.CurrentChain, nonce, signature, publicKey, signingAlgorithm)
			if err != nil {
				go recordDeliverTxFailMetrics(method)
				return app.NewExecTxResult(code.VerifySignatureError, err.Error(), ""), nil
			}
			if !verifyResult {
				go recordDeliverTxFailMetrics(method)
				return app.NewExecTxResult(code.VerifySignatureError, "Invalid Tx signature", ""), nil
			}
		}
	}

	res = app.DeliverTxRouter(method, param, nonce, signature, nodeID)
	app.logger.Infof(
		`DeliverTx response: {"code":%d,"log":"%s","attributes":[{"key":"%s","value":"%s"}]}`,
		res.Code,
		res.Log,
		string(res.Events[0].Attributes[0].Key), string(res.Events[0].Attributes[0].Value),
	)
	if res.Code != code.OK {
		go recordDeliverTxFailMetrics(method)
	}

	return res, nil
}

func (app *ABCIApplication) CheckTx(check *abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	var res *abcitypes.ResponseCheckTx

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = NewResponseCheckTx(code.UnknownError, "Unknown error")
		}
	}()

	var txObj protoTm.Tx
	err := proto.Unmarshal(check.Tx, &txObj)
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

	if mustCheckNodeSignature(method) {
		// ---- Check duplicate nonce ----
		nonceDup := app.isDuplicateNonce(nonce, true)
		if nonceDup {
			res.Code = code.DuplicateNonce
			res.Log = "Duplicate nonce"
			go recordCheckTxFailMetrics(method)
			return res, nil
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
			return res, nil
		}
	}

	app.logger.Infof("CheckTx: %s, NodeID: %s", method, nodeID)

	if method == "" || param == nil || nodeID == "" {
		res.Code = code.InvalidTransactionFormat
		res.Log = "Invalid transaction format"
		go recordCheckTxFailMetrics(method)
		return res, nil
	}
	if mustCheckNodeSignature(method) {
		if nonce == nil || signature == nil {
			res.Code = code.InvalidTransactionFormat
			res.Log = "Invalid transaction format"
			go recordCheckTxFailMetrics(method)
			return res, nil
		}
	}

	// Check has function in system
	if !IsMethod[method] {
		res.Code = code.UnknownMethod
		res.Log = "Unknown method name"
		go recordCheckTxFailMetrics(method)
		return res, nil
	}

	verifiedSignatureKey := string(signature) + "|" + nodeID
	if mustCheckNodeSignature(method) {
		// Check signature
		publicKey, signingAlgorithm, retCode, retLog := app.getNodePublicKeyForSignatureVerification(method, param, nodeID, true)
		if retCode != code.OK {
			return NewResponseCheckTx(retCode, retLog), nil
		}

		verifyResult, err := verifySignature(method, param, app.CurrentChain, nonce, signature, publicKey, signingAlgorithm)
		if err != nil {
			go recordCheckTxFailMetrics(method)
			return NewResponseCheckTx(code.VerifySignatureError, err.Error()), nil
		}
		if !verifyResult {
			go recordCheckTxFailMetrics(method)
			return NewResponseCheckTx(code.VerifySignatureError, "Invalid Tx signature"), nil
		}
		app.verifiedSignatures.Store(verifiedSignatureKey, publicKey)
	}

	result := app.CheckTxRouter(method, param, nonce, signature, nodeID, true)
	if result.Code != code.OK {
		app.verifiedSignatures.Delete(verifiedSignatureKey)
		go recordCheckTxFailMetrics(method)
	}

	return result, nil
}

func (app *ABCIApplication) Commit(commit *abcitypes.RequestCommit) (*abcitypes.ResponseCommit, error) {
	startTime := time.Now()
	app.logger.Infof("Commit")

	app.state.Height = app.state.Height + 1
	app.state.Save()
	dbSaveDuration := time.Since(startTime)
	go recordDBSaveDurationMetrics(dbSaveDuration)

	for key := range app.deliverTxNonceState {
		app.checkTxNonceState.Delete(key)
	}
	app.deliverTxNonceState = make(map[string][]byte)

	duration := time.Since(startTime)
	go recordCommitDurationMetrics(duration)

	var retainHeight int64 = 0
	if app.retainBlockCount > 0 && app.state.CurrentBlockHeight > app.retainBlockCount {
		retainHeight = app.state.CurrentBlockHeight - app.retainBlockCount
	}

	return &abcitypes.ResponseCommit{
		RetainHeight: retainHeight,
	}, nil
}

func (app *ABCIApplication) Query(req *abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	var res *abcitypes.ResponseQuery

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorf("Recovered in %s, %s", r, identifyPanic())
			res = app.NewResponseQuery(nil, "Unknown error", app.state.Height)
		}
	}()

	var query protoTm.Query
	err := proto.Unmarshal(req.Data, &query)
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

	height := req.Height
	if height == 0 {
		height = app.state.Height
	}

	if method == "" {
		return app.NewResponseQuery(nil, "method can't be empty", app.state.Height), nil
	}

	res = app.QueryRouter(method, param, height)

	return res, nil
}

func mustCheckNodeSignature(method string) bool {
	switch method {
	case "SetInitData":
	case "SetInitData_pb":
		return false
	}
	return true
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
