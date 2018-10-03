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

	"github.com/gogo/protobuf/proto"
	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/ndidplatform/smart-contract/protos/data"
	"github.com/tendermint/tendermint/abci/types"
)

func (app *DIDApplication) createRequest(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateRequest, Parameter: %s", param)
	var funcParam Request
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	var request data.Request

	// set request data
	request.RequestId = funcParam.RequestID
	request.MinIdp = int64(funcParam.MinIdp)
	request.MinAal = funcParam.MinAal
	request.MinIal = funcParam.MinIal
	request.RequestTimeout = int64(funcParam.Timeout)
	// request.DataRequestList = funcParam.DataRequestList
	request.RequestMessageHash = funcParam.MessageHash
	request.Mode = int64(funcParam.Mode)
	request.IdpIdList = funcParam.IdPIDList

	// set data request
	request.DataRequestList = make([]*data.DataRequest, 0)
	for index := range funcParam.DataRequestList {
		var newRow data.DataRequest
		newRow.ServiceId = funcParam.DataRequestList[index].ServiceID
		newRow.RequestParamsHash = funcParam.DataRequestList[index].RequestParamsHash
		newRow.MinAs = int64(funcParam.DataRequestList[index].Count)
		newRow.AsIdList = funcParam.DataRequestList[index].As
		if funcParam.DataRequestList[index].As == nil {
			newRow.AsIdList = make([]string, 0)
		}
		newRow.AnsweredAsIdList = make([]string, 0)
		newRow.ReceivedDataFromList = make([]string, 0)
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

	// check duplicate service ID in Data Request
	serviceIDCount := make(map[string]int)
	for _, dataRequest := range request.DataRequestList {
		serviceIDCount[dataRequest.ServiceId]++
	}
	for _, count := range serviceIDCount {
		if count > 1 {
			return app.ReturnDeliverTxLog(code.DuplicateServiceIDInDataRequest, "Duplicate Service ID In Data Request", "")
		}
	}

	// set creation_block_height
	request.CreationBlockHeight = app.CurrentBlock

	key := "Request" + "|" + request.RequestId

	value, err := proto.Marshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	_, existValue := app.state.db.Get(prefixKey([]byte(key)))
	if existValue != nil {
		return app.ReturnDeliverTxLog(code.DuplicateRequestID, "Duplicate Request ID", "")
	}
	app.SetStateDB([]byte(key), []byte(value))
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
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Can not set time out a closed request", "")
	}

	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Can not close a timed out request", "")
	}

	// // Check valid list
	// if len(funcParam.ResponseValidList) != len(request.Responses) {
	// 	return app.ReturnDeliverTxLog(code.IncompleteValidList, "Incomplete valid list", "")
	// }

	for _, valid := range funcParam.ResponseValidList {
		for index := range request.ResponseList {
			if valid.IdpID == request.ResponseList[index].IdpId {
				if valid.ValidProof != nil {
					if *valid.ValidProof {
						request.ResponseList[index].ValidProof = "true"
					} else {
						request.ResponseList[index].ValidProof = "false"
					}
				}
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
	value, err = proto.Marshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
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
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Can not close a timed out request", "")
	}

	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Can not set time out a closed request", "")
	}

	// // Check valid list
	// if len(funcParam.ResponseValidList) != len(request.Responses) {
	// 	return app.ReturnDeliverTxLog(code.IncompleteValidList, "Incomplete valid list", "")
	// }

	for _, valid := range funcParam.ResponseValidList {
		for index := range request.ResponseList {
			if valid.IdpID == request.ResponseList[index].IdpId {
				if valid.ValidProof != nil {
					if *valid.ValidProof {
						request.ResponseList[index].ValidProof = "true"
					} else {
						request.ResponseList[index].ValidProof = "false"
					}
				}
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
	value, err = proto.Marshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(key), []byte(value))
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
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check as_id is exist in as_id_list
	exist := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == funcParam.ServiceID {
			for _, as := range dataRequest.AnsweredAsIdList {
				if as == funcParam.AsID {
					exist = true
					break
				}
			}
		}
	}
	if exist == false {
		return app.ReturnDeliverTxLog(code.AsIDIsNotExistInASList, "AS ID is not exist in answered AS list", "")
	}

	// Check Duplicate AS ID
	duplicate := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == funcParam.ServiceID {
			for _, as := range dataRequest.ReceivedDataFromList {
				if as == funcParam.AsID {
					duplicate = true
					break
				}
			}
		}
	}
	if duplicate == true {
		return app.ReturnDeliverTxLog(code.DuplicateASInDataRequest, "Duplicate AS ID in data request", "")
	}

	// Update received_data_from_list in request
	for index, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == funcParam.ServiceID {
			request.DataRequestList[index].ReceivedDataFromList = append(dataRequest.ReceivedDataFromList, funcParam.AsID)
		}
	}

	// Request has data request. If received data, signed answer > data request count on each data request
	// dataRequestCompletedCount := 0
	// for _, dataRequest := range request.DataRequestList {
	// 	if len(dataRequest.AnsweredAsIdList) >= dataRequest.Count &&
	// 		len(dataRequest.ReceivedDataFromList) >= dataRequest.Count {
	// 		dataRequestCompletedCount++
	// 	}
	// }
	// if dataRequestCompletedCount == len(request.DataRequestList) {
	// 	app.logger.Info("Auto close")
	// 	request.IsClosed = true
	// } else {
	// 	app.logger.Info("Auto close")
	// }

	value, err = proto.Marshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}
