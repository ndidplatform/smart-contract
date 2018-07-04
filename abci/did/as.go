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

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/tendermint/abci/types"
)

func signData(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SignData, Parameter: %s", param)
	var signData SignDataParam
	err := json.Unmarshal([]byte(param), &signData)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	requestKey := "Request" + "|" + signData.RequestID
	_, requestJSON := app.state.db.Get(prefixKey([]byte(requestKey)))
	if requestJSON == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request Request
	err = json.Unmarshal([]byte(requestJSON), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check IsClosed
	if request.IsClosed {
		return ReturnDeliverTxLog(code.RequestIsClosed, "Request is closed", "")
	}

	// Check IsTimedOut
	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.RequestIsTimedOut, "Request is timed out", "")
	}

	// if AS != [], Check nodeID is exist in as_id_list
	exist := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == signData.ServiceID {
			if len(dataRequest.As) == 0 {
				exist = true
				break
			} else {
				for _, as := range dataRequest.As {
					if as == nodeID {
						exist = true
						break
					}
				}
			}
		}
	}
	if exist == false {
		return ReturnDeliverTxLog(code.NodeIDIsNotExistInASList, "Node ID is not exist in AS list", "")
	}

	// Check Duplicate AS ID
	duplicate := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == signData.ServiceID {
			for _, as := range dataRequest.AnsweredAsIdList {
				if as == nodeID {
					duplicate = true
					break
				}
			}
		}
	}
	if duplicate == true {
		return ReturnDeliverTxLog(code.DuplicateAnsweredAsIDList, "Duplicate AS ID in answered AS list", "")
	}

	// Check min_as
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == signData.ServiceID {
			if len(dataRequest.AnsweredAsIdList) >= dataRequest.Count {
				return ReturnDeliverTxLog(code.DataRequestIsCompleted, "Can't sign data to data request that's enough data", "")
			}
		}
	}

	signDataKey := "SignData" + "|" + nodeID + "|" + signData.ServiceID + "|" + signData.RequestID
	signDataValue := signData.Signature
	// signDataJSON, err := json.Marshal(signData)
	// if err != nil {
	// 	return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	// }

	// Update answered_as_id_list in request
	for index, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == signData.ServiceID {
			request.DataRequestList[index].AnsweredAsIdList = append(dataRequest.AnsweredAsIdList, nodeID)
		}
	}

	requestJSON, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(requestKey), []byte(requestJSON))
	app.SetStateDB([]byte(signDataKey), []byte(signDataValue))
	return ReturnDeliverTxLog(code.OK, "success", signData.RequestID)
}

func registerServiceDestination(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterServiceDestination, Parameter: %s", param)
	var funcParam RegisterServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check Service ID
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceJSON := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceJSON == nil {
		return ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var service ServiceDetail
	err = json.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	provideServiceKey := "ProvideService" + "|" + nodeID
	_, provideServiceValue := app.state.db.Get(prefixKey([]byte(provideServiceKey)))
	var services []Service
	if provideServiceValue != nil {
		err := json.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	// Check duplicate service ID
	for _, service := range services {
		if service.ServiceID == funcParam.ServiceID {
			return ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID in provide service list", "")
		}
	}
	// Append to ProvideService list
	var newService Service
	newService.ServiceID = funcParam.ServiceID
	newService.MinAal = funcParam.MinAal
	newService.MinIal = funcParam.MinIal
	newService.Active = true
	services = append(services, newService)

	provideServiceJSON, err := json.Marshal(services)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	// Add ServiceDestination
	serviceDestinationKey := "ServiceDestination" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceDestinationKey)))

	if chkExists != nil {
		var nodes GetAsNodesByServiceIdResult
		err := json.Unmarshal([]byte(chkExists), &nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// Check duplicate node ID before add
		for _, node := range nodes.Node {
			if node.ID == nodeID {
				return ReturnDeliverTxLog(code.DuplicateNodeID, "Duplicate node ID", "")
			}
		}

		var newNode = ASNode{
			nodeID,
			getNodeNameByNodeID(nodeID, app),
			funcParam.MinIal,
			funcParam.MinAal,
			funcParam.ServiceID,
			true,
		}
		nodes.Node = append(nodes.Node, newNode)
		value, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(serviceDestinationKey), []byte(value))
	} else {
		var nodes GetAsNodesByServiceIdResult
		var newNode = ASNode{
			nodeID,
			getNodeNameByNodeID(nodeID, app),
			funcParam.MinIal,
			funcParam.MinAal,
			funcParam.ServiceID,
			true,
		}
		nodes.Node = append(nodes.Node, newNode)
		value, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(serviceDestinationKey), []byte(value))
	}
	app.SetStateDB([]byte(provideServiceKey), []byte(provideServiceJSON))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func updateServiceDestination(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateServiceDestination, Parameter: %s", param)
	var funcParam UpdateServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check Service ID
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceJSON := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceJSON == nil {
		return ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var service ServiceDetail
	err = json.Unmarshal([]byte(serviceJSON), &service)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := "ServiceDestination" + "|" + funcParam.ServiceID
	_, serviceDestinationValue := app.state.db.Get(prefixKey([]byte(serviceDestinationKey)))

	if serviceDestinationValue == nil {
		return ReturnDeliverTxLog(code.ServiceDestinationNotFound, "Service destination not found", "")
	}

	var nodes GetAsNodesByServiceIdResult
	err = json.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	for index := range nodes.Node {
		if nodes.Node[index].ID == nodeID {
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
	var services []Service
	if provideServiceValue != nil {
		err := json.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	for index, service := range services {
		if service.ServiceID == funcParam.ServiceID {
			if funcParam.MinAal > 0 {
				services[index].MinAal = funcParam.MinAal
			}
			if funcParam.MinIal > 0 {
				services[index].MinIal = funcParam.MinIal
			}
			break
		}
	}
	provideServiceJSON, err := json.Marshal(services)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	serviceDestinationJSON, err := json.Marshal(nodes)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(provideServiceKey), []byte(provideServiceJSON))
	app.SetStateDB([]byte(serviceDestinationKey), []byte(serviceDestinationJSON))
	return ReturnDeliverTxLog(code.OK, "success", "")
}
