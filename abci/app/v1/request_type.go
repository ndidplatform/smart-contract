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

	goleveldbutil "github.com/syndtr/goleveldb/leveldb/util"
	"github.com/tendermint/tendermint/abci/types"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type AddRequestTypeParam struct {
	Name string `json:"name"`
}

func (app *ABCIApplication) validateAddRequestType(funcParam AddRequestTypeParam, callerNodeID string, committedState bool) error {
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

	key := requestTypeKeyPrefix + keySeparator + funcParam.Name
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if exists {
		return &ApplicationError{
			Code:    code.RequestTypeAlreadyExists,
			Message: "request type already exists",
		}
	}

	return nil
}

func (app *ABCIApplication) addRequestTypeCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam AddRequestTypeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddRequestType(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

// regulator only
func (app *ABCIApplication) addRequestType(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddRequestType, Parameter: %s", param)
	var funcParam AddRequestTypeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddRequestType(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	key := requestTypeKeyPrefix + keySeparator + funcParam.Name

	var requestType data.RequestType
	value, err := utils.ProtoDeterministicMarshal(&requestType)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type RemoveRequestTypeParam struct {
	Name string `json:"name"`
}

func (app *ABCIApplication) validateRemoveRequestType(funcParam RemoveRequestTypeParam, callerNodeID string, committedState bool) error {
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

	key := requestTypeKeyPrefix + keySeparator + funcParam.Name
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !exists {
		return &ApplicationError{
			Code:    code.RequestTypeDoesNotExist,
			Message: "request type does not exist",
		}
	}

	return nil
}

func (app *ABCIApplication) removeRequestTypeCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam RemoveRequestTypeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveRequestType(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

// regulator only
func (app *ABCIApplication) removeRequestType(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RemoveRequestType, Parameter: %s", param)
	var funcParam RemoveRequestTypeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveRequestType(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	key := requestTypeKeyPrefix + keySeparator + funcParam.Name

	app.state.Delete([]byte(key))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type GetRequestTypeListParam struct {
	Prefix string `json:"prefix"`
}

func (app *ABCIApplication) getRequestTypeList(param []byte, height int64) types.ResponseQuery {
	app.logger.Infof("GetRequestTypeList, Parameter: %s", param)
	var funcParam GetRequestTypeListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result := make([]string, 0)

	requestTypeKeyIteratorBasePrefix := requestTypeKeyPrefix + keySeparator
	requestTypeKeyIteratorPrefix := requestTypeKeyIteratorBasePrefix + funcParam.Prefix
	r := goleveldbutil.BytesPrefix([]byte(requestTypeKeyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		runes := []rune(string(key))
		requestType := string(runes[len(requestTypeKeyIteratorBasePrefix):])

		result = append(result, requestType)
	}
	iter.Close()

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
