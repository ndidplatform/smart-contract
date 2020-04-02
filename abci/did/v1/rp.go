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

	"github.com/golang/protobuf/proto"
	"github.com/ndidplatform/smart-contract/v4/abci/code"
	"github.com/ndidplatform/smart-contract/v4/abci/utils"
	"github.com/ndidplatform/smart-contract/v4/protos/data"
	"github.com/tendermint/tendermint/abci/types"
)

func (app *DIDApplication) createRequest(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateRequest, Parameter: %s", param)
	var funcParam CreateRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// get RP node detail
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, nodeDetaiValue := app.GetStateDB([]byte(nodeDetailKey))
	if nodeDetaiValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	var rpNodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetaiValue), &rpNodeDetail)

	// log chain ID
	app.logger.Infof("CreateRequest, Chain ID: %s", app.CurrentChain)
	var request data.Request
	// set request data
	request.RequestId = funcParam.RequestID

	key := "Request" + "|" + request.RequestId
	requestIDExist := app.HasVersionedStateDB([]byte(key))
	if requestIDExist {
		return app.ReturnDeliverTxLog(code.DuplicateRequestID, "Duplicate Request ID", "")
	}

	request.MinIdp = int64(funcParam.MinIdp)
	request.MinAal = funcParam.MinAal
	request.MinIal = funcParam.MinIal
	request.RequestTimeout = int64(funcParam.Timeout)
	// request.DataRequestList = funcParam.DataRequestList
	request.RequestMessageHash = funcParam.MessageHash
	request.Mode = funcParam.Mode
	// Check valid mode
	allowedMode := app.GetAllowedModeFromStateDB(funcParam.Purpose, false)
	validMode := false
	for _, mode := range allowedMode {
		if mode == request.Mode {
			validMode = true
			break
		}
	}
	if !validMode {
		return app.ReturnDeliverTxLog(code.InvalidMode, "Must be create request on valid mode", "")
	}
	request.IdpIdList = funcParam.IdPIDList
	// Check all IdP in list is active
	for _, idp := range request.IdpIdList {
		// Check idp is in the rp whitelist
		if rpNodeDetail.UseWhitelist && !contains(idp, rpNodeDetail.Whitelist) {
			return app.ReturnDeliverTxLog(code.NodeNotInWhitelist, "IdP is not in RP whitelist", "")
		}

		// Get node detail
		nodeDetailKey := "NodeID" + "|" + idp
		_, nodeDetaiValue := app.GetStateDB([]byte(nodeDetailKey))
		if nodeDetaiValue == nil {
			return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
		}
		var node data.NodeDetail
		err = proto.Unmarshal([]byte(nodeDetaiValue), &node)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// Check node is active
		if !node.Active {
			return app.ReturnDeliverTxLog(code.NodeIDInIdPListIsNotActive, "Node ID in IdP list is not active", "")
		}

		// Check rp is in the idp whitelist
		if node.UseWhitelist && !contains(nodeID, node.Whitelist) {
			return app.ReturnDeliverTxLog(code.NodeNotInWhitelist, "RP is not in IdP whitelist", "")
		}

		// If node is behind proxy
		if node.ProxyNodeId != "" {
			proxyNodeID := node.ProxyNodeId
			// Get proxy node detail
			proxyNodeDetailKey := "NodeID" + "|" + string(proxyNodeID)
			_, proxyNodeDetailValue := app.GetStateDB([]byte(proxyNodeDetailKey))
			if proxyNodeDetailValue == nil {
				return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
			}
			var proxyNode data.NodeDetail
			err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
			// Check proxy node is active
			if !proxyNode.Active {
				return app.ReturnDeliverTxLog(code.NodeIDInIdPListIsNotActive, "Node ID in IdP list is not active", "")
			}
		}
	}
	// set data request
	request.DataRequestList = make([]*data.DataRequest, 0)
	serviceIDInDataRequestList := make(map[string]int)
	nodeDetailMap := make(map[string]*data.NodeDetail, 0)
	for index := range funcParam.DataRequestList {
		var newRow data.DataRequest
		newRow.ServiceId = funcParam.DataRequestList[index].ServiceID

		// Check for duplicate service ID in data request list
		_, exist := serviceIDInDataRequestList[newRow.ServiceId]
		if exist {
			return app.ReturnDeliverTxLog(code.DuplicateServiceIDInDataRequest, "Duplicate Service ID In Data Request", "")
		}
		serviceIDInDataRequestList[newRow.ServiceId]++

		newRow.RequestParamsHash = funcParam.DataRequestList[index].RequestParamsHash
		newRow.MinAs = int64(funcParam.DataRequestList[index].Count)
		newRow.AsIdList = funcParam.DataRequestList[index].As
		if funcParam.DataRequestList[index].As == nil {
			newRow.AsIdList = make([]string, 0)
		}
		newRow.ResponseList = make([]*data.ASResponse, 0)
		// Check all as in as_list is active
		for _, as := range newRow.AsIdList {
			var node data.NodeDetail
			if nodeDetailMap[as] == nil {
				// Get node detail
				nodeDetailKey := "NodeID" + "|" + as
				_, nodeDetaiValue := app.GetStateDB([]byte(nodeDetailKey))
				if nodeDetaiValue == nil {
					return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
				}
				err = proto.Unmarshal([]byte(nodeDetaiValue), &node)
				if err != nil {
					return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
				}
				// Save node detail to mapping
				nodeDetailMap[as] = &node
			} else {
				// Get node detail from mapping
				node = *nodeDetailMap[as]
			}

			// Check node is active
			if !node.Active {
				return app.ReturnDeliverTxLog(code.NodeIDInASListIsNotActive, "Node ID in AS list is not active", "")
			}

			// If node is behind proxy
			if node.ProxyNodeId != "" {
				proxyNodeID := node.ProxyNodeId
				// Get proxy node detail
				proxyNodeDetailKey := "NodeID" + "|" + string(proxyNodeID)
				_, proxyNodeDetailValue := app.GetStateDB([]byte(proxyNodeDetailKey))
				if proxyNodeDetailValue == nil {
					return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
				}
				var proxyNode data.NodeDetail
				err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
				if err != nil {
					return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
				}
				// Check proxy node is active
				if !proxyNode.Active {
					return app.ReturnDeliverTxLog(code.NodeIDInASListIsNotActive, "Node ID in AS list is not active", "")
				}
			}
		}
		request.DataRequestList = append(request.DataRequestList, &newRow)
	}
	// set default value
	request.Closed = false
	request.TimedOut = false
	request.Purpose = ""
	request.UseCount = 0
	// set Owner
	request.Owner = nodeID
	// set Can add accossor
	ownerRole := app.getRoleFromNodeID(nodeID)
	if string(ownerRole) == "IdP" {
		request.Purpose = funcParam.Purpose
	}
	// set default value
	request.ResponseList = make([]*data.Response, 0)
	// set creation_block_height
	request.CreationBlockHeight = app.CurrentBlock
	// set chain_id
	request.ChainId = app.CurrentChain

	value, err := utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetVersionedStateDB([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", request.RequestId)
}

func (app *DIDApplication) closeRequest(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CloseRequest, Parameter: %s", param)
	var funcParam CloseRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := "Request" + "|" + funcParam.RequestID
	_, value := app.GetVersionedStateDB([]byte(key), 0)
	if value == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Can not close a closed request", "")
	}
	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Can not close a timed out request", "")
	}
	for _, valid := range funcParam.ResponseValidList {
		for index := range request.ResponseList {
			if valid.IdpID == request.ResponseList[index].IdpId {
				if valid.ValidIal != nil {
					if *valid.ValidIal {
						request.ResponseList[index].ValidIal = "true"
					} else {
						request.ResponseList[index].ValidIal = "false"
					}
				}
				if valid.ValidSignature != nil {
					if *valid.ValidSignature {
						request.ResponseList[index].ValidSignature = "true"
					} else {
						request.ResponseList[index].ValidSignature = "false"
					}
				}
			}
		}
	}
	request.Closed = true
	value, err = utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetVersionedStateDB([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

func (app *DIDApplication) timeOutRequest(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("TimeOutRequest, Parameter: %s", param)
	var funcParam TimeOutRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := "Request" + "|" + funcParam.RequestID
	_, value := app.GetVersionedStateDB([]byte(key), 0)
	if value == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Can not set time out a timed out request", "")
	}
	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Can not set time out a closed request", "")
	}
	for _, valid := range funcParam.ResponseValidList {
		for index := range request.ResponseList {
			if valid.IdpID == request.ResponseList[index].IdpId {
				if valid.ValidIal != nil {
					if *valid.ValidIal {
						request.ResponseList[index].ValidIal = "true"
					} else {
						request.ResponseList[index].ValidIal = "false"
					}
				}
				if valid.ValidSignature != nil {
					if *valid.ValidSignature {
						request.ResponseList[index].ValidSignature = "true"
					} else {
						request.ResponseList[index].ValidSignature = "false"
					}
				}
			}
		}
	}
	request.TimedOut = true
	value, err = utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetVersionedStateDB([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

func (app *DIDApplication) setDataReceived(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetDataReceived, Parameter: %s", param)
	var funcParam SetDataReceivedParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := "Request" + "|" + funcParam.RequestID
	_, value := app.GetVersionedStateDB([]byte(key), 0)
	if value == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
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

	var targetAsResponse *data.ASResponse
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == funcParam.ServiceID {
			for _, asResponse := range dataRequest.ResponseList {
				if asResponse.AsId == funcParam.AsID {
					targetAsResponse = asResponse
					break
				}
			}
		}
	}
	// Check as_id is exist in as_id_list
	if targetAsResponse == nil {
		return app.ReturnDeliverTxLog(code.AsIDDoesNotExistInASList, "AS ID does not exist in answered AS list", "")
	}
	// Check Duplicate
	if targetAsResponse.ReceivedData {
		return app.ReturnDeliverTxLog(code.DuplicateASInDataRequest, "Duplicate AS ID in data request", "")
	}
	// Update targetAsResponse status
	targetAsResponse.ReceivedData = true

	value, err = utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetVersionedStateDB([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}
