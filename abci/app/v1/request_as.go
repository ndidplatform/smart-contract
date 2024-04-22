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

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type CreateAsResponseParam struct {
	ServiceID string `json:"service_id"`
	RequestID string `json:"request_id"`
	Signature string `json:"signature"`
	ErrorCode *int32 `json:"error_code"`
}

func (app *ABCIApplication) validateCreateAsResponse(funcParam CreateAsResponseParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isASNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallASMethod,
			Message: "This node does not have permission to call AS method",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	requestKey := requestKeyPrefix + keySeparator + funcParam.RequestID
	requestValue, err := app.state.GetVersioned([]byte(requestKey), 0, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if requestValue == nil {
		return &ApplicationError{
			Code:    code.RequestIDNotFound,
			Message: "Request ID not found",
		}
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestValue), &request)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	// Check closed request
	if request.Closed {
		return &ApplicationError{
			Code:    code.RequestIsClosed,
			Message: "Request is closed",
		}
	}

	// Check timed out request
	if request.TimedOut {
		return &ApplicationError{
			Code:    code.RequestIsTimedOut,
			Message: "Request is timed out",
		}
	}

	// Check nodeID is exist in as_id_list
	exist := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == funcParam.ServiceID {
			for _, as := range dataRequest.AsIdList {
				if as == callerNodeID {
					exist = true
					break
				}
			}
		}
	}
	if !exist {
		return &ApplicationError{
			Code:    code.NodeIDDoesNotExistInASList,
			Message: "Node ID does not exist in AS list",
		}
	}

	// Check Duplicate AS ID
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == funcParam.ServiceID {
			for _, asResponse := range dataRequest.ResponseList {
				if asResponse.AsId == callerNodeID {
					return &ApplicationError{
						Code:    code.DuplicateASResponse,
						Message: "Duplicate AS response",
					}
				}
			}
		}
	}

	// Check min_as
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == funcParam.ServiceID {
			if dataRequest.MinAs > 0 {
				var nonErrorResponseCount int64 = 0
				for _, asResponse := range dataRequest.ResponseList {
					if asResponse.ErrorCode == 0 {
						nonErrorResponseCount++
					}
				}
				if nonErrorResponseCount >= dataRequest.MinAs {
					return &ApplicationError{
						Code:    code.DataRequestIsCompleted,
						Message: "Can't create AS response to a request with enough AS responses",
					}
				}
				var remainingPossibleResponseCount int64 = int64(len(dataRequest.AsIdList)) - int64(len(dataRequest.ResponseList))
				if nonErrorResponseCount+remainingPossibleResponseCount < dataRequest.MinAs {
					return &ApplicationError{
						Code:    code.DataRequestCannotBeFulfilled,
						Message: "Can't create AS response to a data request that cannot be fulfilled",
					}
				}
			}
		}
	}

	// Check error code exists
	if funcParam.ErrorCode != nil {
		errorCodeKey := errorCodeKeyPrefix + keySeparator + "as" + keySeparator + fmt.Sprintf("%d", *funcParam.ErrorCode)
		hasErrorCodeKey, err := app.state.Has([]byte(errorCodeKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if !hasErrorCodeKey {
			return &ApplicationError{
				Code:    code.InvalidErrorCode,
				Message: "ErrorCode does not exist",
			}
		}
	}

	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceValue == nil {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	// Check service is active
	if !service.Active {
		return &ApplicationError{
			Code:    code.ServiceIsNotActive,
			Message: "Service is not active",
		}
	}

	// Check service destination is approved by NDID
	approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + callerNodeID
	approveServiceValue, err := app.state.Get([]byte(approveServiceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if approveServiceValue == nil {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}
	var approveService data.ApproveService
	err = proto.Unmarshal([]byte(approveServiceValue), &approveService)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if !approveService.Active {
		return &ApplicationError{
			Code:    code.ServiceDestinationIsNotActive,
			Message: "Service destination is not approved by NDID",
		}
	}

	// Check service destination is active
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}

	if serviceDestinationValue == nil {
		return &ApplicationError{
			Code:    code.ServiceDestinationNotFound,
			Message: "Service destination not found",
		}
	}

	var nodes data.ServiceDesList
	err = proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	for index := range nodes.Node {
		if nodes.Node[index].NodeId == callerNodeID {
			if !nodes.Node[index].Active {
				return &ApplicationError{
					Code:    code.ServiceDestinationIsNotActive,
					Message: "Service destination is not active",
				}
			}
			break
		}
	}

	return nil
}

func (app *ABCIApplication) createAsResponseCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam CreateAsResponseParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateCreateAsResponse(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) createAsResponse(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("CreateAsResponse, Parameter: %s", param)
	var funcParam CreateAsResponseParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateCreateAsResponse(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	requestKey := requestKeyPrefix + keySeparator + funcParam.RequestID
	requestValue, err := app.state.GetVersioned([]byte(requestKey), 0, false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestValue), &request)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	var signDataKey string
	var signDataValue string
	if funcParam.ErrorCode == nil {
		signDataKey = dataSignatureKeyPrefix + keySeparator + callerNodeID + keySeparator + funcParam.ServiceID + keySeparator + funcParam.RequestID
		signDataValue = funcParam.Signature
	}

	// Update answered_as_id_list in request
	for index, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceId == funcParam.ServiceID {
			var asResponse data.ASResponse
			if funcParam.ErrorCode == nil {
				asResponse = data.ASResponse{
					AsId:         callerNodeID,
					Signed:       true,
					ReceivedData: false,
				}
			} else {
				asResponse = data.ASResponse{
					AsId:      callerNodeID,
					ErrorCode: *funcParam.ErrorCode,
				}
			}
			request.DataRequestList[index].ResponseList = append(dataRequest.ResponseList, &asResponse)
		}
	}

	requestValue, err = utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	err = app.state.SetVersioned([]byte(requestKey), []byte(requestValue))
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if funcParam.ErrorCode == nil {
		app.state.Set([]byte(signDataKey), []byte(signDataValue))
	}

	return app.NewExecTxResult(code.OK, "success", funcParam.RequestID)
}
