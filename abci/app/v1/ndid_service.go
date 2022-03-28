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

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

func (app *ABCIApplication) addService(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddService, Parameter: %s", param)
	var funcParam AddServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	chkExists, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if chkExists != nil {
		return app.ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID", "")
	}
	// Add new service
	var service data.ServiceDetail
	service.ServiceId = funcParam.ServiceID
	service.ServiceName = funcParam.ServiceName
	service.Active = true
	service.DataSchema = funcParam.DataSchema
	service.DataSchemaVersion = funcParam.DataSchemaVersion
	serviceJSON, err := utils.ProtoDeterministicMarshal(&service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	// Add detail to service directory
	allServiceKey := "AllService"
	allServiceValue, err := app.state.Get([]byte(allServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceDetailList
	if allServiceValue != nil {
		err = proto.Unmarshal([]byte(allServiceValue), &services)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// Check duplicate service
		for _, service := range services.Services {
			if service.ServiceId == funcParam.ServiceID {
				return app.ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID", "")
			}
		}
	}
	var newService data.ServiceDetail
	newService.ServiceId = funcParam.ServiceID
	newService.ServiceName = funcParam.ServiceName
	newService.Active = true
	services.Services = append(services.Services, &newService)
	allServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allServiceKey), []byte(allServiceJSON))
	app.state.Set([]byte(serviceKey), []byte(serviceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) enableService(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableService, Parameter: %s", param)
	var funcParam DisableServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	chkExists, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if chkExists == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	// Delete detail in service directory
	allServiceKey := "AllService"
	allServiceValue, err := app.state.Get([]byte(allServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceDetailList
	err = proto.Unmarshal([]byte(allServiceValue), &services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for index, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			services.Services[index].Active = true
			break
		}
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(chkExists), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	service.Active = true
	allServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	serviceJSON, err := utils.ProtoDeterministicMarshal(&service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(serviceKey), []byte(serviceJSON))
	app.state.Set([]byte(allServiceKey), []byte(allServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) disableService(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableService, Parameter: %s", param)
	var funcParam DisableServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	chkExists, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if chkExists == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	// Delete detail in service directory
	allServiceKey := "AllService"
	allServiceValue, err := app.state.Get([]byte(allServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceDetailList
	if allServiceValue == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "List of Service not found", "")
	}
	err = proto.Unmarshal([]byte(allServiceValue), &services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for index, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			services.Services[index].Active = false
			break
		}
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(chkExists), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	service.Active = false
	allServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	serviceJSON, err := utils.ProtoDeterministicMarshal(&service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(serviceKey), []byte(serviceJSON))
	app.state.Set([]byte(allServiceKey), []byte(allServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) updateService(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateService, Parameter: %s", param)
	var funcParam UpdateServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if serviceValue == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	// Update service
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if funcParam.ServiceName != "" {
		service.ServiceName = funcParam.ServiceName
	}
	if funcParam.DataSchema != "" {
		service.DataSchema = funcParam.DataSchema
	}
	if funcParam.DataSchemaVersion != "" {
		service.DataSchemaVersion = funcParam.DataSchemaVersion
	}
	// Update detail in service directory
	allServiceKey := "AllService"
	allServiceValue, err := app.state.Get([]byte(allServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceDetailList
	if allServiceValue != nil {
		err = proto.Unmarshal([]byte(allServiceValue), &services)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// Update service
		for index, service := range services.Services {
			if service.ServiceId == funcParam.ServiceID {
				if funcParam.ServiceName != "" {
					services.Services[index].ServiceName = funcParam.ServiceName
				}
			}
		}
	}
	serviceJSON, err := utils.ProtoDeterministicMarshal(&service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	allServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allServiceKey), []byte(allServiceJSON))
	app.state.Set([]byte(serviceKey), []byte(serviceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) registerServiceDestinationByNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterServiceDestinationByNDID, Parameter: %s", param)
	var funcParam RegisterServiceDestinationByNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check node ID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check role is AS
	if nodeDetail.Role != "AS" {
		return app.ReturnDeliverTxLog(code.RoleIsNotAS, "Role of node ID is not AS", "")
	}
	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceJSON, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if serviceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID
	var approveService data.ApproveService
	approveService.Active = true
	approveServiceJSON, err := utils.ProtoDeterministicMarshal(&approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(approveServiceKey), []byte(approveServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) disableServiceDestinationByNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableServiceDestinationByNDID, Parameter: %s", param)
	var funcParam DisableServiceDestinationByNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceJSON, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if serviceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	// Check node ID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check role is AS
	if nodeDetail.Role != "AS" {
		return app.ReturnDeliverTxLog(code.RoleIsNotAS, "Role of node ID is not AS", "")
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID
	approveServiceJSON, err := app.state.Get([]byte(approveServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if approveServiceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var approveService data.ApproveService
	err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	approveService.Active = false
	approveServiceJSON, err = utils.ProtoDeterministicMarshal(&approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(approveServiceKey), []byte(approveServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) enableServiceDestinationByNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableServiceDestinationByNDID, Parameter: %s", param)
	var funcParam DisableServiceDestinationByNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceJSON, err := app.state.Get([]byte(serviceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if serviceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	// Check node ID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check role is AS
	if nodeDetail.Role != "AS" {
		return app.ReturnDeliverTxLog(code.RoleIsNotAS, "Role of node ID is not AS", "")
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + funcParam.NodeID
	approveServiceJSON, err := app.state.Get([]byte(approveServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if approveServiceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var approveService data.ApproveService
	err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	approveService.Active = true
	approveServiceJSON, err = utils.ProtoDeterministicMarshal(&approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(approveServiceKey), []byte(approveServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
