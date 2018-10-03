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
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/ndidplatform/smart-contract/protos/data"
	"github.com/tendermint/tendermint/abci/types"
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
	"DisableNode":                      true,
	"DisableServiceDestinationByNDID":  true,
	"EnableNode":                       true,
	"EnableServiceDestinationByNDID":   true,
	"EnableNamespace":                  true,
	"EnableService":                    true,
	"SetTimeOutBlockRegisterIdentity":  true,
	"AddNodeToProxyNode":               true,
	"UpdateNodeProxyNode":              true,
	"RemoveNodeFromProxyNode":          true,
}

func (app *DIDApplication) initNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("InitNDID, Parameter: %s", param)
	var funcParam InitNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	var nodeDetail data.NodeDetail
	nodeDetail.PublicKey = funcParam.PublicKey
	nodeDetail.MasterPublicKey = funcParam.MasterPublicKey
	nodeDetail.NodeName = "NDID"
	nodeDetail.Role = "NDID"
	nodeDetail.Active = true
	nodeDetailByte, err := proto.Marshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	masterNDIDKey := "MasterNDID"
	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	app.SetStateDB([]byte(masterNDIDKey), []byte(nodeID))
	app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailByte))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) registerNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterNode, Parameter: %s", param)
	var funcParam RegisterNode
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "NodeID" + "|" + funcParam.NodeID
	// check Duplicate Node ID
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return app.ReturnDeliverTxLog(code.DuplicateNodeID, "Duplicate Node ID", "")
	}

	// check role is valid
	if !(funcParam.Role == "RP" ||
		funcParam.Role == "IdP" ||
		funcParam.Role == "AS" ||
		strings.ToLower(funcParam.Role) == "proxy") {
		return app.ReturnDeliverTxLog(code.WrongRole, "Wrong Role", "")
	}

	if strings.ToLower(funcParam.Role) == "proxy" {
		funcParam.Role = "Proxy"
	}

	// create node detail
	var nodeDetail data.NodeDetail
	nodeDetail.PublicKey = funcParam.PublicKey
	nodeDetail.MasterPublicKey = funcParam.MasterPublicKey
	nodeDetail.NodeName = funcParam.NodeName
	nodeDetail.Role = funcParam.Role
	nodeDetail.Active = true

	// if node is IdP, set max_aal, min_ial
	if funcParam.Role == "IdP" {
		nodeDetail.MaxAal = funcParam.MaxAal
		nodeDetail.MaxIal = funcParam.MaxIal
	}

	// if node is IdP, add node id to IdPList
	var idpsList data.IdPList
	idpsKey := "IdPList"
	if funcParam.Role == "IdP" {
		_, idpsValue := app.state.db.Get(prefixKey([]byte(idpsKey)))
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
		}
		idpsList.NodeId = append(idpsList.NodeId, funcParam.NodeID)
		idpsListByte, err := proto.Marshal(&idpsList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(idpsKey), []byte(idpsListByte))
	}

	// if node is rp, add node id to rpList
	var rpsList data.RPList
	rpsKey := "rpList"
	if funcParam.Role == "RP" {
		_, rpsValue := app.state.db.Get(prefixKey([]byte(rpsKey)))
		if rpsValue != nil {
			err := proto.Unmarshal(rpsValue, &rpsList)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
		}
		rpsList.NodeId = append(rpsList.NodeId, funcParam.NodeID)
		rpsListByte, err := proto.Marshal(&rpsList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(rpsKey), []byte(rpsListByte))
	}

	// if node is as, add node id to asList
	var asList data.ASList
	asKey := "asList"
	if funcParam.Role == "AS" {
		_, asValue := app.state.db.Get(prefixKey([]byte(asKey)))
		if asValue != nil {
			err := proto.Unmarshal(asValue, &asList)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
		}
		asList.NodeId = append(asList.NodeId, funcParam.NodeID)
		asListByte, err := proto.Marshal(&asList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(asKey), []byte(asListByte))
	}

	var allList data.AllList
	allKey := "allList"
	_, allValue := app.state.db.Get(prefixKey([]byte(allKey)))
	if allValue != nil {
		err := proto.Unmarshal(allValue, &allList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	allList.NodeId = append(allList.NodeId, funcParam.NodeID)
	allListByte, err := proto.Marshal(&allList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(allKey), []byte(allListByte))

	nodeDetailByte, err := proto.Marshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.createTokenAccount(funcParam.NodeID)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) addNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNamespace, Parameter: %s", param)
	var funcParam Namespace
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "AllNamespace"
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces data.NamespaceList

	if chkExists != nil {
		err = proto.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// Check duplicate namespace
		for _, namespace := range namespaces.Namespaces {
			if namespace.Namespace == funcParam.Namespace {
				return app.ReturnDeliverTxLog(code.DuplicateNamespace, "Duplicate namespace", "")
			}
		}
	}

	var newNamespace data.Namespace
	newNamespace.Namespace = funcParam.Namespace
	newNamespace.Description = funcParam.Description
	// set active flag
	newNamespace.Active = true
	namespaces.Namespaces = append(namespaces.Namespaces, &newNamespace)
	value, err := proto.Marshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) disableNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableNamespace, Parameter: %s", param)
	var funcParam DisableNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "AllNamespace"
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces data.NamespaceList

	if chkExists != nil {
		err = proto.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index, namespace := range namespaces.Namespaces {
			if namespace.Namespace == funcParam.Namespace {
				namespaces.Namespaces[index].Active = false
				break
			}
		}

		value, err := proto.Marshal(&namespaces)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
		return app.ReturnDeliverTxLog(code.OK, "success", "")
	}

	return app.ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
}

func (app *DIDApplication) addService(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddService, Parameter: %s", param)
	var funcParam AddServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
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
	serviceJSON, err := proto.Marshal(&service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	// Add detail to service directory
	allServiceKey := "AllService"
	_, allServiceValue := app.state.db.Get(prefixKey([]byte(allServiceKey)))

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
	allServiceJSON, err := proto.Marshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) disableService(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableService, Parameter: %s", param)
	var funcParam DisableServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if chkExists == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}

	// Delete detail in service directory
	allServiceKey := "AllService"
	_, allServiceValue := app.state.db.Get(prefixKey([]byte(allServiceKey)))

	var services data.ServiceDetailList

	if allServiceValue != nil {
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

		allServiceJSON, err := proto.Marshal(&services)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		serviceJSON, err := proto.Marshal(&service)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
		app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	}

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) updateNodeByNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNodeByNDID, Parameter: %s", param)
	var funcParam UpdateNodeByNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Get node detail by NodeID
	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))

	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	var node data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &node)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Selective update
	if funcParam.NodeName != "" {
		node.NodeName = funcParam.NodeName
	}
	// If node is IdP then update max_ial, max_aal
	if node.Role == "IdP" {
		if funcParam.MaxIal > 0 {
			node.MaxIal = funcParam.MaxIal
		}
		if funcParam.MaxAal > 0 {
			node.MaxAal = funcParam.MaxAal
		}
	}
	nodeDetailJSON, err := proto.Marshal(&node)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) updateService(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateService, Parameter: %s", param)
	var funcParam UpdateServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceValue := app.state.db.Get(prefixKey([]byte(serviceKey)))
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
	_, allServiceValue := app.state.db.Get(prefixKey([]byte(allServiceKey)))

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

	serviceJSON, err := proto.Marshal(&service)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	allServiceJSON, err := proto.Marshal(&services)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) registerServiceDestinationByNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterServiceDestinationByNDID, Parameter: %s", param)
	var funcParam RegisterServiceDestinationByNDIDParam
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

	approveServiceKey := "ApproveKey" + "|" + funcParam.ServiceID + "|" + funcParam.NodeID
	var approveService data.ApproveService
	approveService.Active = true
	approveServiceJSON, err := proto.Marshal(&approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(approveServiceKey), []byte(approveServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) disableNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableNode, Parameter: %s", param)
	var funcParam DisableNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))

	if nodeDetailValue != nil {
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		nodeDetail.Active = false

		nodeDetailValue, err := proto.Marshal(&nodeDetail)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))
		return app.ReturnDeliverTxLog(code.OK, "success", "")
	}

	return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
}

func (app *DIDApplication) disableServiceDestinationByNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableServiceDestinationByNDID, Parameter: %s", param)
	var funcParam DisableServiceDestinationByNDIDParam
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

	approveServiceKey := "ApproveKey" + "|" + funcParam.ServiceID + "|" + funcParam.NodeID
	_, approveServiceJSON := app.state.db.Get(prefixKey([]byte(approveServiceKey)))
	if approveServiceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var approveService data.ApproveService
	err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	approveService.Active = false
	approveServiceJSON, err = proto.Marshal(&approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(approveServiceKey), []byte(approveServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) enableNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableNode, Parameter: %s", param)
	var funcParam DisableNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))

	if nodeDetailValue != nil {
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		nodeDetail.Active = true

		nodeDetailValue, err := proto.Marshal(&nodeDetail)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))
		return app.ReturnDeliverTxLog(code.OK, "success", "")
	}

	return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
}

func (app *DIDApplication) enableServiceDestinationByNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableServiceDestinationByNDID, Parameter: %s", param)
	var funcParam DisableServiceDestinationByNDIDParam
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

	approveServiceKey := "ApproveKey" + "|" + funcParam.ServiceID + "|" + funcParam.NodeID
	_, approveServiceJSON := app.state.db.Get(prefixKey([]byte(approveServiceKey)))
	if approveServiceJSON == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	var approveService data.ApproveService
	err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	approveService.Active = true
	approveServiceJSON, err = proto.Marshal(&approveService)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(approveServiceKey), []byte(approveServiceJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) enableNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableNamespace, Parameter: %s", param)
	var funcParam DisableNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "AllNamespace"
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces data.NamespaceList

	if chkExists != nil {
		err = proto.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index, namespace := range namespaces.Namespaces {
			if namespace.Namespace == funcParam.Namespace {
				namespaces.Namespaces[index].Active = true
				break
			}
		}

		value, err := proto.Marshal(&namespaces)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
		return app.ReturnDeliverTxLog(code.OK, "success", "")
	}

	return app.ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
}

func (app *DIDApplication) enableService(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableService, Parameter: %s", param)
	var funcParam DisableServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if chkExists == nil {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}

	// Delete detail in service directory
	allServiceKey := "AllService"
	_, allServiceValue := app.state.db.Get(prefixKey([]byte(allServiceKey)))

	var services data.ServiceDetailList

	if allServiceValue != nil {
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

		allServiceJSON, err := proto.Marshal(&services)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		serviceJSON, err := proto.Marshal(&service)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
		app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	}

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) setTimeOutBlockRegisterIdentity(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetTimeOutBlockRegisterIdentity, Parameter: %s", param)
	var funcParam TimeOutBlockRegisterIdentity
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "TimeOutBlockRegisterIdentity"
	var timeOut data.TimeOutBlockRegisterIdentity
	timeOut.TimeOutBlock = funcParam.TimeOutBlock
	// Check time out block > 0
	if timeOut.TimeOutBlock <= 0 {
		return app.ReturnDeliverTxLog(code.TimeOutBlockIsMustGreaterThanZero, "Time out block is must greater than 0", "")
	}
	value, err := proto.Marshal(&timeOut)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) addNodeToProxyNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNodeToProxyNode, Parameter: %s", param)
	var funcParam AddNodeToProxyNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	proxyKey := "Proxy" + "|" + funcParam.NodeID
	behindProxyNodeKey := "BehindProxyNode" + "|" + funcParam.ProxyNodeID
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)

	// Get node detail by NodeID
	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))

	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}

	// Check already associated with a proxy
	_, proxyValue := app.state.db.Get(prefixKey([]byte(proxyKey)))
	if proxyValue != nil {
		return app.ReturnDeliverTxLog(code.NodeIDIsAlreadyAssociatedWithProxyNode, "This node ID is already associated with a proxy node", "")
	}

	// Check is not proxy node
	if app.checkIsProxyNode(funcParam.NodeID) {
		return app.ReturnDeliverTxLog(code.NodeIDisProxyNode, "This node ID is an ID of a proxy node", "")
	}

	// Check ProxyNodeID is proxy node
	if !app.checkIsProxyNode(funcParam.ProxyNodeID) {
		return app.ReturnDeliverTxLog(code.ProxyNodeNotFound, "Proxy node ID not found", "")
	}

	_, behindProxyNodeValue := app.state.db.Get(prefixKey([]byte(behindProxyNodeKey)))
	if behindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}

	var proxy data.Proxy
	proxy.ProxyNodeId = funcParam.ProxyNodeID
	proxy.Config = funcParam.Config
	proxyJSON, err := proto.Marshal(&proxy)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	nodes.Nodes = append(nodes.Nodes, funcParam.NodeID)
	behindProxyNodeJSON, err := proto.Marshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	// Delete msq address
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	msqAddres := make([]*data.MQ, 0)
	nodeDetail.Mq = msqAddres
	nodeDetailByte, err := proto.Marshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.SetStateDB([]byte(proxyKey), []byte(proxyJSON))
	app.SetStateDB([]byte(behindProxyNodeKey), []byte(behindProxyNodeJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) updateNodeProxyNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNodeProxyNode, Parameter: %s", param)
	var funcParam UpdateNodeProxyNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	proxyKey := "Proxy" + "|" + funcParam.NodeID
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)

	var newProxyNodes data.BehindNodeList
	newProxyNodes.Nodes = make([]string, 0)

	// Get node detail by NodeID
	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))

	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}

	// Check already associated with a proxy
	_, proxyValue := app.state.db.Get(prefixKey([]byte(proxyKey)))
	if proxyValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDHasNotBeenAssociatedWithProxyNode, "This node has not been associated with a proxy node", "")
	}

	if funcParam.ProxyNodeID != "" {
		// Check ProxyNodeID is proxy node
		if !app.checkIsProxyNode(funcParam.ProxyNodeID) {
			return app.ReturnDeliverTxLog(code.ProxyNodeNotFound, "Proxy node ID not found", "")
		}
	}

	var proxy data.Proxy
	err = proto.Unmarshal([]byte(proxyValue), &proxy)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	behindProxyNodeKey := "BehindProxyNode" + "|" + proxy.ProxyNodeId
	_, behindProxyNodeValue := app.state.db.Get(prefixKey([]byte(behindProxyNodeKey)))
	if behindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}

	newBehindProxyNodeKey := "BehindProxyNode" + "|" + funcParam.ProxyNodeID
	_, newBehindProxyNodeValue := app.state.db.Get(prefixKey([]byte(newBehindProxyNodeKey)))
	if newBehindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(newBehindProxyNodeValue), &newProxyNodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}

	if funcParam.ProxyNodeID != "" {
		if proxy.ProxyNodeId != funcParam.ProxyNodeID {
			// Delete from old proxy list
			for i, node := range nodes.Nodes {
				if node == funcParam.NodeID {
					copy(nodes.Nodes[i:], nodes.Nodes[i+1:])
					nodes.Nodes[len(nodes.Nodes)-1] = ""
					nodes.Nodes = nodes.Nodes[:len(nodes.Nodes)-1]
				}
			}

			// Add to new proxy list
			newProxyNodes.Nodes = append(newProxyNodes.Nodes, funcParam.NodeID)
		}
		proxy.ProxyNodeId = funcParam.ProxyNodeID
	}

	if funcParam.Config != "" {
		proxy.Config = funcParam.Config
	}

	proxyJSON, err := proto.Marshal(&proxy)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	proxyValue = proxyJSON
	behindProxyNodeJSON, err := proto.Marshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	newBehindProxyNodeJSON, err := proto.Marshal(&newProxyNodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(proxyKey), []byte(proxyValue))
	app.SetStateDB([]byte(behindProxyNodeKey), []byte(behindProxyNodeJSON))
	app.SetStateDB([]byte(newBehindProxyNodeKey), []byte(newBehindProxyNodeJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) removeNodeFromProxyNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RemoveNodeFromProxyNode, Parameter: %s", param)
	var funcParam RemoveNodeFromProxyNode
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	proxyKey := "Proxy" + "|" + funcParam.NodeID
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)

	// Get node detail by NodeID
	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))

	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}

	// Check is not proxy node
	if app.checkIsProxyNode(funcParam.NodeID) {
		return app.ReturnDeliverTxLog(code.NodeIDisProxyNode, "This node ID is an ID of a proxy node", "")
	}

	// Check already associated with a proxy
	_, proxyValue := app.state.db.Get(prefixKey([]byte(proxyKey)))
	if proxyValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDHasNotBeenAssociatedWithProxyNode, "This node has not been associated with a proxy node", "")
	}

	var proxy data.Proxy
	err = proto.Unmarshal([]byte(proxyValue), &proxy)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	behindProxyNodeKey := "BehindProxyNode" + "|" + proxy.ProxyNodeId
	_, behindProxyNodeValue := app.state.db.Get(prefixKey([]byte(behindProxyNodeKey)))
	if behindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// Delete from old proxy list
		for i, node := range nodes.Nodes {
			if node == funcParam.NodeID {
				copy(nodes.Nodes[i:], nodes.Nodes[i+1:])
				nodes.Nodes[len(nodes.Nodes)-1] = ""
				nodes.Nodes = nodes.Nodes[:len(nodes.Nodes)-1]
			}
		}
	}

	behindProxyNodeJSON, err := proto.Marshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.DeleteStateDB([]byte(proxyKey))
	app.SetStateDB([]byte(behindProxyNodeKey), []byte(behindProxyNodeJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
