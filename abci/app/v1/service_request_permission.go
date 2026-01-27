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

type ServiceRequestPermission struct {
	ServiceID      string   `json:"service_id"`
	Enabled        bool     `json:"enabled"`
	AllowedNodeIDs []string `json:"allowed_node_id_list"`
}

type GetServiceRequesterNodeWhitelistResult struct {
	ServiceRequestPermissionList []ServiceRequestPermission `json:"service_request_permission_list"`
}

func (app *ABCIApplication) getServiceRequesterNodeWhitelist(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServiceRequesterNodeWhitelist, Parameter: %s", param)

	serviceRequestPermissionList := make([]ServiceRequestPermission, 0)

	keyIteratorPrefix := serviceKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		var service data.ServiceDetail
		err = proto.Unmarshal(value, &service)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}

		runes := []rune(string(key))
		serviceID := string(runes[len(keyIteratorPrefix):])

		serviceRequestPermissionList = append(serviceRequestPermissionList, ServiceRequestPermission{
			ServiceID: serviceID,
			Enabled:   service.RequesterNodeWhitelistEnabled,
		})
	}
	iter.Close()

	// get enabled state
	for idx, serviceRequestPermission := range serviceRequestPermissionList {
		allowedNodeIDs := make([]string, 0)

		keyIteratorPrefix := serviceRequesterNodeWhitelistKeyPrefix + keySeparator + serviceRequestPermission.ServiceID + keySeparator
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

		serviceRequestPermissionList[idx].AllowedNodeIDs = allowedNodeIDs
	}

	result := &GetServiceRequesterNodeWhitelistResult{
		ServiceRequestPermissionList: serviceRequestPermissionList,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetServiceRequesterNodeWhitelistByServiceIDParam struct {
	ServiceID string `json:"service_id"`
}

type GetServiceRequesterNodeWhitelistByServiceIDResult struct {
	NodeIDs []string `json:"node_id_list"`
	Enabled bool     `json:"enabled"`
}

func (app *ABCIApplication) getServiceRequesterNodeWhitelistByServiceID(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServiceRequesterNodeWhitelistByServiceID, Parameter: %s", param)

	var funcParam GetServiceRequesterNodeWhitelistByServiceIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	key := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		return app.NewResponseQuery(nil, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal(value, &service)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	nodeIDs := make([]string, 0)

	keyIteratorPrefix := serviceRequesterNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator
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

	result := &GetServiceRequesterNodeWhitelistByServiceIDResult{
		NodeIDs: nodeIDs,
		Enabled: service.RequesterNodeWhitelistEnabled,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetServiceRequesterNodePermissionParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
}

type GetServiceRequesterNodePermissionResult struct {
	Allowed bool `json:"allowed"`
}

func (app *ABCIApplication) getServiceRequesterNodePermission(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServiceRequesterNodePermission, Parameter: %s", param)

	var funcParam GetServiceRequesterNodePermissionParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	key := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		return app.NewResponseQuery(nil, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal(value, &service)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	allowed := false
	if service.RequesterNodeWhitelistEnabled {
		key := serviceRequesterNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID
		allowed, err = app.state.Has([]byte(key), true)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
	} else {
		allowed = true
	}

	result := &GetServiceRequesterNodePermissionResult{
		Allowed: allowed,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) hasServiceRequestPermission(service ServiceDetail, nodeID string) (allowed bool, err error) {
	// check if whitelist is active
	if !service.RequesterNodeWhitelistEnabled {
		return true, nil
	}

	key := serviceRequesterNodeWhitelistKeyPrefix + keySeparator + service.ServiceID + keySeparator + nodeID
	allowed, err = app.state.Has([]byte(key), true)
	if err != nil {
		return false, err
	}

	return allowed, nil
}
