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

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type TimeOutBlockRegisterIdentity struct {
	TimeOutBlock int64 `json:"time_out_block"`
}

func (app *ABCIApplication) validateSetTimeOutBlockRegisterIdentity(funcParam TimeOutBlockRegisterIdentity, callerNodeID string, committedState bool, checktx bool) error {
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

	return nil
}

func (app *ABCIApplication) setTimeOutBlockRegisterIdentityCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam TimeOutBlockRegisterIdentity
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetTimeOutBlockRegisterIdentity(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setTimeOutBlockRegisterIdentity(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetTimeOutBlockRegisterIdentity, Parameter: %s", param)
	var funcParam TimeOutBlockRegisterIdentity
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetTimeOutBlockRegisterIdentity(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := "TimeOutBlockRegisterIdentity"
	var timeOut data.TimeOutBlockRegisterIdentity
	timeOut.TimeOutBlock = funcParam.TimeOutBlock
	// Check time out block > 0
	if timeOut.TimeOutBlock <= 0 {
		return app.NewExecTxResult(code.TimeOutBlockIsMustGreaterThanZero, "Time out block is must greater than 0", "")
	}
	value, err := utils.ProtoDeterministicMarshal(&timeOut)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), []byte(value))
	return app.NewExecTxResult(code.OK, "success", "")
}

type SetAllowedMinIalForRegisterIdentityAtFirstIdpParam struct {
	MinIal float64 `json:"min_ial"`
}

func (app *ABCIApplication) validateSetAllowedMinIalForRegisterIdentityAtFirstIdp(funcParam SetAllowedMinIalForRegisterIdentityAtFirstIdpParam, callerNodeID string, committedState bool, checktx bool) error {
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

	return nil
}

func (app *ABCIApplication) setAllowedMinIalForRegisterIdentityAtFirstIdpCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam SetAllowedMinIalForRegisterIdentityAtFirstIdpParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetAllowedMinIalForRegisterIdentityAtFirstIdp(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setAllowedMinIalForRegisterIdentityAtFirstIdp(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetAllowedMinIalForRegisterIdentityAtFirstIdp, Parameter: %s", param)
	var funcParam SetAllowedMinIalForRegisterIdentityAtFirstIdpParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetAllowedMinIalForRegisterIdentityAtFirstIdp(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	allowedMinIalKey := "AllowedMinIalForRegisterIdentityAtFirstIdp"
	var allowedMinIal data.AllowedMinIalForRegisterIdentityAtFirstIdp
	allowedMinIal.MinIal = funcParam.MinIal
	allowedMinIalByte, err := utils.ProtoDeterministicMarshal(&allowedMinIal)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allowedMinIalKey), allowedMinIalByte)
	return app.NewExecTxResult(code.OK, "success", "")
}
