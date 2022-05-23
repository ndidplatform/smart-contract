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

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

type AddRequestTypeParam struct {
	Name string `json:"name"`
}

// regulator only
func (app *ABCIApplication) addRequestType(param []byte, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddRequestType, Parameter: %s", param)
	var funcParam AddRequestTypeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := requestTypeKeyPrefix + keySeparator + funcParam.Name
	exists, err := app.state.Has([]byte(key), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if exists {
		return app.ReturnDeliverTxLog(code.RequestTypeAlreadyExists, "request type already exists", "")
	}

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

// regulator only
func (app *ABCIApplication) removeRequestType(param []byte, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RemoveRequestType, Parameter: %s", param)
	var funcParam RemoveRequestTypeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := requestTypeKeyPrefix + keySeparator + funcParam.Name
	exists, err := app.state.Has([]byte(key), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if !exists {
		return app.ReturnDeliverTxLog(code.RequestTypeDoesNotExist, "request type does not exist", "")
	}

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
