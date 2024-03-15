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

type SetSupportedIALListParam struct {
	SupportedIALList []float64 `json:"supported_ial_list"`
}

func (app *ABCIApplication) validateSetSupportedIALList(funcParam SetSupportedIALListParam, callerNodeID string, committedState bool) error {
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

func (app *ABCIApplication) setSupportedIALListCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetSupportedIALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetSupportedIALList(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setSupportedIALList(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetSupportedIALList, Parameter: %s", param)
	var funcParam SetSupportedIALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetSupportedIALList(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	var supportedIALList data.SupportedIALList
	supportedIALList.IalList = funcParam.SupportedIALList
	supportedIALListByte, err := utils.ProtoDeterministicMarshal(&supportedIALList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(supportedIALListKeyBytes, supportedIALListByte)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

// type GetSupportedIALListParam struct {
// }

type GetSupportedIALListResult struct {
	SupportedIALList []float64 `json:"supported_ial_list"`
}

func (app *ABCIApplication) GetSupportedIALList(param []byte, committedState bool) types.ResponseQuery {
	app.logger.Infof("GetSupportedIALList, Parameter: %s", param)
	// var funcParam GetSupportedIALListParam
	// err := json.Unmarshal(param, &funcParam)
	// if err != nil {
	// 	return app.ReturnQuery(nil, err.Error(), app.state.Height)
	// }

	var result GetSupportedIALListResult
	result.SupportedIALList = make([]float64, 0)

	supportedIALValue, err := app.state.Get(supportedIALListKeyBytes, committedState)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if supportedIALValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		return app.ReturnQuery(resultJSON, "success", app.state.Height)
	}

	var supportedIALList data.SupportedIALList
	err = proto.Unmarshal(supportedIALValue, &supportedIALList)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result.SupportedIALList = supportedIALList.IalList

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
