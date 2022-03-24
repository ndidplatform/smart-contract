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
	"strconv"
	"strings"

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v6/abci/code"
	"github.com/ndidplatform/smart-contract/v6/abci/utils"
	data "github.com/ndidplatform/smart-contract/v6/protos/data"
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
	"AddErrorCode":                     true,
	"RemoveErrorCode":                  true,
	"SetInitData":                      true,
	"EndInit":                          true,
	"SetLastBlock":                     true,
	"SetAllowedModeList":               true,
	"UpdateNamespace":                  true,
	"SetAllowedMinIalForRegisterIdentityAtFirstIdp":        true,
	"SetServicePriceCeiling":                               true,
	"SetServicePriceMinEffectiveDatetimeDelay":             true,
	"AddRequestType":                                       true,
	"RemoveRequestType":                                    true,
	"AddSuppressedIdentityModificationNotificationNode":    true,
	"RemoveSuppressedIdentityModificationNotificationNode": true,
}

func (app *ABCIApplication) initNDID(param string, nodeID string) types.ResponseDeliverTx {
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
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	chainHistoryInfoKey := "ChainHistoryInfo"
	app.state.Set(masterNDIDKeyBytes, []byte(nodeID))
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set(initStateKeyBytes, []byte("true"))
	app.state.Set([]byte(chainHistoryInfoKey), []byte(funcParam.ChainHistoryInfo))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) registerNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterNode, Parameter: %s", param)
	var funcParam RegisterNode
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	// check Duplicate Node ID
	chkExists, err := app.state.Get([]byte(key), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
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
	// if node is IdP, set max_aal, min_ial, on_the_fly_support, is_idp_agent, and supported_request_message_type_list
	if funcParam.Role == "IdP" {
		nodeDetail.MaxAal = funcParam.MaxAal
		nodeDetail.MaxIal = funcParam.MaxIal
		nodeDetail.OnTheFlySupport = funcParam.OnTheFlySupport != nil && *funcParam.OnTheFlySupport
		nodeDetail.IsIdpAgent = funcParam.IsIdPAgent != nil && *funcParam.IsIdPAgent
		nodeDetail.SupportedRequestMessageDataUrlTypeList = make([]string, 0)
	}
	// if node is Idp or rp, set use_whitelist and whitelist
	if funcParam.Role == "IdP" || funcParam.Role == "RP" {
		if funcParam.UseWhitelist != nil && *funcParam.UseWhitelist {
			nodeDetail.UseWhitelist = true
		} else {
			nodeDetail.UseWhitelist = false
		}
		if funcParam.Whitelist != nil {
			// check if all node in whitelist exists
			for _, whitelistNode := range funcParam.Whitelist {
				whitelistKey := nodeIDKeyPrefix + keySeparator + whitelistNode
				hasWhitelistKey, err := app.state.Has([]byte(whitelistKey), false)
				if err != nil {
					return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
				}
				if !hasWhitelistKey {
					return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Whitelist node not exists", "")
				}
			}
			nodeDetail.Whitelist = funcParam.Whitelist
		} else {
			nodeDetail.Whitelist = []string{}
		}
	}
	// if node is IdP, add node id to IdPList
	var idpsList data.IdPList
	if funcParam.Role == "IdP" {
		idpsKey := "IdPList"
		idpsValue, err := app.state.Get([]byte(idpsKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
		}
		idpsList.NodeId = append(idpsList.NodeId, funcParam.NodeID)
		idpsListByte, err := utils.ProtoDeterministicMarshal(&idpsList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set(idpListKeyBytes, []byte(idpsListByte))
	}
	// if node is rp, add node id to rpList
	var rpsList data.RPList
	rpsKey := "rpList"
	if funcParam.Role == "RP" {
		rpsValue, err := app.state.Get([]byte(rpsKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if rpsValue != nil {
			err := proto.Unmarshal(rpsValue, &rpsList)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
		}
		rpsList.NodeId = append(rpsList.NodeId, funcParam.NodeID)
		rpsListByte, err := utils.ProtoDeterministicMarshal(&rpsList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(rpsKey), []byte(rpsListByte))
	}
	// if node is as, add node id to asList
	var asList data.ASList
	asKey := "asList"
	if funcParam.Role == "AS" {
		asValue, err := app.state.Get([]byte(asKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if asValue != nil {
			err := proto.Unmarshal(asValue, &asList)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
		}
		asList.NodeId = append(asList.NodeId, funcParam.NodeID)
		asListByte, err := utils.ProtoDeterministicMarshal(&asList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(asKey), []byte(asListByte))
	}
	var allList data.AllList
	allKey := "allList"
	allValue, err := app.state.Get([]byte(allKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if allValue != nil {
		err := proto.Unmarshal(allValue, &allList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	allList.NodeId = append(allList.NodeId, funcParam.NodeID)
	allListByte, err := utils.ProtoDeterministicMarshal(&allList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allKey), []byte(allListByte))
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.createTokenAccount(funcParam.NodeID)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) addNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNamespace, Parameter: %s", param)
	var funcParam Namespace
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	chkExists, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
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
	if funcParam.AllowedIdentifierCountInReferenceGroup != 0 {
		newNamespace.AllowedIdentifierCountInReferenceGroup = funcParam.AllowedIdentifierCountInReferenceGroup
	}
	if funcParam.AllowedActiveIdentifierCountInReferenceGroup != 0 {
		newNamespace.AllowedActiveIdentifierCountInReferenceGroup = funcParam.AllowedActiveIdentifierCountInReferenceGroup
	}
	// set active flag
	newNamespace.Active = true
	namespaces.Namespaces = append(namespaces.Namespaces, &newNamespace)
	value, err := utils.ProtoDeterministicMarshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(allNamespaceKeyBytes, []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) disableNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableNamespace, Parameter: %s", param)
	var funcParam DisableNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	chkExists, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if chkExists == nil {
		return app.ReturnDeliverTxLog(code.NamespaceNotFound, "List of namespace not found", "")
	}
	var namespaces data.NamespaceList
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
	value, err := utils.ProtoDeterministicMarshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(allNamespaceKeyBytes, []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

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

func (app *ABCIApplication) updateNodeByNDID(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNodeByNDID, Parameter: %s", param)
	var funcParam UpdateNodeByNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
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
	// If node is IdP then update max_ial, max_aal and is_idp_agent
	if node.Role == "IdP" {
		if funcParam.MaxIal > 0 {
			node.MaxIal = funcParam.MaxIal
		}
		if funcParam.MaxAal > 0 {
			node.MaxAal = funcParam.MaxAal
		}
		if funcParam.OnTheFlySupport != nil {
			node.OnTheFlySupport = *funcParam.OnTheFlySupport
		}
		if funcParam.IsIdPAgent != nil {
			node.IsIdpAgent = *funcParam.IsIdPAgent
		}
	}
	// If node is Idp or rp, update use_whitelist and whitelist
	if node.Role == "IdP" || node.Role == "RP" {
		if funcParam.UseWhitelist != nil {
			node.UseWhitelist = *funcParam.UseWhitelist
		}
		if funcParam.Whitelist != nil {
			// check if all node in whitelist exists
			for _, whitelistNode := range funcParam.Whitelist {
				whitelistKey := nodeIDKeyPrefix + keySeparator + whitelistNode
				hasWhitelistKey, err := app.state.Has([]byte(whitelistKey), false)
				if err != nil {
					return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
				}
				if !hasWhitelistKey {
					return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Whitelist node not exists", "")
				}
			}
			node.Whitelist = funcParam.Whitelist
		}
	}
	nodeDetailJSON, err := utils.ProtoDeterministicMarshal(&node)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailJSON))
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

func (app *ABCIApplication) disableNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableNode, Parameter: %s", param)
	var funcParam DisableNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
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
	nodeDetail.Active = false
	nodeDetailValue, err = utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailValue))
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

func (app *ABCIApplication) enableNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableNode, Parameter: %s", param)
	var funcParam DisableNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
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
	nodeDetail.Active = true
	nodeDetailValue, err = utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailValue))
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

func (app *ABCIApplication) enableNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableNamespace, Parameter: %s", param)
	var funcParam DisableNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	chkExists, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var namespaces data.NamespaceList
	if chkExists == nil {
		return app.ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
	}
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
	value, err := utils.ProtoDeterministicMarshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(allNamespaceKeyBytes, []byte(value))
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

func (app *ABCIApplication) setTimeOutBlockRegisterIdentity(param string, nodeID string) types.ResponseDeliverTx {
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
	value, err := utils.ProtoDeterministicMarshal(&timeOut)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) addNodeToProxyNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNodeToProxyNode, Parameter: %s", param)
	var funcParam AddNodeToProxyNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	behindProxyNodeKey := behindProxyNodeKeyPrefix + keySeparator + funcParam.ProxyNodeID
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)
	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check already associated with a proxy
	if nodeDetail.ProxyNodeId != "" {
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
	behindProxyNodeValue, err := app.state.Get([]byte(behindProxyNodeKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if behindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}

	// Set proxy node ID and proxy config
	nodeDetail.ProxyNodeId = funcParam.ProxyNodeID
	nodeDetail.ProxyConfig = funcParam.Config

	nodes.Nodes = append(nodes.Nodes, funcParam.NodeID)
	behindProxyNodeJSON, err := utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	// Delete msq address
	msqAddres := make([]*data.MQ, 0)
	nodeDetail.Mq = msqAddres
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set([]byte(behindProxyNodeKey), []byte(behindProxyNodeJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) updateNodeProxyNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNodeProxyNode, Parameter: %s", param)
	var funcParam UpdateNodeProxyNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)
	var newProxyNodes data.BehindNodeList
	newProxyNodes.Nodes = make([]string, 0)
	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check already associated with a proxy
	if nodeDetail.ProxyNodeId == "" {
		return app.ReturnDeliverTxLog(code.NodeIDHasNotBeenAssociatedWithProxyNode, "This node has not been associated with a proxy node", "")
	}
	if funcParam.ProxyNodeID != "" {
		// Check ProxyNodeID is proxy node
		if !app.checkIsProxyNode(funcParam.ProxyNodeID) {
			return app.ReturnDeliverTxLog(code.ProxyNodeNotFound, "Proxy node ID not found", "")
		}
	}
	behindProxyNodeKey := behindProxyNodeKeyPrefix + keySeparator + nodeDetail.ProxyNodeId
	behindProxyNodeValue, err := app.state.Get([]byte(behindProxyNodeKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if behindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	newBehindProxyNodeKey := behindProxyNodeKeyPrefix + keySeparator + funcParam.ProxyNodeID
	newBehindProxyNodeValue, err := app.state.Get([]byte(newBehindProxyNodeKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if newBehindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(newBehindProxyNodeValue), &newProxyNodes)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	if funcParam.ProxyNodeID != "" {
		if nodeDetail.ProxyNodeId != funcParam.ProxyNodeID {
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
		nodeDetail.ProxyNodeId = funcParam.ProxyNodeID
	}
	if funcParam.Config != "" {
		nodeDetail.ProxyConfig = funcParam.Config
	}
	behindProxyNodeJSON, err := utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	newBehindProxyNodeJSON, err := utils.ProtoDeterministicMarshal(&newProxyNodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set([]byte(behindProxyNodeKey), []byte(behindProxyNodeJSON))
	app.state.Set([]byte(newBehindProxyNodeKey), []byte(newBehindProxyNodeJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) removeNodeFromProxyNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RemoveNodeFromProxyNode, Parameter: %s", param)
	var funcParam RemoveNodeFromProxyNode
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)
	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	// Check is not proxy node
	if app.checkIsProxyNode(funcParam.NodeID) {
		return app.ReturnDeliverTxLog(code.NodeIDisProxyNode, "This node ID is an ID of a proxy node", "")
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check already associated with a proxy
	if nodeDetail.ProxyNodeId == "" {
		return app.ReturnDeliverTxLog(code.NodeIDHasNotBeenAssociatedWithProxyNode, "This node has not been associated with a proxy node", "")
	}
	behindProxyNodeKey := behindProxyNodeKeyPrefix + keySeparator + nodeDetail.ProxyNodeId
	behindProxyNodeValue, err := app.state.Get([]byte(behindProxyNodeKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
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
	// Delete node proxy ID and proxy config
	nodeDetail.ProxyNodeId = ""
	nodeDetail.ProxyConfig = ""
	behindProxyNodeJSON, err := utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set([]byte(behindProxyNodeKey), []byte(behindProxyNodeJSON))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) setLastBlock(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetLastBlock, Parameter: %s", param)
	var funcParam SetLastBlockParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	lastBlockValue := funcParam.BlockHeight
	if funcParam.BlockHeight == 0 {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	if funcParam.BlockHeight < -1 {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	if funcParam.BlockHeight > 0 && funcParam.BlockHeight < app.state.CurrentBlockHeight {
		lastBlockValue = app.state.CurrentBlockHeight
	}
	app.state.Set(lastBlockKeyBytes, []byte(strconv.FormatInt(lastBlockValue, 10)))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) SetInitData(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetInitData, Parameter: %s", param)
	var funcParam SetInitDataParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for _, kv := range funcParam.KVList {
		app.state.Set(kv.Key, kv.Value)
	}
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) EndInit(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EndInit, Parameter: %s", param)
	var funcParam EndInitParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	app.state.Set(initStateKeyBytes, []byte("false"))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) SetAllowedModeList(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetAllowedModeList, Parameter: %s", param)
	var funcParam SetAllowedModeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	allowedModeKey := allowedModeListKeyPrefix + keySeparator + funcParam.Purpose
	var allowedModeList data.AllowedModeList
	allowedModeList.Mode = funcParam.AllowedModeList
	allowedModeListByte, err := utils.ProtoDeterministicMarshal(&allowedModeList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allowedModeKey), allowedModeListByte)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) SetAllowedMinIalForRegisterIdentityAtFirstIdp(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetAllowedMinIalForRegisterIdentityAtFirstIdp, Parameter: %s", param)
	var funcParam SetAllowedMinIalForRegisterIdentityAtFirstIdpParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	allowedMinIalKey := "AllowedMinIalForRegisterIdentityAtFirstIdp"
	var allowedMinIal data.AllowedMinIalForRegisterIdentityAtFirstIdp
	allowedMinIal.MinIal = funcParam.MinIal
	allowedMinIalByte, err := utils.ProtoDeterministicMarshal(&allowedMinIal)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allowedMinIalKey), allowedMinIalByte)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) updateNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNamespace, Parameter: %s", param)
	var funcParam UpdateNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	allNamespaceValue, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if allNamespaceValue == nil {
		return app.ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespaceValue), &namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for index, namespace := range namespaces.Namespaces {
		if namespace.Namespace == funcParam.Namespace {
			if funcParam.Description != "" {
				namespaces.Namespaces[index].Description = funcParam.Description
			}
			if funcParam.AllowedIdentifierCountInReferenceGroup != 0 {
				namespaces.Namespaces[index].AllowedIdentifierCountInReferenceGroup = funcParam.AllowedIdentifierCountInReferenceGroup
			}
			if funcParam.AllowedActiveIdentifierCountInReferenceGroup != 0 {
				namespaces.Namespaces[index].AllowedActiveIdentifierCountInReferenceGroup = funcParam.AllowedActiveIdentifierCountInReferenceGroup
			}
			break
		}
	}
	allNamespaceValue, err = utils.ProtoDeterministicMarshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.state.Set(allNamespaceKeyBytes, []byte(allNamespaceValue))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (*ABCIApplication) checkErrorCodeType(errorCodeType string) bool {
	return contains(errorCodeType, []string{"idp", "as"})
}

func (app *ABCIApplication) addErrorCode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddErrorCode, Parameter: %s", param)
	var funcParam AddErrorCodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// convert error type to lower case
	funcParam.Type = strings.ToLower(funcParam.Type)
	if !app.checkErrorCodeType(funcParam.Type) {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "Invalid error code type", "")
	}
	if funcParam.ErrorCode == 0 {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode cannot be 0", "")
	}

	errorCode := data.ErrorCode{
		ErrorCode:   funcParam.ErrorCode,
		Description: funcParam.Description,
	}

	// add error code
	errorCodeBytes, err := utils.ProtoDeterministicMarshal(&errorCode)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	errorKey := errorCodeKeyPrefix + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", errorCode.ErrorCode)
	hasErrorKey, err := app.state.Has([]byte(errorKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if hasErrorKey {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode is already in the database", "")
	}
	app.state.Set([]byte(errorKey), []byte(errorCodeBytes))

	// add error code to ErrorCodeList
	var errorCodeList data.ErrorCodeList
	errorsKey := errorCodeListKeyPrefix + keySeparator + funcParam.Type
	errorCodeListBytes, err := app.state.Get([]byte(errorsKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if errorCodeListBytes != nil {
		err := proto.Unmarshal(errorCodeListBytes, &errorCodeList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}
	errorCodeList.ErrorCode = append(errorCodeList.ErrorCode, &errorCode)
	errorCodeListBytes, err = utils.ProtoDeterministicMarshal(&errorCodeList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(errorsKey), []byte(errorCodeListBytes))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) removeErrorCode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RemoveErrorCode, Parameter: %s", param)
	var funcParam RemoveErrorCodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// remove error code from ErrorCode index
	errorKey := errorCodeKeyPrefix + keySeparator + funcParam.Type + keySeparator + fmt.Sprintf("%d", funcParam.ErrorCode)
	hasErrorKey, err := app.state.Has([]byte(errorKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if !hasErrorKey {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode not exists", "")
	}
	err = app.state.Delete([]byte(errorKey))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}

	// remove ErrorCode from ErrorCodeList
	var errorCodeList data.ErrorCodeList
	errorsKey := errorCodeListKeyPrefix + keySeparator + funcParam.Type
	errorCodeListBytes, err := app.state.Get([]byte(errorsKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if errorCodeListBytes != nil {
		err := proto.Unmarshal(errorCodeListBytes, &errorCodeList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
	}

	newErrorCodeList := data.ErrorCodeList{
		ErrorCode: make([]*data.ErrorCode, 0, len(errorCodeList.ErrorCode)),
	}
	for _, errorCode := range errorCodeList.ErrorCode {
		if errorCode.ErrorCode != funcParam.ErrorCode {
			newErrorCodeList.ErrorCode = append(newErrorCodeList.ErrorCode, errorCode)
		}
	}

	if len(newErrorCodeList.ErrorCode) != len(errorCodeList.ErrorCode)-1 {
		return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode not exists", "")
	}

	errorCodeListBytes, err = utils.ProtoDeterministicMarshal(&newErrorCodeList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	app.state.Set([]byte(errorsKey), []byte(errorCodeListBytes))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
