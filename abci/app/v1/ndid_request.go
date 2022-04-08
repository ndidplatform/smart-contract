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

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

type SetAllowedModeListParam struct {
	Purpose         string  `json:"purpose"`
	AllowedModeList []int32 `json:"allowed_mode_list"`
}

func (app *ABCIApplication) SetAllowedModeList(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetAllowedModeList, Parameter: %s", param)
	var funcParam SetAllowedModeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
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
