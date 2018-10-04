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

package did

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/ndidplatform/smart-contract/protos/data"
	"github.com/tendermint/tendermint/abci/types"
)

func (app *DIDApplication) signData(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SignData, Parameter: %s", param)
	var signData SignDataParam
	err := json.Unmarshal([]byte(param), &signData)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	requestKey := "Request" + "|" + signData.RequestID
	_, requestJSON := app.state.db.Get(prefixKey([]byte(requestKey)))
	if requestJSON == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestJSON), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check IsClosed
	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Request is closed", "")
	}

	// Check IsTimedOut
	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Request is timed out", "")
	}

	// Check Service ID
	serviceKey := "Service" + "|" + signData.ServiceID
	_, serviceJSON := app.state.db.Get(prefixKey([]byte(serviceKey)))
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

	// Check service destination is approved by NDID
	approveServiceKey := "ApproveKey" + "|" + signData.ServiceID + "|" + nodeID
	_, approveServiceJSON := app.state.db.Get(prefixKey([]byte(approveServiceKey)))
	if approveServiceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var approveService data.ApproveService
	err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if !approveService.Active {
		return app.ReturnDeliverTxLog(code.ServiceDestinationIsNotActive, "Service destination is not approved by NDID", "")
	}

	// Check service destination is active
	serviceDestinationKey := "ServiceDestination" + "|" + signData.ServiceID
	_, serviceDestinationValue := app.state.db.Get(prefixKey([]byte(serviceDestinationKey)))

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
			if !nodes.Node[index].Active {
				return app.ReturnDeliverTxLog(code.ServiceDestinationIsNotActive, "Service destination is not active", "")
			}
			break
		}
	}

	// Check nodeID is exist in as_id_list
	exist := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == signData.ServiceID {
			for _, as := range dataRequest.AsIdList {
				if as == nodeID {
					exist = true
					break
				}
			}
		}
	}
	if exist == false {
		return app.ReturnDeliverTxLog(code.NodeIDIsNotExistInASList, "Node ID is not exist in AS list", "")
	}

	// Check Duplicate AS ID
	duplicate := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == signData.ServiceID {
			for _, as := range dataRequest.AnsweredAsIdList {
				if as == nodeID {
					duplicate = true
					break
				}
			}
		}
	}
	if duplicate == true {
		return app.ReturnDeliverTxLog(code.DuplicateAnsweredAsIDList, "Duplicate AS ID in answered AS list", "")
	}

	// Check min_as
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == signData.ServiceID {
			if int64(len(dataRequest.AnsweredAsIdList)) >= dataRequest.MinAs {
				return app.ReturnDeliverTxLog(code.DataRequestIsCompleted, "Can't sign data to data request that's enough data", "")
			}
		}
	}

	signDataKey := "SignData" + "|" + nodeID + "|" + signData.ServiceID + "|" + signData.RequestID
	signDataValue := signData.Signature

	// Update answered_as_id_list in request
	for index, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == signData.ServiceID {
			request.DataRequestList[index].AnsweredAsIdList = append(dataRequest.AnsweredAsIdList, nodeID)
		}
	}

	requestJSON, err = proto.Marshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(requestKey), []byte(requestJSON))
	app.SetStateDB([]byte(signDataKey), []byte(signDataValue))
	return app.ReturnDeliverTxLog(code.OK, "success", signData.RequestID)
}

func (app *DIDApplication) registerServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterServiceDestination, Parameter: %s", param)
	var funcParam RegisterServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check Service ID
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceJSON := app.state.db.Get(prefixKey([]byte(serviceKey)))
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

	provideServiceKey := "ProvideService" + "|" + nodeID
	_, provideServiceValue := app.state.db.Get(prefixKey([]byte(provideServiceKey)))
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
	approveServiceKey := "ApproveKey" + "|" + funcParam.ServiceID + "|" + nodeID
	_, approveServiceJSON := app.state.db.Get(prefixKey([]byte(approveServiceKey)))
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
	services.Services = append(services.Services, &newService)

	provideServiceJSON, err := proto.Marshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	// Add ServiceDestination
	serviceDestinationKey := "ServiceDestination" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceDestinationKey)))

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
		newNode.Active = true
		nodes.Node = append(nodes.Node, &newNode)
		value, err := proto.Marshal(&nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(serviceDestinationKey), []byte(value))
	} else {
		var nodes data.ServiceDesList
		var newNode data.ASNode
		newNode.NodeId = nodeID
		newNode.MinIal = funcParam.MinIal
		newNode.MinAal = funcParam.MinAal
		newNode.ServiceId = funcParam.ServiceID
		newNode.Active = true
		nodes.Node = append(nodes.Node, &newNode)
		value, err := proto.Marshal(&nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(serviceDestinationKey), []byte(value))
	}
	app.SetStateDB([]byte(provideServiceKey), []byte(provideServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) updateServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateServiceDestination, Parameter: %s", param)
	var funcParam UpdateServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check Service ID
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceJSON := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := "ServiceDestination" + "|" + funcParam.ServiceID
	_, serviceDestinationValue := app.state.db.Get(prefixKey([]byte(serviceDestinationKey)))

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
			break
		}
	}

	// Update PrivideService
	provideServiceKey := "ProvideService" + "|" + nodeID
	_, provideServiceValue := app.state.db.Get(prefixKey([]byte(provideServiceKey)))
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
			break
		}
	}
	provideServiceJSON, err := proto.Marshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	serviceDestinationJSON, err := proto.Marshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(provideServiceKey), []byte(provideServiceJSON))
	app.SetStateDB([]byte(serviceDestinationKey), []byte(serviceDestinationJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) disableServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableServiceDestination, Parameter: %s", param)
	var funcParam DisableServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check Service ID
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceJSON := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := "ServiceDestination" + "|" + funcParam.ServiceID
	_, serviceDestinationValue := app.state.db.Get(prefixKey([]byte(serviceDestinationKey)))

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
	provideServiceKey := "ProvideService" + "|" + nodeID
	_, provideServiceValue := app.state.db.Get(prefixKey([]byte(provideServiceKey)))
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
	provideServiceJSON, err := proto.Marshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	serviceDestinationJSON, err := proto.Marshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(provideServiceKey), []byte(provideServiceJSON))
	app.SetStateDB([]byte(serviceDestinationKey), []byte(serviceDestinationJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) enableServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableServiceDestination, Parameter: %s", param)
	var funcParam DisableServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check Service ID
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceJSON := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := "ServiceDestination" + "|" + funcParam.ServiceID
	_, serviceDestinationValue := app.state.db.Get(prefixKey([]byte(serviceDestinationKey)))

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
	provideServiceKey := "ProvideService" + "|" + nodeID
	_, provideServiceValue := app.state.db.Get(prefixKey([]byte(provideServiceKey)))
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
	provideServiceJSON, err := proto.Marshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	serviceDestinationJSON, err := proto.Marshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(provideServiceKey), []byte(provideServiceJSON))
	app.SetStateDB([]byte(serviceDestinationKey), []byte(serviceDestinationJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
