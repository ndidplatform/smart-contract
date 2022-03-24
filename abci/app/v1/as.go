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
	"fmt"

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v6/abci/code"
	"github.com/ndidplatform/smart-contract/v6/abci/utils"
	data "github.com/ndidplatform/smart-contract/v6/protos/data"
)

func (app *ABCIApplication) createAsResponse(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateAsResponse, Parameter: %s", param)
	var createAsResponseParam CreateAsResponseParam
	err := json.Unmarshal([]byte(param), &createAsResponseParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	requestKey := requestKeyPrefix + keySeparator + createAsResponseParam.RequestID
	requestJSON, err := app.state.GetVersioned([]byte(requestKey), 0, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if requestJSON == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestJSON), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check error code exists
	if createAsResponseParam.ErrorCode != nil {
		errorCodeKey := errorCodeKeyPrefix + keySeparator + "as" + keySeparator + fmt.Sprintf("%d", *createAsResponseParam.ErrorCode)
		hasErrorCodeKey, err := app.state.Has([]byte(errorCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if !hasErrorCodeKey {
			return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode does not exist", "")
		}
	}

	// Check closed request
	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Request is closed", "")
	}

	// Check timed out request
	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Request is timed out", "")
	}

	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + createAsResponseParam.ServiceID
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

	// Check service destination is approved by NDID
	approveServiceKey := approvedServiceKeyPrefix + keySeparator + createAsResponseParam.ServiceID + keySeparator + nodeID
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
	if !approveService.Active {
		return app.ReturnDeliverTxLog(code.ServiceDestinationIsNotActive, "Service destination is not approved by NDID", "")
	}

	// Check service destination is active
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + createAsResponseParam.ServiceID
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
			if !nodes.Node[index].Active {
				return app.ReturnDeliverTxLog(code.ServiceDestinationIsNotActive, "Service destination is not active", "")
			}
			break
		}
	}

	// Check nodeID is exist in as_id_list
	exist := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == createAsResponseParam.ServiceID {
			for _, as := range dataRequest.AsIdList {
				if as == nodeID {
					exist = true
					break
				}
			}
		}
	}
	if exist == false {
		return app.ReturnDeliverTxLog(code.NodeIDDoesNotExistInASList, "Node ID does not exist in AS list", "")
	}

	// Check Duplicate AS ID
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == createAsResponseParam.ServiceID {
			for _, asResponse := range dataRequest.ResponseList {
				if asResponse.AsId == nodeID {
					return app.ReturnDeliverTxLog(code.DuplicateASResponse, "Duplicate AS response", "")
				}
			}
		}
	}

	// Check min_as
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == createAsResponseParam.ServiceID {
			if dataRequest.MinAs > 0 {
				var nonErrorResponseCount int64 = 0
				for _, asResponse := range dataRequest.ResponseList {
					if asResponse.ErrorCode == 0 {
						nonErrorResponseCount++
					}
				}
				if nonErrorResponseCount >= dataRequest.MinAs {
					return app.ReturnDeliverTxLog(code.DataRequestIsCompleted, "Can't create AS response to a request with enough AS responses", "")
				}
				var remainingPossibleResponseCount int64 = int64(len(dataRequest.AsIdList)) - int64(len(dataRequest.ResponseList))
				if nonErrorResponseCount+remainingPossibleResponseCount < dataRequest.MinAs {
					return app.ReturnDeliverTxLog(code.DataRequestCannotBeFulfilled, "Can't create AS response to a data request that cannot be fulfilled", "")
				}
			}
		}
	}

	var signDataKey string
	var signDataValue string
	if createAsResponseParam.ErrorCode == nil {
		signDataKey = dataSignatureKeyPrefix + keySeparator + nodeID + keySeparator + createAsResponseParam.ServiceID + keySeparator + createAsResponseParam.RequestID
		signDataValue = createAsResponseParam.Signature
	}

	// Update answered_as_id_list in request
	for index, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == createAsResponseParam.ServiceID {
			var asResponse data.ASResponse
			if createAsResponseParam.ErrorCode == nil {
				asResponse = data.ASResponse{
					AsId:         nodeID,
					Signed:       true,
					ReceivedData: false,
				}
			} else {
				asResponse = data.ASResponse{
					AsId:      nodeID,
					ErrorCode: *createAsResponseParam.ErrorCode,
				}
			}
			request.DataRequestList[index].ResponseList = append(dataRequest.ResponseList, &asResponse)
		}
	}

	requestJSON, err = utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	err = app.state.SetVersioned([]byte(requestKey), []byte(requestJSON))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if createAsResponseParam.ErrorCode == nil {
		app.state.Set([]byte(signDataKey), []byte(signDataValue))
	}
	return app.ReturnDeliverTxLog(code.OK, "success", createAsResponseParam.RequestID)
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

func (app *ABCIApplication) enableServiceDestination(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableServiceDestination, Parameter: %s", param)
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
