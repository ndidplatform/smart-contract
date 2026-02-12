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

type DomainNodePermission struct {
	Domain         string   `json:"domain"`
	Enabled        bool     `json:"enabled"`
	AllowedNodeIDs []string `json:"allowed_node_id_list"`
}

type GetDomainNodeWhitelistResult struct {
	DomainNodePermissionList []DomainNodePermission `json:"domain_node_permission_list"`
}

func (app *ABCIApplication) getDomainNodeWhitelist(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetDomainNodeWhitelist, Parameter: %s", param)

	domainNodePermissionList := make([]DomainNodePermission, 0)

	// get all domains
	keyIteratorPrefix := domainKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		var domain data.Domain
		err = proto.Unmarshal(value, &domain)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}

		runes := []rune(string(key))
		domainName := string(runes[len(keyIteratorPrefix):])

		domainNodePermissionList = append(domainNodePermissionList, DomainNodePermission{
			Domain:  domainName,
			Enabled: domain.NodeWhitelistEnabled,
		})
	}
	iter.Close()

	// get allowed node IDs for each domain
	for idx, domainNodePermission := range domainNodePermissionList {
		allowedNodeIDs := make([]string, 0)

		keyIteratorPrefix := domainNodeWhitelistKeyPrefix + keySeparator + domainNodePermission.Domain + keySeparator
		r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
		iter, err := app.state.db.Iterator(r.Start, r.Limit)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		for ; iter.Valid(); iter.Next() {
			key := iter.Key()

			runes := []rune(string(key))
			nodeID := string(runes[len(keyIteratorPrefix):])

			allowedNodeIDs = append(allowedNodeIDs, nodeID)
		}
		iter.Close()

		domainNodePermissionList[idx].AllowedNodeIDs = allowedNodeIDs
	}

	result := &GetDomainNodeWhitelistResult{
		DomainNodePermissionList: domainNodePermissionList,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetDomainNodeWhitelistByDomainParam struct {
	Domain string `json:"domain"`
}

type GetDomainNodeWhitelistByDomainResult struct {
	NodeIDs []string `json:"node_id_list"`
	Enabled bool     `json:"enabled"`
}

func (app *ABCIApplication) getDomainNodeWhitelistByDomain(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetDomainNodeWhitelistByDomain, Parameter: %s", param)

	var funcParam GetDomainNodeWhitelistByDomainParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	key := domainKeyPrefix + keySeparator + funcParam.Domain
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		return app.NewResponseQuery(nil, "not found", app.state.Height)
	}
	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	nodeIDs := make([]string, 0)

	keyIteratorPrefix := domainNodeWhitelistKeyPrefix + keySeparator + funcParam.Domain + keySeparator
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

	result := &GetDomainNodeWhitelistByDomainResult{
		NodeIDs: nodeIDs,
		Enabled: domain.NodeWhitelistEnabled,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetDomainNodePermissionParam struct {
	NodeID string `json:"node_id"`
	Domain string `json:"domain"`
}

type GetDomainNodePermissionResult struct {
	Allowed bool `json:"allowed"`
}

func (app *ABCIApplication) getDomainNodePermission(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetDomainNodePermission, Parameter: %s", param)

	var funcParam GetDomainNodePermissionParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	allowed, err := app.hasDomainPermission(funcParam.Domain, funcParam.NodeID)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := &GetDomainNodePermissionResult{
		Allowed: allowed,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) hasDomainPermission(domainName string, nodeID string) (allowed bool, err error) {
	key := domainKeyPrefix + keySeparator + domainName
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return false, err
	}
	if value == nil {
		return false, nil
	}
	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return false, err
	}

	if !domain.Active {
		return false, nil
	}

	if !domain.NodeWhitelistEnabled {
		return true, nil
	}

	key = domainNodeWhitelistKeyPrefix + keySeparator + domainName + keySeparator + nodeID
	allowed, err = app.state.Has([]byte(key), true)
	if err != nil {
		return false, err
	}

	return allowed, nil
}
