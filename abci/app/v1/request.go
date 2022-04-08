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

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

type IdPResponse struct {
	Ial            *float64 `json:"ial,omitempty"`
	Aal            *float64 `json:"aal,omitempty"`
	Status         *string  `json:"status,omitempty"`
	Signature      *string  `json:"signature,omitempty"`
	IdpID          string   `json:"idp_id"`
	ValidIal       *bool    `json:"valid_ial"`
	ValidSignature *bool    `json:"valid_signature"`
	ErrorCode      *int32   `json:"error_code,omitempty"`
}

type ASResponse struct {
	AsID         string `json:"as_id"`
	Signed       *bool  `json:"signed,omitempty"`
	ReceivedData *bool  `json:"received_data,omitempty"`
	ErrorCode    *int32 `json:"error_code,omitempty"`
}

type DataRequest struct {
	ServiceID         string       `json:"service_id"`
	As                []string     `json:"as_id_list"`
	Count             int          `json:"min_as"`
	RequestParamsHash string       `json:"request_params_hash"`
	ResponseList      []ASResponse `json:"response_list"`
}

type CreateRequestParam struct {
	RequestID       string        `json:"request_id"`
	MinIdp          int           `json:"min_idp"`
	MinAal          float64       `json:"min_aal"`
	MinIal          float64       `json:"min_ial"`
	Timeout         int           `json:"request_timeout"`
	IdPIDList       []string      `json:"idp_id_list"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"request_message_hash"`
	Purpose         string        `json:"purpose"`
	Mode            int32         `json:"mode"`
	RequestType     *string       `json:"request_type"`
}

func (app *ABCIApplication) createRequest(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateRequest, Parameter: %s", param)
	var funcParam CreateRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// get RP node detail
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	nodeDetaiValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if nodeDetaiValue == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	var requesterNodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetaiValue), &requesterNodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// log chain ID
	app.logger.Infof("CreateRequest, Chain ID: %s", app.CurrentChain)
	var request data.Request
	// set request data
	request.RequestId = funcParam.RequestID
	request.Mode = funcParam.Mode

	if requesterNodeDetail.Role == "IdP" {
		// IdP must not be able to create request with mode 1 or 2
		if request.Mode == 1 {
			return app.ReturnDeliverTxLog(code.IdPCreateRequestMode1And2NotAllowed, "IdP cannot create request with mode 1 or 2", "")
		}
		// IdP must not be able to create request with data request to AS
		if len(funcParam.DataRequestList) > 0 {
			return app.ReturnDeliverTxLog(code.IdPCreateRequestWithDataRequestNotAllowed, "IdP cannot create request with data request", "")
		}
	}

	key := requestKeyPrefix + keySeparator + request.RequestId
	requestIDExist, err := app.state.HasVersioned([]byte(key), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if requestIDExist {
		return app.ReturnDeliverTxLog(code.DuplicateRequestID, "Duplicate Request ID", "")
	}

	request.MinIdp = int64(funcParam.MinIdp)
	request.MinAal = funcParam.MinAal
	request.MinIal = funcParam.MinIal
	request.RequestTimeout = int64(funcParam.Timeout)
	// request.DataRequestList = funcParam.DataRequestList
	request.RequestMessageHash = funcParam.MessageHash
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
		if requesterNodeDetail.UseWhitelist && !contains(idp, requesterNodeDetail.Whitelist) {
			return app.ReturnDeliverTxLog(code.NodeNotInWhitelist, "IdP is not in RP whitelist", "")
		}

		// Get node detail
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp
		nodeDetaiValue, err := app.state.Get([]byte(nodeDetailKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
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
			proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
			proxyNodeDetailValue, err := app.state.Get([]byte(proxyNodeDetailKey), false)
			if err != nil {
				return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
			}
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
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + as
				nodeDetaiValue, err := app.state.Get([]byte(nodeDetailKey), false)
				if err != nil {
					return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
				}
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
				proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
				proxyNodeDetailValue, err := app.state.Get([]byte(proxyNodeDetailKey), false)
				if err != nil {
					return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
				}
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

	if funcParam.RequestType != nil {
		key := requestTypeKeyPrefix + keySeparator + *funcParam.RequestType
		requestTypeExists, err := app.state.Has([]byte(key), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if !requestTypeExists {
			return app.ReturnDeliverTxLog(code.RequestTypeDoesNotExist, err.Error(), "")
		}

		request.RequestType = *funcParam.RequestType
	}

	// set Owner
	request.Owner = nodeID
	// set Can add accossor
	if requesterNodeDetail.Role == "IdP" {
		request.Purpose = funcParam.Purpose
	}
	// set default value
	request.ResponseList = make([]*data.Response, 0)
	// set creation_block_height
	request.CreationBlockHeight = app.state.CurrentBlockHeight
	// set chain_id
	request.ChainId = app.CurrentChain

	value, err := utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	err = app.state.SetVersioned([]byte(key), []byte(value))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", request.RequestId)
}

type ResponseValid struct {
	IdpID          string `json:"idp_id"`
	ValidIal       *bool  `json:"valid_ial"`
	ValidSignature *bool  `json:"valid_signature"`
}

type CloseRequestParam struct {
	RequestID         string          `json:"request_id"`
	ResponseValidList []ResponseValid `json:"response_valid_list"`
}

func (app *ABCIApplication) closeRequest(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CloseRequest, Parameter: %s", param)
	var funcParam CloseRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := requestKeyPrefix + keySeparator + funcParam.RequestID
	value, err := app.state.GetVersioned([]byte(key), 0, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
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
	err = app.state.SetVersioned([]byte(key), []byte(value))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

type TimeOutRequestParam struct {
	RequestID         string          `json:"request_id"`
	ResponseValidList []ResponseValid `json:"response_valid_list"`
}

func (app *ABCIApplication) timeOutRequest(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("TimeOutRequest, Parameter: %s", param)
	var funcParam TimeOutRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := requestKeyPrefix + keySeparator + funcParam.RequestID
	value, err := app.state.GetVersioned([]byte(key), 0, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
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
	err = app.state.SetVersioned([]byte(key), []byte(value))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

type SetDataReceivedParam struct {
	RequestID string `json:"request_id"`
	ServiceID string `json:"service_id"`
	AsID      string `json:"as_id"`
}

func (app *ABCIApplication) setDataReceived(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetDataReceived, Parameter: %s", param)
	var funcParam SetDataReceivedParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := requestKeyPrefix + keySeparator + funcParam.RequestID
	value, err := app.state.GetVersioned([]byte(key), 0, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
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
	err = app.state.SetVersioned([]byte(key), []byte(value))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

type GetRequestParam struct {
	RequestID string `json:"request_id"`
}

type GetRequestResult struct {
	IsClosed    bool   `json:"closed"`
	IsTimedOut  bool   `json:"timed_out"`
	MessageHash string `json:"request_message_hash"`
	Mode        int32  `json:"mode"`
}

func (app *ABCIApplication) getRequest(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetRequest, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := requestKeyPrefix + keySeparator + funcParam.RequestID
	value, err := app.state.GetVersioned([]byte(key), height, true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	if value == nil {
		valueJSON := []byte("{}")
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var res GetRequestResult
	res.IsClosed = request.Closed
	res.IsTimedOut = request.TimedOut
	res.MessageHash = request.RequestMessageHash
	res.Mode = request.Mode

	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)
}

type GetRequestDetailResult struct {
	RequestID           string        `json:"request_id"`
	MinIdp              int           `json:"min_idp"`
	MinAal              float64       `json:"min_aal"`
	MinIal              float64       `json:"min_ial"`
	Timeout             int           `json:"request_timeout"`
	IdPIDList           []string      `json:"idp_id_list"`
	DataRequestList     []DataRequest `json:"data_request_list"`
	MessageHash         string        `json:"request_message_hash"`
	Responses           []IdPResponse `json:"response_list"`
	IsClosed            bool          `json:"closed"`
	IsTimedOut          bool          `json:"timed_out"`
	Purpose             string        `json:"purpose"`
	Mode                int32         `json:"mode"`
	RequestType         *string       `json:"request_type"`
	RequesterNodeID     string        `json:"requester_node_id"`
	CreationBlockHeight int64         `json:"creation_block_height"`
	CreationChainID     string        `json:"creation_chain_id"`
}

func (app *ABCIApplication) getRequestDetail(param string, height int64, committedState bool) types.ResponseQuery {
	app.logger.Infof("GetRequestDetail, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	key := requestKeyPrefix + keySeparator + funcParam.RequestID
	var value []byte
	value, err = app.state.GetVersioned([]byte(key), height, committedState)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	if value == nil {
		valueJSON := []byte("{}")
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}

	var result GetRequestDetailResult
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		value = []byte("")
		return app.ReturnQuery(value, err.Error(), app.state.Height)
	}

	result.RequestID = request.RequestId
	result.MinIdp = int(request.MinIdp)
	result.MinAal = float64(request.MinAal)
	result.MinIal = float64(request.MinIal)
	result.Timeout = int(request.RequestTimeout)
	result.IdPIDList = request.IdpIdList
	result.DataRequestList = make([]DataRequest, 0)
	for _, dataRequest := range request.DataRequestList {
		newRow := DataRequest{
			ServiceID:         dataRequest.ServiceId,
			As:                dataRequest.AsIdList,
			Count:             int(dataRequest.MinAs),
			ResponseList:      make([]ASResponse, 0, len(dataRequest.ResponseList)),
			RequestParamsHash: dataRequest.RequestParamsHash,
		}
		for _, asResponse := range dataRequest.ResponseList {
			if asResponse.ErrorCode == 0 {
				newRow.ResponseList = append(newRow.ResponseList, ASResponse{
					AsID:         asResponse.AsId,
					Signed:       &asResponse.Signed,
					ReceivedData: &asResponse.ReceivedData,
				})
			} else {
				newRow.ResponseList = append(newRow.ResponseList, ASResponse{
					AsID:      asResponse.AsId,
					ErrorCode: &asResponse.ErrorCode,
				})
			}
		}
		result.DataRequestList = append(result.DataRequestList, newRow)
	}
	result.MessageHash = request.RequestMessageHash
	result.Responses = make([]IdPResponse, 0)
	for _, response := range request.ResponseList {
		var newRow IdPResponse
		if response.ErrorCode == 0 {
			var validIal *bool
			if response.ValidIal != "" {
				tValue := response.ValidIal == "true"
				validIal = &tValue
			}
			var validSignature *bool
			if response.ValidSignature != "" {
				tValue := response.ValidSignature == "true"
				validSignature = &tValue
			}
			ial := float64(response.Ial)
			aal := float64(response.Aal)
			newRow = IdPResponse{
				IdpID:          response.IdpId,
				Ial:            &ial,
				Aal:            &aal,
				Status:         &response.Status,
				Signature:      &response.Signature,
				ValidIal:       validIal,
				ValidSignature: validSignature,
			}
		} else {
			newRow = IdPResponse{
				IdpID:     response.IdpId,
				ErrorCode: &response.ErrorCode,
			}
		}
		result.Responses = append(result.Responses, newRow)
	}
	result.IsClosed = request.Closed
	result.IsTimedOut = request.TimedOut
	result.Mode = request.Mode

	// Set purpose
	result.Purpose = request.Purpose

	// make nil to array len 0
	if result.IdPIDList == nil {
		result.IdPIDList = make([]string, 0)
	}

	if request.RequestType != "" {
		result.RequestType = &request.RequestType
	}

	// Set requester_node_id
	result.RequesterNodeID = request.Owner

	// Set creation_block_height
	result.CreationBlockHeight = request.CreationBlockHeight

	// Set creation_chain_id
	result.CreationChainID = request.ChainId

	resultJSON, err := json.Marshal(result)
	if err != nil {
		value = []byte("")
		return app.ReturnQuery(value, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

type GetDataSignatureParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
	RequestID string `json:"request_id"`
}

type GetDataSignatureResult struct {
	Signature string `json:"signature"`
}

func (app *ABCIApplication) getDataSignature(param string) types.ResponseQuery {
	app.logger.Infof("GetDataSignature, Parameter: %s", param)
	var funcParam GetDataSignatureParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	signDataKey := dataSignatureKeyPrefix + keySeparator + funcParam.NodeID + keySeparator + funcParam.ServiceID + keySeparator + funcParam.RequestID
	signDataValue, err := app.state.Get([]byte(signDataKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if signDataValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var result GetDataSignatureResult
	result.Signature = string(signDataValue)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

type GetAllowedModeListParam struct {
	Purpose string `json:"purpose"`
}

type GetAllowedModeListResult struct {
	AllowedModeList []int32 `json:"allowed_mode_list"`
}

func (app *ABCIApplication) GetAllowedModeList(param string) types.ResponseQuery {
	app.logger.Infof("GetAllowedModeList, Parameter: %s", param)
	var funcParam GetAllowedModeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAllowedModeListResult
	result.AllowedModeList = app.GetAllowedModeFromStateDB(funcParam.Purpose, true)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetAllowedModeFromStateDB(purpose string, committedState bool) (result []int32) {
	allowedModeKey := "AllowedModeList" + keySeparator + purpose
	var allowedModeList data.AllowedModeList
	allowedModeValue, err := app.state.Get([]byte(allowedModeKey), committedState)
	if err != nil {
		return nil
	}
	if allowedModeValue == nil {
		// return default value
		if !modeFunctionMap[purpose] {
			result = append(result, 1)
		}
		result = append(result, 2)
		result = append(result, 3)
		return result
	}
	err = proto.Unmarshal(allowedModeValue, &allowedModeList)
	if err != nil {
		return result
	}
	result = allowedModeList.Mode
	return result
}
