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
	"InitNDID":         true,
	"RegisterNode":     true,
	"AddNodeToken":     true,
	"ReduceNodeToken":  true,
	"SetNodeToken":     true,
	"SetPriceFunc":     true,
	"AddNamespace":     true,
	"DeleteNamespace":  true,
	"SetValidator":     true,
	"AddService":       true,
	"DeleteService":    true,
	"UpdateNodeByNDID": true,
	"UpdateService":    true,
}

func initNDID(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("InitNDID, Parameter: %s", param)
	var funcParam InitNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := "NodePublicKeyRole" + "|" + funcParam.PublicKey
	value := []byte("MasterNDID")
	app.SetStateDB([]byte(key), []byte(value))

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	// TODO: fix param InitNDID
	var nodeDetail = NodeDetail{
		funcParam.PublicKey,
		funcParam.PublicKey,
		"NDID",
	}
	nodeDetailValue, err := json.Marshal(nodeDetail)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))

	key = "MasterNDID"
	value = []byte(funcParam.PublicKey)
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

	// check Duplicate Master Key
	key = "NodePublicKeyRole" + "|" + funcParam.MasterPublicKey
	_, chkExists = app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicatePublicKey, "Duplicate Public Key", "")
	}

	// check Duplicate Key
	key = "NodePublicKeyRole" + "|" + funcParam.PublicKey
	_, chkExists = app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicatePublicKey, "Duplicate Public Key", "")
	}

	if funcParam.Role == "RP" ||
		funcParam.Role == "IdP" ||
		funcParam.Role == "AS" {
		nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
		var nodeDetail = NodeDetail{
			funcParam.PublicKey,
			funcParam.MasterPublicKey,
			funcParam.NodeName,
		}
		nodeDetailValue, err := json.Marshal(nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))

		// Set master Role
		publicKeyRoleKey := "NodePublicKeyRole" + "|" + funcParam.MasterPublicKey
		publicKeyRoleValue := "Master" + funcParam.Role
		app.SetStateDB([]byte(publicKeyRoleKey), []byte(publicKeyRoleValue))

		// Set Role
		publicKeyRoleKey = "NodePublicKeyRole" + "|" + funcParam.PublicKey
		publicKeyRoleValue = funcParam.Role
		app.SetStateDB([]byte(publicKeyRoleKey), []byte(publicKeyRoleValue))

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
	namespaces = append(namespaces, funcParam)
	value, err := json.Marshal(namespaces)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func deleteNamespace(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DeleteNamespace, Parameter: %s", param)
	var funcParam DeleteNamespaceParam
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
				namespaces = append(namespaces[:index], namespaces[index+1:]...)
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
	var service = Service{
		funcParam.ServiceName,
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

func deleteService(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DeleteService, Parameter: %s", param)
	var funcParam DeleteServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if chkExists == nil {
		return ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}

	// Dekete detail in service directory
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
				services = append(services[:index], services[index+1:]...)
				break
			}
		}

		allServiceJSON, err := json.Marshal(services)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(allServiceKey), []byte(allServiceJSON))
	}

	app.DeleteStateDB([]byte(serviceKey))
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
	var service Service
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
