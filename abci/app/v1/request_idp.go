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

type CreateIdpResponseParam struct {
	Aal       float64 `json:"aal"`
	Ial       float64 `json:"ial"`
	RequestID string  `json:"request_id"`
	Signature string  `json:"signature"`
	Status    string  `json:"status"`
	ErrorCode *int32  `json:"error_code"`
}

func (app *ABCIApplication) createIdpResponse(param []byte, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateIdpResponse, Parameter: %s", param)
	var funcParam CreateIdpResponseParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// get request
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

	// Check min_idp
	var nonErrorResponseCount int64 = 0
	for _, response := range request.ResponseList {
		if response.Status != "" {
			nonErrorResponseCount++
		}
	}
	if nonErrorResponseCount >= request.MinIdp {
		return app.ReturnDeliverTxLog(code.RequestIsCompleted, "Can't response to a request that is completed", "")
	}
	var remainingPossibleResponseCount int64 = int64(len(request.IdpIdList)) - int64(len(request.ResponseList))
	if nonErrorResponseCount+remainingPossibleResponseCount < request.MinIdp {
		return app.ReturnDeliverTxLog(code.RequestCannotBeFulfilled, "Can't response to a request that cannot be fulfilled", "")
	}

	response := data.Response{
		IdpId: nodeID,
	}

	// Check closed request
	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Can't response a request that's closed", "")
	}
	// Check timed out request
	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Can't response a request that's timed out", "")
	}

	if funcParam.ErrorCode == nil {
		response.Ial = funcParam.Ial
		response.Aal = funcParam.Aal
		response.Status = funcParam.Status
		response.Signature = funcParam.Signature

		// Check AAL
		if request.MinAal > response.Aal {
			return app.ReturnDeliverTxLog(code.AALError, "Response's AAL is less than min AAL", "")
		}
		// Check IAL
		if request.MinIal > response.Ial {
			return app.ReturnDeliverTxLog(code.IALError, "Response's IAL is less than min IAL", "")
		}
		// Check AAL, IAL with MaxIalAal
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
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
		if response.Aal > nodeDetail.MaxAal {
			return app.ReturnDeliverTxLog(code.AALError, "Response's AAL is greater than max AAL", "")
		}
		if response.Ial > nodeDetail.MaxIal {
			return app.ReturnDeliverTxLog(code.IALError, "Response's IAL is greater than max IAL", "")
		}
	} else {
		// Check error code exists
		errorCodeKey := errorCodeKeyPrefix + keySeparator + "idp" + keySeparator + fmt.Sprintf("%d", *funcParam.ErrorCode)
		hasErrorCodeKey, err := app.state.Has([]byte(errorCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if !hasErrorCodeKey {
			return app.ReturnDeliverTxLog(code.InvalidErrorCode, "ErrorCode does not exist", "")
		}
		response.ErrorCode = *funcParam.ErrorCode
	}

	// Check nodeID is exist in idp_id_list
	exist := false
	for _, idpID := range request.IdpIdList {
		if idpID == nodeID {
			exist = true
			break
		}
	}
	if exist == false {
		return app.ReturnDeliverTxLog(code.NodeIDDoesNotExistInIdPList, "Node ID does not exist in IdP list", "")
	}

	// Check duplicate response from the same IdP
	for _, oldResponse := range request.ResponseList {
		if oldResponse.IdpId == nodeID {
			return app.ReturnDeliverTxLog(code.DuplicateIdPResponse, "Duplicate IdP response", "")
		}
	}

	request.ResponseList = append(request.ResponseList, &response)
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
