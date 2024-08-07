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
	"encoding/json"
	"strconv"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	appTypes "github.com/ndidplatform/smart-contract/v9/abci/app/v1/types"
	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
	protoParam "github.com/ndidplatform/smart-contract/v9/protos/param"
)

type InitNDIDParam struct {
	NodeID                 string `json:"node_id"`
	SigningPublicKey       string `json:"signing_public_key"`
	SigningAlgorithm       string `json:"signing_algorithm"`
	SigningMasterPublicKey string `json:"signing_master_public_key"`
	SigningMasterAlgorithm string `json:"signing_master_algorithm"`
	EncryptionPublicKey    string `json:"encryption_public_key"`
	EncryptionAlgorithm    string `json:"encryption_algorithm"`
	ChainHistoryInfo       string `json:"chain_history_info"`
}

func (app *ABCIApplication) validateInitNDID(funcParam InitNDIDParam, callerNodeID string, committedState bool) error {
	exist, err := app.state.Has(masterNDIDKeyBytes, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if exist {
		// NDID node (first node of the network) is already existed
		return &ApplicationError{
			Code:    code.NDIDisAlreadyExisted,
			Message: "NDID node is already existed",
		}
	}

	// Validate master public key format
	err = checkPubKeyForSigning(
		funcParam.SigningMasterPublicKey,
		appTypes.SignatureAlgorithm(funcParam.SigningMasterAlgorithm),
	)
	if err != nil {
		return err
	}

	// Validate public key format
	err = checkPubKeyForSigning(
		funcParam.SigningPublicKey,
		appTypes.SignatureAlgorithm(funcParam.SigningAlgorithm),
	)
	if err != nil {
		return err
	}

	// Validate encryption public key format
	err = checkPubKeyForEncryption(funcParam.EncryptionPublicKey)
	if err != nil {
		return err
	}

	return nil
}

func (app *ABCIApplication) initNDIDCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam InitNDIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateInitNDID(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) initNDID(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("InitNDID, Parameter: %s", param)
	var funcParam InitNDIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateInitNDID(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	var nodeDetail data.NodeDetail
	nodeDetail.SigningPublicKey = &data.NodeKey{
		PublicKey:           funcParam.SigningPublicKey,
		Algorithm:           funcParam.SigningAlgorithm,
		Version:             1,
		CreationBlockHeight: app.state.CurrentBlockHeight,
		CreationChainId:     app.CurrentChain,
		Active:              true,
	}
	// create key version history
	nodeKeyKey :=
		nodeKeyKeyPrefix + keySeparator +
			"signing" + keySeparator +
			funcParam.NodeID + keySeparator +
			strconv.FormatInt(nodeDetail.SigningPublicKey.Version, 10)
	nodeKeyValue, err := utils.ProtoDeterministicMarshal(nodeDetail.SigningPublicKey)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

	nodeDetail.SigningMasterPublicKey = &data.NodeKey{
		PublicKey:           funcParam.SigningMasterPublicKey,
		Algorithm:           funcParam.SigningMasterAlgorithm,
		Version:             1,
		CreationBlockHeight: app.state.CurrentBlockHeight,
		CreationChainId:     app.CurrentChain,
		Active:              true,
	}
	// create key version history
	nodeKeyKey =
		nodeKeyKeyPrefix + keySeparator +
			"signing_master" + keySeparator +
			funcParam.NodeID + keySeparator +
			strconv.FormatInt(nodeDetail.SigningMasterPublicKey.Version, 10)
	nodeKeyValue, err = utils.ProtoDeterministicMarshal(nodeDetail.SigningMasterPublicKey)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

	nodeDetail.EncryptionPublicKey = &data.NodeKey{
		PublicKey:           funcParam.EncryptionPublicKey,
		Algorithm:           funcParam.EncryptionAlgorithm,
		Version:             1,
		CreationBlockHeight: app.state.CurrentBlockHeight,
		CreationChainId:     app.CurrentChain,
		Active:              true,
	}
	// create key version history
	nodeKeyKey =
		nodeKeyKeyPrefix + keySeparator +
			"encryption" + keySeparator +
			funcParam.NodeID + keySeparator +
			strconv.FormatInt(nodeDetail.EncryptionPublicKey.Version, 10)
	nodeKeyValue, err = utils.ProtoDeterministicMarshal(nodeDetail.EncryptionPublicKey)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

	nodeDetail.NodeName = "NDID"
	nodeDetail.Role = string(appTypes.NodeRoleNdid)
	nodeDetail.Active = true
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	chainHistoryInfoKey := "ChainHistoryInfo"
	app.state.Set(masterNDIDKeyBytes, []byte(callerNodeID))
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set(initStateKeyBytes, []byte("true"))
	app.state.Set([]byte(chainHistoryInfoKey), []byte(funcParam.ChainHistoryInfo))

	return app.NewExecTxResult(code.OK, "success", "")
}

func (app *ABCIApplication) checkCanSetInitData(committedState bool) error {
	value, err := app.state.Get(initStateKeyBytes, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if string(value) != "true" {
		return &ApplicationError{
			Code:    code.ChainIsAlreadyInitialized,
			Message: "Chain is already initialized",
		}
	}

	return nil
}

type SetInitDataParam struct {
	KVList []KeyValue `json:"kv_list"`
}

type KeyValue struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

func (app *ABCIApplication) validateSetInitData(funcParam SetInitDataParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	err = app.checkCanSetInitData(committedState)
	if err != nil {
		return err
	}

	return nil
}

func (app *ABCIApplication) setInitDataCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam SetInitDataParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetInitData(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setInitData(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetInitData, Parameter: %s", param)
	var funcParam SetInitDataParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetInitData(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	for _, kv := range funcParam.KVList {
		app.state.Set(kv.Key, kv.Value)
	}

	return app.NewExecTxResult(code.OK, "success", "")
}

func (app *ABCIApplication) validateSetInitData_pb(funcParam protoParam.SetInitDataParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	err = app.checkCanSetInitData(committedState)
	if err != nil {
		return err
	}

	return nil
}

func (app *ABCIApplication) setInitData_pbCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam protoParam.SetInitDataParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetInitData_pb(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setInitData_pb(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetInitData_pb, Parameter: %s", param)
	var funcParam protoParam.SetInitDataParam
	err := proto.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetInitData_pb(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	for _, kv := range funcParam.KvList {
		app.state.Set(kv.Key, kv.Value)
	}

	return app.NewExecTxResult(code.OK, "success", "")
}

type EndInitParam struct{}

func (app *ABCIApplication) validateEndInit(funcParam EndInitParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	return nil
}

func (app *ABCIApplication) endInitCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EndInitParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEndInit(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) endInit(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EndInit, Parameter: %s", param)
	var funcParam EndInitParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEndInit(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	app.state.Set(initStateKeyBytes, []byte("false"))
	return app.NewExecTxResult(code.OK, "success", "")
}

type SetLastBlockParam struct {
	BlockHeight int64 `json:"block_height"`
}

func (app *ABCIApplication) validateSetLastBlock(funcParam SetLastBlockParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	return nil
}

func (app *ABCIApplication) setLastBlockCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam SetLastBlockParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetLastBlock(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setLastBlock(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetLastBlock, Parameter: %s", param)
	var funcParam SetLastBlockParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetLastBlock(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	lastBlockValue := funcParam.BlockHeight
	if funcParam.BlockHeight == 0 {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	if funcParam.BlockHeight < -1 {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	if funcParam.BlockHeight > 0 && funcParam.BlockHeight < app.state.CurrentBlockHeight {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	app.state.Set(lastBlockKeyBytes, []byte(strconv.FormatInt(lastBlockValue, 10)))

	return app.NewExecTxResult(code.OK, "success", "")
}
