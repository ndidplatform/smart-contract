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

	data "github.com/ndidplatform/smart-contract/v10/protos/data"
)

type ServiceRequestTypePermission struct {
	ServiceID           string    `json:"service_id"`
	Enabled             bool      `json:"enabled"`
	AllowedRequestTypes []*string `json:"allowed_request_type_list"`
}

type GetServiceRequestTypeWhitelistResult struct {
	ServiceRequestPermissionList []ServiceRequestTypePermission `json:"service_request_permission_list"`
}

func (app *ABCIApplication) getServiceRequestTypeWhitelist(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServiceRequestTypeWhitelist, Parameter: %s", param)

	serviceRequestPermissionList := make([]ServiceRequestTypePermission, 0)

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
			iter.Close()
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}

		runes := []rune(string(key))
		serviceID := string(runes[len(keyIteratorPrefix):])

		serviceRequestPermissionList = append(serviceRequestPermissionList, ServiceRequestTypePermission{
			ServiceID: serviceID,
			Enabled:   service.RequestTypeWhitelistEnabled,
		})
	}
	iter.Close()

	// get allowed request types for each service
	for idx, serviceRequestPermission := range serviceRequestPermissionList {
		allowedRequestTypes := make([]*string, 0)

		keyIteratorPrefix := serviceRequestTypeWhitelistKeyPrefix + keySeparator + serviceRequestPermission.ServiceID + keySeparator
		r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
		iter, err := app.state.db.Iterator(r.Start, r.Limit)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		for ; iter.Valid(); iter.Next() {
			key := iter.Key()

			runes := []rune(string(key))
			requestType := string(runes[len(keyIteratorPrefix):])

			if requestType == "" {
				allowedRequestTypes = append(allowedRequestTypes, nil)
			} else {
				allowedRequestTypes = append(allowedRequestTypes, &requestType)
			}
		}
		iter.Close()

		serviceRequestPermissionList[idx].AllowedRequestTypes = allowedRequestTypes
	}

	result := &GetServiceRequestTypeWhitelistResult{
		ServiceRequestPermissionList: serviceRequestPermissionList,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetServiceRequestTypeWhitelistByServiceIDParam struct {
	ServiceID string `json:"service_id"`
}

type GetServiceRequestTypeWhitelistByServiceIDResult struct {
	RequestTypes []*string `json:"request_type_list"`
	Enabled      bool      `json:"enabled"`
}

func (app *ABCIApplication) getServiceRequestTypeWhitelistByServiceID(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServiceRequestTypeWhitelistByServiceID, Parameter: %s", param)

	var funcParam GetServiceRequestTypeWhitelistByServiceIDParam
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

	requestTypes := make([]*string, 0)

	keyIteratorPrefix := serviceRequestTypeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		runes := []rune(string(key))
		requestType := string(runes[len(keyIteratorPrefix):])

		if requestType == "" {
			requestTypes = append(requestTypes, nil)
		} else {
			requestTypes = append(requestTypes, &requestType)
		}
	}
	iter.Close()

	result := &GetServiceRequestTypeWhitelistByServiceIDResult{
		RequestTypes: requestTypes,
		Enabled:      service.RequestTypeWhitelistEnabled,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetRequestTypeWhitelistedServiceListParam struct {
	RequestType *string `json:"request_type"`
}

type ServicePermissionForRequestType struct {
	ServiceID string `json:"service_id"`
	Enabled   bool   `json:"enabled"`
}

type GetRequestTypeWhitelistedServiceListResult struct {
	ServicePermissionList []ServicePermissionForRequestType `json:"service_permission_list"`
}

func (app *ABCIApplication) getRequestTypeWhitelistedServiceList(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetRequestTypeWhitelistedServiceList, Parameter: %s", param)

	var funcParam GetRequestTypeWhitelistedServiceListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	var requestType string
	if funcParam.RequestType != nil {
		requestType = *funcParam.RequestType
	} else {
		// default (request without request type specified)
		requestType = ""
	}

	servicePermissionList := make([]ServicePermissionForRequestType, 0)

	keyIteratorPrefix := serviceRequestTypeWhitelistKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(keyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		// Key structure: prefix + separator + serviceID + separator + requestType
		runes := []rune(string(key))
		keyContent := string(runes[len(keyIteratorPrefix):])

		// Split or parse out the components.
		// Check if the key ends with input requestType
		expectedSuffix := keySeparator + requestType
		if !strings.HasSuffix(keyContent, expectedSuffix) {
			continue
		}

		// Extract the serviceID from the middle of the key string
		serviceID := keyContent[:len(keyContent)-len(expectedSuffix)]

		// Get service's request type whitelist enabled status
		serviceKey := serviceKeyPrefix + keySeparator + serviceID
		serviceValue, err := app.state.Get([]byte(serviceKey), true)
		if err != nil {
			iter.Close()
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}

		var service data.ServiceDetail
		err = proto.Unmarshal(serviceValue, &service)
		if err != nil {
			iter.Close()
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}

		servicePermissionList = append(servicePermissionList, ServicePermissionForRequestType{
			ServiceID: serviceID,
			Enabled:   service.RequestTypeWhitelistEnabled,
		})
	}
	iter.Close()

	result := &GetRequestTypeWhitelistedServiceListResult{
		ServicePermissionList: servicePermissionList,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetServiceRequestTypePermissionParam struct {
	RequestType *string `json:"request_type"`
	ServiceID   string  `json:"service_id"`
}

type GetServiceRequestTypePermissionResult struct {
	Allowed bool `json:"allowed"`
}

func (app *ABCIApplication) getServiceRequestTypePermission(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServiceRequestTypePermission, Parameter: %s", param)

	var funcParam GetServiceRequestTypePermissionParam
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
	if service.RequestTypeWhitelistEnabled {
		var requestType string
		if funcParam.RequestType != nil {
			requestType = *funcParam.RequestType
		} else {
			// default (request without request type specified)
			requestType = ""
		}

		key := serviceRequestTypeWhitelistKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + requestType
		allowed, err = app.state.Has([]byte(key), true)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
	} else {
		allowed = true
	}

	result := &GetServiceRequestTypePermissionResult{
		Allowed: allowed,
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) hasServiceRequestTypePermission(service ServiceDetail, requestType *string) (allowed bool, err error) {
	// check if whitelist is active
	if !service.RequestTypeWhitelistEnabled {
		return true, nil
	}

	requestTypeInKey := ""
	if requestType != nil {
		requestTypeInKey = *requestType
	}

	key := serviceRequestTypeWhitelistKeyPrefix + keySeparator + service.ServiceID + keySeparator + requestTypeInKey
	allowed, err = app.state.Has([]byte(key), true)
	if err != nil {
		return false, err
	}

	return allowed, nil
}
