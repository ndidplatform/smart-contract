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
	"strings"

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

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
