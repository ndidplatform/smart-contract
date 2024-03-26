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

func (app *ABCIApplication) setSupportedAALListCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam SetSupportedAALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetSupportedAALList(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setSupportedAALList(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetSupportedAALList, Parameter: %s", param)
	var funcParam SetSupportedAALListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetSupportedAALList(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	var supportedAALList data.SupportedAALList
	supportedAALList.AalList = funcParam.SupportedAALList
	supportedAALListByte, err := utils.ProtoDeterministicMarshal(&supportedAALList)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set(supportedAALListKeyBytes, supportedAALListByte)
	return app.NewExecTxResult(code.OK, "success", "")
}

// type GetSupportedAALListParam struct {
// }

type GetSupportedAALListResult struct {
	SupportedAALList []float64 `json:"supported_aal_list"`
}

func (app *ABCIApplication) GetSupportedAALList(param []byte, committedState bool) *abcitypes.ResponseQuery {
	app.logger.Infof("GetSupportedAALList, Parameter: %s", param)
	// var funcParam GetSupportedAALListParam
	// err := json.Unmarshal(param, &funcParam)
	// if err != nil {
	// 	return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	// }

	var result GetSupportedAALListResult
	result.SupportedAALList = make([]float64, 0)

	supportedAALValue, err := app.state.Get(supportedAALListKeyBytes, committedState)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if supportedAALValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}

		return app.NewResponseQuery(resultJSON, "success", app.state.Height)
	}

	var supportedAALList data.SupportedAALList
	err = proto.Unmarshal(supportedAALValue, &supportedAALList)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result.SupportedAALList = supportedAALList.AalList

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(resultJSON, "success", app.state.Height)
}
