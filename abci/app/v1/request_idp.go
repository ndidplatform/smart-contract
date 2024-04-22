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

type CreateIdpResponseParam struct {
	Aal       float64 `json:"aal"`
	Ial       float64 `json:"ial"`
	RequestID string  `json:"request_id"`
	Signature string  `json:"signature"`
	Status    string  `json:"status"`
	ErrorCode *int32  `json:"error_code"`
}

func (app *ABCIApplication) validateCreateIdpResponse(funcParam CreateIdpResponseParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + callerNodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if nodeDetailValue == nil {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	ok := app.isIDPorIDPAgentNode(&nodeDetail)
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP or IdP agent method",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	// get request
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
			Message: "Can't response a request that's closed",
		}
	}
	// Check timed out request
	if request.TimedOut {
		return &ApplicationError{
			Code:    code.RequestIsTimedOut,
			Message: "Can't response a request that's timed out",
		}
	}

	// Check nodeID exists in idp_id_list
	exist := false
	for _, idpID := range request.IdpIdList {
		if idpID == callerNodeID {
			exist = true
			break
		}
	}
	if !exist {
		return &ApplicationError{
			Code:    code.NodeIDDoesNotExistInIdPList,
			Message: "Node ID does not exist in IdP list",
		}
	}

	// Check duplicate response from the same IdP
	for _, currentResponse := range request.ResponseList {
		if currentResponse.IdpId == callerNodeID {
			return &ApplicationError{
				Code:    code.DuplicateIdPResponse,
				Message: "Duplicate IdP response",
			}
		}
	}

	// Check min_idp
	var nonErrorResponseCount int64 = 0
	for _, response := range request.ResponseList {
		if response.Status != "" {
			nonErrorResponseCount++
		}
	}
	if nonErrorResponseCount >= request.MinIdp {
		return &ApplicationError{
			Code:    code.RequestIsCompleted,
			Message: "Can't response to a request that is completed",
		}
	}
	var remainingPossibleResponseCount int64 = int64(len(request.IdpIdList)) - int64(len(request.ResponseList))
	if nonErrorResponseCount+remainingPossibleResponseCount < request.MinIdp {
		return &ApplicationError{
			Code:    code.RequestCannotBeFulfilled,
			Message: "Can't response to a request that cannot be fulfilled",
		}
	}

	if funcParam.ErrorCode == nil {
		// Check AAL
		if request.MinAal > funcParam.Aal {
			return &ApplicationError{
				Code:    code.AALError,
				Message: "Response's AAL is less than min AAL",
			}
		}
		// Check IAL
		if request.MinIal > funcParam.Ial {
			return &ApplicationError{
				Code:    code.IALError,
				Message: "Response's IAL is less than min IAL",
			}
		}

		// Check IAL and AAL with response node's MaxIal and MaxAal
		if funcParam.Aal > nodeDetail.MaxAal {
			return &ApplicationError{
				Code:    code.AALError,
				Message: "Response's AAL is greater than max AAL",
			}
		}
		if funcParam.Ial > nodeDetail.MaxIal {
			return &ApplicationError{
				Code:    code.IALError,
				Message: "Response's IAL is greater than max IAL",
			}
		}
	} else {
		// Check error code exists
		errorCodeKey := errorCodeKeyPrefix + keySeparator + "idp" + keySeparator + fmt.Sprintf("%d", *funcParam.ErrorCode)
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

	return nil
}

func (app *ABCIApplication) createIdpResponseCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam CreateIdpResponseParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateCreateIdpResponse(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) createIdpResponse(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("CreateIdpResponse, Parameter: %s", param)
	var funcParam CreateIdpResponseParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateCreateIdpResponse(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// get request
	requestKey := requestKeyPrefix + keySeparator + funcParam.RequestID
	value, err := app.state.GetVersioned([]byte(requestKey), 0, false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	response := data.Response{
		IdpId: callerNodeID,
	}

	if funcParam.ErrorCode == nil {
		// normal response
		response.Ial = funcParam.Ial
		response.Aal = funcParam.Aal
		response.Status = funcParam.Status
		response.Signature = funcParam.Signature
	} else {
		// error response
		response.ErrorCode = *funcParam.ErrorCode
	}

	request.ResponseList = append(request.ResponseList, &response)

	value, err = utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	err = app.state.SetVersioned([]byte(requestKey), []byte(value))
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	return app.NewExecTxResult(code.OK, "success", funcParam.RequestID)
}
