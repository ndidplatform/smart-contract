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

type AddRequestTypeToServiceRequestTypeWhitelistParam struct {
	RequestType *string `json:"request_type"`
	ServiceID   string  `json:"service_id"`
}

func (app *ABCIApplication) validateAddRequestTypeToServiceRequestTypeWhitelist(funcParam AddRequestTypeToServiceRequestTypeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	var requestType string
	if funcParam.RequestType != nil {
		requestType = *funcParam.RequestType

		requestTypeKey := requestTypeKeyPrefix + keySeparator + requestType
		requestTypeValue, err := app.state.Get([]byte(requestTypeKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if requestTypeValue == nil {
			return &ApplicationError{
				Code:    code.RequestTypeDoesNotExist,
				Message: "Request type not found",
			}
		}

	} else {
		// default (request without request type specified)
		requestType = ""

		// no need to check request type existence
	}

	key := serviceRequestTypeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + requestType
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

func (app *ABCIApplication) addRequestTypeToServiceRequestTypeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddRequestTypeToServiceRequestTypeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddRequestTypeToServiceRequestTypeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addRequestTypeToServiceRequestTypeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddRequestTypeToServiceRequestTypeWhitelist, Parameter: %s", param)
	var funcParam AddRequestTypeToServiceRequestTypeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddRequestTypeToServiceRequestTypeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	var requestType string
	if funcParam.RequestType != nil {
		requestType = *funcParam.RequestType
	} else {
		// default (request without request type specified)
		requestType = ""
	}

	key := serviceRequestTypeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + requestType

	var nodePermission data.ServiceRequestTypePermission

	value, err := utils.ProtoDeterministicMarshal(&nodePermission)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveRequestTypeFromServiceRequestTypeWhitelistParam struct {
	RequestType *string `json:"request_type"`
	ServiceID   string  `json:"service_id"`
}

func (app *ABCIApplication) validateRemoveRequestTypeFromServiceRequestTypeWhitelist(funcParam RemoveRequestTypeFromServiceRequestTypeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	var requestType string
	if funcParam.RequestType != nil {
		requestType = *funcParam.RequestType
	} else {
		// default (request without request type specified)
		requestType = ""
	}

	key := serviceRequestTypeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + requestType
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

func (app *ABCIApplication) removeRequestTypeFromServiceRequestTypeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveRequestTypeFromServiceRequestTypeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveRequestTypeFromServiceRequestTypeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) removeRequestTypeFromServiceRequestTypeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveRequestTypeFromServiceRequestTypeWhitelist, Parameter: %s", param)
	var funcParam RemoveRequestTypeFromServiceRequestTypeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveRequestTypeFromServiceRequestTypeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	var requestType string
	if funcParam.RequestType != nil {
		requestType = *funcParam.RequestType
	} else {
		// default (request without request type specified)
		requestType = ""
	}

	key := serviceRequestTypeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + requestType

	app.state.Delete([]byte(key))

	return app.NewExecTxResult(code.OK, "success", "")
}

//

type EnableServiceRequestTypeWhitelistParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateEnableServiceRequestTypeWhitelist(funcParam EnableServiceRequestTypeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	if service.RequestTypeWhitelistEnabled {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already enabled",
		}
	}

	return nil
}

func (app *ABCIApplication) enableServiceRequestTypeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableServiceRequestTypeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableServiceRequestTypeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableServiceRequestTypeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableServiceRequestTypeWhitelist, Parameter: %s", param)
	var funcParam EnableServiceRequestTypeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableServiceRequestTypeWhitelist(funcParam, callerNodeID, false, false)
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

	service.RequestTypeWhitelistEnabled = true

	newValue, err := utils.ProtoDeterministicMarshal(&service)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(serviceKey), newValue)

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableServiceRequestTypeWhitelistParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateDisableServiceRequestTypeWhitelist(funcParam DisableServiceRequestTypeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	if !service.RequestTypeWhitelistEnabled {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already disabled",
		}
	}

	return nil
}

func (app *ABCIApplication) disableServiceRequestTypeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableServiceRequestTypeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableServiceRequestTypeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableServiceRequestTypeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableServiceRequestTypeWhitelist, Parameter: %s", param)
	var funcParam DisableServiceRequestTypeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableServiceRequestTypeWhitelist(funcParam, callerNodeID, false, false)
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

	service.RequestTypeWhitelistEnabled = false

	newValue, err := utils.ProtoDeterministicMarshal(&service)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(serviceKey), newValue)

	return app.NewExecTxResult(code.OK, "success", "")
}
