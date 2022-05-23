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
	"google.golang.org/protobuf/proto"

	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

type ServiceDetail struct {
	ServiceID         string `json:"service_id"`
	ServiceName       string `json:"service_name"`
	DataSchema        string `json:"data_schema"`
	DataSchemaVersion string `json:"data_schema_version"`
	Active            bool   `json:"active"`
}

type GetServiceDetailParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) getServiceDetail(param []byte) types.ResponseQuery {
	app.logger.Infof("GetServiceDetail, Parameter: %s", param)
	var funcParam GetServiceDetailParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		value = []byte("{}")
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal(value, &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	returnValue, err := json.Marshal(&service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServiceList(param []byte) types.ResponseQuery {
	app.logger.Infof("GetServiceList, Parameter: %s", param)
	key := "AllService"
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		result := make([]ServiceDetail, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	result := make([]*data.ServiceDetail, 0)
	// filter flag==true
	var services data.ServiceDetailList
	err = proto.Unmarshal([]byte(value), &services)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for _, service := range services.Services {
		if service.Active {
			result = append(result, service)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServiceNameByServiceID(serviceID string) string {
	key := serviceKeyPrefix + keySeparator + serviceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		panic(err)
	}
	if value == nil {
		return ""
	}
	var result ServiceDetail
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		return ""
	}
	return result.ServiceName
}

type GetServicesByAsIDParam struct {
	AsID string `json:"as_id"`
}

type GetServicesByAsIDResult struct {
	Services []Service `json:"services"`
}

type Service struct {
	ServiceID              string   `json:"service_id"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	Active                 bool     `json:"active"`
	Suspended              bool     `json:"suspended"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

func (app *ABCIApplication) getServicesByAsID(param []byte) types.ResponseQuery {
	app.logger.Infof("GetServicesByAsID, Parameter: %s", param)
	var funcParam GetServicesByAsIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetServicesByAsIDResult
	result.Services = make([]Service, 0)
	provideServiceKey := providedServicesKeyPrefix + keySeparator + funcParam.AsID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if provideServiceValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	var services data.ServiceList
	err = proto.Unmarshal([]byte(provideServiceValue), &services)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.AsID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if nodeDetailValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for index, provideService := range services.Services {
		serviceKey := serviceKeyPrefix + keySeparator + provideService.ServiceId
		serviceValue, err := app.state.Get([]byte(serviceKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if serviceValue == nil {
			continue
		}
		var service data.ServiceDetail
		err = proto.Unmarshal([]byte(serviceValue), &service)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if nodeDetail.Active && service.Active {
			// Set suspended from NDID
			approveServiceKey := approvedServiceKeyPrefix + keySeparator + provideService.ServiceId + keySeparator + funcParam.AsID
			approveServiceJSON, err := app.state.Get([]byte(approveServiceKey), true)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			if approveServiceJSON == nil {
				continue
			}
			var approveService data.ApproveService
			err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
			if err == nil {
				services.Services[index].Suspended = !approveService.Active
			}
			var newRow Service
			newRow.Active = services.Services[index].Active
			newRow.MinAal = services.Services[index].MinAal
			newRow.MinIal = services.Services[index].MinIal
			newRow.ServiceID = services.Services[index].ServiceId
			newRow.Suspended = services.Services[index].Suspended
			newRow.SupportedNamespaceList = services.Services[index].SupportedNamespaceList
			result.Services = append(result.Services, newRow)
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.Services) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
