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

type SetSupportedAALListParam struct {
	SupportedAALList []float64 `json:"supported_aal_list"`
}

func (app *ABCIApplication) validateSetSupportedAALList(funcParam SetSupportedAALListParam, callerNodeID string, committedState bool) error {
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

func (app *ABCIApplication) setSupportedAALListCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetSupportedAALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetSupportedAALList(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setSupportedAALList(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetSupportedAALList, Parameter: %s", param)
	var funcParam SetSupportedAALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetSupportedAALList(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	var supportedAALList data.SupportedAALList
	supportedAALList.AalList = funcParam.SupportedAALList
	supportedAALListByte, err := utils.ProtoDeterministicMarshal(&supportedAALList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(supportedAALListKeyBytes, supportedAALListByte)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

// type GetSupportedAALListParam struct {
// }

type GetSupportedAALListResult struct {
	SupportedAALList []float64 `json:"supported_aal_list"`
}

func (app *ABCIApplication) GetSupportedAALList(param []byte, committedState bool) types.ResponseQuery {
	app.logger.Infof("GetSupportedAALList, Parameter: %s", param)
	// var funcParam GetSupportedAALListParam
	// err := json.Unmarshal(param, &funcParam)
	// if err != nil {
	// 	return app.ReturnQuery(nil, err.Error(), app.state.Height)
	// }

	var result GetSupportedAALListResult
	result.SupportedAALList = make([]float64, 0)

	supportedAALValue, err := app.state.Get(supportedAALListKeyBytes, committedState)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if supportedAALValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		return app.ReturnQuery(resultJSON, "success", app.state.Height)
	}

	var supportedAALList data.SupportedAALList
	err = proto.Unmarshal(supportedAALValue, &supportedAALList)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result.SupportedAALList = supportedAALList.AalList

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
