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
	goleveldbutil "github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"

	appTypes "github.com/ndidplatform/smart-contract/v9/abci/app/v1/types"
	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type AddSuppressedIdentityModificationNotificationNodeParam struct {
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateAddSuppressedIdentityModificationNotificationNode(funcParam AddSuppressedIdentityModificationNotificationNodeParam, callerNodeID string, committedState bool, checktx bool) error {
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

	// check if node ID exists and is an IdP
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &node)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if appTypes.NodeRole(node.Role) != appTypes.NodeRoleIdp {
		return &ApplicationError{
			Code:    code.NotIdPNode,
			Message: "not an IdP node",
		}
	}

	key := suppressedIdentityModificationNotificationNodePrefix + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if exists {
		return &ApplicationError{
			Code:    code.SuppressedIdentityModificationNotificationNodeIDAlreadyExists,
			Message: "suppressed identity modification notification node ID already exists",
		}
	}

	return nil
}

func (app *ABCIApplication) addSuppressedIdentityModificationNotificationNodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddSuppressedIdentityModificationNotificationNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddSuppressedIdentityModificationNotificationNode(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

// regulator only
func (app *ABCIApplication) addSuppressedIdentityModificationNotificationNode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddSuppressedIdentityModificationNotificationNode, Parameter: %s", param)
	var funcParam AddSuppressedIdentityModificationNotificationNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddSuppressedIdentityModificationNotificationNode(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := suppressedIdentityModificationNotificationNodePrefix + keySeparator + funcParam.NodeID

	var suppressedIdentityModificationNotificationNode data.SuppressedIdentityModificationNotificationNode
	value, err := utils.ProtoDeterministicMarshal(&suppressedIdentityModificationNotificationNode)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveSuppressedIdentityModificationNotificationNodeParam struct {
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateRemoveSuppressedIdentityModificationNotificationNode(funcParam RemoveSuppressedIdentityModificationNotificationNodeParam, callerNodeID string, committedState bool, checktx bool) error {
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

	key := suppressedIdentityModificationNotificationNodePrefix + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !exists {
		return &ApplicationError{
			Code:    code.SuppressedIdentityModificationNotificationNodeIDDoesNotExist,
			Message: "suppressed identity modification notification node ID does not exist",
		}
	}

	return nil
}

func (app *ABCIApplication) removeSuppressedIdentityModificationNotificationNodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveSuppressedIdentityModificationNotificationNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveSuppressedIdentityModificationNotificationNode(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

// regulator only
func (app *ABCIApplication) removeSuppressedIdentityModificationNotificationNode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveSuppressedIdentityModificationNotificationNode, Parameter: %s", param)
	var funcParam RemoveSuppressedIdentityModificationNotificationNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveSuppressedIdentityModificationNotificationNode(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := suppressedIdentityModificationNotificationNodePrefix + keySeparator + funcParam.NodeID

	app.state.Delete([]byte(key))

	return app.NewExecTxResult(code.OK, "success", "")
}

type GetSuppressedIdentityModificationNotificationNodeListParam struct {
	Prefix string `json:"prefix"`
}

func (app *ABCIApplication) getSuppressedIdentityModificationNotificationNodeList(param []byte, height int64) *abcitypes.ResponseQuery {
	app.logger.Infof("GetSuppressedIdentityModificationNotificationNodeList, Parameter: %s", param)
	var funcParam GetSuppressedIdentityModificationNotificationNodeListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := make([]string, 0)

	keyIteratorBasePrefix := suppressedIdentityModificationNotificationNodePrefix + keySeparator
	keyIteratorPrefix := keyIteratorBasePrefix + funcParam.Prefix
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		runes := []rune(string(key))
		requestType := string(runes[len(keyIteratorBasePrefix):])

		result = append(result, requestType)
	}
	iter.Close()

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(resultJSON, "success", app.state.Height)
}

type IsSuppressedIdentityModificationNotificationNodeParams struct {
	NodeID string `json:"node_id"`
}

type IsSuppressedIdentityModificationNotificationNodeResult struct {
	Suppressed bool `json:"suppressed"`
}

func (app *ABCIApplication) isSuppressedIdentityModificationNotificationNode(param []byte, height int64) *abcitypes.ResponseQuery {
	app.logger.Infof("IsSuppressedIdentityModificationNotificationNode, Parameter: %s", param)
	var funcParam IsSuppressedIdentityModificationNotificationNodeParams
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	var result IsSuppressedIdentityModificationNotificationNodeResult

	key := suppressedIdentityModificationNotificationNodePrefix + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result.Suppressed = exists

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(resultJSON, "success", app.state.Height)
}
