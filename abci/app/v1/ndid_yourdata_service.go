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

type AllowYourDataServiceToBeMixedInRequestParam struct {
}

func (app *ABCIApplication) validateAllowYourDataServiceToBeMixedInRequest(funcParam AllowYourDataServiceToBeMixedInRequestParam, callerNodeID string, committedState bool, checktx bool) error {
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

	// stateless

	if checktx {
		return nil
	}

	// stateful

	value, err := app.state.Get(yourDataServiceMixedInRequestPermissionKey, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var yourDataServiceMixedInRequestPermission data.YourDataServiceMixedInRequestPermission
	err = proto.Unmarshal(value, &yourDataServiceMixedInRequestPermission)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if yourDataServiceMixedInRequestPermission.Allowed {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already allowed",
		}
	}

	return nil
}

func (app *ABCIApplication) allowYourDataServiceToBeMixedInRequestCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AllowYourDataServiceToBeMixedInRequestParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAllowYourDataServiceToBeMixedInRequest(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) allowYourDataServiceToBeMixedInRequest(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AllowYourDataServiceToBeMixedInRequest, Parameter: %s", param)
	var funcParam AllowYourDataServiceToBeMixedInRequestParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAllowYourDataServiceToBeMixedInRequest(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	value, err := app.state.Get(yourDataServiceMixedInRequestPermissionKey, false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var yourDataServiceMixedInRequestPermission data.YourDataServiceMixedInRequestPermission
	err = proto.Unmarshal(value, &yourDataServiceMixedInRequestPermission)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	yourDataServiceMixedInRequestPermission.Allowed = true

	newValue, err := utils.ProtoDeterministicMarshal(&yourDataServiceMixedInRequestPermission)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set(yourDataServiceMixedInRequestPermissionKey, newValue)

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisallowYourDataServiceToBeMixedInRequestParam struct {
}

func (app *ABCIApplication) validateDisallowYourDataServiceToBeMixedInRequest(funcParam DisallowYourDataServiceToBeMixedInRequestParam, callerNodeID string, committedState bool, checktx bool) error {
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

	if checktx {
		return nil
	}

	// stateful

	value, err := app.state.Get(yourDataServiceMixedInRequestPermissionKey, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var yourDataServiceMixedInRequestPermission data.YourDataServiceMixedInRequestPermission
	err = proto.Unmarshal(value, &yourDataServiceMixedInRequestPermission)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if !yourDataServiceMixedInRequestPermission.Allowed {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already disallowed",
		}
	}

	return nil
}

func (app *ABCIApplication) disallowYourDataServiceToBeMixedInRequestCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisallowYourDataServiceToBeMixedInRequestParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisallowYourDataServiceToBeMixedInRequest(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disallowYourDataServiceToBeMixedInRequest(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisallowYourDataServiceToBeMixedInRequest, Parameter: %s", param)
	var funcParam DisallowYourDataServiceToBeMixedInRequestParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisallowYourDataServiceToBeMixedInRequest(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	value, err := app.state.Get(yourDataServiceMixedInRequestPermissionKey, false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var yourDataServiceMixedInRequestPermission data.YourDataServiceMixedInRequestPermission
	err = proto.Unmarshal(value, &yourDataServiceMixedInRequestPermission)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	yourDataServiceMixedInRequestPermission.Allowed = false

	newValue, err := utils.ProtoDeterministicMarshal(&yourDataServiceMixedInRequestPermission)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set(yourDataServiceMixedInRequestPermissionKey, newValue)

	return app.NewExecTxResult(code.OK, "success", "")
}
