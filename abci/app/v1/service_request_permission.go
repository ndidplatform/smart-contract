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

	abcitypes "github.com/cometbft/cometbft/abci/types"
	goleveldbutil "github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"

	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type ServiceRequestPermission struct {
	ServiceID      string   `json:"service_id"`
	AllowedNodeIDs []string `json:"allowed_node_id_list"`
}

type GetServiceRequestNodeWhitelistResult struct {
	ServiceRequestPermissionList []ServiceRequestPermission `json:"service_request_permission_list"`
	Enabled                      bool                       `json:"enabled"`
}

func (app *ABCIApplication) getServiceRequestNodeWhitelist(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServiceRequestNodeWhitelist, Parameter: %s", param)

	serviceRequestPermissionList := make([]ServiceRequestPermission, 0)

	keyIteratorPrefix := serviceRequestNodeWhitelistKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		// key format: <prefix>|<serviceID>|<nodeID>
		keyParts := strings.Split(string(key), keySeparator)
		serviceID := keyParts[len(keyParts)-2]
		nodeID := keyParts[len(keyParts)-1]

		if len(serviceRequestPermissionList) > 0 {
			lastIdx := len(serviceRequestPermissionList) - 1
			if serviceRequestPermissionList[lastIdx].ServiceID == serviceID {
				serviceRequestPermissionList[lastIdx].AllowedNodeIDs = append(serviceRequestPermissionList[lastIdx].AllowedNodeIDs, nodeID)
			}
		} else {
			serviceRequestPermissionList = append(serviceRequestPermissionList, ServiceRequestPermission{
				ServiceID:      serviceID,
				AllowedNodeIDs: []string{nodeID},
			})
		}
	}
	iter.Close()

	enabled := false
	metadataValue, err := app.state.Get(serviceRequestNodeWhitelistMetadataKey, true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if metadataValue != nil {
		var serviceRequestNodeWhitelist data.ServiceRequestNodeWhitelist
		err = proto.Unmarshal(metadataValue, &serviceRequestNodeWhitelist)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		enabled = serviceRequestNodeWhitelist.Active
	}

	result := &GetServiceRequestNodeWhitelistResult{
		ServiceRequestPermissionList: serviceRequestPermissionList,
		Enabled:                      enabled,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetServiceRequestNodeWhitelistByServiceIDParam struct {
	ServiceID string `json:"service_id"`
}

type GetServiceRequestNodeWhitelistByServiceIDResult struct {
	NodeIDs []string `json:"node_id_list"`
	Enabled bool     `json:"enabled"`
}

func (app *ABCIApplication) getServiceRequestNodeWhitelistByServiceID(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServiceRequestNodeWhitelistByServiceID, Parameter: %s", param)

	var funcParam GetServiceRequestNodeWhitelistByServiceIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	nodeIDs := make([]string, 0)

	keyIteratorPrefix := serviceRequestNodeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator
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

	metadataValue, err := app.state.Get(serviceRequestNodeWhitelistMetadataKey, true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var serviceRequestNodeWhitelist data.ServiceRequestNodeWhitelist
	err = proto.Unmarshal(metadataValue, &serviceRequestNodeWhitelist)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	result := &GetServiceRequestNodeWhitelistByServiceIDResult{
		NodeIDs: nodeIDs,
		Enabled: serviceRequestNodeWhitelist.Active,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetAllowedServiceRequestListParam struct {
	NodeID string `json:"node_id"`
}

type GetAllowedServiceRequestListReturn struct {
	AllowedServiceIDList []string `json:"allowed_service_id_list"`
}

func (app *ABCIApplication) getAllowedServiceRequestList(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetAllowedServiceRequestList, Parameter: %s", param)

	var funcParam GetAllowedServiceRequestListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	serviceIDs := make([]string, 0)

	keyIteratorPrefix := serviceRequestNodeWhitelistKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		// key format: <prefix>|<serviceID>|<nodeID>
		keyParts := strings.Split(string(key), keySeparator)
		serviceID := keyParts[len(keyParts)-2]
		nodeID := keyParts[len(keyParts)-1]

		if nodeID == funcParam.NodeID {
			serviceIDs = append(serviceIDs, serviceID)
		}
	}
	iter.Close()

	result := &GetAllowedServiceRequestListReturn{
		AllowedServiceIDList: serviceIDs,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) hasServiceRequestPermission(serviceID string, nodeID string) (allowed bool, err error) {
	// check if whitelist is active
	metadataValue, err := app.state.Get(serviceRequestNodeWhitelistMetadataKey, true)
	if err != nil {
		return false, err
	}
	if metadataValue == nil {
		return true, nil
	}

	var serviceRequestNodeWhitelist data.ServiceRequestNodeWhitelist
	err = proto.Unmarshal(metadataValue, &serviceRequestNodeWhitelist)
	if err != nil {
		return false, err
	}

	if !serviceRequestNodeWhitelist.Active {
		return true, nil
	}

	key := serviceRequestNodeWhitelistKeyPrefix + keySeparator + serviceID + keySeparator + nodeID
	allowed, err = app.state.Has([]byte(key), true)
	if err != nil {
		return false, err
	}

	return allowed, nil
}
