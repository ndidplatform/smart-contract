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

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v8/abci/code"
	"github.com/ndidplatform/smart-contract/v8/abci/utils"
	data "github.com/ndidplatform/smart-contract/v8/protos/data"
)

type CreateAsResponseParam struct {
	ServiceID string `json:"service_id"`
	RequestID string `json:"request_id"`
	Signature string `json:"signature"`
	ErrorCode *int32 `json:"error_code"`
}

func (app *ABCIApplication) createAsResponse(param []byte, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateAsResponse, Parameter: %s", param)
	var createAsResponseParam CreateAsResponseParam
	err := json.Unmarshal(param, &createAsResponseParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	requestKey := requestKeyPrefix + keySeparator + createAsResponseParam.RequestID
	requestJSON, err := app.state.GetVersioned([]byte(requestKey), 0, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if requestJSON == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestJSON), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check error code exists
	if createAsResponseParam.ErrorCode != nil {
		errorCodeKey := errorCodeKeyPrefix + keySeparator + "as" + keySeparator + fmt.Sprintf("%d", *createAsResponseParam.ErrorCode)
		hasErrorCodeKey, err := app.state.Has([]byte(errorCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if !hasErrorCodeKey {
			return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode does not exist", "")
		}
	}

	// Check closed request
	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Request is closed", "")
	}

	// Check timed out request
	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Request is timed out", "")
	}

	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + createAsResponseParam.ServiceID
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

	// Check service is active
	if !service.Active {
		return app.ReturnDeliverTxLog(code.ServiceIsNotActive, "Service is not active", "")
	}

	// Check service destination is approved by NDID
	approveServiceKey := approvedServiceKeyPrefix + keySeparator + createAsResponseParam.ServiceID + keySeparator + nodeID
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
	if !approveService.Active {
		return app.ReturnDeliverTxLog(code.ServiceDestinationIsNotActive, "Service destination is not approved by NDID", "")
	}

	// Check service destination is active
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + createAsResponseParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}

	if serviceDestinationValue == nil {
		return app.ReturnDeliverTxLog(code.ServiceDestinationNotFound, "Service destination not found", "")
	}

	var nodes data.ServiceDesList
	err = proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	for index := range nodes.Node {
		if nodes.Node[index].NodeId == nodeID {
			if !nodes.Node[index].Active {
				return app.ReturnDeliverTxLog(code.ServiceDestinationIsNotActive, "Service destination is not active", "")
			}
			break
		}
	}

	// Check nodeID is exist in as_id_list
	exist := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == createAsResponseParam.ServiceID {
			for _, as := range dataRequest.AsIdList {
				if as == nodeID {
					exist = true
					break
				}
			}
		}
	}
	if exist == false {
		return app.ReturnDeliverTxLog(code.NodeIDDoesNotExistInASList, "Node ID does not exist in AS list", "")
	}

	// Check Duplicate AS ID
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == createAsResponseParam.ServiceID {
			for _, asResponse := range dataRequest.ResponseList {
				if asResponse.AsId == nodeID {
					return app.ReturnDeliverTxLog(code.DuplicateASResponse, "Duplicate AS response", "")
				}
			}
		}
	}

	// Check min_as
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == createAsResponseParam.ServiceID {
			if dataRequest.MinAs > 0 {
				var nonErrorResponseCount int64 = 0
				for _, asResponse := range dataRequest.ResponseList {
					if asResponse.ErrorCode == 0 {
						nonErrorResponseCount++
					}
				}
				if nonErrorResponseCount >= dataRequest.MinAs {
					return app.ReturnDeliverTxLog(code.DataRequestIsCompleted, "Can't create AS response to a request with enough AS responses", "")
				}
				var remainingPossibleResponseCount int64 = int64(len(dataRequest.AsIdList)) - int64(len(dataRequest.ResponseList))
				if nonErrorResponseCount+remainingPossibleResponseCount < dataRequest.MinAs {
					return app.ReturnDeliverTxLog(code.DataRequestCannotBeFulfilled, "Can't create AS response to a data request that cannot be fulfilled", "")
				}
			}
		}
	}

	var signDataKey string
	var signDataValue string
	if createAsResponseParam.ErrorCode == nil {
		signDataKey = dataSignatureKeyPrefix + keySeparator + nodeID + keySeparator + createAsResponseParam.ServiceID + keySeparator + createAsResponseParam.RequestID
		signDataValue = createAsResponseParam.Signature
	}

	// Update answered_as_id_list in request
	for index, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == createAsResponseParam.ServiceID {
			var asResponse data.ASResponse
			if createAsResponseParam.ErrorCode == nil {
				asResponse = data.ASResponse{
					AsId:         nodeID,
					Signed:       true,
					ReceivedData: false,
				}
			} else {
				asResponse = data.ASResponse{
					AsId:      nodeID,
					ErrorCode: *createAsResponseParam.ErrorCode,
				}
			}
			request.DataRequestList[index].ResponseList = append(dataRequest.ResponseList, &asResponse)
		}
	}

	requestJSON, err = utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	err = app.state.SetVersioned([]byte(requestKey), []byte(requestJSON))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if createAsResponseParam.ErrorCode == nil {
		app.state.Set([]byte(signDataKey), []byte(signDataValue))
	}
	return app.ReturnDeliverTxLog(code.OK, "success", createAsResponseParam.RequestID)
}
