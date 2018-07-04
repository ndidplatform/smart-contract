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
	"github.com/tendermint/tendermint/abci/types"
)

func createRequest(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateRequest, Parameter: %s", param)
	var request Request
	err := json.Unmarshal([]byte(param), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// set default value
	request.IsClosed = false
	request.IsTimedOut = false
	request.CanAddAccessor = false

	// set Owner
	request.Owner = nodeID

	// set Can add accossor
	ownerRole := getRoleFromNodeID(nodeID, app)
	if string(ownerRole) == "IdP" || string(ownerRole) == "MasterIdP" {
		request.CanAddAccessor = true
	}

	// set default value
	request.Responses = make([]Response, 0)
	for index := range request.DataRequestList {
		if request.DataRequestList[index].As == nil {
			request.DataRequestList[index].As = make([]string, 0)
		}
		request.DataRequestList[index].AnsweredAsIdList = make([]string, 0)
		request.DataRequestList[index].ReceivedDataFromList = make([]string, 0)
	}

	// check duplicate service ID in Data Request
	serviceIDCount := make(map[string]int)
	for _, dataRequest := range request.DataRequestList {
		serviceIDCount[dataRequest.ServiceID]++
	}
	for _, count := range serviceIDCount {
		if count > 1 {
			return ReturnDeliverTxLog(code.DuplicateServiceIDInDataRequest, "Duplicate Service ID In Data Request", "")
		}
	}

	key := "Request" + "|" + request.RequestID

	value, err := json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	_, existValue := app.state.db.Get(prefixKey([]byte(key)))
	if existValue != nil {
		return ReturnDeliverTxLog(code.DuplicateRequestID, "Duplicate Request ID", "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", request.RequestID)
}

func closeRequest(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CloseRequest, Parameter: %s", param)
	var funcParam CloseRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "Request" + "|" + funcParam.RequestID
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.RequestIsTimedOut, "Can not close a timed out request", "")
	}

	for _, valid := range funcParam.ResponseValidList {
		for index := range request.Responses {
			if valid.IdpID == request.Responses[index].IdpID {
				request.Responses[index].ValidProof = &valid.ValidProof
				request.Responses[index].ValidIal = &valid.ValidIal
			}
		}
	}

	request.IsClosed = true
	value, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

func timeOutRequest(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("TimeOutRequest, Parameter: %s", param)
	var funcParam TimeOutRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "Request" + "|" + funcParam.RequestID
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	if request.IsClosed {
		return ReturnDeliverTxLog(code.RequestIsClosed, "Can not set time out a closed request", "")
	}

	for _, valid := range funcParam.ResponseValidList {
		for index := range request.Responses {
			if valid.IdpID == request.Responses[index].IdpID {
				request.Responses[index].ValidProof = &valid.ValidProof
				request.Responses[index].ValidIal = &valid.ValidIal
			}
		}
	}

	request.IsTimedOut = true
	value, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

func setDataReceived(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetDataReceived, Parameter: %s", param)
	var funcParam SetDataReceivedParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "Request" + "|" + funcParam.RequestID
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check as_id is exist in as_id_list
	exist := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == funcParam.ServiceID {
			for _, as := range dataRequest.AnsweredAsIdList {
				if as == funcParam.AsID {
					exist = true
					break
				}
			}
		}
	}
	if exist == false {
		return ReturnDeliverTxLog(code.AsIDIsNotExistInASList, "AS ID is not exist in answered AS list", "")
	}

	// Check Duplicate AS ID
	duplicate := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == funcParam.ServiceID {
			for _, as := range dataRequest.ReceivedDataFromList {
				if as == funcParam.AsID {
					duplicate = true
					break
				}
			}
		}
	}
	if duplicate == true {
		return ReturnDeliverTxLog(code.DuplicateASInDataRequest, "Duplicate AS ID in data request", "")
	}

	// Update received_data_from_list in request
	for index, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == funcParam.ServiceID {
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

	value, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}
