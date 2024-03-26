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

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type AddAllowedNodeSupportedFeatureParam struct {
	Name string `json:"name"`
}

func (app *ABCIApplication) validateAddAllowedNodeSupportedFeature(funcParam AddAllowedNodeSupportedFeatureParam, callerNodeID string, committedState bool) error {
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

	key := nodeSupportedFeatureKeyPrefix + keySeparator + funcParam.Name
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if exists {
		return &ApplicationError{
			Code:    code.NodeSupportedFeatureAlreadyExists,
			Message: "node supported feature already exists",
		}
	}

	return nil
}

func (app *ABCIApplication) addAllowedNodeSupportedFeatureCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddAllowedNodeSupportedFeatureParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddAllowedNodeSupportedFeature(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

// regulator only
func (app *ABCIApplication) addAllowedNodeSupportedFeature(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddAllowedNodeSupportedFeature, Parameter: %s", param)
	var funcParam AddAllowedNodeSupportedFeatureParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddAllowedNodeSupportedFeature(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := nodeSupportedFeatureKeyPrefix + keySeparator + funcParam.Name

	var nodeSupportedFeature data.NodeSupportedFeature
	value, err := utils.ProtoDeterministicMarshal(&nodeSupportedFeature)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveAllowedNodeSupportedFeatureParam struct {
	Name string `json:"name"`
}

func (app *ABCIApplication) validateRemoveAllowedNodeSupportedFeature(funcParam RemoveAllowedNodeSupportedFeatureParam, callerNodeID string, committedState bool) error {
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

	key := nodeSupportedFeatureKeyPrefix + keySeparator + funcParam.Name
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !exists {
		return &ApplicationError{
			Code:    code.NodeSupportedFeatureDoesNotExist,
			Message: "node supported feature does not exist",
		}
	}

	return nil
}

func (app *ABCIApplication) removeAllowedNodeSupportedFeatureCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveAllowedNodeSupportedFeatureParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveAllowedNodeSupportedFeature(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

// regulator only
func (app *ABCIApplication) removeAllowedNodeSupportedFeature(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveAllowedNodeSupportedFeature, Parameter: %s", param)
	var funcParam RemoveAllowedNodeSupportedFeatureParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveAllowedNodeSupportedFeature(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := nodeSupportedFeatureKeyPrefix + keySeparator + funcParam.Name

	app.state.Delete([]byte(key))

	return app.NewExecTxResult(code.OK, "success", "")
}

type GetAllowedNodeSupportedFeatureListParam struct {
	Prefix string `json:"prefix"`
}

func (app *ABCIApplication) getAllowedNodeSupportedFeatureList(param []byte, height int64) *abcitypes.ResponseQuery {
	app.logger.Infof("GetAllowedNodeSupportedFeatureList, Parameter: %s", param)
	var funcParam GetAllowedNodeSupportedFeatureListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := make([]string, 0)

	nodeSupportedFeatureKeyIteratorBasePrefix := nodeSupportedFeatureKeyPrefix + keySeparator
	nodeSupportedFeatureKeyIteratorPrefix := nodeSupportedFeatureKeyIteratorBasePrefix + funcParam.Prefix
	r := goleveldbutil.BytesPrefix([]byte(nodeSupportedFeatureKeyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		runes := []rune(string(key))
		nodeSupportedFeature := string(runes[len(nodeSupportedFeatureKeyIteratorBasePrefix):])

		result = append(result, nodeSupportedFeature)
	}
	iter.Close()

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(resultJSON, "success", app.state.Height)
}
