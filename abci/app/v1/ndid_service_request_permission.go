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

type AddNodeToServiceRequestNodeWhitelistParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateAddNodeToServiceRequestNodeWhitelist(funcParam AddNodeToServiceRequestNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	key := serviceRequestNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if exists {
		return &ApplicationError{
			Code:    code.DuplicateEntry,
			Message: "Duplicate entry",
		}
	}

	return nil
}

func (app *ABCIApplication) addNodeToServiceRequestNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddNodeToServiceRequestNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddNodeToServiceRequestNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addNodeToServiceRequestNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddNodeToServiceRequestNodeWhitelist, Parameter: %s", param)
	var funcParam AddNodeToServiceRequestNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddNodeToServiceRequestNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := serviceRequestNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID

	var nodePermission data.ServiceRequestNodePermission

	value, err := utils.ProtoDeterministicMarshal(&nodePermission)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveNodeFromServiceRequestNodeWhitelistParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateRemoveNodeFromServiceRequestNodeWhitelist(funcParam RemoveNodeFromServiceRequestNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	key := serviceRequestNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !exists {
		return &ApplicationError{
			Code:    code.NotFound,
			Message: "entry not found",
		}
	}

	return nil
}

func (app *ABCIApplication) removeNodeFromServiceRequestNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveNodeFromServiceRequestNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveNodeFromServiceRequestNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) removeNodeFromServiceRequestNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveNodeFromServiceRequestNodeWhitelist, Parameter: %s", param)
	var funcParam RemoveNodeFromServiceRequestNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveNodeFromServiceRequestNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := serviceRequestNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID

	app.state.Delete([]byte(key))

	return app.NewExecTxResult(code.OK, "success", "")
}

//

type EnableServiceRequestNodeWhitelistParam struct {
}

func (app *ABCIApplication) validateEnableServiceRequestNodeWhitelist(funcParam EnableServiceRequestNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	value, err := app.state.Get(serviceRequestNodeWhitelistMetadataKey, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var serviceRequestNodeWhitelist data.ServiceRequestNodeWhitelist
	err = proto.Unmarshal(value, &serviceRequestNodeWhitelist)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if serviceRequestNodeWhitelist.Active {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already active/enabled",
		}
	}

	return nil
}

func (app *ABCIApplication) enableServiceRequestNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableServiceRequestNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableServiceRequestNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableServiceRequestNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableServiceRequestNodeWhitelist, Parameter: %s", param)
	var funcParam EnableServiceRequestNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableServiceRequestNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	value, err := app.state.Get(serviceRequestNodeWhitelistMetadataKey, false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var serviceRequestNodeWhitelist data.ServiceRequestNodeWhitelist
	err = proto.Unmarshal(value, &serviceRequestNodeWhitelist)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	serviceRequestNodeWhitelist.Active = true

	value, err = utils.ProtoDeterministicMarshal(&serviceRequestNodeWhitelist)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set(serviceRequestNodeWhitelistMetadataKey, value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableServiceRequestNodeWhitelistParam struct {
}

func (app *ABCIApplication) validateDisableServiceRequestNodeWhitelist(funcParam DisableServiceRequestNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	value, err := app.state.Get(serviceRequestNodeWhitelistMetadataKey, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var serviceRequestNodeWhitelist data.ServiceRequestNodeWhitelist
	err = proto.Unmarshal(value, &serviceRequestNodeWhitelist)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if !serviceRequestNodeWhitelist.Active {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already inactive/disabled",
		}
	}

	return nil
}

func (app *ABCIApplication) disableServiceRequestNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableServiceRequestNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableServiceRequestNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableServiceRequestNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableServiceRequestNodeWhitelist, Parameter: %s", param)
	var funcParam DisableServiceRequestNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableServiceRequestNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	value, err := app.state.Get(serviceRequestNodeWhitelistMetadataKey, false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var serviceRequestNodeWhitelist data.ServiceRequestNodeWhitelist
	err = proto.Unmarshal(value, &serviceRequestNodeWhitelist)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	serviceRequestNodeWhitelist.Active = false

	value, err = utils.ProtoDeterministicMarshal(&serviceRequestNodeWhitelist)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set(serviceRequestNodeWhitelistMetadataKey, value)

	return app.NewExecTxResult(code.OK, "success", "")
}
