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

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type AddNodeToYourDataNodeWhitelistParam struct {
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateAddNodeToYourDataNodeWhitelist(funcParam AddNodeToYourDataNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
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

	if checktx {
		return nil
	}

	// stateful

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if nodeDetailValue == nil {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}

	key := yourDataNodeWhitelistKeyPrefix + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if exists {
		return &ApplicationError{
			Code:    code.DuplicateNodeID,
			Message: "Duplicate node ID",
		}
	}

	return nil
}

func (app *ABCIApplication) addNodeToYourDataNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddNodeToYourDataNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddNodeToYourDataNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addNodeToYourDataNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddNodeToYourDataNodeWhitelist, Parameter: %s", param)
	var funcParam AddNodeToYourDataNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddNodeToYourDataNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := yourDataNodeWhitelistKeyPrefix + keySeparator + funcParam.NodeID

	var nodePermission data.YourDataNodePermission

	value, err := utils.ProtoDeterministicMarshal(&nodePermission)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveNodeFromYourDataNodeWhitelistParam struct {
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateRemoveNodeFromYourDataNodeWhitelist(funcParam RemoveNodeFromYourDataNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
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

	if checktx {
		return nil
	}

	// stateful

	key := yourDataNodeWhitelistKeyPrefix + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !exists {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "node ID not found",
		}
	}

	return nil
}

func (app *ABCIApplication) removeNodeFromYourDataNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveNodeFromYourDataNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveNodeFromYourDataNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) removeNodeFromYourDataNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveNodeFromYourDataNodeWhitelist, Parameter: %s", param)
	var funcParam RemoveNodeFromYourDataNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveNodeFromYourDataNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := yourDataNodeWhitelistKeyPrefix + keySeparator + funcParam.NodeID

	app.state.Delete([]byte(key))

	return app.NewExecTxResult(code.OK, "success", "")
}

//

type EnableYourDataNodeWhitelistParam struct {
}

func (app *ABCIApplication) validateEnableYourDataNodeWhitelist(funcParam EnableYourDataNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
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

	if checktx {
		return nil
	}

	// stateful

	value, err := app.state.Get(yourDataNodeWhitelistMetadataKey, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var yourDataNodeWhitelist data.YourDataNodeWhitelist
	err = proto.Unmarshal(value, &yourDataNodeWhitelist)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if yourDataNodeWhitelist.Active {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already active/enabled",
		}
	}

	return nil
}

func (app *ABCIApplication) enableYourDataNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableYourDataNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableYourDataNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableYourDataNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableYourDataNodeWhitelist, Parameter: %s", param)
	var funcParam EnableYourDataNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableYourDataNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	value, err := app.state.Get(yourDataNodeWhitelistMetadataKey, false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var yourDataNodeWhitelist data.YourDataNodeWhitelist
	err = proto.Unmarshal(value, &yourDataNodeWhitelist)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	yourDataNodeWhitelist.Active = true

	value, err = utils.ProtoDeterministicMarshal(&yourDataNodeWhitelist)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set(yourDataNodeWhitelistMetadataKey, value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableYourDataNodeWhitelistParam struct {
}

func (app *ABCIApplication) validateDisableYourDataNodeWhitelist(funcParam DisableYourDataNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
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

	if checktx {
		return nil
	}

	// stateful

	value, err := app.state.Get(yourDataNodeWhitelistMetadataKey, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var yourDataNodeWhitelist data.YourDataNodeWhitelist
	err = proto.Unmarshal(value, &yourDataNodeWhitelist)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if !yourDataNodeWhitelist.Active {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already inactive/disabled",
		}
	}

	return nil
}

func (app *ABCIApplication) disableYourDataNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableYourDataNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableYourDataNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableYourDataNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableYourDataNodeWhitelist, Parameter: %s", param)
	var funcParam DisableYourDataNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableYourDataNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	value, err := app.state.Get(yourDataNodeWhitelistMetadataKey, false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var yourDataNodeWhitelist data.YourDataNodeWhitelist
	err = proto.Unmarshal(value, &yourDataNodeWhitelist)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	yourDataNodeWhitelist.Active = false

	value, err = utils.ProtoDeterministicMarshal(&yourDataNodeWhitelist)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set(yourDataNodeWhitelistMetadataKey, value)

	return app.NewExecTxResult(code.OK, "success", "")
}
