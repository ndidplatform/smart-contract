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

type SetAllowedIALListParam struct {
	AllowedIALList []float64 `json:"allowed_ial_list"`
}

func (app *ABCIApplication) validateSetAllowedIALList(funcParam SetAllowedIALListParam, callerNodeID string, committedState bool) error {
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

func (app *ABCIApplication) setAllowedIALListCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetAllowedIALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetAllowedIALList(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setAllowedIALList(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetAllowedIALList, Parameter: %s", param)
	var funcParam SetAllowedIALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetAllowedIALList(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	var allowedIALList data.AllowedIALList
	allowedIALList.IalList = funcParam.AllowedIALList
	allowedIALListByte, err := utils.ProtoDeterministicMarshal(&allowedIALList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(allowedIALListKeyBytes, allowedIALListByte)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

// type GetAllowedIALListParam struct {
// }

type GetAllowedIALListResult struct {
	AllowedIALList []float64 `json:"allowed_ial_list"`
}

func (app *ABCIApplication) GetAllowedIALList(param []byte, committedState bool) types.ResponseQuery {
	app.logger.Infof("GetAllowedIALList, Parameter: %s", param)
	// var funcParam GetAllowedIALListParam
	// err := json.Unmarshal(param, &funcParam)
	// if err != nil {
	// 	return app.ReturnQuery(nil, err.Error(), app.state.Height)
	// }

	var result GetAllowedIALListResult
	result.AllowedIALList = make([]float64, 0)

	allowedIALValue, err := app.state.Get(allowedIALListKeyBytes, committedState)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if allowedIALValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		return app.ReturnQuery(resultJSON, "success", app.state.Height)
	}

	var allowedIALList data.AllowedIALList
	err = proto.Unmarshal(allowedIALValue, &allowedIALList)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result.AllowedIALList = allowedIALList.IalList

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
