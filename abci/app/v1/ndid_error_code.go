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
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

func (*ABCIApplication) checkErrorCodeType(errorCodeType string) bool {
	return contains(errorCodeType, []string{"idp", "as"})
}

type AddErrorCodeParam struct {
	ErrorCode   int32  `json:"error_code"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func (app *ABCIApplication) validateAddErrorCode(funcParam AddErrorCodeParam, callerNodeID string, committedState bool) error {
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

	funcParam.Type = strings.ToLower(funcParam.Type)
	if !app.checkErrorCodeType(funcParam.Type) {
		return &ApplicationError{
			Code:    code.InvalidErrorCode,
			Message: "Invalid error code type",
		}
	}
	if funcParam.ErrorCode == 0 {
		return &ApplicationError{
			Code:    code.InvalidErrorCode,
			Message: "ErrorCode cannot be 0",
		}
	}

	errorKey := errorCodeKeyPrefix + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", funcParam.ErrorCode)
	hasErrorKey, err := app.state.Has([]byte(errorKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if hasErrorKey {
		return &ApplicationError{
			Code:    code.InvalidErrorCode,
			Message: "ErrorCode already exists",
		}
	}

	return nil
}

func (app *ABCIApplication) addErrorCodeCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam AddErrorCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddErrorCode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) addErrorCode(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddErrorCode, Parameter: %s", param)
	var funcParam AddErrorCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddErrorCode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	// convert error type to lower case
	funcParam.Type = strings.ToLower(funcParam.Type)

	errorCode := data.ErrorCode{
		ErrorCode:   funcParam.ErrorCode,
		Description: funcParam.Description,
	}

	// add error code
	errorCodeBytes, err := utils.ProtoDeterministicMarshal(&errorCode)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	errorKey := errorCodeKeyPrefix + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", errorCode.ErrorCode)
	app.state.Set([]byte(errorKey), []byte(errorCodeBytes))

	// add error code to ErrorCodeList
	var errorCodeList data.ErrorCodeList
	errorsKey := errorCodeListKeyPrefix + keySeparator + funcParam.Type
	errorCodeListBytes, err := app.state.Get([]byte(errorsKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if errorCodeListBytes != nil {
		err := proto.Unmarshal(errorCodeListBytes, &errorCodeList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	errorCodeList.ErrorCode = append(errorCodeList.ErrorCode, &errorCode)
	errorCodeListBytes, err = utils.ProtoDeterministicMarshal(&errorCodeList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(errorsKey), []byte(errorCodeListBytes))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type RemoveErrorCodeParam struct {
	ErrorCode int32  `json:"error_code"`
	Type      string `json:"type"`
}

func (app *ABCIApplication) validateRemoveErrorCode(funcParam RemoveErrorCodeParam, callerNodeID string, committedState bool) error {
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

	errorKey := errorCodeKeyPrefix + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", funcParam.ErrorCode)
	hasErrorKey, err := app.state.Has([]byte(errorKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !hasErrorKey {
		return &ApplicationError{
			Code:    code.InvalidErrorCode,
			Message: "ErrorCode does not exist",
		}
	}

	return nil
}

func (app *ABCIApplication) removeErrorCodeCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam RemoveErrorCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveErrorCode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) removeErrorCode(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RemoveErrorCode, Parameter: %s", param)
	var funcParam RemoveErrorCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveErrorCode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	// remove error code from ErrorCode index
	errorKey := errorCodeKeyPrefix + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", funcParam.ErrorCode)
	err = app.state.Delete([]byte(errorKey))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}

	// remove ErrorCode from ErrorCodeList
	var errorCodeList data.ErrorCodeList
	errorsKey := errorCodeListKeyPrefix + keySeparator + funcParam.Type
	errorCodeListBytes, err := app.state.Get([]byte(errorsKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if errorCodeListBytes != nil {
		err := proto.Unmarshal(errorCodeListBytes, &errorCodeList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}

	newErrorCodeList := data.ErrorCodeList{
		ErrorCode: make([]*data.ErrorCode, 0, len(errorCodeList.ErrorCode)),
	}
	for _, errorCode := range errorCodeList.ErrorCode {
		if errorCode.ErrorCode != funcParam.ErrorCode {
			newErrorCodeList.ErrorCode = append(newErrorCodeList.ErrorCode, errorCode)
		}
	}

	if len(newErrorCodeList.ErrorCode) != len(errorCodeList.ErrorCode)-1 {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode does not exist", "")
	}

	errorCodeListBytes, err = utils.ProtoDeterministicMarshal(&newErrorCodeList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	app.state.Set([]byte(errorsKey), []byte(errorCodeListBytes))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
