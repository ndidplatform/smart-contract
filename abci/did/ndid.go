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
	"github.com/tendermint/abci/types"
)

var isNDIDMethod = map[string]bool{
	"InitNDID":                         true,
	"RegisterNode":                     true,
	"AddNodeToken":                     true,
	"ReduceNodeToken":                  true,
	"SetNodeToken":                     true,
	"SetPriceFunc":                     true,
	"AddNamespace":                     true,
	"DisableNamespace":                 true,
	"SetValidator":                     true,
	"AddService":                       true,
	"DisableService":                   true,
	"UpdateNodeByNDID":                 true,
	"UpdateService":                    true,
	"RegisterServiceDestinationByNDID": true,
	"UpdateServiceDestinationByNDID":   true,
	"DisableNode":                      true,
	"DisableServiceDestinationByNDID":  true,
	"EnableNode":                       true,
	"EnableServiceDestinationByNDID":   true,
	"EnableNamespace":                  true,
	"EnableService":                    true,
}

func initNDID(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("InitNDID, Parameter: %s", param)
	var funcParam InitNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// key := "NodePublicKeyRole" + "|" + funcParam.PublicKey
	// value := []byte("MasterNDID")
	// app.SetStateDB([]byte(key), []byte(value))

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	// TODO: fix param InitNDID
	var nodeDetail = NodeDetail{
		funcParam.PublicKey,
		funcParam.PublicKey,
		"NDID",
		"NDID",
		true,
	}
	nodeDetailValue, err := json.Marshal(nodeDetail)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))

	key := "MasterNDID"
	value := []byte(funcParam.PublicKey)
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func registerNode(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterNode, Parameter: %s", param)
	var funcParam RegisterNode
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "NodeID" + "|" + funcParam.NodeID
	// check Duplicate Node ID
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicateNodeID, "Duplicate Node ID", "")
	}

	if funcParam.Role == "RP" ||
		funcParam.Role == "IdP" ||
		funcParam.Role == "AS" {
		nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
		var nodeDetail = NodeDetail{
			funcParam.PublicKey,
			funcParam.MasterPublicKey,
			funcParam.NodeName,
			funcParam.Role,
			true,
		}
		nodeDetailValue, err := json.Marshal(nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))

		createTokenAccount(funcParam.NodeID, app)

		// Add max_aal, min_ial when node is IdP
		if funcParam.Role == "IdP" {
			maxIalAalKey := "MaxIalAalNode" + "|" + funcParam.NodeID
			var maxIalAal MaxIalAal
			maxIalAal.MaxAal = funcParam.MaxAal
			maxIalAal.MaxIal = funcParam.MaxIal
			maxIalAalValue, err := json.Marshal(maxIalAal)
			if err != nil {
				return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
			}
			app.SetStateDB([]byte(maxIalAalKey), []byte(maxIalAalValue))

			// Save all IdP's nodeID for GetIdpNodes
			idpsKey := "IdPList"
			_, idpsValue := app.state.db.Get(prefixKey([]byte(idpsKey)))
			var idpsList []string
			if idpsValue != nil {
				err := json.Unmarshal([]byte(idpsValue), &idpsList)
				if err != nil {
					return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
				}
			}
			idpsList = append(idpsList, funcParam.NodeID)
			idpsValue, err = json.Marshal(idpsList)
			if err != nil {
				return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
			}
			app.SetStateDB([]byte(idpsKey), []byte(idpsValue))
		}

		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.WrongRole, "Wrong Role", "")
}

func addNamespace(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNamespace, Parameter: %s", param)
	var funcParam Namespace
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "AllNamespace"
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces []Namespace

	if chkExists != nil {
		err = json.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// Check duplicate namespace
		for _, namespace := range namespaces {
			if namespace.Namespace == funcParam.Namespace {
				return ReturnDeliverTxLog(code.DuplicateNamespace, "Duplicate namespace", "")
			}
		}
	}
	// set active flag
	funcParam.Active = true
	namespaces = append(namespaces, funcParam)
	value, err := json.Marshal(namespaces)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func disableNamespace(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableNamespace, Parameter: %s", param)
	var funcParam DisableNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "AllNamespace"
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces []Namespace

	if chkExists != nil {
		err = json.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index, namespace := range namespaces {
			if namespace.Namespace == funcParam.Namespace {
				namespaces[index].Active = false
				break
			}
		}

		value, err := json.Marshal(namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}

	return ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
}

func addService(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddService, Parameter: %s", param)
	var funcParam AddServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID", "")
	}

	// Add new service
	var service = ServiceDetail{
		funcParam.ServiceID,
		funcParam.ServiceName,
		true,
	}
	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	// Add detail to service directory
	allServiceKey := "AllService"
	_, allServiceValue := app.state.db.Get(prefixKey([]byte(allServiceKey)))

	var services []ServiceDetail

	if allServiceValue != nil {
		err = json.Unmarshal([]byte(allServiceValue), &services)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// Check duplicate service
		for _, service := range services {
			if service.ServiceID == funcParam.ServiceID {
				return ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID", "")
			}
		}
	}
	newService := ServiceDetail{
		funcParam.ServiceID,
		funcParam.ServiceName,
		true,
	}
	services = append(services, newService)
	allServiceJSON, err := json.Marshal(services)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func disableService(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableService, Parameter: %s", param)
	var funcParam DisableServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if chkExists == nil {
		return ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}

	// Delete detail in service directory
	allServiceKey := "AllService"
	_, allServiceValue := app.state.db.Get(prefixKey([]byte(allServiceKey)))

	var services []ServiceDetail

	if allServiceValue != nil {
		err = json.Unmarshal([]byte(allServiceValue), &services)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index, service := range services {
			if service.ServiceID == funcParam.ServiceID {
				services[index].Active = false
				break
			}
		}

		var service ServiceDetail
		err = json.Unmarshal([]byte(chkExists), &service)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		service.Active = false

		allServiceJSON, err := json.Marshal(services)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		serviceJSON, err := json.Marshal(service)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
		app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	}

	return ReturnDeliverTxLog(code.OK, "success", "")
}

func updateNodeByNDID(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNodeByNDID, Parameter: %s", param)
	var funcParam UpdateNodeByNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	maxIalAalKey := "MaxIalAalNode" + "|" + funcParam.NodeID
	_, maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
	if maxIalAalValue != nil {
		var maxIalAal MaxIalAal
		err = json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// Selective update
		if funcParam.MaxIal > 0 {
			maxIalAal.MaxIal = funcParam.MaxIal
		}
		if funcParam.MaxAal > 0 {
			maxIalAal.MaxAal = funcParam.MaxAal
		}
		maxIalAalJSON, err := json.Marshal(maxIalAal)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(maxIalAalKey), []byte(maxIalAalJSON))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
}

func updateService(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateService, Parameter: %s", param)
	var funcParam UpdateServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceValue := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceValue == nil {
		return ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	// Update service
	var service ServiceDetail
	err = json.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if funcParam.ServiceName != "" {
		service.ServiceName = funcParam.ServiceName
	}

	// Update detail in service directory
	allServiceKey := "AllService"
	_, allServiceValue := app.state.db.Get(prefixKey([]byte(allServiceKey)))

	var services []ServiceDetail

	if allServiceValue != nil {
		err = json.Unmarshal([]byte(allServiceValue), &services)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// Update service
		for index, service := range services {
			if service.ServiceID == funcParam.ServiceID {
				if funcParam.ServiceName != "" {
					services[index].ServiceName = funcParam.ServiceName
				}
			}
		}
	}

	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	allServiceJSON, err := json.Marshal(services)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func registerServiceDestinationByNDID(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterServiceDestinationByNDID, Parameter: %s", param)
	var funcParam RegisterServiceDestinationByNDIDParam
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

	provideServiceKey := "ProvideService" + "|" + funcParam.NodeID
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
			if node.ID == funcParam.NodeID {
				return ReturnDeliverTxLog(code.DuplicateNodeID, "Duplicate node ID", "")
			}
		}

		var newNode = ASNode{
			funcParam.NodeID,
			getNodeNameByNodeID(funcParam.NodeID, app),
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
			funcParam.NodeID,
			getNodeNameByNodeID(funcParam.NodeID, app),
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

func updateServiceDestinationByNDID(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateServiceDestinationByNDID, Parameter: %s", param)
	var funcParam UpdateServiceDestinationByNDIDParam
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
		if nodes.Node[index].ID == funcParam.NodeID {
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
	provideServiceKey := "ProvideService" + "|" + funcParam.NodeID
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

func disableNode(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableNode, Parameter: %s", param)
	var funcParam DisableNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))

	if nodeDetailValue != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		nodeDetail.Active = false

		nodeDetailValue, err := json.Marshal(nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}

	return ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
}

func disableServiceDestinationByNDID(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableServiceDestinationByNDID, Parameter: %s", param)
	var funcParam DisableServiceDestinationByNDIDParam
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
		if nodes.Node[index].ID == funcParam.NodeID {
			nodes.Node[index].Active = false
			break
		}
	}

	// Update PrivideService
	provideServiceKey := "ProvideService" + "|" + funcParam.NodeID
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
			services[index].Active = false
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

func enableNode(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableNode, Parameter: %s", param)
	var funcParam DisableNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))

	if nodeDetailValue != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		nodeDetail.Active = true

		nodeDetailValue, err := json.Marshal(nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}

	return ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
}

func enableServiceDestinationByNDID(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableServiceDestinationByNDID, Parameter: %s", param)
	var funcParam DisableServiceDestinationByNDIDParam
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
		if nodes.Node[index].ID == funcParam.NodeID {
			nodes.Node[index].Active = true
			break
		}
	}

	// Update PrivideService
	provideServiceKey := "ProvideService" + "|" + funcParam.NodeID
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
			services[index].Active = true
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

func enableNamespace(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableNamespace, Parameter: %s", param)
	var funcParam DisableNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "AllNamespace"
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces []Namespace

	if chkExists != nil {
		err = json.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index, namespace := range namespaces {
			if namespace.Namespace == funcParam.Namespace {
				namespaces[index].Active = true
				break
			}
		}

		value, err := json.Marshal(namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}

	return ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
}

func enableService(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableService, Parameter: %s", param)
	var funcParam DisableServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if chkExists == nil {
		return ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}

	// Delete detail in service directory
	allServiceKey := "AllService"
	_, allServiceValue := app.state.db.Get(prefixKey([]byte(allServiceKey)))

	var services []ServiceDetail

	if allServiceValue != nil {
		err = json.Unmarshal([]byte(allServiceValue), &services)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index, service := range services {
			if service.ServiceID == funcParam.ServiceID {
				services[index].Active = true
				break
			}
		}

		var service ServiceDetail
		err = json.Unmarshal([]byte(chkExists), &service)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		service.Active = true

		allServiceJSON, err := json.Marshal(services)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		serviceJSON, err := json.Marshal(service)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
		app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	}

	return ReturnDeliverTxLog(code.OK, "success", "")
}
