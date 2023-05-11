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

	"github.com/ndidplatform/smart-contract/v8/abci/code"
	"github.com/ndidplatform/smart-contract/v8/abci/utils"
	data "github.com/ndidplatform/smart-contract/v8/protos/data"
)

type SetAllowedModeListParam struct {
	Purpose         string  `json:"purpose"`
	AllowedModeList []int32 `json:"allowed_mode_list"`
}

func (app *ABCIApplication) validateSetAllowedModeList(funcParam SetAllowedModeListParam, callerNodeID string, committedState bool) error {
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

func (app *ABCIApplication) setAllowedModeListCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetAllowedModeListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetAllowedModeList(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setAllowedModeList(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetAllowedModeList, Parameter: %s", param)
	var funcParam SetAllowedModeListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetAllowedModeList(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	allowedModeKey := allowedModeListKeyPrefix + keySeparator + funcParam.Purpose
	var allowedModeList data.AllowedModeList
	allowedModeList.Mode = funcParam.AllowedModeList
	allowedModeListByte, err := utils.ProtoDeterministicMarshal(&allowedModeList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allowedModeKey), allowedModeListByte)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
