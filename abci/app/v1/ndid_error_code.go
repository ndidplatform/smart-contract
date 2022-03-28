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

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

func (*ABCIApplication) checkErrorCodeType(errorCodeType string) bool {
	return contains(errorCodeType, []string{"idp", "as"})
}

func (app *ABCIApplication) addErrorCode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddErrorCode, Parameter: %s", param)
	var funcParam AddErrorCodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// convert error type to lower case
	funcParam.Type = strings.ToLower(funcParam.Type)
	if !app.checkErrorCodeType(funcParam.Type) {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "Invalid error code type", "")
	}
	if funcParam.ErrorCode == 0 {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode cannot be 0", "")
	}

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
	hasErrorKey, err := app.state.Has([]byte(errorKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if hasErrorKey {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode is already in the database", "")
	}
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

func (app *ABCIApplication) removeErrorCode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RemoveErrorCode, Parameter: %s", param)
	var funcParam RemoveErrorCodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// remove error code from ErrorCode index
	errorKey := errorCodeKeyPrefix + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", funcParam.ErrorCode)
	hasErrorKey, err := app.state.Has([]byte(errorKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if !hasErrorKey {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode not exists", "")
	}
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
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode not exists", "")
	}

	errorCodeListBytes, err = utils.ProtoDeterministicMarshal(&newErrorCodeList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	app.state.Set([]byte(errorsKey), []byte(errorCodeListBytes))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
