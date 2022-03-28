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
	"strconv"

	"github.com/tendermint/tendermint/abci/types"

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

func (app *ABCIApplication) initNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("InitNDID, Parameter: %s", param)
	var funcParam InitNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	var nodeDetail data.NodeDetail
	nodeDetail.PublicKey = funcParam.PublicKey
	nodeDetail.MasterPublicKey = funcParam.MasterPublicKey
	nodeDetail.NodeName = "NDID"
	nodeDetail.Role = "NDID"
	nodeDetail.Active = true
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	chainHistoryInfoKey := "ChainHistoryInfo"
	app.state.Set(masterNDIDKeyBytes, []byte(nodeID))
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set(initStateKeyBytes, []byte("true"))
	app.state.Set([]byte(chainHistoryInfoKey), []byte(funcParam.ChainHistoryInfo))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) SetInitData(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetInitData, Parameter: %s", param)
	var funcParam SetInitDataParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for _, kv := range funcParam.KVList {
		app.state.Set(kv.Key, kv.Value)
	}
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) EndInit(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EndInit, Parameter: %s", param)
	var funcParam EndInitParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	app.state.Set(initStateKeyBytes, []byte("false"))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) setLastBlock(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetLastBlock, Parameter: %s", param)
	var funcParam SetLastBlockParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	lastBlockValue := funcParam.BlockHeight
	if funcParam.BlockHeight == 0 {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	if funcParam.BlockHeight < -1 {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	if funcParam.BlockHeight > 0 && funcParam.BlockHeight < app.state.CurrentBlockHeight {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	app.state.Set(lastBlockKeyBytes, []byte(strconv.FormatInt(lastBlockValue, 10)))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
