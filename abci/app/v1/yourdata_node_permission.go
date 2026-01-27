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
	goleveldbutil "github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"

	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type GetYourDataNodeWhitelistResult struct {
	NodeIDs []string `json:"node_id_list"`
	Enabled bool     `json:"enabled"`
}

func (app *ABCIApplication) getYourDataNodeWhitelist(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetYourDataNodeWhitelist, Parameter: %s", param)

	nodeIDs := make([]string, 0)

	keyIteratorPrefix := yourDataNodeWhitelistKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		runes := []rune(string(key))
		nodeID := string(runes[len(keyIteratorPrefix):])

		nodeIDs = append(nodeIDs, nodeID)
	}
	iter.Close()

	metadataValue, err := app.state.Get(yourDataNodeWhitelistMetadataKey, true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var yourDataNodeWhitelist data.YourDataNodeWhitelist
	err = proto.Unmarshal(metadataValue, &yourDataNodeWhitelist)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := &GetYourDataNodeWhitelistResult{
		NodeIDs: nodeIDs,
		Enabled: yourDataNodeWhitelist.Active,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetYourDataPermissionStatusParam struct {
	NodeID string `json:"node_id"`
}

type GetYourDataPermissionStatusReturn struct {
	Allowed bool `json:"allowed"`
}

func (app *ABCIApplication) getYourDataPermissionStatus(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetYourDataPermissionStatus, Parameter: %s", param)

	var funcParam GetYourDataPermissionStatusParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	allowed, err := app.hasYourDataPermission(funcParam.NodeID)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := &GetYourDataPermissionStatusReturn{
		Allowed: allowed,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) hasYourDataPermission(nodeID string) (allowed bool, err error) {
	// check if whitelist is active
	metadataValue, err := app.state.Get(yourDataNodeWhitelistMetadataKey, true)
	if err != nil {
		return false, err
	}
	if metadataValue == nil {
		return true, nil
	}

	var yourDataNodeWhitelist data.YourDataNodeWhitelist
	err = proto.Unmarshal(metadataValue, &yourDataNodeWhitelist)
	if err != nil {
		return false, err
	}

	if !yourDataNodeWhitelist.Active {
		return true, nil
	}

	key := yourDataNodeWhitelistKeyPrefix + keySeparator + nodeID
	allowed, err = app.state.Has([]byte(key), true)
	if err != nil {
		return false, err
	}

	return allowed, nil
}
