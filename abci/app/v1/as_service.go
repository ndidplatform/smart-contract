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

type RegisterServiceDestinationParam struct {
	MinAal                 float64  `json:"min_aal"`
	MinIal                 float64  `json:"min_ial"`
	ServiceID              string   `json:"service_id"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

func (app *ABCIApplication) registerServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterServiceDestination, Parameter: %s", param)
	var funcParam RegisterServiceDestinationParam
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
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check service is active
	if !service.Active {
		return app.ReturnDeliverTxLog(code.ServiceIsNotActive, "Service is not active", "")
	}

	provideServiceKey := providedServicesKeyPrefix + keySeparator + nodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	// Check duplicate service ID
	for _, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			return app.ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID in provide service list", "")
		}
	}

	// Check approve register service destination from NDID
	approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + nodeID
	approveServiceJSON, err := app.state.Get([]byte(approveServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if approveServiceJSON == nil {
		return app.ReturnDeliverTxLog(code.NoPermissionForRegisterServiceDestination, "This node does not have permission to register service destination", "")
	}
	var approveService data.ApproveService
	err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if approveService.Active == false {
		return app.ReturnDeliverTxLog(code.NoPermissionForRegisterServiceDestination, "This node does not have permission to register service destination", "")
	}

	// Append to ProvideService list
	var newService data.Service
	newService.ServiceId = funcParam.ServiceID
	newService.MinAal = funcParam.MinAal
	newService.MinIal = funcParam.MinIal
	newService.Active = true
	newService.SupportedNamespaceList = funcParam.SupportedNamespaceList
	services.Services = append(services.Services, &newService)

	provideServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	// Add ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	chkExists, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}

	if chkExists != nil {
		var nodes data.ServiceDesList
		err := proto.Unmarshal([]byte(chkExists), &nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// Check duplicate node ID before add
		for _, node := range nodes.Node {
			if node.NodeId == nodeID {
				return app.ReturnDeliverTxLog(code.DuplicateNodeID, "Duplicate node ID", "")
			}
		}

		var newNode data.ASNode
		newNode.NodeId = nodeID
		newNode.MinIal = funcParam.MinIal
		newNode.MinAal = funcParam.MinAal
		newNode.ServiceId = funcParam.ServiceID
		newNode.SupportedNamespaceList = funcParam.SupportedNamespaceList
		newNode.Active = true
		nodes.Node = append(nodes.Node, &newNode)
		value, err := utils.ProtoDeterministicMarshal(&nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(serviceDestinationKey), []byte(value))
	} else {
		var nodes data.ServiceDesList
		var newNode data.ASNode
		newNode.NodeId = nodeID
		newNode.MinIal = funcParam.MinIal
		newNode.MinAal = funcParam.MinAal
		newNode.ServiceId = funcParam.ServiceID
		newNode.SupportedNamespaceList = funcParam.SupportedNamespaceList
		newNode.Active = true
		nodes.Node = append(nodes.Node, &newNode)
		value, err := utils.ProtoDeterministicMarshal(&nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(serviceDestinationKey), []byte(value))
	}
	app.state.Set([]byte(provideServiceKey), []byte(provideServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type UpdateServiceDestinationParam struct {
	ServiceID              string   `json:"service_id"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

func (app *ABCIApplication) updateServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateServiceDestination, Parameter: %s", param)
	var funcParam UpdateServiceDestinationParam
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
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}

	if serviceDestinationValue == nil {
		return app.ReturnDeliverTxLog(code.ServiceDestinationNotFound, "Service destination not found", "")
	}

	var nodes data.ServiceDesList
	err = proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	for index := range nodes.Node {
		if nodes.Node[index].NodeId == nodeID {
			// selective update
			if funcParam.MinAal > 0 {
				nodes.Node[index].MinAal = funcParam.MinAal
			}
			if funcParam.MinIal > 0 {
				nodes.Node[index].MinIal = funcParam.MinIal
			}
			if len(funcParam.SupportedNamespaceList) > 0 {
				nodes.Node[index].SupportedNamespaceList = funcParam.SupportedNamespaceList
			}
			break
		}
	}

	// Update ProvideService
	provideServiceKey := providedServicesKeyPrefix + keySeparator + nodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	for index, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			if funcParam.MinAal > 0 {
				services.Services[index].MinAal = funcParam.MinAal
			}
			if funcParam.MinIal > 0 {
				services.Services[index].MinIal = funcParam.MinIal
			}
			if len(funcParam.SupportedNamespaceList) > 0 {
				services.Services[index].SupportedNamespaceList = funcParam.SupportedNamespaceList
			}
			break
		}
	}
	provideServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	serviceDestinationJSON, err := utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(provideServiceKey), []byte(provideServiceJSON))
	app.state.Set([]byte(serviceDestinationKey), []byte(serviceDestinationJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type DisableServiceDestinationParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) disableServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableServiceDestination, Parameter: %s", param)
	var funcParam DisableServiceDestinationParam
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
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}

	if serviceDestinationValue == nil {
		return app.ReturnDeliverTxLog(code.ServiceDestinationNotFound, "Service destination not found", "")
	}

	var nodes data.ServiceDesList
	err = proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	for index := range nodes.Node {
		if nodes.Node[index].NodeId == nodeID {
			nodes.Node[index].Active = false
			break
		}
	}

	// Update ProvideService
	provideServiceKey := providedServicesKeyPrefix + keySeparator + nodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	for index, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			services.Services[index].Active = false
			break
		}
	}
	provideServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	serviceDestinationJSON, err := utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(provideServiceKey), []byte(provideServiceJSON))
	app.state.Set([]byte(serviceDestinationKey), []byte(serviceDestinationJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type EnableServiceDestinationParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) enableServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableServiceDestination, Parameter: %s", param)
	var funcParam EnableServiceDestinationParam
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
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if serviceDestinationValue == nil {
		return app.ReturnDeliverTxLog(code.ServiceDestinationNotFound, "Service destination not found", "")
	}

	var nodes data.ServiceDesList
	err = proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	for index := range nodes.Node {
		if nodes.Node[index].NodeId == nodeID {
			nodes.Node[index].Active = true
			break
		}
	}

	// Update ProvideService
	provideServiceKey := providedServicesKeyPrefix + keySeparator + nodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	for index, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			services.Services[index].Active = true
			break
		}
	}
	provideServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	serviceDestinationJSON, err := utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(provideServiceKey), []byte(provideServiceJSON))
	app.state.Set([]byte(serviceDestinationKey), []byte(serviceDestinationJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
