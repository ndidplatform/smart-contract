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

	goleveldbutil "github.com/syndtr/goleveldb/leveldb/util"
	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v6/abci/code"
	"github.com/ndidplatform/smart-contract/v6/abci/utils"
	data "github.com/ndidplatform/smart-contract/v6/protos/data"
)

type AddSuppressedIdentityModificationNotificationNodeParam struct {
	NodeID string `json:"node_id"`
}

// regulator only
func (app *ABCIApplication) addSuppressedIdentityModificationNotificationNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddSuppressedIdentityModificationNotificationNode, Parameter: %s", param)
	var funcParam AddSuppressedIdentityModificationNotificationNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// check if node ID exists and is an IdP
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var node data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &node)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if node.Role != "IdP" {
		return app.ReturnDeliverTxLog(code.NotIdPNode, "not an IdP node", "")
	}

	key := suppressedIdentityModificationNotificationNodePrefix + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if exists {
		return app.ReturnDeliverTxLog(
			code.SuppressedIdentityModificationNotificationNodeIDAlreadyExists,
			"suppressed identity modification notification node ID already exists",
			"",
		)
	}

	var suppressedIdentityModificationNotificationNode data.SuppressedIdentityModificationNotificationNode
	value, err := utils.ProtoDeterministicMarshal(&suppressedIdentityModificationNotificationNode)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type RemoveSuppressedIdentityModificationNotificationNodeParam struct {
	NodeID string `json:"node_id"`
}

// regulator only
func (app *ABCIApplication) removeSuppressedIdentityModificationNotificationNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RemoveSuppressedIdentityModificationNotificationNode, Parameter: %s", param)
	var funcParam RemoveSuppressedIdentityModificationNotificationNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := suppressedIdentityModificationNotificationNodePrefix + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if !exists {
		return app.ReturnDeliverTxLog(
			code.SuppressedIdentityModificationNotificationNodeIDDoesNotExist,
			"suppressed identity modification notification node ID does not exist",
			"",
		)
	}

	app.state.Delete([]byte(key))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type GetSuppressedIdentityModificationNotificationNodeListParam struct {
	Prefix string `json:"prefix"`
}

func (app *ABCIApplication) getSuppressedIdentityModificationNotificationNodeList(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetSuppressedIdentityModificationNotificationNodeList, Parameter: %s", param)
	var funcParam GetSuppressedIdentityModificationNotificationNodeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result := make([]string, 0)

	keyIteratorBasePrefix := suppressedIdentityModificationNotificationNodePrefix + keySeparator
	keyIteratorPrefix := keyIteratorBasePrefix + funcParam.Prefix
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		runes := []rune(string(key))
		requestType := string(runes[len(keyIteratorBasePrefix):])

		result = append(result, requestType)
	}
	iter.Close()

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

type IsSuppressedIdentityModificationNotificationNodeParams struct {
	NodeID string `json:"node_id"`
}

type IsSuppressedIdentityModificationNotificationNodeResult struct {
	Suppressed bool `json:"suppressed"`
}

func (app *ABCIApplication) isSuppressedIdentityModificationNotificationNode(param string, height int64) types.ResponseQuery {
	app.logger.Infof("IsSuppressedIdentityModificationNotificationNode, Parameter: %s", param)
	var funcParam IsSuppressedIdentityModificationNotificationNodeParams
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var result IsSuppressedIdentityModificationNotificationNodeResult

	key := suppressedIdentityModificationNotificationNodePrefix + keySeparator + funcParam.NodeID
	exists, err := app.state.Has([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result.Suppressed = exists

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
