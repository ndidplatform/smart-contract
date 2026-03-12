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

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v10/abci/code"
	"github.com/ndidplatform/smart-contract/v10/abci/utils"
	data "github.com/ndidplatform/smart-contract/v10/protos/data"
)

func (*ABCIApplication) isValidDomainErrorCodeType(errorCodeType string) bool {
	return contains(errorCodeType, []string{"as"})
}

type AddDomainErrorCodeParam struct {
	Domain      string `json:"domain"`
	ErrorCode   int32  `json:"error_code"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func (app *ABCIApplication) validateAddDomainErrorCode(funcParam AddDomainErrorCodeParam, callerNodeID string, committedState bool, checktx bool) error {
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

	funcParam.Type = strings.ToLower(funcParam.Type)
	if !app.isValidDomainErrorCodeType(funcParam.Type) {
		return &ApplicationError{
			Code:    code.InvalidErrorCode,
			Message: "Invalid error code type",
		}
	}
	if funcParam.ErrorCode == 0 {
		return &ApplicationError{
			Code:    code.InvalidErrorCode,
			Message: "error code cannot be 0",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	domainKey := domainKeyPrefix + keySeparator + funcParam.Domain
	domainExists, err := app.state.Has([]byte(domainKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !domainExists {
		return &ApplicationError{
			Code:    code.DomainDoesNotExist,
			Message: "Domain does not exist",
		}
	}

	errorKey := domainErrorCodeListKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", funcParam.ErrorCode)
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
			Message: "error code already exists",
		}
	}

	return nil
}

func (app *ABCIApplication) addDomainErrorCodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddDomainErrorCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddDomainErrorCode(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addDomainErrorCode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddDomainErrorCode, Parameter: %s", param)
	var funcParam AddDomainErrorCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddDomainErrorCode(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
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
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	errorKey := domainErrorCodeListKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", errorCode.ErrorCode)
	app.state.Set([]byte(errorKey), []byte(errorCodeBytes))

	// add error code to ErrorCodeList
	var errorCodeList data.ErrorCodeList
	errorsKey := domainErrorCodeListKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.Type
	errorCodeListBytes, err := app.state.Get([]byte(errorsKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if errorCodeListBytes != nil {
		err := proto.Unmarshal(errorCodeListBytes, &errorCodeList)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	}
	errorCodeList.ErrorCode = append(errorCodeList.ErrorCode, &errorCode)
	errorCodeListBytes, err = utils.ProtoDeterministicMarshal(&errorCodeList)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(errorsKey), []byte(errorCodeListBytes))

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveDomainErrorCodeParam struct {
	Domain    string `json:"domain"`
	ErrorCode int32  `json:"error_code"`
	Type      string `json:"type"`
}

func (app *ABCIApplication) validateRemoveDomainErrorCode(funcParam RemoveDomainErrorCodeParam, callerNodeID string, committedState bool, checktx bool) error {
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

	errorKey := domainErrorCodeListKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", funcParam.ErrorCode)
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
			Message: "error code does not exist",
		}
	}

	return nil
}

func (app *ABCIApplication) removeDomainErrorCodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveDomainErrorCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveDomainErrorCode(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) removeDomainErrorCode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveDomainErrorCode, Parameter: %s", param)
	var funcParam RemoveDomainErrorCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveDomainErrorCode(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// remove error code from ErrorCode index
	errorKey := domainErrorCodeListKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", funcParam.ErrorCode)
	err = app.state.Delete([]byte(errorKey))
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	// remove ErrorCode from ErrorCodeList
	var errorCodeList data.ErrorCodeList
	errorsKey := domainErrorCodeListKeyPrefix + keySeparator + funcParam.Domain + keySeparator + funcParam.Type
	errorCodeListBytes, err := app.state.Get([]byte(errorsKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if errorCodeListBytes != nil {
		err := proto.Unmarshal(errorCodeListBytes, &errorCodeList)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
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
		return app.NewExecTxResult(code.InvalidErrorCode, "error code does not exist", "")
	}

	errorCodeListBytes, err = utils.ProtoDeterministicMarshal(&newErrorCodeList)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}
	app.state.Set([]byte(errorsKey), []byte(errorCodeListBytes))

	return app.NewExecTxResult(code.OK, "success", "")
}
