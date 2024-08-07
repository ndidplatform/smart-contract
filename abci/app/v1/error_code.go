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
	"strings"

	"google.golang.org/protobuf/proto"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type GetErrorCodeListParam struct {
	Type string `json:"type"`
}

type GetErrorCodeListResult struct {
	ErrorCode   int32  `json:"error_code"`
	Description string `json:"description"`
}

func (app *ABCIApplication) getErrorCodeList(param []byte) *abcitypes.ResponseQuery {
	var funcParam GetErrorCodeListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	// convert funcParam to lowercase and fetch the code list
	funcParam.Type = strings.ToLower(funcParam.Type)
	errorCodeListKey := errorCodeListKeyPrefix + keySeparator + funcParam.Type
	errorCodeListBytes, err := app.state.Get([]byte(errorCodeListKey), false)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	var errorCodeList data.ErrorCodeList
	err = proto.Unmarshal(errorCodeListBytes, &errorCodeList)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	// parse result into response format
	result := make([]*GetErrorCodeListResult, 0, len(errorCodeList.ErrorCode))
	for _, errorCode := range errorCodeList.ErrorCode {
		result = append(result, &GetErrorCodeListResult{
			ErrorCode:   errorCode.ErrorCode,
			Description: errorCode.Description,
		})
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}
