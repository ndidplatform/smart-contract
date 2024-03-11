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
	"strconv"
	"strings"

	goleveldbutil "github.com/syndtr/goleveldb/leveldb/util"
	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	appTypes "github.com/ndidplatform/smart-contract/v9/abci/app/v1/types"
	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type MsqAddress struct {
	IP   string `json:"ip"`
	Port int64  `json:"port"`
}

type SetMqAddressesParam struct {
	Addresses []MsqAddress `json:"addresses"`
}

func (app *ABCIApplication) validateSetMqAddresses(funcParam SetMqAddressesParam, callerNodeID string, committedState bool) error {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + callerNodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if appTypes.NodeRole(node.Role) != appTypes.NodeRoleRp &&
		appTypes.NodeRole(node.Role) != appTypes.NodeRoleIdp &&
		appTypes.NodeRole(node.Role) != appTypes.NodeRoleAs &&
		appTypes.NodeRole(node.Role) != appTypes.NodeRoleProxy {
		return &ApplicationError{
			Code:    code.NoPermissionForSetMqAddresses,
			Message: "This node does not have permission to set MQ addresses",
		}
	}

	return nil
}

func (app *ABCIApplication) setMqAddressesCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetMqAddressesParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetMqAddresses(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setMqAddresses(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetMqAddresses, Parameter: %s", param)
	var funcParam SetMqAddressesParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetMqAddresses(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + callerNodeID
	value, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	var msqAddress []*data.MQ
	for _, address := range funcParam.Addresses {
		var msq data.MQ
		msq.Ip = address.IP
		msq.Port = address.Port
		msqAddress = append(msqAddress, &msq)
	}
	nodeDetail.Mq = msqAddress

	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type GetNodeSigningMasterPublicKeyParam struct {
	NodeID  string `json:"node_id"`
	Version int64  `json:"version"` // TODO
}

type GetNodeSigningMasterPublicKeyResult NodeKey

func (app *ABCIApplication) getNodeSigningMasterPublicKey(param []byte) types.ResponseQuery {
	app.logger.Infof("GetNodeSigningMasterPublicKey, Parameter: %s", param)
	var funcParam GetNodeSigningMasterPublicKeyParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var res NodeKey
	if value == nil {
		valueJSON, err := json.Marshal(res)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	res = NodeKey{
		PublicKey:           nodeDetail.SigningMasterPublicKey.PublicKey,
		Algorithm:           nodeDetail.SigningMasterPublicKey.Algorithm,
		Version:             nodeDetail.SigningMasterPublicKey.Version,
		CreationBlockHeight: nodeDetail.SigningMasterPublicKey.CreationBlockHeight,
		CreationChainID:     nodeDetail.SigningMasterPublicKey.CreationChainId,
		Active:              nodeDetail.SigningMasterPublicKey.Active,
	}
	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)
}

type GetNodeSigningPublicKeyParam struct {
	NodeID  string `json:"node_id"`
	Version int64  `json:"version"` // TODO
}

type GetNodeSigningPublicKeyResult NodeKey

func (app *ABCIApplication) getNodeSigningPublicKey(param []byte) types.ResponseQuery {
	app.logger.Infof("GetNodeSigningPublicKey, Parameter: %s", param)
	var funcParam GetNodeSigningPublicKeyParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var res NodeKey
	if value == nil {
		valueJSON, err := json.Marshal(res)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	res = NodeKey{
		PublicKey:           nodeDetail.SigningPublicKey.PublicKey,
		Algorithm:           nodeDetail.SigningPublicKey.Algorithm,
		Version:             nodeDetail.SigningPublicKey.Version,
		CreationBlockHeight: nodeDetail.SigningPublicKey.CreationBlockHeight,
		CreationChainID:     nodeDetail.SigningPublicKey.CreationChainId,
		Active:              nodeDetail.SigningPublicKey.Active,
	}
	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)
}

type GetNodeEncryptionPublicKeyParam struct {
	NodeID  string `json:"node_id"`
	Version int64  `json:"version"` // TODO
}

type GetNodeEncryptionPublicKeyResult NodeKey

func (app *ABCIApplication) getNodeEncryptionPublicKey(param []byte) types.ResponseQuery {
	app.logger.Infof("GetNodeEncryptionPublicKey, Parameter: %s", param)
	var funcParam GetNodeEncryptionPublicKeyParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var res NodeKey
	if value == nil {
		valueJSON, err := json.Marshal(res)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	res = NodeKey{
		PublicKey:           nodeDetail.EncryptionPublicKey.PublicKey,
		Algorithm:           nodeDetail.EncryptionPublicKey.Algorithm,
		Version:             nodeDetail.EncryptionPublicKey.Version,
		CreationBlockHeight: nodeDetail.EncryptionPublicKey.CreationBlockHeight,
		CreationChainID:     nodeDetail.EncryptionPublicKey.CreationChainId,
		Active:              nodeDetail.EncryptionPublicKey.Active,
	}
	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getNodeNameByNodeID(nodeID string) string {
	key := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		panic(err)
	}
	if value == nil {
		return ""
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(value), &nodeDetail)
	if err != nil {
		return ""
	}
	return nodeDetail.NodeName
}

type GetIdpNodesParam struct {
	ReferenceGroupCode                     string   `json:"reference_group_code"`
	IdentityNamespace                      string   `json:"identity_namespace"`
	IdentityIdentifierHash                 string   `json:"identity_identifier_hash"`
	FilterForNodeID                        *string  `json:"filter_for_node_id"`
	IsIdpAgent                             *bool    `json:"agent"`
	MinAal                                 float64  `json:"min_aal"`
	MinIal                                 float64  `json:"min_ial"`
	SupportedFeatureList                   []string `json:"supported_feature_list"`
	NodeIDList                             []string `json:"node_id_list"`
	SupportedRequestMessageDataUrlTypeList []string `json:"supported_request_message_data_url_type_list"`
	ModeList                               []int32  `json:"mode_list"`
}

type GetIdpNodesResult struct {
	Node []MsqDestinationNode `json:"node"`
}

type MsqDestinationNode struct {
	ID                                     string   `json:"node_id"`
	Name                                   string   `json:"node_name"`
	MaxIal                                 float64  `json:"max_ial"`
	MaxAal                                 float64  `json:"max_aal"`
	SupportedFeatureList                   []string `json:"supported_feature_list"`
	Ial                                    *float64 `json:"ial,omitempty"`
	Lial                                   *bool    `json:"lial"`
	Laal                                   *bool    `json:"laal"`
	ModeList                               *[]int32 `json:"mode_list,omitempty"`
	SupportedRequestMessageDataUrlTypeList []string `json:"supported_request_message_data_url_type_list"`
	IsIdpAgent                             bool     `json:"agent"`
}

func (app *ABCIApplication) getIdpNodes(param []byte) types.ResponseQuery {
	app.logger.Infof("GetIdpNodes, Parameter: %s", param)
	var funcParam GetIdpNodesParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	// fetch Filter RP node detail
	var nodeToFilterForDetail *data.NodeDetail
	if funcParam.FilterForNodeID != nil {
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + *funcParam.FilterForNodeID
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if nodeDetailValue == nil {
			return app.ReturnQuery(nil, "Filter RP does not exists", app.state.Height)
		}
		nodeToFilterForDetail = &data.NodeDetail{}
		if err := proto.Unmarshal(nodeDetailValue, nodeToFilterForDetail); err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
	}

	// getMsqDestionationNode returns MsqDestinationNode if nodeID is valid
	// otherwise return nil
	getMsqDestinationNode := func(nodeID string) *MsqDestinationNode {
		// check if Idp in Filter node whitelist
		if nodeToFilterForDetail != nil && nodeToFilterForDetail.UseWhitelist &&
			!contains(nodeID, nodeToFilterForDetail.Whitelist) {
			return nil
		}

		nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
		if err != nil {
			return nil
		}
		if nodeDetailValue == nil {
			return nil
		}
		var nodeDetail data.NodeDetail
		err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			return nil
		}
		// check node is active
		if !nodeDetail.Active {
			return nil
		}
		// check Max IAL && AAL
		if !(nodeDetail.MaxIal >= funcParam.MinIal &&
			nodeDetail.MaxAal >= funcParam.MinAal) {
			return nil
		}
		// Filter by node_id_list
		if len(funcParam.NodeIDList) > 0 && !contains(nodeID, funcParam.NodeIDList) {
			return nil
		}
		// Filter by SupportedFeatureList
		if funcParam.SupportedFeatureList != nil {
			for _, supportedFeature := range funcParam.SupportedFeatureList {
				if !contains(supportedFeature, nodeDetail.SupportedFeatureList) {
					return nil
				}
			}
		}
		// Filter by IsIdpAgent
		if funcParam.IsIdpAgent != nil && *funcParam.IsIdpAgent != nodeDetail.IsIdpAgent {
			return nil
		}
		// Filter by supported_request_message_data_url_type_list
		if len(funcParam.SupportedRequestMessageDataUrlTypeList) > 0 {
			// foundSupported := false
			supportedCount := 0
			for _, supportedType := range nodeDetail.SupportedRequestMessageDataUrlTypeList {
				if contains(supportedType, funcParam.SupportedRequestMessageDataUrlTypeList) {
					supportedCount++
				}
			}
			if supportedCount < len(funcParam.SupportedRequestMessageDataUrlTypeList) {
				return nil
			}
		}
		// Check if Filter RP is in Idp whitelist
		if funcParam.FilterForNodeID != nil && nodeDetail.UseWhitelist &&
			!contains(*funcParam.FilterForNodeID, nodeDetail.Whitelist) {
			return nil
		}

		return &MsqDestinationNode{
			ID:                                     nodeID,
			Name:                                   nodeDetail.NodeName,
			MaxIal:                                 nodeDetail.MaxIal,
			MaxAal:                                 nodeDetail.MaxAal,
			SupportedFeatureList:                   append(make([]string, 0), nodeDetail.SupportedFeatureList...),
			SupportedRequestMessageDataUrlTypeList: append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...),
			IsIdpAgent:                             nodeDetail.IsIdpAgent,
		}
	}

	var returnNodes GetIdpNodesResult

	if funcParam.ReferenceGroupCode == "" && funcParam.IdentityNamespace == "" && funcParam.IdentityIdentifierHash == "" {
		// fetch every idp nodes from IdPList
		idpsValue, err := app.state.Get(idpListKeyBytes, true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if idpsValue != nil {
			var idpsList data.IdPList
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			returnNodes.Node = make([]MsqDestinationNode, 0, len(idpsList.NodeId))
			for _, idp := range idpsList.NodeId {
				if msqDesNode := getMsqDestinationNode(idp); msqDesNode != nil {
					returnNodes.Node = append(returnNodes.Node, *msqDesNode)
				}
			}
		}
	} else {
		// fetch idp nodes from reference group
		refGroupCode := ""
		if funcParam.ReferenceGroupCode != "" {
			refGroupCode = funcParam.ReferenceGroupCode
		} else {
			identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
			refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), true)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			if refGroupCodeFromDB == nil {
				return app.ReturnQuery(nil, "not found", app.state.Height)
			}
			refGroupCode = string(refGroupCodeFromDB)
		}
		refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
		refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if refGroupValue == nil {
			return app.ReturnQuery(nil, "not found", app.state.Height)
		}
		var refGroup data.ReferenceGroup
		err = proto.Unmarshal(refGroupValue, &refGroup)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		returnNodes.Node = make([]MsqDestinationNode, 0, len(refGroup.Idps))
		for _, idp := range refGroup.Idps {
			// check IdP has Association with Identity
			if !idp.Active {
				continue
			}
			// check Ial > min ial
			if idp.Ial < funcParam.MinIal {
				continue
			}
			// Filter by node_id_list
			if len(funcParam.NodeIDList) > 0 {
				if !contains(idp.NodeId, funcParam.NodeIDList) {
					continue
				}
			}
			// Filter by mode_list
			if len(funcParam.ModeList) > 0 {
				supportedModeCount := 0
				for _, mode := range idp.Mode {
					if containsInt32(mode, funcParam.ModeList) {
						supportedModeCount++
					}
				}
				if supportedModeCount < len(funcParam.ModeList) {
					continue
				}
			}
			if msqDesNode := getMsqDestinationNode(idp.NodeId); msqDesNode != nil {
				msqDesNode.Ial = &idp.Ial
				if idp.Lial != nil {
					msqDesNode.Lial = &idp.Lial.Value
				}
				if idp.Laal != nil {
					msqDesNode.Laal = &idp.Laal.Value
				}
				msqDesNode.ModeList = &idp.Mode
				returnNodes.Node = append(returnNodes.Node, *msqDesNode)
			}
		}
	}

	value, err := json.Marshal(returnNodes)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(returnNodes.Node) == 0 {
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	return app.ReturnQuery(value, "success", app.state.Height)
}

type GetAsNodesByServiceIdParam struct {
	ServiceID  string   `json:"service_id"`
	NodeIDList []string `json:"node_id_list"`
}

type GetAsNodesByServiceIdResult struct {
	Node []ASNode `json:"node"`
}

type ASNode struct {
	ID        string  `json:"node_id"`
	Name      string  `json:"node_name"`
	MinIal    float64 `json:"min_ial"`
	MinAal    float64 `json:"min_aal"`
	ServiceID string  `json:"service_id"`
	Active    bool    `json:"active"`
}

type GetAsNodesByServiceIdWithNameResult struct {
	Node []ASNodeResult `json:"node"`
}

type ASNodeResult struct {
	ID                     string   `json:"node_id"`
	Name                   string   `json:"node_name"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

func (app *ABCIApplication) getAsNodesByServiceId(param []byte) types.ResponseQuery {
	app.logger.Infof("GetAsNodesByServiceId, Parameter: %s", param)
	var funcParam GetAsNodesByServiceIdParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	if value == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}

	// filter serive is active
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if serviceValue == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if service.Active == false {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "service is not active", app.state.Height)
	}

	var storedData data.ServiceDesList
	err = proto.Unmarshal([]byte(value), &storedData)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var result GetAsNodesByServiceIdWithNameResult
	result.Node = make([]ASNodeResult, 0)
	for index := range storedData.Node {

		// filter service destination is Active
		if !storedData.Node[index].Active {
			continue
		}

		// Filter approve from NDID
		approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + storedData.Node[index].NodeId
		approveServiceJSON, err := app.state.Get([]byte(approveServiceKey), true)
		if err != nil {
			continue
		}
		if approveServiceJSON == nil {
			continue
		}
		var approveService data.ApproveService
		err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
		if err != nil {
			continue
		}
		if !approveService.Active {
			continue
		}

		nodeDetailKey := nodeIDKeyPrefix + keySeparator + storedData.Node[index].NodeId
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
		if err != nil {
			continue
		}
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			continue
		}

		// filter node is active
		if !nodeDetail.Active {
			continue
		}
		var newRow = ASNodeResult{
			storedData.Node[index].NodeId,
			nodeDetail.NodeName,
			storedData.Node[index].MinIal,
			storedData.Node[index].MinAal,
			storedData.Node[index].SupportedNamespaceList,
		}
		result.Node = append(result.Node, newRow)
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.Node) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

type GetMqAddressesParam struct {
	NodeID string `json:"node_id"`
}

type GetMqAddressesResult []MsqAddress

func (app *ABCIApplication) getMqAddresses(param []byte) types.ResponseQuery {
	app.logger.Infof("GetMqAddresses, Parameter: %s", param)
	var funcParam GetMqAddressesParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	value, err := app.state.Get([]byte(nodeDetailKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		value = []byte("[]")
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var result GetMqAddressesResult
	for _, msq := range nodeDetail.Mq {
		var newRow MsqAddress
		newRow.IP = msq.Ip
		newRow.Port = msq.Port
		result = append(result, newRow)
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

type UpdateNodeParam struct {
	SigningPublicKey                       string   `json:"signing_public_key"`
	SigningAlgorithm                       string   `json:"signing_algorithm"`
	SigningMasterPublicKey                 string   `json:"signing_master_public_key"`
	SigningMasterAlgorithm                 string   `json:"signing_master_algorithm"`
	EncryptionPublicKey                    string   `json:"encryption_public_key"`
	EncryptionAlgorithm                    string   `json:"encryption_algorithm"`
	SupportedRequestMessageDataUrlTypeList []string `json:"supported_request_message_data_url_type_list"`
}

func (app *ABCIApplication) validateUpdateNode(funcParam UpdateNodeParam, callerNodeID string, committedState bool) error {
	// Validate master public key format
	if funcParam.SigningMasterPublicKey != "" {
		err := checkPubKeyForSigning(
			funcParam.SigningMasterPublicKey,
			appTypes.SignatureAlgorithm(funcParam.SigningMasterAlgorithm),
		)
		if err != nil {
			return err
		}
	}

	// Validate public key format
	if funcParam.SigningPublicKey != "" {
		err := checkPubKeyForSigning(
			funcParam.SigningPublicKey,
			appTypes.SignatureAlgorithm(funcParam.SigningAlgorithm),
		)
		if err != nil {
			return err
		}
	}

	// Validate encryption public key format
	if funcParam.EncryptionPublicKey != "" {
		err := checkPubKeyForEncryption(funcParam.EncryptionPublicKey)
		if err != nil {
			return err
		}
	}

	key := nodeIDKeyPrefix + keySeparator + callerNodeID
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}

	return nil
}

func (app *ABCIApplication) updateNodeCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam UpdateNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateUpdateNode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) updateNode(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNode, Parameter: %s", param)
	var funcParam UpdateNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateUpdateNode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	key := nodeIDKeyPrefix + keySeparator + callerNodeID
	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(value), &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// update MasterPublicKey
	if funcParam.SigningMasterPublicKey != "" {
		// set current key version inactive
		nodeKeyKey :=
			nodeKeyKeyPrefix + keySeparator +
				"signing_master" + keySeparator +
				callerNodeID + keySeparator +
				strconv.FormatInt(nodeDetail.SigningMasterPublicKey.Version, 10)
		nodeDetail.SigningMasterPublicKey.Active = false
		nodeKeyValue, err := utils.ProtoDeterministicMarshal(nodeDetail.SigningMasterPublicKey)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

		// new key
		nodeDetail.SigningMasterPublicKey = &data.NodeKey{
			PublicKey:           funcParam.SigningMasterPublicKey,
			Algorithm:           funcParam.SigningMasterAlgorithm,
			Version:             nodeDetail.SigningMasterPublicKey.Version + 1,
			CreationBlockHeight: app.state.CurrentBlockHeight,
			CreationChainId:     app.CurrentChain,
			Active:              true,
		}
		// create key version history
		nodeKeyKey =
			nodeKeyKeyPrefix + keySeparator +
				"signing_master" + keySeparator +
				callerNodeID + keySeparator +
				strconv.FormatInt(nodeDetail.SigningMasterPublicKey.Version, 10)
		nodeKeyValue, err = utils.ProtoDeterministicMarshal(nodeDetail.SigningMasterPublicKey)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))
	}
	// update PublicKey
	if funcParam.SigningPublicKey != "" {
		// set current key version inactive
		nodeKeyKey :=
			nodeKeyKeyPrefix + keySeparator +
				"signing" + keySeparator +
				callerNodeID + keySeparator +
				strconv.FormatInt(nodeDetail.SigningPublicKey.Version, 10)
		nodeDetail.SigningPublicKey.Active = false
		nodeKeyValue, err := utils.ProtoDeterministicMarshal(nodeDetail.SigningPublicKey)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

		// new key
		nodeDetail.SigningPublicKey = &data.NodeKey{
			PublicKey:           funcParam.SigningPublicKey,
			Algorithm:           funcParam.SigningAlgorithm,
			Version:             nodeDetail.SigningPublicKey.Version + 1,
			CreationBlockHeight: app.state.CurrentBlockHeight,
			CreationChainId:     app.CurrentChain,
			Active:              true,
		}
		// create key version history
		nodeKeyKey =
			nodeKeyKeyPrefix + keySeparator +
				"signing" + keySeparator +
				callerNodeID + keySeparator +
				strconv.FormatInt(nodeDetail.SigningPublicKey.Version, 10)
		nodeKeyValue, err = utils.ProtoDeterministicMarshal(nodeDetail.SigningPublicKey)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))
	}
	// update EncryptionPublicKey
	if funcParam.EncryptionPublicKey != "" {
		// set current key version inactive
		nodeKeyKey :=
			nodeKeyKeyPrefix + keySeparator +
				"encryption" + keySeparator +
				callerNodeID + keySeparator +
				strconv.FormatInt(nodeDetail.EncryptionPublicKey.Version, 10)
		nodeDetail.EncryptionPublicKey.Active = false
		nodeKeyValue, err := utils.ProtoDeterministicMarshal(nodeDetail.EncryptionPublicKey)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

		// new key
		nodeDetail.EncryptionPublicKey = &data.NodeKey{
			PublicKey:           funcParam.EncryptionPublicKey,
			Algorithm:           funcParam.EncryptionAlgorithm,
			Version:             nodeDetail.EncryptionPublicKey.Version + 1,
			CreationBlockHeight: app.state.CurrentBlockHeight,
			CreationChainId:     app.CurrentChain,
			Active:              true,
		}
		// create key version history
		nodeKeyKey =
			nodeKeyKeyPrefix + keySeparator +
				"encryption" + keySeparator +
				callerNodeID + keySeparator +
				strconv.FormatInt(nodeDetail.EncryptionPublicKey.Version, 10)
		nodeKeyValue, err = utils.ProtoDeterministicMarshal(nodeDetail.EncryptionPublicKey)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))
	}
	// update SupportedRequestMessageDataUrlTypeList and Role of node ID is IdP
	if funcParam.SupportedRequestMessageDataUrlTypeList != nil && appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleIdp {
		nodeDetail.SupportedRequestMessageDataUrlTypeList = funcParam.SupportedRequestMessageDataUrlTypeList
	}
	nodeDetailValue, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), []byte(nodeDetailValue))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type GetNodeInfoParam struct {
	NodeID string `json:"node_id"`
}

type GetNodeInfoResult struct {
	SigningPublicKey       NodeKey `json:"signing_public_key"`
	SigningMasterPublicKey NodeKey `json:"signing_master_public_key"`
	EncryptionPublicKey    NodeKey `json:"encryption_public_key"`
	NodeName               string  `json:"node_name"`
	Role                   string  `json:"role"`
	// for IdP
	MaxIal                                 *float64  `json:"max_ial,omitempty"`
	MaxAal                                 *float64  `json:"max_aal,omitempty"`
	SupportedFeatureList                   []string  `json:"supported_feature_list"`
	SupportedRequestMessageDataUrlTypeList *[]string `json:"supported_request_message_data_url_type_list,omitempty"`
	IsIdpAgent                             *bool     `json:"agent,omitempty"`
	// for IdP and RP
	UseWhitelist *bool     `json:"node_id_whitelist_active,omitempty"`
	Whitelist    *[]string `json:"node_id_whitelist,omitempty"`
	// for node behind proxy
	Proxy *ProxyNodeInfo `json:"proxy,omitempty"`
	// for all
	Mq     []MsqAddress `json:"mq"`
	Active bool         `json:"active"`
}

type NodeKey struct {
	PublicKey           string `json:"public_key"`
	Algorithm           string `json:"algorithm"`
	Version             int64  `json:"version"`
	CreationBlockHeight int64  `json:"creation_block_height"`
	CreationChainID     string `json:"creation_chain_id"`
	Active              bool   `json:"active"`
}

type ProxyNodeInfo struct {
	NodeID                 string       `json:"node_id"`
	NodeName               string       `json:"node_name"`
	SigningPublicKey       NodeKey      `json:"signing_public_key"`
	SigningMasterPublicKey NodeKey      `json:"signing_master_public_key"`
	EncryptionPublicKey    NodeKey      `json:"encryption_public_key"`
	Mq                     []MsqAddress `json:"mq"`
	Config                 string       `json:"config"`
}

func (app *ABCIApplication) getNodeInfo(param []byte) types.ResponseQuery {
	app.logger.Infof("GetNodeInfo, Parameter: %s", param)
	var funcParam GetNodeInfoParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if nodeDetailValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result := GetNodeInfoResult{
		SigningPublicKey: NodeKey{
			PublicKey:           nodeDetail.SigningPublicKey.PublicKey,
			Algorithm:           nodeDetail.SigningPublicKey.Algorithm,
			Version:             nodeDetail.SigningPublicKey.Version,
			CreationBlockHeight: nodeDetail.SigningPublicKey.CreationBlockHeight,
			CreationChainID:     nodeDetail.SigningPublicKey.CreationChainId,
			Active:              nodeDetail.SigningPublicKey.Active,
		},
		SigningMasterPublicKey: NodeKey{
			PublicKey:           nodeDetail.SigningMasterPublicKey.PublicKey,
			Algorithm:           nodeDetail.SigningMasterPublicKey.Algorithm,
			Version:             nodeDetail.SigningMasterPublicKey.Version,
			CreationBlockHeight: nodeDetail.SigningMasterPublicKey.CreationBlockHeight,
			CreationChainID:     nodeDetail.SigningMasterPublicKey.CreationChainId,
			Active:              nodeDetail.SigningMasterPublicKey.Active,
		},
		EncryptionPublicKey: NodeKey{
			PublicKey:           nodeDetail.EncryptionPublicKey.PublicKey,
			Algorithm:           nodeDetail.EncryptionPublicKey.Algorithm,
			Version:             nodeDetail.EncryptionPublicKey.Version,
			CreationBlockHeight: nodeDetail.EncryptionPublicKey.CreationBlockHeight,
			CreationChainID:     nodeDetail.EncryptionPublicKey.CreationChainId,
			Active:              nodeDetail.EncryptionPublicKey.Active,
		},
		NodeName: nodeDetail.NodeName,
		Role:     nodeDetail.Role,
		Active:   nodeDetail.Active,
	}
	for _, mq := range nodeDetail.Mq {
		result.Mq = append(result.Mq, MsqAddress{
			IP:   mq.Ip,
			Port: mq.Port,
		})
	}

	// If node behind proxy
	if nodeDetail.ProxyNodeId != "" {
		proxyNodeID := nodeDetail.ProxyNodeId
		// Get proxy node detail
		proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
		proxyNodeDetailValue, err := app.state.Get([]byte(proxyNodeDetailKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if proxyNodeDetailValue == nil {
			return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
		}
		var proxyNode data.NodeDetail
		err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		proxy := ProxyNodeInfo{
			NodeID:   string(proxyNodeID),
			NodeName: proxyNode.NodeName,
			SigningPublicKey: NodeKey{
				PublicKey:           proxyNode.SigningPublicKey.PublicKey,
				Algorithm:           proxyNode.SigningPublicKey.Algorithm,
				Version:             proxyNode.SigningPublicKey.Version,
				CreationBlockHeight: proxyNode.SigningPublicKey.CreationBlockHeight,
				CreationChainID:     proxyNode.SigningPublicKey.CreationChainId,
				Active:              proxyNode.SigningPublicKey.Active,
			},
			SigningMasterPublicKey: NodeKey{
				PublicKey:           proxyNode.SigningMasterPublicKey.PublicKey,
				Algorithm:           proxyNode.SigningMasterPublicKey.Algorithm,
				Version:             proxyNode.SigningMasterPublicKey.Version,
				CreationBlockHeight: proxyNode.SigningMasterPublicKey.CreationBlockHeight,
				CreationChainID:     proxyNode.SigningMasterPublicKey.CreationChainId,
				Active:              proxyNode.SigningMasterPublicKey.Active,
			},
			EncryptionPublicKey: NodeKey{
				PublicKey:           proxyNode.EncryptionPublicKey.PublicKey,
				Algorithm:           proxyNode.EncryptionPublicKey.Algorithm,
				Version:             proxyNode.EncryptionPublicKey.Version,
				CreationBlockHeight: proxyNode.EncryptionPublicKey.CreationBlockHeight,
				CreationChainID:     proxyNode.EncryptionPublicKey.CreationChainId,
				Active:              proxyNode.EncryptionPublicKey.Active,
			},
			Config: nodeDetail.ProxyConfig,
		}
		for _, mq := range proxyNode.Mq {
			proxy.Mq = append(proxy.Mq, MsqAddress{
				IP:   mq.Ip,
				Port: mq.Port,
			})
		}

		result.Proxy = &proxy
	}

	supportedFeatureList := append(make([]string, 0), nodeDetail.SupportedFeatureList...)
	result.SupportedFeatureList = supportedFeatureList

	if appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleIdp {
		result.MaxIal = &nodeDetail.MaxIal
		result.MaxAal = &nodeDetail.MaxAal
		supportedRequestMessageDataUrlTypeList := append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
		result.SupportedRequestMessageDataUrlTypeList = &supportedRequestMessageDataUrlTypeList
		result.IsIdpAgent = &nodeDetail.IsIdpAgent
	}

	if appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleIdp ||
		appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleRp {
		result.UseWhitelist = &nodeDetail.UseWhitelist
		if nodeDetail.UseWhitelist {
			result.Whitelist = &nodeDetail.Whitelist
		}
	}

	value, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(value, "success", app.state.Height)
}

type GetIdpNodesInfoResult struct {
	Node []IdpNode `json:"node"`
}

type IdpNode struct {
	NodeID                                 string        `json:"node_id"`
	Name                                   string        `json:"name"`
	MaxIal                                 float64       `json:"max_ial"`
	MaxAal                                 float64       `json:"max_aal"`
	SupportedFeatureList                   []string      `json:"supported_feature_list"`
	SigningPublicKey                       NodeKey       `json:"signing_public_key"`
	EncryptionPublicKey                    NodeKey       `json:"encryption_public_key"`
	Mq                                     []MsqAddress  `json:"mq"`
	IsIdpAgent                             bool          `json:"agent"`
	UseWhitelist                           *bool         `json:"node_id_whitelist_active,omitempty"`
	Whitelist                              *[]string     `json:"node_id_whitelist,omitempty"`
	Ial                                    *float64      `json:"ial,omitempty"`
	ModeList                               *[]int32      `json:"mode_list,omitempty"`
	SupportedRequestMessageDataUrlTypeList []string      `json:"supported_request_message_data_url_type_list"`
	Proxy                                  *IdpNodeProxy `json:"proxy,omitempty"`
}

type IdpNodeProxy struct {
	NodeID              string       `json:"node_id"`
	SigningPublicKey    NodeKey      `json:"signing_public_key"`
	EncryptionPublicKey NodeKey      `json:"encryption_public_key"`
	Mq                  []MsqAddress `json:"mq"`
	Config              string       `json:"config"`
}

func (app *ABCIApplication) getIdpNodesInfo(param []byte) types.ResponseQuery {
	app.logger.Infof("GetIdpNodesInfo, Parameter: %s", param)
	var funcParam GetIdpNodesParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	// fetch Filter RP node detail
	var nodeToFilterForDetail *data.NodeDetail
	if funcParam.FilterForNodeID != nil {
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + *funcParam.FilterForNodeID
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if nodeDetailValue == nil {
			return app.ReturnQuery(nil, "Filter RP does not exists", app.state.Height)
		}
		nodeToFilterForDetail = &data.NodeDetail{}
		if err := proto.Unmarshal(nodeDetailValue, nodeToFilterForDetail); err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
	}

	// return IdpNode if nodeID valid and within funcParam
	// return nil otherwise
	getIdpNode := func(nodeID string) *IdpNode {
		// check if Idp in Filter RP whitelist
		if nodeToFilterForDetail != nil && nodeToFilterForDetail.UseWhitelist &&
			!contains(nodeID, nodeToFilterForDetail.Whitelist) {
			return nil
		}

		nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
		if err != nil {
			return nil
		}
		if nodeDetailValue == nil {
			return nil
		}
		var nodeDetail data.NodeDetail
		err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			return nil
		}
		// check node is active
		if !nodeDetail.Active {
			return nil
		}
		// check Max IAL && AAL
		if !(nodeDetail.MaxIal >= funcParam.MinIal &&
			nodeDetail.MaxAal >= funcParam.MinAal) {
			return nil
		}
		// Filter by node_id_list
		if len(funcParam.NodeIDList) > 0 && !contains(nodeID, funcParam.NodeIDList) {
			return nil
		}
		// Filter by SupportedFeatureList
		if funcParam.SupportedFeatureList != nil {
			for _, supportedFeature := range funcParam.SupportedFeatureList {
				if !contains(supportedFeature, nodeDetail.SupportedFeatureList) {
					return nil
				}
			}
		}
		// Filter by IsIdpAgent
		if funcParam.IsIdpAgent != nil && *funcParam.IsIdpAgent != nodeDetail.IsIdpAgent {
			return nil
		}
		// Filter by supported_request_message_data_url_type_list
		if len(funcParam.SupportedRequestMessageDataUrlTypeList) > 0 {
			// foundSupported := false
			supportedCount := 0
			for _, supportedType := range nodeDetail.SupportedRequestMessageDataUrlTypeList {
				if contains(supportedType, funcParam.SupportedRequestMessageDataUrlTypeList) {
					supportedCount++
				}
			}
			if supportedCount < len(funcParam.SupportedRequestMessageDataUrlTypeList) {
				return nil
			}
		}
		// Check if Filter RP is in Idp whitelist
		if funcParam.FilterForNodeID != nil && nodeDetail.UseWhitelist &&
			!contains(*funcParam.FilterForNodeID, nodeDetail.Whitelist) {
			return nil
		}

		var proxy *IdpNodeProxy
		if nodeDetail.ProxyNodeId != "" {
			proxyNodeID := nodeDetail.ProxyNodeId
			// Get proxy node detail
			proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
			proxyNodeDetailValue, err := app.state.Get([]byte(proxyNodeDetailKey), true)
			if err != nil {
				return nil
			}
			if proxyNodeDetailValue == nil {
				return nil
			}
			var proxyNode data.NodeDetail
			err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
			if err != nil {
				return nil
			}
			// Check proxy node is active
			if !proxyNode.Active {
				return nil
			}
			proxy = &IdpNodeProxy{
				NodeID: string(proxyNodeID),
				SigningPublicKey: NodeKey{
					PublicKey:           proxyNode.SigningPublicKey.PublicKey,
					Algorithm:           proxyNode.SigningPublicKey.Algorithm,
					Version:             proxyNode.SigningPublicKey.Version,
					CreationBlockHeight: proxyNode.SigningPublicKey.CreationBlockHeight,
					CreationChainID:     proxyNode.SigningPublicKey.CreationChainId,
					Active:              proxyNode.SigningPublicKey.Active,
				},
				EncryptionPublicKey: NodeKey{
					PublicKey:           proxyNode.EncryptionPublicKey.PublicKey,
					Algorithm:           proxyNode.EncryptionPublicKey.Algorithm,
					Version:             proxyNode.EncryptionPublicKey.Version,
					CreationBlockHeight: proxyNode.EncryptionPublicKey.CreationBlockHeight,
					CreationChainID:     proxyNode.EncryptionPublicKey.CreationChainId,
					Active:              proxyNode.EncryptionPublicKey.Active,
				},
				Config: nodeDetail.ProxyConfig,
			}
			for _, mq := range proxyNode.Mq {
				proxy.Mq = append(proxy.Mq, MsqAddress{
					IP:   mq.Ip,
					Port: mq.Port,
				})
			}
		}

		var whitelist *[]string
		if nodeDetail.UseWhitelist {
			whitelist = &nodeDetail.Whitelist
		}

		idpNode := &IdpNode{
			NodeID: nodeID,
			Name:   nodeDetail.NodeName,
			MaxIal: nodeDetail.MaxIal,
			MaxAal: nodeDetail.MaxAal,
			SigningPublicKey: NodeKey{
				PublicKey:           nodeDetail.SigningPublicKey.PublicKey,
				Algorithm:           nodeDetail.SigningPublicKey.Algorithm,
				Version:             nodeDetail.SigningPublicKey.Version,
				CreationBlockHeight: nodeDetail.SigningPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.SigningPublicKey.CreationChainId,
				Active:              nodeDetail.SigningPublicKey.Active,
			},
			EncryptionPublicKey: NodeKey{
				PublicKey:           nodeDetail.EncryptionPublicKey.PublicKey,
				Algorithm:           nodeDetail.EncryptionPublicKey.Algorithm,
				Version:             nodeDetail.EncryptionPublicKey.Version,
				CreationBlockHeight: nodeDetail.EncryptionPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.EncryptionPublicKey.CreationChainId,
				Active:              nodeDetail.EncryptionPublicKey.Active,
			},
			SupportedFeatureList:                   append(make([]string, 0), nodeDetail.SupportedFeatureList...),
			IsIdpAgent:                             nodeDetail.IsIdpAgent,
			UseWhitelist:                           &nodeDetail.UseWhitelist,
			Whitelist:                              whitelist,
			SupportedRequestMessageDataUrlTypeList: append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...),
			Proxy:                                  proxy,
		}

		for _, mq := range nodeDetail.Mq {
			idpNode.Mq = append(idpNode.Mq, MsqAddress{
				IP:   mq.Ip,
				Port: mq.Port,
			})
		}

		return idpNode
	}

	var returnNodes GetIdpNodesInfoResult

	if funcParam.ReferenceGroupCode == "" && funcParam.IdentityNamespace == "" && funcParam.IdentityIdentifierHash == "" {
		var idpsList data.IdPList
		idpsValue, err := app.state.Get(idpListKeyBytes, true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if idpsValue != nil {
			if err := proto.Unmarshal(idpsValue, &idpsList); err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			returnNodes.Node = make([]IdpNode, 0, len(idpsList.NodeId))
			for _, idp := range idpsList.NodeId {
				if idpNode := getIdpNode(idp); idpNode != nil {
					returnNodes.Node = append(returnNodes.Node, *idpNode)
				}
			}
		}
	} else {
		refGroupCode := ""
		if funcParam.ReferenceGroupCode != "" {
			refGroupCode = funcParam.ReferenceGroupCode
		} else {
			identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
			refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), true)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			if refGroupCodeFromDB == nil {
				return app.ReturnQuery(nil, "not found", app.state.Height)
			}
			refGroupCode = string(refGroupCodeFromDB)
		}
		refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
		refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if refGroupValue == nil {
			return app.ReturnQuery(nil, "not found", app.state.Height)
		}
		var refGroup data.ReferenceGroup
		err = proto.Unmarshal(refGroupValue, &refGroup)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		for _, idp := range refGroup.Idps {
			// check IdP has Association with Identity
			if !idp.Active {
				continue
			}
			// check Ial > min ial
			if idp.Ial < funcParam.MinIal {
				continue
			}
			// Filter by node_id_list
			if len(funcParam.NodeIDList) > 0 {
				if !contains(idp.NodeId, funcParam.NodeIDList) {
					continue
				}
			}
			// Filter by mode_list
			if len(funcParam.ModeList) > 0 {
				supportedModeCount := 0
				for _, mode := range idp.Mode {
					if containsInt32(mode, funcParam.ModeList) {
						supportedModeCount++
					}
				}
				if supportedModeCount < len(funcParam.ModeList) {
					continue
				}
			}
			if idpNode := getIdpNode(idp.NodeId); idpNode != nil {
				idpNode.Ial = &idp.Ial
				idpNode.ModeList = &idp.Mode
				returnNodes.Node = append(returnNodes.Node, *idpNode)
			}
		}
	}
	value, err := json.Marshal(returnNodes)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(returnNodes.Node) == 0 {
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	return app.ReturnQuery(value, "success", app.state.Height)
}

type GetAsNodesInfoByServiceIdResult struct {
	Node []interface{} `json:"node"`
}

type ASWithMqNode struct {
	ID                     string       `json:"node_id"`
	Name                   string       `json:"name"`
	MinIal                 float64      `json:"min_ial"`
	MinAal                 float64      `json:"min_aal"`
	SigningPublicKey       NodeKey      `json:"signing_public_key"`
	EncryptionPublicKey    NodeKey      `json:"encryption_public_key"`
	Mq                     []MsqAddress `json:"mq"`
	SupportedNamespaceList []string     `json:"supported_namespace_list"`
}

type ASWithMqNodeBehindProxy struct {
	NodeID                 string   `json:"node_id"`
	Name                   string   `json:"name"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	SigningPublicKey       NodeKey  `json:"signing_public_key"`
	EncryptionPublicKey    NodeKey  `json:"encryption_public_key"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
	Proxy                  struct {
		NodeID              string       `json:"node_id"`
		SigningPublicKey    NodeKey      `json:"signing_public_key"`
		EncryptionPublicKey NodeKey      `json:"encryption_public_key"`
		Mq                  []MsqAddress `json:"mq"`
		Config              string       `json:"config"`
	} `json:"proxy"`
}

func (app *ABCIApplication) getAsNodesInfoByServiceId(param []byte) types.ResponseQuery {
	app.logger.Infof("GetAsNodesInfoByServiceId, Parameter: %s", param)
	var funcParam GetAsNodesByServiceIdParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		var result GetAsNodesInfoByServiceIdResult
		result.Node = make([]interface{}, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	// filter serive is active
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if serviceValue == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if service.Active == false {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "service is not active", app.state.Height)
	}
	var storedData data.ServiceDesList
	err = proto.Unmarshal([]byte(value), &storedData)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	// Make mapping
	mapNodeIDList := map[string]bool{}
	for _, nodeID := range funcParam.NodeIDList {
		mapNodeIDList[nodeID] = true
	}
	var result GetAsNodesInfoByServiceIdResult
	result.Node = make([]interface{}, 0)
	for index := range storedData.Node {
		// filter from node_id_list
		if len(mapNodeIDList) > 0 {
			if mapNodeIDList[storedData.Node[index].NodeId] == false {
				continue
			}
		}
		// filter service destination is Active
		if !storedData.Node[index].Active {
			continue
		}
		// Filter approve from NDID
		approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + storedData.Node[index].NodeId
		approveServiceJSON, err := app.state.Get([]byte(approveServiceKey), true)
		if err != nil {
			continue
		}
		if approveServiceJSON == nil {
			continue
		}
		var approveService data.ApproveService
		err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
		if err != nil {
			continue
		}
		if !approveService.Active {
			continue
		}
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + storedData.Node[index].NodeId
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
		if err != nil {
			continue
		}
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			continue
		}
		// filter node is active
		if !nodeDetail.Active {
			continue
		}
		// If node is behind proxy
		if nodeDetail.ProxyNodeId != "" {
			proxyNodeID := nodeDetail.ProxyNodeId
			// Get proxy node detail
			proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
			proxyNodeDetailValue, err := app.state.Get([]byte(proxyNodeDetailKey), true)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			if proxyNodeDetailValue == nil {
				return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
			}
			var proxyNode data.NodeDetail
			err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			// Check proxy node is active
			if !proxyNode.Active {
				continue
			}
			var as ASWithMqNodeBehindProxy
			as.NodeID = storedData.Node[index].NodeId
			as.Name = nodeDetail.NodeName
			as.MinIal = storedData.Node[index].MinIal
			as.MinAal = storedData.Node[index].MinAal
			as.SigningPublicKey = NodeKey{
				PublicKey:           nodeDetail.SigningPublicKey.PublicKey,
				Algorithm:           nodeDetail.SigningPublicKey.Algorithm,
				Version:             nodeDetail.SigningPublicKey.Version,
				CreationBlockHeight: nodeDetail.SigningPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.SigningPublicKey.CreationChainId,
				Active:              nodeDetail.SigningPublicKey.Active,
			}
			as.EncryptionPublicKey = NodeKey{
				PublicKey:           nodeDetail.EncryptionPublicKey.PublicKey,
				Algorithm:           nodeDetail.EncryptionPublicKey.Algorithm,
				Version:             nodeDetail.EncryptionPublicKey.Version,
				CreationBlockHeight: nodeDetail.EncryptionPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.EncryptionPublicKey.CreationChainId,
				Active:              nodeDetail.EncryptionPublicKey.Active,
			}
			as.SupportedNamespaceList = storedData.Node[index].SupportedNamespaceList
			as.Proxy.NodeID = string(proxyNodeID)
			as.Proxy.SigningPublicKey = NodeKey{
				PublicKey:           proxyNode.SigningPublicKey.PublicKey,
				Algorithm:           proxyNode.SigningPublicKey.Algorithm,
				Version:             proxyNode.SigningPublicKey.Version,
				CreationBlockHeight: proxyNode.SigningPublicKey.CreationBlockHeight,
				CreationChainID:     proxyNode.SigningPublicKey.CreationChainId,
				Active:              proxyNode.SigningPublicKey.Active,
			}
			as.Proxy.EncryptionPublicKey = NodeKey{
				PublicKey:           proxyNode.EncryptionPublicKey.PublicKey,
				Algorithm:           proxyNode.EncryptionPublicKey.Algorithm,
				Version:             proxyNode.EncryptionPublicKey.Version,
				CreationBlockHeight: proxyNode.EncryptionPublicKey.CreationBlockHeight,
				CreationChainID:     proxyNode.EncryptionPublicKey.CreationChainId,
				Active:              proxyNode.EncryptionPublicKey.Active,
			}
			if proxyNode.Mq != nil {
				for _, mq := range proxyNode.Mq {
					var msq MsqAddress
					msq.IP = mq.Ip
					msq.Port = mq.Port
					as.Proxy.Mq = append(as.Proxy.Mq, msq)
				}
			}
			as.Proxy.Config = nodeDetail.ProxyConfig
			result.Node = append(result.Node, as)
		} else {
			var msqAddress []MsqAddress
			for _, mq := range nodeDetail.Mq {
				var msq MsqAddress
				msq.IP = mq.Ip
				msq.Port = mq.Port
				msqAddress = append(msqAddress, msq)
			}
			var newRow = ASWithMqNode{
				ID:     storedData.Node[index].NodeId,
				Name:   nodeDetail.NodeName,
				MinIal: storedData.Node[index].MinIal,
				MinAal: storedData.Node[index].MinAal,
				SigningPublicKey: NodeKey{
					PublicKey:           nodeDetail.SigningPublicKey.PublicKey,
					Algorithm:           nodeDetail.SigningPublicKey.Algorithm,
					Version:             nodeDetail.SigningPublicKey.Version,
					CreationBlockHeight: nodeDetail.SigningPublicKey.CreationBlockHeight,
					CreationChainID:     nodeDetail.SigningPublicKey.CreationChainId,
					Active:              nodeDetail.SigningPublicKey.Active,
				},
				EncryptionPublicKey: NodeKey{
					PublicKey:           nodeDetail.EncryptionPublicKey.PublicKey,
					Algorithm:           nodeDetail.EncryptionPublicKey.Algorithm,
					Version:             nodeDetail.EncryptionPublicKey.Version,
					CreationBlockHeight: nodeDetail.EncryptionPublicKey.CreationBlockHeight,
					CreationChainID:     nodeDetail.EncryptionPublicKey.CreationChainId,
					Active:              nodeDetail.EncryptionPublicKey.Active,
				},
				Mq:                     msqAddress,
				SupportedNamespaceList: storedData.Node[index].SupportedNamespaceList,
			}
			result.Node = append(result.Node, newRow)
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

type GetNodesBehindProxyNodeParam struct {
	ProxyNodeID string `json:"proxy_node_id"`
}

type GetNodesBehindProxyNodeResult struct {
	Nodes []interface{} `json:"nodes"`
}

type IdPBehindProxy struct {
	NodeID                                 string   `json:"node_id"`
	NodeName                               string   `json:"node_name"`
	Role                                   string   `json:"role"`
	SigningPublicKey                       NodeKey  `json:"signing_public_key"`
	SigningMasterPublicKey                 NodeKey  `json:"signing_master_public_key"`
	EncryptionPublicKey                    NodeKey  `json:"encryption_public_key"`
	MaxIal                                 float64  `json:"max_ial"`
	MaxAal                                 float64  `json:"max_aal"`
	SupportedFeatureList                   []string `json:"supported_feature_list"`
	IsIdpAgent                             bool     `json:"agent"`
	Config                                 string   `json:"config"`
	SupportedRequestMessageDataUrlTypeList []string `json:"supported_request_message_data_url_type_list"`
}

type ASorRPBehindProxy struct {
	NodeID                 string   `json:"node_id"`
	NodeName               string   `json:"node_name"`
	Role                   string   `json:"role"`
	SigningPublicKey       NodeKey  `json:"signing_public_key"`
	SigningMasterPublicKey NodeKey  `json:"signing_master_public_key"`
	EncryptionPublicKey    NodeKey  `json:"encryption_public_key"`
	SupportedFeatureList   []string `json:"supported_feature_list"`
	Config                 string   `json:"config"`
}

func (app *ABCIApplication) getNodesBehindProxyNode(param []byte) types.ResponseQuery {
	app.logger.Infof("GetNodesBehindProxyNode, Parameter: %s", param)
	var funcParam GetNodesBehindProxyNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetNodesBehindProxyNodeResult
	result.Nodes = make([]interface{}, 0)
	behindProxyNodeKey := "BehindProxyNode" + keySeparator + funcParam.ProxyNodeID
	behindProxyNodeValue, err := app.state.Get([]byte(behindProxyNodeKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if behindProxyNodeValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)
	err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for _, node := range nodes.Nodes {
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + node
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
		if err != nil {
			continue
		}
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			continue
		}

		// Check node has proxy node ID
		if nodeDetail.ProxyNodeId == "" {
			continue
		}

		if appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleIdp {
			var row IdPBehindProxy
			row.NodeID = node
			row.NodeName = nodeDetail.NodeName
			row.Role = nodeDetail.Role
			row.SigningPublicKey = NodeKey{
				PublicKey:           nodeDetail.SigningPublicKey.PublicKey,
				Algorithm:           nodeDetail.SigningPublicKey.Algorithm,
				Version:             nodeDetail.SigningPublicKey.Version,
				CreationBlockHeight: nodeDetail.SigningPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.SigningPublicKey.CreationChainId,
				Active:              nodeDetail.SigningPublicKey.Active,
			}
			row.SigningMasterPublicKey = NodeKey{
				PublicKey:           nodeDetail.SigningMasterPublicKey.PublicKey,
				Algorithm:           nodeDetail.SigningMasterPublicKey.Algorithm,
				Version:             nodeDetail.SigningMasterPublicKey.Version,
				CreationBlockHeight: nodeDetail.SigningMasterPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.SigningMasterPublicKey.CreationChainId,
				Active:              nodeDetail.SigningMasterPublicKey.Active,
			}
			row.EncryptionPublicKey = NodeKey{
				PublicKey:           nodeDetail.EncryptionPublicKey.PublicKey,
				Algorithm:           nodeDetail.EncryptionPublicKey.Algorithm,
				Version:             nodeDetail.EncryptionPublicKey.Version,
				CreationBlockHeight: nodeDetail.EncryptionPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.EncryptionPublicKey.CreationChainId,
				Active:              nodeDetail.EncryptionPublicKey.Active,
			}
			row.MaxIal = nodeDetail.MaxIal
			row.MaxAal = nodeDetail.MaxAal
			row.SupportedFeatureList = nodeDetail.SupportedFeatureList
			row.IsIdpAgent = nodeDetail.IsIdpAgent
			row.Config = nodeDetail.ProxyConfig
			row.SupportedRequestMessageDataUrlTypeList = nodeDetail.SupportedRequestMessageDataUrlTypeList
			result.Nodes = append(result.Nodes, row)
		} else {
			var row ASorRPBehindProxy
			row.NodeID = node
			row.NodeName = nodeDetail.NodeName
			row.Role = nodeDetail.Role
			row.SigningPublicKey = NodeKey{
				PublicKey:           nodeDetail.SigningPublicKey.PublicKey,
				Algorithm:           nodeDetail.SigningPublicKey.Algorithm,
				Version:             nodeDetail.SigningPublicKey.Version,
				CreationBlockHeight: nodeDetail.SigningPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.SigningPublicKey.CreationChainId,
				Active:              nodeDetail.SigningPublicKey.Active,
			}
			row.SigningMasterPublicKey = NodeKey{
				PublicKey:           nodeDetail.SigningMasterPublicKey.PublicKey,
				Algorithm:           nodeDetail.SigningMasterPublicKey.Algorithm,
				Version:             nodeDetail.SigningMasterPublicKey.Version,
				CreationBlockHeight: nodeDetail.SigningMasterPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.SigningMasterPublicKey.CreationChainId,
				Active:              nodeDetail.SigningMasterPublicKey.Active,
			}
			row.EncryptionPublicKey = NodeKey{
				PublicKey:           nodeDetail.EncryptionPublicKey.PublicKey,
				Algorithm:           nodeDetail.EncryptionPublicKey.Algorithm,
				Version:             nodeDetail.EncryptionPublicKey.Version,
				CreationBlockHeight: nodeDetail.EncryptionPublicKey.CreationBlockHeight,
				CreationChainID:     nodeDetail.EncryptionPublicKey.CreationChainId,
				Active:              nodeDetail.EncryptionPublicKey.Active,
			}
			row.SupportedFeatureList = nodeDetail.SupportedFeatureList
			row.Config = nodeDetail.ProxyConfig
			result.Nodes = append(result.Nodes, row)
		}

	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.Nodes) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

type GetNodeIDListParam struct {
	Role string `json:"role"`
}

type GetNodeIDListResult struct {
	NodeIDList []string `json:"node_id_list"`
}

func (app *ABCIApplication) getNodeIDList(param []byte) types.ResponseQuery {
	app.logger.Infof("GetNodeIDList, Parameter: %s", param)
	var funcParam GetNodeIDListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetNodeIDListResult
	result.NodeIDList = make([]string, 0)
	if strings.ToLower(funcParam.Role) == "rp" {
		var rpsList data.RPList
		rpsKey := "rpList"
		rpsValue, err := app.state.Get([]byte(rpsKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if rpsValue != nil {
			err := proto.Unmarshal(rpsValue, &rpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, nodeID := range rpsList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
				nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
				if err != nil {
					return app.ReturnQuery(nil, err.Error(), app.state.Height)
				}
				if nodeDetailValue != nil {
					var nodeDetail data.NodeDetail
					err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
					if err != nil {
						continue
					}
					if nodeDetail.Active {
						result.NodeIDList = append(result.NodeIDList, nodeID)
					}
				}
			}
		}
	} else if strings.ToLower(funcParam.Role) == "idp" {
		var idpsList data.IdPList
		idpsValue, err := app.state.Get(idpListKeyBytes, true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, nodeID := range idpsList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
				nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
				if err != nil {
					return app.ReturnQuery(nil, err.Error(), app.state.Height)
				}
				if nodeDetailValue != nil {
					var nodeDetail data.NodeDetail
					err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
					if err != nil {
						continue
					}
					if nodeDetail.Active {
						result.NodeIDList = append(result.NodeIDList, nodeID)
					}
				}
			}
		}
	} else if strings.ToLower(funcParam.Role) == "as" {
		var asList data.ASList
		asKey := "asList"
		asValue, err := app.state.Get([]byte(asKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if asValue != nil {
			err := proto.Unmarshal(asValue, &asList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, nodeID := range asList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
				nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
				if err != nil {
					return app.ReturnQuery(nil, err.Error(), app.state.Height)
				}
				if nodeDetailValue != nil {
					var nodeDetail data.NodeDetail
					err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
					if err != nil {
						continue
					}
					if nodeDetail.Active {
						result.NodeIDList = append(result.NodeIDList, nodeID)
					}
				}
			}
		}
	} else {
		var allList data.AllList
		allKey := "allList"
		allValue, err := app.state.Get([]byte(allKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if allValue != nil {
			err := proto.Unmarshal(allValue, &allList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, nodeID := range allList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
				nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
				if err != nil {
					return app.ReturnQuery(nil, err.Error(), app.state.Height)
				}
				if nodeDetailValue != nil {
					var nodeDetail data.NodeDetail
					err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
					if err != nil {
						continue
					}
					if nodeDetail.Active {
						result.NodeIDList = append(result.NodeIDList, nodeID)
					}
				}
			}
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.NodeIDList) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

type getNodePublicKeyListParam struct {
	NodeID string `json:"node_id"`
}

type getNodePublicKeyListResult struct {
	SigningPublicKeyList       []NodeKey `json:"signing_public_key_list"`
	SigningMasterPublicKeyList []NodeKey `json:"signing_master_public_key_list"`
	EncryptionPublicKeyList    []NodeKey `json:"encryption_public_key_list"`
}

func (app *ABCIApplication) getNodePublicKeyList(param []byte) types.ResponseQuery {
	app.logger.Infof("GetNodePublicKeyList, Parameter: %s", param)
	var funcParam getNodePublicKeyListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	// check node ID exists
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if nodeDetailValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}

	// signing public keys
	signingPublicKeyList := make([]NodeKey, 0)

	nodeKeyKeyIteratorPrefix :=
		nodeKeyKeyPrefix + keySeparator +
			"signing" + keySeparator +
			funcParam.NodeID + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(nodeKeyKeyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		value := iter.Value()

		var nodeKey data.NodeKey
		err = proto.Unmarshal(value, &nodeKey)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		retValNodeKey := NodeKey{
			PublicKey:           nodeKey.PublicKey,
			Algorithm:           nodeKey.Algorithm,
			Version:             nodeKey.Version,
			CreationBlockHeight: nodeKey.CreationBlockHeight,
			CreationChainID:     nodeKey.CreationChainId,
			Active:              nodeKey.Active,
		}

		signingPublicKeyList = append(signingPublicKeyList, retValNodeKey)
	}
	iter.Close()

	// signing master public keys
	signingMasterPublicKeyList := make([]NodeKey, 0)

	nodeKeyKeyIteratorPrefix =
		nodeKeyKeyPrefix + keySeparator +
			"signing_master" + keySeparator +
			funcParam.NodeID + keySeparator
	r = goleveldbutil.BytesPrefix([]byte(nodeKeyKeyIteratorPrefix))
	iter, err = app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		value := iter.Value()

		var nodeKey data.NodeKey
		err = proto.Unmarshal(value, &nodeKey)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		retValNodeKey := NodeKey{
			PublicKey:           nodeKey.PublicKey,
			Algorithm:           nodeKey.Algorithm,
			Version:             nodeKey.Version,
			CreationBlockHeight: nodeKey.CreationBlockHeight,
			CreationChainID:     nodeKey.CreationChainId,
			Active:              nodeKey.Active,
		}

		signingMasterPublicKeyList = append(signingMasterPublicKeyList, retValNodeKey)
	}
	iter.Close()

	// encryption public keys
	encryptionPublicKeyList := make([]NodeKey, 0)

	nodeKeyKeyIteratorPrefix =
		nodeKeyKeyPrefix + keySeparator +
			"encryption" + keySeparator +
			funcParam.NodeID + keySeparator
	r = goleveldbutil.BytesPrefix([]byte(nodeKeyKeyIteratorPrefix))
	iter, err = app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		value := iter.Value()

		var nodeKey data.NodeKey
		err = proto.Unmarshal(value, &nodeKey)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		retValNodeKey := NodeKey{
			PublicKey:           nodeKey.PublicKey,
			Algorithm:           nodeKey.Algorithm,
			Version:             nodeKey.Version,
			CreationBlockHeight: nodeKey.CreationBlockHeight,
			CreationChainID:     nodeKey.CreationChainId,
			Active:              nodeKey.Active,
		}

		encryptionPublicKeyList = append(encryptionPublicKeyList, retValNodeKey)
	}
	iter.Close()

	result := &getNodePublicKeyListResult{
		SigningPublicKeyList:       signingPublicKeyList,
		SigningMasterPublicKeyList: signingMasterPublicKeyList,
		EncryptionPublicKeyList:    encryptionPublicKeyList,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
