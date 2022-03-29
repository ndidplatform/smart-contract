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

func (app *ABCIApplication) setTimeOutBlockRegisterIdentity(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetTimeOutBlockRegisterIdentity, Parameter: %s", param)
	var funcParam TimeOutBlockRegisterIdentity
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := "TimeOutBlockRegisterIdentity"
	var timeOut data.TimeOutBlockRegisterIdentity
	timeOut.TimeOutBlock = funcParam.TimeOutBlock
	// Check time out block > 0
	if timeOut.TimeOutBlock <= 0 {
		return app.ReturnDeliverTxLog(code.TimeOutBlockIsMustGreaterThanZero, "Time out block is must greater than 0", "")
	}
	value, err := utils.ProtoDeterministicMarshal(&timeOut)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) SetAllowedMinIalForRegisterIdentityAtFirstIdp(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetAllowedMinIalForRegisterIdentityAtFirstIdp, Parameter: %s", param)
	var funcParam SetAllowedMinIalForRegisterIdentityAtFirstIdpParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	allowedMinIalKey := "AllowedMinIalForRegisterIdentityAtFirstIdp"
	var allowedMinIal data.AllowedMinIalForRegisterIdentityAtFirstIdp
	allowedMinIal.MinIal = funcParam.MinIal
	allowedMinIalByte, err := utils.ProtoDeterministicMarshal(&allowedMinIal)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allowedMinIalKey), allowedMinIalByte)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
