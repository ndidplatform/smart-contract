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

type AddNodeToDomainNodeWhitelistParam struct {
	Domain string `json:"domain"`
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateAddNodeToDomainNodeWhitelist(funcParam AddNodeToDomainNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	domainKey := domainKeyPrefix + keySeparator + funcParam.Domain
	exists, err := app.state.Has([]byte(domainKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !exists {
		return &ApplicationError{
			Code:    code.DomainDoesNotExist,
			Message: "Domain does not exist",
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

	key := domainNodeWhitelistKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.NodeID
	exists, err = app.state.Has([]byte(key), committedState)
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

func (app *ABCIApplication) addNodeToDomainNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddNodeToDomainNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddNodeToDomainNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addNodeToDomainNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddNodeToDomainNodeWhitelist, Parameter: %s", param)
	var funcParam AddNodeToDomainNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddNodeToDomainNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := domainNodeWhitelistKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.NodeID

	var nodePermission data.DomainNodePermission

	value, err := utils.ProtoDeterministicMarshal(&nodePermission)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveNodeFromDomainNodeWhitelistParam struct {
	Domain string `json:"domain"`
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateRemoveNodeFromDomainNodeWhitelist(funcParam RemoveNodeFromDomainNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	key := domainNodeWhitelistKeyPrefix + keySeparator + funcParam.NodeID
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
			Message: "not found",
		}
	}

	return nil
}

func (app *ABCIApplication) removeNodeFromDomainNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveNodeFromDomainNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveNodeFromDomainNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) removeNodeFromDomainNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveNodeFromDomainNodeWhitelist, Parameter: %s", param)
	var funcParam RemoveNodeFromDomainNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveNodeFromDomainNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := domainNodeWhitelistKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.NodeID

	app.state.Delete([]byte(key))

	return app.NewExecTxResult(code.OK, "success", "")
}

//

type EnableDomainNodeWhitelistParam struct {
	Domain string `json:"domain"`
}

func (app *ABCIApplication) validateEnableDomainNodeWhitelist(funcParam EnableDomainNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	key := domainKeyPrefix + keySeparator + funcParam.Domain
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return &ApplicationError{
			Code:    code.DomainDoesNotExist,
			Message: "Domain does not exist",
		}
	}

	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if domain.NodeWhitelistEnabled {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already enabled",
		}
	}

	return nil
}

func (app *ABCIApplication) enableDomainNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableDomainNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableDomainNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableDomainNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableDomainNodeWhitelist, Parameter: %s", param)
	var funcParam EnableDomainNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableDomainNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := domainKeyPrefix + keySeparator + funcParam.Domain

	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	domain.NodeWhitelistEnabled = true

	value, err = utils.ProtoDeterministicMarshal(&domain)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableDomainNodeWhitelistParam struct {
	Domain string `json:"domain"`
}

func (app *ABCIApplication) validateDisableDomainNodeWhitelist(funcParam DisableDomainNodeWhitelistParam, callerNodeID string, committedState bool, checktx bool) error {
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

	key := domainKeyPrefix + keySeparator + funcParam.Domain
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return &ApplicationError{
			Code:    code.DomainDoesNotExist,
			Message: "Domain does not exist",
		}
	}

	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if !domain.NodeWhitelistEnabled {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already inactive/disabled",
		}
	}

	return nil
}

func (app *ABCIApplication) disableDomainNodeWhitelistCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableDomainNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableDomainNodeWhitelist(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableDomainNodeWhitelist(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableDomainNodeWhitelist, Parameter: %s", param)
	var funcParam DisableDomainNodeWhitelistParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableDomainNodeWhitelist(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := domainKeyPrefix + keySeparator + funcParam.Domain

	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	domain.NodeWhitelistEnabled = false

	value, err = utils.ProtoDeterministicMarshal(&domain)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}
