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

// RP

type GetYourDataRPNodeWhitelistResult struct {
	RPNodeIDs []string `json:"rp_node_id_list"`
	Enabled   bool     `json:"enabled"`
}

func (app *ABCIApplication) getYourDataRPNodeWhitelist(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetYourDataRPNodeWhitelist, Parameter: %s", param)

	rpNodeIDs := make([]string, 0)

	keyIteratorPrefix := yourDataRPNodeWhitelistKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		runes := []rune(string(key))
		nodeID := string(runes[len(keyIteratorPrefix):])

		rpNodeIDs = append(rpNodeIDs, nodeID)
	}
	iter.Close()

	metadataValue, err := app.state.Get(yourDataRPNodeWhitelistMetadataKey, true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var yourDataRPNodeWhitelist data.YourDataRPNodeWhitelist
	err = proto.Unmarshal(metadataValue, &yourDataRPNodeWhitelist)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := &GetYourDataRPNodeWhitelistResult{
		RPNodeIDs: rpNodeIDs,
		Enabled:   yourDataRPNodeWhitelist.Active,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetYourDataRPPermissionStatusParam struct {
	RPNodeID string `json:"rp_node_id"`
}

type GetYourDataRPPermissionStatusReturn struct {
	Allowed bool `json:"allowed"`
}

func (app *ABCIApplication) getYourDataRPPermissionStatus(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetYourDataRPPermissionStatus, Parameter: %s", param)

	var funcParam GetYourDataRPPermissionStatusParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	key := yourDataRPNodeWhitelistKeyPrefix + keySeparator + funcParam.RPNodeID
	allowed, err := app.state.Has([]byte(key), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := &GetYourDataRPPermissionStatusReturn{
		Allowed: allowed,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

// AS

type GetYourDataASNodeWhitelistResult struct {
	ASNodeIDs []string `json:"as_node_id_list"`
	Enabled   bool     `json:"enabled"`
}

func (app *ABCIApplication) getYourDataASNodeWhitelist(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetYourDataASNodeWhitelist, Parameter: %s", param)

	asNodeIDs := make([]string, 0)

	keyIteratorPrefix := yourDataASNodeWhitelistKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		runes := []rune(string(key))
		nodeID := string(runes[len(keyIteratorPrefix):])

		asNodeIDs = append(asNodeIDs, nodeID)
	}
	iter.Close()

	metadataValue, err := app.state.Get(yourDataASNodeWhitelistMetadataKey, true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var yourDataASNodeWhitelist data.YourDataASNodeWhitelist
	err = proto.Unmarshal(metadataValue, &yourDataASNodeWhitelist)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := &GetYourDataASNodeWhitelistResult{
		ASNodeIDs: asNodeIDs,
		Enabled:   yourDataASNodeWhitelist.Active,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetYourDataASPermissionStatusParam struct {
	ASNodeID string `json:"as_node_id"`
}

type GetYourDataASPermissionStatusReturn struct {
	Allowed bool `json:"allowed"`
}

func (app *ABCIApplication) getYourDataASPermissionStatus(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetYourDataASPermissionStatus, Parameter: %s", param)

	var funcParam GetYourDataASPermissionStatusParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	key := yourDataASNodeWhitelistKeyPrefix + keySeparator + funcParam.ASNodeID
	allowed, err := app.state.Has([]byte(key), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := &GetYourDataASPermissionStatusReturn{
		Allowed: allowed,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}
