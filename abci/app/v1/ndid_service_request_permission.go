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

	"github.com/ndidplatform/smart-contract/v10/abci/code"
	"github.com/ndidplatform/smart-contract/v10/abci/utils"
	data "github.com/ndidplatform/smart-contract/v10/protos/data"
)

type AddNodeToServiceRequesterNodeWhitelistParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateAddNodeToServiceRequesterNodeWhitelist(funcParam AddNodeToServiceRequesterNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceExists, err := app.state.Has([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !serviceExists {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}

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

	key := serviceRequesterNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID
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

func (app *ABCIApplication) addNodeToServiceRequesterNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddNodeToServiceRequesterNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddNodeToServiceRequesterNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addNodeToServiceRequesterNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddNodeToServiceRequesterNodeWhitelist, Parameter: %s", param)
	var funcParam AddNodeToServiceRequesterNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddNodeToServiceRequesterNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := serviceRequesterNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID

	var nodePermission data.ServiceRequestNodePermission

	value, err := utils.ProtoDeterministicMarshal(&nodePermission)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveNodeFromServiceRequesterNodeWhitelistParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateRemoveNodeFromServiceRequesterNodeWhitelist(funcParam RemoveNodeFromServiceRequesterNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	key := serviceRequesterNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID
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

func (app *ABCIApplication) removeNodeFromServiceRequesterNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveNodeFromServiceRequesterNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveNodeFromServiceRequesterNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) removeNodeFromServiceRequesterNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveNodeFromServiceRequesterNodeWhitelist, Parameter: %s", param)
	var funcParam RemoveNodeFromServiceRequesterNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveNodeFromServiceRequesterNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := serviceRequesterNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID

	app.state.Delete([]byte(key))

	return app.NewExecTxResult(code.OK, "success", "")
}

//

type EnableServiceRequesterNodeWhitelistParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateEnableServiceRequesterNodeWhitelist(funcParam EnableServiceRequesterNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceValue == nil {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}

	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	if service.RequesterNodeWhitelistEnabled {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already enabled",
		}
	}

	return nil
}

func (app *ABCIApplication) enableServiceRequesterNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableServiceRequesterNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableServiceRequesterNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableServiceRequesterNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableServiceRequesterNodeWhitelist, Parameter: %s", param)
	var funcParam EnableServiceRequesterNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableServiceRequesterNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	service.RequesterNodeWhitelistEnabled = true

	newValue, err := utils.ProtoDeterministicMarshal(&service)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(serviceKey), newValue)

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableServiceRequesterNodeWhitelistParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateDisableServiceRequesterNodeWhitelist(funcParam DisableServiceRequesterNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceValue == nil {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}

	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	if !service.RequesterNodeWhitelistEnabled {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already disabled",
		}
	}

	return nil
}

func (app *ABCIApplication) disableServiceRequesterNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableServiceRequesterNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableServiceRequesterNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableServiceRequesterNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableServiceRequesterNodeWhitelist, Parameter: %s", param)
	var funcParam DisableServiceRequesterNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableServiceRequesterNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	service.RequesterNodeWhitelistEnabled = false

	newValue, err := utils.ProtoDeterministicMarshal(&service)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(serviceKey), newValue)

	return app.NewExecTxResult(code.OK, "success", "")
}
