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

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type SetAllowedAALListParam struct {
	AllowedAALList []float64 `json:"allowed_aal_list"`
}

func (app *ABCIApplication) validateSetAllowedAALList(funcParam SetAllowedAALListParam, callerNodeID string, committedState bool) error {
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

func (app *ABCIApplication) setAllowedAALListCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetAllowedAALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetAllowedAALList(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setAllowedAALList(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetAllowedAALList, Parameter: %s", param)
	var funcParam SetAllowedAALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetAllowedAALList(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	var allowedAALList data.AllowedAALList
	allowedAALList.AalList = funcParam.AllowedAALList
	allowedAALListByte, err := utils.ProtoDeterministicMarshal(&allowedAALList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(allowedAALListKeyBytes, allowedAALListByte)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

// type GetAllowedAALListParam struct {
// }

type GetAllowedAALListResult struct {
	AllowedAALList []float64 `json:"allowed_aal_list"`
}

func (app *ABCIApplication) GetAllowedAALList(param []byte, committedState bool) types.ResponseQuery {
	app.logger.Infof("GetAllowedAALList, Parameter: %s", param)
	// var funcParam GetAllowedAALListParam
	// err := json.Unmarshal(param, &funcParam)
	// if err != nil {
	// 	return app.ReturnQuery(nil, err.Error(), app.state.Height)
	// }

	var result GetAllowedAALListResult
	result.AllowedAALList = make([]float64, 0)

	allowedAALValue, err := app.state.Get(allowedAALListKeyBytes, committedState)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if allowedAALValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		return app.ReturnQuery(resultJSON, "success", app.state.Height)
	}

	var allowedAALList data.AllowedAALList
	err = proto.Unmarshal(allowedAALValue, &allowedAALList)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result.AllowedAALList = allowedAALList.AalList

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
