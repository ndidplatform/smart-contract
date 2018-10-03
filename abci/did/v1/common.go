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

func (app *DIDApplication) setMqAddresses(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetMqAddresses, Parameter: %s", param)
	var funcParam SetMqAddressesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
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
	nodeDetailByte, err := proto.Marshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailByte))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) getNodeMasterPublicKey(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodeMasterPublicKey, Parameter: %s", param)
	var funcParam GetNodeMasterPublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	key := "NodeID" + "|" + funcParam.NodeID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	var res GetNodeMasterPublicKeyResult
	if value == nil {
		valueJSON, err := json.Marshal(res)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(valueJSON, "not found", app.state.db.Version64())
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	res.MasterPublicKey = nodeDetail.MasterPublicKey
	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(valueJSON, "success", app.state.db.Version64())

}

func (app *DIDApplication) getNodePublicKey(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodePublicKey, Parameter: %s", param)
	var funcParam GetNodePublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	key := "NodeID" + "|" + funcParam.NodeID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	var res GetNodePublicKeyResult
	if value == nil {
		valueJSON, err := json.Marshal(res)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(valueJSON, "not found", app.state.db.Version64())
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	res.PublicKey = nodeDetail.PublicKey
	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(valueJSON, "success", app.state.db.Version64())
}

func (app *DIDApplication) getNodeNameByNodeID(nodeID string) string {
	key := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ""
		}
		return nodeDetail.NodeName
	}
	return ""
}

func (app *DIDApplication) getIdpNodes(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetIdpNodes, Parameter: %s", param)
	var funcParam GetIdpNodesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var returnNodes GetIdpNodesResult
	returnNodes.Node = make([]MsqDestinationNode, 0)

	if funcParam.HashID == "" {
		idpsKey := "IdPList"
		_, idpsValue := app.state.db.GetVersioned(prefixKey([]byte(idpsKey)), height)
		var idpsList data.IdPList
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			for _, idp := range idpsList.NodeId {
				nodeDetailKey := "NodeID" + "|" + idp
				_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
				if nodeDetailValue == nil {
					continue
				}
				var nodeDetail data.NodeDetail
				err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
				if err != nil {
					continue
				}
				// check node is active
				if !nodeDetail.Active {
					continue
				}
				// check Max IAL && AAL
				if !(nodeDetail.MaxIal >= funcParam.MinIal &&
					nodeDetail.MaxAal >= funcParam.MinAal) {
					continue
				}
				var msqDesNode = MsqDestinationNode{
					idp,
					nodeDetail.NodeName,
					nodeDetail.MaxIal,
					nodeDetail.MaxAal,
				}
				returnNodes.Node = append(returnNodes.Node, msqDesNode)
			}
		}
	} else {
		key := "MsqDestination" + "|" + funcParam.HashID
		_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

		if value != nil {
			var nodes data.MsqDesList
			err = proto.Unmarshal([]byte(value), &nodes)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}

			for _, node := range nodes.Nodes {
				// check msq destination is not active
				if !node.Active {
					continue
				}
				// check Ial > min ial
				if node.Ial < funcParam.MinIal {
					continue
				}
				// check msq destination is not timed out
				if node.TimeoutBlock != 0 && app.CurrentBlock > node.TimeoutBlock {
					continue
				}
				nodeDetailKey := "NodeID" + "|" + node.NodeId
				_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
				if nodeDetailValue == nil {
					continue
				}
				var nodeDetail data.NodeDetail
				err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
				if err != nil {
					continue
				}
				// check node is active
				if !nodeDetail.Active {
					continue
				}
				// check Max IAL && AAL
				if !(nodeDetail.MaxIal >= funcParam.MinIal &&
					nodeDetail.MaxAal >= funcParam.MinAal) {
					continue
				}
				var msqDesNode = MsqDestinationNode{
					node.NodeId,
					nodeDetail.NodeName,
					nodeDetail.MaxIal,
					nodeDetail.MaxAal,
				}
				returnNodes.Node = append(returnNodes.Node, msqDesNode)
			}
		}
	}

	value, err := json.Marshal(returnNodes)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	// return app.ReturnQuery(value, "success", app.state.db.Version64())
	if len(returnNodes.Node) > 0 {
		return app.ReturnQuery(value, "success", app.state.db.Version64())
	}
	return app.ReturnQuery(value, "not found", app.state.db.Version64())
}

func (app *DIDApplication) getAsNodesByServiceId(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetAsNodesByServiceId, Parameter: %s", param)
	var funcParam GetAsNodesByServiceIdParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	key := "ServiceDestination" + "|" + funcParam.ServiceID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(value, "not found", app.state.db.Version64())
	}

	// filter serive is active
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceValue := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceValue != nil {
		var service data.ServiceDetail
		err = proto.Unmarshal([]byte(serviceValue), &service)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		if service.Active == false {
			var result GetAsNodesByServiceIdResult
			result.Node = make([]ASNode, 0)
			value, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			return app.ReturnQuery(value, "service is not active", app.state.db.Version64())
		}
	} else {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(value, "not found", app.state.db.Version64())
	}

	var storedData data.ServiceDesList
	err = proto.Unmarshal([]byte(value), &storedData)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result GetAsNodesByServiceIdWithNameResult
	result.Node = make([]ASNodeResult, 0)
	for index := range storedData.Node {

		// filter service destination is Active
		if !storedData.Node[index].Active {
			continue
		}

		// Filter approve from NDID
		approveServiceKey := "ApproveKey" + "|" + funcParam.ServiceID + "|" + storedData.Node[index].NodeId
		_, approveServiceJSON := app.state.db.Get(prefixKey([]byte(approveServiceKey)))
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

		nodeDetailKey := "NodeID" + "|" + storedData.Node[index].NodeId
		_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
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
		}
		result.Node = append(result.Node, newRow)
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(resultJSON, "success", app.state.db.Version64())
}

func (app *DIDApplication) getMqAddresses(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetMqAddresses, Parameter: %s", param)
	var funcParam GetMqAddressesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	if value == nil {
		value = []byte("[]")
		return app.ReturnQuery(value, "not found", app.state.db.Version64())
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
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	if len(result) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.db.Version64())
	}
	return app.ReturnQuery(resultJSON, "success", app.state.db.Version64())
}

func (app *DIDApplication) getRequest(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetRequest, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	key := "Request" + "|" + funcParam.RequestID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		valueJSON := []byte("{}")
		return app.ReturnQuery(valueJSON, "not found", app.state.db.Version64())
	}
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var res GetRequestResult
	res.IsClosed = request.Closed
	res.IsTimedOut = request.TimedOut
	res.MessageHash = request.RequestMessageHash
	res.Mode = int(request.Mode)

	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(valueJSON, "success", app.state.db.Version64())
}

func (app *DIDApplication) getRequestDetail(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetRequestDetail, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	key := "Request" + "|" + funcParam.RequestID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		valueJSON := []byte("{}")
		return app.ReturnQuery(valueJSON, "not found", app.state.db.Version64())
	}

	var result GetRequestDetailResult
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		value = []byte("")
		return app.ReturnQuery(value, err.Error(), app.state.db.Version64())
	}

	result.RequestID = request.RequestId
	result.MinIdp = int(request.MinIdp)
	result.MinAal = float64(request.MinAal)
	result.MinIal = float64(request.MinIal)
	result.Timeout = int(request.RequestTimeout)
	result.IdPIDList = request.IdpIdList
	for _, dataRequest := range request.DataRequestList {
		var newRow DataRequest
		newRow.ServiceID = dataRequest.ServiceId
		newRow.As = dataRequest.AsIdList
		newRow.Count = int(dataRequest.MinAs)
		newRow.AnsweredAsIdList = dataRequest.AnsweredAsIdList
		newRow.ReceivedDataFromList = dataRequest.ReceivedDataFromList
		newRow.RequestParamsHash = dataRequest.RequestParamsHash
		if newRow.As == nil {
			newRow.As = make([]string, 0)
		}
		if newRow.AnsweredAsIdList == nil {
			newRow.AnsweredAsIdList = make([]string, 0)
		}
		if newRow.ReceivedDataFromList == nil {
			newRow.ReceivedDataFromList = make([]string, 0)
		}
		result.DataRequestList = append(result.DataRequestList, newRow)
	}
	result.MessageHash = request.RequestMessageHash
	for _, response := range request.ResponseList {
		var newRow Response
		newRow.Ial = float64(response.Ial)
		newRow.Aal = float64(response.Aal)
		newRow.Status = response.Status
		newRow.Signature = response.Signature
		newRow.IdentityProof = response.IdentityProof
		newRow.PrivateProofHash = response.PrivateProofHash
		newRow.IdpID = response.IdpId
		if response.ValidProof != "" {
			if response.ValidProof == "true" {
				tValue := true
				newRow.ValidProof = &tValue
			} else {
				fValue := false
				newRow.ValidProof = &fValue
			}

		}
		if response.ValidIal != "" {
			if response.ValidIal == "true" {
				tValue := true
				newRow.ValidIal = &tValue
			} else {
				fValue := false
				newRow.ValidIal = &fValue
			}

		}
		if response.ValidSignature != "" {
			if response.ValidSignature == "true" {
				tValue := true
				newRow.ValidSignature = &tValue
			} else {
				fValue := false
				newRow.ValidSignature = &fValue
			}

		}
		result.Responses = append(result.Responses, newRow)
	}
	result.IsClosed = request.Closed
	result.IsTimedOut = request.TimedOut
	result.Mode = int(request.Mode)

	// Set purpose
	result.Purpose = request.Purpose

	// make nil to array len 0
	if result.IdPIDList == nil {
		result.IdPIDList = make([]string, 0)
	}
	if result.DataRequestList == nil {
		result.DataRequestList = make([]DataRequest, 0)
	}

	// Set requester_node_id
	result.RequesterNodeID = request.Owner

	// Set creation_block_height
	result.CreationBlockHeight = request.CreationBlockHeight

	resultJSON, err := json.Marshal(result)
	if err != nil {
		value = []byte("")
		return app.ReturnQuery(value, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(resultJSON, "success", app.state.db.Version64())
}

func (app *DIDApplication) getNamespaceList(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetNamespaceList, Parameter: %s", param)
	key := "AllNamespace"
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	if value == nil {
		value = []byte("[]")
		return app.ReturnQuery(value, "not found", app.state.db.Version64())
	}

	result := make([]*data.Namespace, 0)
	// filter flag==true
	var namespaces data.NamespaceList
	err := proto.Unmarshal([]byte(value), &namespaces)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	for _, namespace := range namespaces.Namespaces {
		if namespace.Active {
			result = append(result, namespace)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
}

func (app *DIDApplication) getServiceDetail(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetServiceDetail, Parameter: %s", param)
	var funcParam GetServiceDetailParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	key := "Service" + "|" + funcParam.ServiceID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	if value == nil {
		value = []byte("{}")
		return app.ReturnQuery(value, "not found", app.state.db.Version64())
	}
	var service data.ServiceDetail
	err = proto.Unmarshal(value, &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	returnValue, err := json.Marshal(service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
}

func (app *DIDApplication) updateNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNode, Parameter: %s", param)
	var funcParam UpdateNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value != nil {
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// update MasterPublicKey
		if funcParam.MasterPublicKey != "" {
			nodeDetail.MasterPublicKey = funcParam.MasterPublicKey
		}

		// update PublicKey
		if funcParam.PublicKey != "" {
			nodeDetail.PublicKey = funcParam.PublicKey
		}

		nodeDetailValue, err := proto.Marshal(&nodeDetail)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(nodeDetailValue))
		return app.ReturnDeliverTxLog(code.OK, "success", "")
	}
	return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
}

func (app *DIDApplication) checkExistingIdentity(param string, height int64) types.ResponseQuery {
	app.logger.Infof("CheckExistingIdentity, Parameter: %s", param)
	var funcParam CheckExistingIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result CheckExistingIdentityResult
	result.Exist = false

	key := "MsqDestination" + "|" + funcParam.HashID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value != nil {
		var nodes data.MsqDesList
		err = proto.Unmarshal([]byte(value), &nodes)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}

		msqCount := 0
		for _, node := range nodes.Nodes {
			if node.TimeoutBlock == 0 || node.TimeoutBlock > app.CurrentBlock {
				msqCount++
			}
		}

		if msqCount > 0 {
			result.Exist = true
		}
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
}

func (app *DIDApplication) getAccessorGroupID(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetAccessorGroupID, Parameter: %s", param)
	var funcParam GetAccessorGroupIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result GetAccessorGroupIDResult
	result.AccessorGroupID = ""

	key := "Accessor" + "|" + funcParam.AccessorID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value != nil {
		var accessor data.Accessor
		err = proto.Unmarshal([]byte(value), &accessor)
		if err == nil {
			result.AccessorGroupID = accessor.AccessorGroupId
		}
	}

	returnValue, err := json.Marshal(result)

	// If value == nil set log = "not found"
	if value == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
	}

	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
}

func (app *DIDApplication) getAccessorKey(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetAccessorKey, Parameter: %s", param)
	var funcParam GetAccessorKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result GetAccessorKeyResult
	result.AccessorPublicKey = ""

	key := "Accessor" + "|" + funcParam.AccessorID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value != nil {
		var accessor data.Accessor
		err = proto.Unmarshal([]byte(value), &accessor)
		if err == nil {
			result.AccessorPublicKey = accessor.AccessorPublicKey
			result.Active = accessor.Active
		}
	}

	returnValue, err := json.Marshal(result)

	// If value == nil set log = "not found"
	if value == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
	}

	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
}

func (app *DIDApplication) getServiceList(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetServiceList, Parameter: %s", param)
	key := "AllService"
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	if value == nil {
		result := make([]ServiceDetail, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(value, "not found", app.state.db.Version64())
	}

	result := make([]*data.ServiceDetail, 0)
	// filter flag==true
	var services data.ServiceDetailList
	err := proto.Unmarshal([]byte(value), &services)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	for _, service := range services.Services {
		if service.Active {
			result = append(result, service)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
}

func (app *DIDApplication) getServiceNameByServiceID(serviceID string) string {
	key := "Service" + "|" + serviceID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	var result ServiceDetail
	if value != nil {
		err := json.Unmarshal([]byte(value), &result)
		if err != nil {
			return ""
		}
		return result.ServiceName
	}
	return ""
}

func (app *DIDApplication) checkExistingAccessorID(param string, height int64) types.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorID, Parameter: %s", param)
	var funcParam CheckExistingAccessorIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result CheckExistingResult
	result.Exist = false

	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	_, accessorValue := app.state.db.GetVersioned(prefixKey([]byte(accessorKey)), height)
	if accessorValue != nil {
		var accessor data.Accessor
		err = proto.Unmarshal([]byte(accessorValue), &accessor)
		if err == nil {
			result.Exist = true
		}
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
}

func (app *DIDApplication) checkExistingAccessorGroupID(param string, height int64) types.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorGroupID, Parameter: %s", param)
	var funcParam CheckExistingAccessorGroupIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result CheckExistingResult
	result.Exist = false

	accessorGroupKey := "AccessorGroup" + "|" + funcParam.AccessorGroupID
	_, accessorGroupValue := app.state.db.GetVersioned(prefixKey([]byte(accessorGroupKey)), height)
	if accessorGroupValue != nil {
		result.Exist = true
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
}

func (app *DIDApplication) getNodeInfo(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodeInfo, Parameter: %s", param)
	var funcParam GetNodeInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.GetVersioned(prefixKey([]byte(nodeDetailKey)), height)
	if nodeDetailValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	// If node behind proxy
	proxyKey := "Proxy" + "|" + funcParam.NodeID
	_, proxyValue := app.state.db.Get(prefixKey([]byte(proxyKey)))
	if proxyValue != nil {
		// Get proxy node ID
		var proxy data.Proxy
		err = proto.Unmarshal([]byte(proxyValue), &proxy)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		proxyNodeID := proxy.ProxyNodeId

		// Get proxy node detail
		proxyNodeDetailKey := "NodeID" + "|" + string(proxyNodeID)
		_, proxyNodeDetailValue := app.state.db.GetVersioned(prefixKey([]byte(proxyNodeDetailKey)), height)
		if proxyNodeDetailValue == nil {
			return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
		}
		var proxyNode data.NodeDetail
		err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		if nodeDetail.Role == "IdP" {
			var result GetNodeInfoResultIdPandASBehindProxy
			result.PublicKey = nodeDetail.PublicKey
			result.MasterPublicKey = nodeDetail.MasterPublicKey
			result.NodeName = nodeDetail.NodeName
			result.Role = nodeDetail.Role
			result.MaxIal = nodeDetail.MaxIal
			result.MaxAal = nodeDetail.MaxAal
			result.Proxy.NodeID = string(proxyNodeID)
			result.Proxy.NodeName = proxyNode.NodeName
			result.Proxy.PublicKey = proxyNode.PublicKey
			result.Proxy.MasterPublicKey = proxyNode.MasterPublicKey
			if proxyNode.Mq != nil {
				for _, mq := range proxyNode.Mq {
					var msq MsqAddress
					msq.IP = mq.Ip
					msq.Port = mq.Port
					result.Proxy.Mq = append(result.Proxy.Mq, msq)
				}
			}
			result.Proxy.Config = proxy.Config
			value, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			return app.ReturnQuery(value, "success", app.state.db.Version64())
		}
		var result GetNodeInfoResultRPandASBehindProxy
		result.PublicKey = nodeDetail.PublicKey
		result.MasterPublicKey = nodeDetail.MasterPublicKey
		result.NodeName = nodeDetail.NodeName
		result.Role = nodeDetail.Role
		result.Proxy.NodeID = string(proxyNodeID)
		result.Proxy.NodeName = proxyNode.NodeName
		result.Proxy.PublicKey = proxyNode.PublicKey
		result.Proxy.MasterPublicKey = proxyNode.MasterPublicKey
		if proxyNode.Mq != nil {
			for _, mq := range proxyNode.Mq {
				var msq MsqAddress
				msq.IP = mq.Ip
				msq.Port = mq.Port
				result.Proxy.Mq = append(result.Proxy.Mq, msq)
			}
		}
		result.Proxy.Config = proxy.Config
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(value, "success", app.state.db.Version64())
	} else {
		if nodeDetail.Role == "IdP" {
			var result GetNodeInfoIdPResult
			result.PublicKey = nodeDetail.PublicKey
			result.MasterPublicKey = nodeDetail.MasterPublicKey
			result.NodeName = nodeDetail.NodeName
			result.Role = nodeDetail.Role
			result.MaxIal = nodeDetail.MaxIal
			result.MaxAal = nodeDetail.MaxAal
			if nodeDetail.Mq != nil {
				for _, mq := range nodeDetail.Mq {
					var msq MsqAddress
					msq.IP = mq.Ip
					msq.Port = mq.Port
					result.Mq = append(result.Mq, msq)
				}
			}
			value, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			return app.ReturnQuery(value, "success", app.state.db.Version64())
		}
		var result GetNodeInfoResult
		result.PublicKey = nodeDetail.PublicKey
		result.MasterPublicKey = nodeDetail.MasterPublicKey
		result.NodeName = nodeDetail.NodeName
		result.Role = nodeDetail.Role
		if nodeDetail.Mq != nil {
			for _, mq := range nodeDetail.Mq {
				var msq MsqAddress
				msq.IP = mq.Ip
				msq.Port = mq.Port
				result.Mq = append(result.Mq, msq)
			}
		}
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(value, "success", app.state.db.Version64())
	}
}

func (app *DIDApplication) getIdentityInfo(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetIdentityInfo, Parameter: %s", param)
	var funcParam GetIdentityInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result GetIdentityInfoResult

	key := "MsqDestination" + "|" + funcParam.HashID
	_, chkExists := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if chkExists != nil {
		var nodes data.MsqDesList
		err = proto.Unmarshal([]byte(chkExists), &nodes)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}

		for _, node := range nodes.Nodes {
			if node.NodeId == funcParam.NodeID {
				result.Ial = float64(node.Ial)
				break
			}
		}
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	if result.Ial > 0.0 {
		return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
	}
	return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
}

func (app *DIDApplication) getDataSignature(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetDataSignature, Parameter: %s", param)
	var funcParam GetDataSignatureParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	signDataKey := "SignData" + "|" + funcParam.NodeID + "|" + funcParam.ServiceID + "|" + funcParam.RequestID
	_, signDataValue := app.state.db.GetVersioned(prefixKey([]byte(signDataKey)), height)

	var result GetDataSignatureResult

	if signDataValue != nil {
		result.Signature = string(signDataValue)
	}

	returnValue, err := json.Marshal(result)
	if signDataValue != nil {
		return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
	}
	return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
}

func (app *DIDApplication) getIdentityProof(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetIdentityProof, Parameter: %s", param)
	var funcParam GetIdentityProofParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	identityProofKey := "IdentityProof" + "|" + funcParam.RequestID + "|" + funcParam.IdpID
	_, identityProofValue := app.state.db.GetVersioned(prefixKey([]byte(identityProofKey)), height)
	var result GetIdentityProofResult
	if identityProofValue != nil {
		result.IdentityProof = string(identityProofValue)
	}
	returnValue, err := json.Marshal(result)
	if identityProofValue != nil {
		return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
	}
	return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
}

func (app *DIDApplication) getServicesByAsID(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetServicesByAsID, Parameter: %s", param)
	var funcParam GetServicesByAsIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result GetServicesByAsIDResult
	result.Services = make([]Service, 0)

	provideServiceKey := "ProvideService" + "|" + funcParam.AsID
	_, provideServiceValue := app.state.db.Get(prefixKey([]byte(provideServiceKey)))
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
	}

	nodeDetailKey := "NodeID" + "|" + funcParam.AsID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var nodeDetail data.NodeDetail
	if nodeDetailValue != nil {
		err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
	}

	for index, provideService := range services.Services {
		serviceKey := "Service" + "|" + provideService.ServiceId
		_, serviceValue := app.state.db.Get(prefixKey([]byte(serviceKey)))
		var service data.ServiceDetail
		if serviceValue != nil {
			err = proto.Unmarshal([]byte(serviceValue), &service)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
		}
		if nodeDetail.Active && service.Active {
			// Set suspended from NDID
			approveServiceKey := "ApproveKey" + "|" + provideService.ServiceId + "|" + funcParam.AsID
			_, approveServiceJSON := app.state.db.Get(prefixKey([]byte(approveServiceKey)))
			if approveServiceJSON != nil {
				var approveService data.ApproveService
				err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
				if err == nil {
					services.Services[index].Suspended = !approveService.Active
				}
			}
			var newRow Service
			newRow.Active = services.Services[index].Active
			newRow.MinAal = services.Services[index].MinAal
			newRow.MinIal = services.Services[index].MinIal
			newRow.ServiceID = services.Services[index].ServiceId
			newRow.Suspended = services.Services[index].Suspended
			result.Services = append(result.Services, newRow)
		}
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	if len(result.Services) > 0 {
		return app.ReturnQuery(resultJSON, "success", app.state.db.Version64())
	} else {
		return app.ReturnQuery(resultJSON, "not found", app.state.db.Version64())
	}
}

func (app *DIDApplication) getIdpNodesInfo(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetIdpNodesInfo, Parameter: %s", param)
	var funcParam GetIdpNodesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	var result GetIdpNodesInfoResult
	result.Node = make([]interface{}, 0)

	// Make mapping
	mapNodeIDList := map[string]bool{}
	for _, nodeID := range funcParam.NodeIDList {
		mapNodeIDList[nodeID] = true
	}

	if funcParam.HashID == "" {
		idpsKey := "IdPList"
		_, idpsValue := app.state.db.GetVersioned(prefixKey([]byte(idpsKey)), height)
		var idpsList data.IdPList
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			for _, idp := range idpsList.NodeId {

				// filter from node_id_list
				if len(mapNodeIDList) > 0 {
					if mapNodeIDList[idp] == false {
						continue
					}
				}

				nodeDetailKey := "NodeID" + "|" + idp
				_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
				if nodeDetailValue == nil {
					continue
				}
				var nodeDetail data.NodeDetail
				err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
				if err != nil {
					continue
				}
				// check node is active
				if !nodeDetail.Active {
					continue
				}
				// check Max IAL && AAL
				if !(nodeDetail.MaxIal >= funcParam.MinIal &&
					nodeDetail.MaxAal >= funcParam.MinAal) {
					continue
				}

				// If node is behind proxy
				proxyKey := "Proxy" + "|" + idp
				_, proxyValue := app.state.db.Get(prefixKey([]byte(proxyKey)))
				if proxyValue != nil {

					// Get proxy node ID
					var proxy data.Proxy
					err = proto.Unmarshal([]byte(proxyValue), &proxy)
					if err != nil {
						return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
					}
					proxyNodeID := proxy.ProxyNodeId

					// Get proxy node detail
					proxyNodeDetailKey := "NodeID" + "|" + string(proxyNodeID)
					_, proxyNodeDetailValue := app.state.db.GetVersioned(prefixKey([]byte(proxyNodeDetailKey)), height)
					if proxyNodeDetailValue == nil {
						return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
					}
					var proxyNode data.NodeDetail
					err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
					if err != nil {
						return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
					}
					var msqDesNode IdpNodeBehindProxy
					msqDesNode.NodeID = idp
					msqDesNode.Name = nodeDetail.NodeName
					msqDesNode.MaxIal = nodeDetail.MaxIal
					msqDesNode.MaxAal = nodeDetail.MaxAal
					msqDesNode.PublicKey = nodeDetail.PublicKey
					msqDesNode.Proxy.NodeID = string(proxyNodeID)
					msqDesNode.Proxy.PublicKey = proxyNode.PublicKey
					if proxyNode.Mq != nil {
						for _, mq := range proxyNode.Mq {
							var msq MsqAddress
							msq.IP = mq.Ip
							msq.Port = mq.Port
							msqDesNode.Proxy.Mq = append(msqDesNode.Proxy.Mq, msq)
						}
					}
					msqDesNode.Proxy.Config = proxy.Config
					result.Node = append(result.Node, msqDesNode)
				} else {
					var msq []MsqAddress
					for _, mq := range nodeDetail.Mq {
						var msqAddress MsqAddress
						msqAddress.IP = mq.Ip
						msqAddress.Port = mq.Port
						msq = append(msq, msqAddress)
					}
					var msqDesNode = IdpNode{
						idp,
						nodeDetail.NodeName,
						nodeDetail.MaxIal,
						nodeDetail.MaxAal,
						nodeDetail.PublicKey,
						msq,
					}
					result.Node = append(result.Node, msqDesNode)
				}

			}
		}
	} else {
		key := "MsqDestination" + "|" + funcParam.HashID
		_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
		if value != nil {
			var nodes data.MsqDesList
			err = proto.Unmarshal([]byte(value), &nodes)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			for _, node := range nodes.Nodes {
				// filter from node_id_list
				if len(mapNodeIDList) > 0 {
					if mapNodeIDList[node.NodeId] == false {
						continue
					}
				}
				// check msq destination is not active
				if !node.Active {
					continue
				}
				// check Ial > min ial
				if node.Ial < funcParam.MinIal {
					continue
				}
				// check msq destination is not timed out
				if node.TimeoutBlock != 0 && app.CurrentBlock > node.TimeoutBlock {
					continue
				}
				nodeDetailKey := "NodeID" + "|" + node.NodeId
				_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
				if nodeDetailValue == nil {
					continue
				}
				var nodeDetail data.NodeDetail
				err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
				if err != nil {
					continue
				}
				// check node is active
				if !nodeDetail.Active {
					continue
				}
				// check Max IAL && AAL
				if !(nodeDetail.MaxIal >= funcParam.MinIal &&
					nodeDetail.MaxAal >= funcParam.MinAal) {
					continue
				}

				// If node is behind proxy
				proxyKey := "Proxy" + "|" + node.NodeId
				_, proxyValue := app.state.db.Get(prefixKey([]byte(proxyKey)))
				if proxyValue != nil {

					// Get proxy node ID
					var proxy data.Proxy
					err = proto.Unmarshal([]byte(proxyValue), &proxy)
					if err != nil {
						return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
					}
					proxyNodeID := proxy.ProxyNodeId

					// Get proxy node detail
					proxyNodeDetailKey := "NodeID" + "|" + string(proxyNodeID)
					_, proxyNodeDetailValue := app.state.db.GetVersioned(prefixKey([]byte(proxyNodeDetailKey)), height)
					if proxyNodeDetailValue == nil {
						return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
					}
					var proxyNode data.NodeDetail
					err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
					if err != nil {
						return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
					}
					var msqDesNode IdpNodeBehindProxy
					msqDesNode.NodeID = node.NodeId
					msqDesNode.Name = nodeDetail.NodeName
					msqDesNode.MaxIal = nodeDetail.MaxIal
					msqDesNode.MaxAal = nodeDetail.MaxAal
					msqDesNode.PublicKey = nodeDetail.PublicKey
					msqDesNode.Proxy.NodeID = string(proxyNodeID)
					msqDesNode.Proxy.PublicKey = proxyNode.PublicKey
					if proxyNode.Mq != nil {
						for _, mq := range proxyNode.Mq {
							var msq MsqAddress
							msq.IP = mq.Ip
							msq.Port = mq.Port
							msqDesNode.Proxy.Mq = append(msqDesNode.Proxy.Mq, msq)
						}
					}
					msqDesNode.Proxy.Config = proxy.Config
					result.Node = append(result.Node, msqDesNode)
				} else {
					var msq []MsqAddress
					for _, mq := range nodeDetail.Mq {
						var msqAddress MsqAddress
						msqAddress.IP = mq.Ip
						msqAddress.Port = mq.Port
						msq = append(msq, msqAddress)
					}
					var msqDesNode = IdpNode{
						node.NodeId,
						nodeDetail.NodeName,
						nodeDetail.MaxIal,
						nodeDetail.MaxAal,
						nodeDetail.PublicKey,
						msq,
					}
					result.Node = append(result.Node, msqDesNode)
				}
			}
		}
	}

	value, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	if len(result.Node) > 0 {
		return app.ReturnQuery(value, "success", app.state.db.Version64())
	}
	return app.ReturnQuery(value, "not found", app.state.db.Version64())
}

func (app *DIDApplication) getAsNodesInfoByServiceId(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetAsNodesInfoByServiceId, Parameter: %s", param)
	var funcParam GetAsNodesByServiceIdParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	key := "ServiceDestination" + "|" + funcParam.ServiceID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		var result GetAsNodesInfoByServiceIdResult
		result.Node = make([]interface{}, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(value, "not found", app.state.db.Version64())
	}

	// filter serive is active
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceValue := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceValue != nil {
		var service data.ServiceDetail
		err = proto.Unmarshal([]byte(serviceValue), &service)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		if service.Active == false {
			var result GetAsNodesByServiceIdResult
			result.Node = make([]ASNode, 0)
			value, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			return app.ReturnQuery(value, "service is not active", app.state.db.Version64())
		}
	} else {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(value, "not found", app.state.db.Version64())
	}

	var storedData data.ServiceDesList
	err = proto.Unmarshal([]byte(value), &storedData)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
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
		approveServiceKey := "ApproveKey" + "|" + funcParam.ServiceID + "|" + storedData.Node[index].NodeId
		_, approveServiceJSON := app.state.db.Get(prefixKey([]byte(approveServiceKey)))
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

		nodeDetailKey := "NodeID" + "|" + storedData.Node[index].NodeId
		_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			continue
		}
		// filter node is active
		if !nodeDetail.Active {
			continue
		}

		// If node is behind proxy
		proxyKey := "Proxy" + "|" + storedData.Node[index].NodeId
		_, proxyValue := app.state.db.Get(prefixKey([]byte(proxyKey)))
		if proxyValue != nil {

			// Get proxy node ID
			var proxy data.Proxy
			err = proto.Unmarshal([]byte(proxyValue), &proxy)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			proxyNodeID := proxy.ProxyNodeId

			// Get proxy node detail
			proxyNodeDetailKey := "NodeID" + "|" + string(proxyNodeID)
			_, proxyNodeDetailValue := app.state.db.GetVersioned(prefixKey([]byte(proxyNodeDetailKey)), height)
			if proxyNodeDetailValue == nil {
				return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
			}
			var proxyNode data.NodeDetail
			err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			var as ASWithMqNodeBehindProxy
			as.NodeID = storedData.Node[index].NodeId
			as.Name = nodeDetail.NodeName
			as.MinIal = storedData.Node[index].MinIal
			as.MinAal = storedData.Node[index].MinAal
			as.PublicKey = nodeDetail.PublicKey
			as.Proxy.NodeID = string(proxyNodeID)
			as.Proxy.PublicKey = proxyNode.PublicKey
			if proxyNode.Mq != nil {
				for _, mq := range proxyNode.Mq {
					var msq MsqAddress
					msq.IP = mq.Ip
					msq.Port = mq.Port
					as.Proxy.Mq = append(as.Proxy.Mq, msq)
				}
			}
			as.Proxy.Config = proxy.Config
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
				storedData.Node[index].NodeId,
				nodeDetail.NodeName,
				storedData.Node[index].MinIal,
				storedData.Node[index].MinAal,
				nodeDetail.PublicKey,
				msqAddress,
			}
			result.Node = append(result.Node, newRow)
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(resultJSON, "success", app.state.db.Version64())
}

func (app *DIDApplication) getNodesBehindProxyNode(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodesBehindProxyNode, Parameter: %s", param)
	var funcParam GetNodesBehindProxyNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	var result GetNodesBehindProxyNodeResult
	result.Nodes = make([]interface{}, 0)
	behindProxyNodeKey := "BehindProxyNode" + "|" + funcParam.ProxyNodeID
	_, behindProxyNodeValue := app.state.db.Get(prefixKey([]byte(behindProxyNodeKey)))
	if behindProxyNodeValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.db.Version64())
	}
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)
	err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	for _, node := range nodes.Nodes {
		nodeDetailKey := "NodeID" + "|" + node
		_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			continue
		}

		// Get proxy detail
		proxyKey := "Proxy" + "|" + node
		_, proxyValue := app.state.db.Get(prefixKey([]byte(proxyKey)))
		if proxyValue == nil {
			continue
		}
		var proxy data.Proxy
		err = proto.Unmarshal([]byte(proxyValue), &proxy)
		if err != nil {
			continue
		}

		if nodeDetail.Role == "IdP" {
			var row IdPBehindProxy
			row.NodeID = node
			row.NodeName = nodeDetail.NodeName
			row.Role = nodeDetail.Role
			row.PublicKey = nodeDetail.PublicKey
			row.MasterPublicKey = nodeDetail.MasterPublicKey
			row.MaxIal = nodeDetail.MaxIal
			row.MaxAal = nodeDetail.MaxAal
			row.Config = proxy.Config
			result.Nodes = append(result.Nodes, row)
		} else {
			var row ASorRPBehindProxy
			row.NodeID = node
			row.NodeName = nodeDetail.NodeName
			row.Role = nodeDetail.Role
			row.PublicKey = nodeDetail.PublicKey
			row.MasterPublicKey = nodeDetail.MasterPublicKey
			row.Config = proxy.Config
			result.Nodes = append(result.Nodes, row)
		}

	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	if len(result.Nodes) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.db.Version64())
	}
	return app.ReturnQuery(resultJSON, "success", app.state.db.Version64())
}

func (app *DIDApplication) getNodeIDList(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodeIDList, Parameter: %s", param)
	var funcParam GetNodeIDListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}

	var result GetNodeIDListResult
	result.NodeIDList = make([]string, 0)

	if strings.ToLower(funcParam.Role) == "rp" {
		var rpsList data.RPList
		rpsKey := "rpList"
		_, rpsValue := app.state.db.Get(prefixKey([]byte(rpsKey)))
		if rpsValue != nil {
			err := proto.Unmarshal(rpsValue, &rpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			for _, nodeID := range rpsList.NodeId {
				nodeDetailKey := "NodeID" + "|" + nodeID
				_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
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
		idpsKey := "IdPList"
		_, idpsValue := app.state.db.Get(prefixKey([]byte(idpsKey)))
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			for _, nodeID := range idpsList.NodeId {
				nodeDetailKey := "NodeID" + "|" + nodeID
				_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
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
		_, asValue := app.state.db.Get(prefixKey([]byte(asKey)))
		if asValue != nil {
			err := proto.Unmarshal(asValue, &asList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			for _, nodeID := range asList.NodeId {
				nodeDetailKey := "NodeID" + "|" + nodeID
				_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
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
		_, allValue := app.state.db.Get(prefixKey([]byte(allKey)))
		if allValue != nil {
			err := proto.Unmarshal(allValue, &allList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
			}
			for _, nodeID := range allList.NodeId {
				nodeDetailKey := "NodeID" + "|" + nodeID
				_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
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
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	if len(result.NodeIDList) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.db.Version64())
	}
	return app.ReturnQuery(resultJSON, "success", app.state.db.Version64())
}

func (app *DIDApplication) getAccessorsInAccessorGroup(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetAccessorsInAccessorGroup, Parameter: %s", param)
	var funcParam GetAccessorsInAccessorGroupParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	var result GetAccessorsInAccessorGroupResult
	result.AccessorList = make([]string, 0)
	if funcParam.AccessorGroupID != "" {
		accessorInGroupKey := "AccessorInGroup" + "|" + funcParam.AccessorGroupID
		_, accessorInGroupKeyValue := app.state.db.GetVersioned(prefixKey([]byte(accessorInGroupKey)), height)
		var accessors data.AccessorInGroup
		err := proto.Unmarshal(accessorInGroupKeyValue, &accessors)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
		}
		// If IdpID == "", return all accessors in group
		if funcParam.IdpID == "" {
			for _, accessor := range accessors.Accessors {
				result.AccessorList = append(result.AccessorList, accessor)
			}
		} else {
			// filter by owner of accessor
			for _, accessor := range accessors.Accessors {
				accessorKey := "Accessor" + "|" + accessor
				_, accessorValue := app.state.db.GetVersioned(prefixKey([]byte(accessorKey)), height)
				var accessorObj data.Accessor
				err := proto.Unmarshal(accessorValue, &accessorObj)
				if err != nil {
					return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
				}
				if accessorObj.Owner == funcParam.IdpID {
					result.AccessorList = append(result.AccessorList, accessor)
				}
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if len(result.AccessorList) > 0 {
		return app.ReturnQuery(returnValue, "success", app.state.db.Version64())
	}
	return app.ReturnQuery(returnValue, "not found", app.state.db.Version64())
}
