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
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/tendermint/tendermint/abci/types"
	kv "github.com/tendermint/tendermint/libs/kv"

	"github.com/ndidplatform/smart-contract/v5/abci/code"
	"github.com/ndidplatform/smart-contract/v5/abci/utils"
	"github.com/ndidplatform/smart-contract/v5/protos/data"
)

func (app *ABCIApplication) AddAccessor(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddAccessor, Parameter: %s", param)
	var funcParam AddAccessorParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		return app.ReturnDeliverTxLog(code.GotRefGroupCodeAndIdentity, "Found reference group code and identity detail in parameter", "")
	}
	// Check duplicate accessor ID
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupCodeFromDB != nil {
		return app.ReturnDeliverTxLog(code.DuplicateAccessorID, "Duplicate accessor ID", "")
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	foundThisNodeID := false
	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			foundThisNodeID = true
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			break
		}
	}
	if foundThisNodeID == false {
		return app.ReturnDeliverTxLog(code.IdentityNotFoundInThisIdP, "Identity not found in this IdP", "")
	}

	if mode3 {
		minIdp := 1
		checkRequestResult := app.checkRequest(funcParam.RequestID, "AddAccessor", minIdp)
		if checkRequestResult.Code != code.OK {
			return checkRequestResult
		}
	}

	var accessor data.Accessor
	accessor.AccessorId = funcParam.AccessorID
	accessor.AccessorType = funcParam.AccessorType
	accessor.AccessorPublicKey = funcParam.AccessorPublicKey
	accessor.Active = true
	accessor.Owner = nodeID
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			idp.Accessors = append(idp.Accessors, &accessor)
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	if mode3 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}

	accessorToRefCodeKey = accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	accessorToRefCodeValue := refGroupCode
	app.state.Set([]byte(accessorToRefCodeKey), []byte(accessorToRefCodeValue))
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	var attributes []kv.Pair
	var attribute kv.Pair
	attribute.Key = []byte("reference_group_code")
	attribute.Value = []byte(refGroupCode)
	attributes = append(attributes, attribute)
	return app.ReturnDeliverTxLogWithAttributes(code.OK, "success", attributes)
}

func (app *ABCIApplication) registerIdentity(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterIdentity, Parameter: %s", param)
	var funcParam RegisterIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
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
	// Valid Mode
	var validMode = map[int32]bool{}
	allowedMode := app.GetAllowedModeFromStateDB("RegisterIdentity", false)
	for _, mode := range allowedMode {
		validMode[mode] = true
	}
	user := funcParam
	// Validate user's ial is <= node's max_ial
	if user.Ial > nodeDetail.MaxIal {
		return app.ReturnDeliverTxLog(code.IALError, "IAL must be less than or equals to registered node's MAX IAL", "")
	}
	// Check for identity_namespace and identity_identifier_hash. If exist, error.
	if user.ReferenceGroupCode == "" {
		return app.ReturnDeliverTxLog(code.RefGroupCodeCannotBeEmpty, "Please input reference group code", "")
	}
	// Check accessor
	if user.AccessorID == "" {
		return app.ReturnDeliverTxLog(code.AccessorIDCannotBeEmpty, "Please input accessor ID", "")
	}
	if user.AccessorPublicKey == "" {
		return app.ReturnDeliverTxLog(code.AccessorPublicKeyCannotBeEmpty, "Please input accessor public key", "")
	}
	if user.AccessorType == "" {
		return app.ReturnDeliverTxLog(code.AccessorTypeCannotBeEmpty, "Please input accessor type", "")
	}
	var modeCount = map[int32]int{}
	for _, mode := range allowedMode {
		modeCount[mode] = 0
	}
	for _, mode := range user.ModeList {
		if validMode[mode] {
			modeCount[mode] = modeCount[mode] + 1
		} else {
			return app.ReturnDeliverTxLog(code.InvalidMode, "Must be register identity on valid mode", "")
		}
	}
	user.ModeList = make([]int32, 0)
	for mode, count := range modeCount {
		if count > 0 {
			user.ModeList = append(user.ModeList, mode)
		}
	}
	sort.Slice(user.ModeList, func(i, j int) bool { return user.ModeList[i] < user.ModeList[j] })
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + user.ReferenceGroupCode
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var refGroup data.ReferenceGroup
	// If referenceGroupCode already existed, add new identity to group

	mode3 := false
	for _, mode := range user.ModeList {
		if mode == 3 {
			mode3 = true
			break
		}
	}
	minIdp := 0
	if refGroupValue != nil {
		err := proto.Unmarshal(refGroupValue, &refGroup)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// If have at least one node active
		for _, idp := range refGroup.Idps {
			nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp.NodeId
			nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
			if err != nil {
				return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
			}
			if nodeDetailValue == nil {
				return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
			}
			var nodeDetail data.NodeDetail
			err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
			if nodeDetail.Active && idp.Active {
				minIdp = 1
				break
			}
		}
	}
	if mode3 && minIdp > 0 {
		checkRequestResult := app.checkRequest(user.RequestID, "RegisterIdentity", minIdp)
		if checkRequestResult.Code != code.OK {
			return checkRequestResult
		}
	}
	// Check min_ial when RegisterIdentity when onboard as first IdP
	if refGroupValue == nil {
		if user.Ial < app.GetAllowedMinIalForRegisterIdentityAtFirstIdpFromStateDB(false) {
			return app.ReturnDeliverTxLog(code.IalMustBeGreaterOrEqualMinIal, "Ial must be greater or equal min ial when onboard as first IdP", "")
		}
	}
	// Check number of Identifier in new list and old list in stateDB
	var namespaceCount = map[string]int{}
	var checkDuplicateNamespaceAndHash = map[string]int{}
	validNamespace := app.GetNamespaceMap(false)
	for _, identity := range user.NewIdentityList {
		if identity.IdentityNamespace == "" || identity.IdentityIdentifierHash == "" {
			return app.ReturnDeliverTxLog(code.IdentityCannotBeEmpty, "Please input identity detail", "")
		}
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + identity.IdentityNamespace + keySeparator + identity.IdentityIdentifierHash
		identityToRefCodeValue, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if identityToRefCodeValue != nil {
			return app.ReturnDeliverTxLog(code.IdentityAlreadyExisted, "Identity already existed", "")
		}
		// check namespace is valid
		if !validNamespace[identity.IdentityNamespace] {
			return app.ReturnDeliverTxLog(code.InvalidNamespace, "Namespace is invalid", "")
		}
		namespaceCount[identity.IdentityNamespace] = namespaceCount[identity.IdentityNamespace] + 1
		checkDuplicateNamespaceAndHash[identity.IdentityNamespace+identity.IdentityIdentifierHash] = checkDuplicateNamespaceAndHash[identity.IdentityNamespace+identity.IdentityIdentifierHash] + 1
	}
	// Check duplicate count
	for _, count := range checkDuplicateNamespaceAndHash {
		if count > 1 {
			return app.ReturnDeliverTxLog(code.DuplicateIdentifier, "There are duplicate identifier", "")
		}
	}
	for _, identity := range refGroup.Identities {
		namespaceCount[identity.Namespace] = namespaceCount[identity.Namespace] + 1
	}
	allowedIdentifierCount := app.GetNamespaceAllowedIdentifierCountMap(false)
	for namespace, count := range namespaceCount {
		if count > allowedIdentifierCount[namespace] && allowedIdentifierCount[namespace] > 0 {
			return app.ReturnDeliverTxLog(code.IdentifierCountIsGreaterThanAllowedIdentifierCount, "Identifier count is greater than allowed identifier count", "")
		}
	}
	var accessor data.Accessor
	accessor.AccessorId = user.AccessorID
	accessor.AccessorType = user.AccessorType
	accessor.AccessorPublicKey = user.AccessorPublicKey
	accessor.Active = true
	accessor.Owner = nodeID
	var idp data.IdPInRefGroup
	idp.NodeId = nodeID
	idp.Mode = append(idp.Mode, user.ModeList...)
	idp.Accessors = append(idp.Accessors, &accessor)
	idp.Ial = user.Ial
	idp.Active = true
	for _, identity := range user.NewIdentityList {
		var newIdentity data.IdentityInRefGroup
		newIdentity.Namespace = identity.IdentityNamespace
		newIdentity.IdentifierHash = identity.IdentityIdentifierHash
		refGroup.Identities = append(refGroup.Identities, &newIdentity)
	}
	foundThisNodeID := false
	for iIdp, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			refGroup.Idps[iIdp].Active = true
			refGroup.Idps[iIdp].Mode = funcParam.ModeList
			foundAccessorInThisGroup := false
			for iAcc, accessor := range refGroup.Idps[iIdp].Accessors {
				if accessor.AccessorId == funcParam.AccessorID {
					refGroup.Idps[iIdp].Accessors[iAcc].AccessorType = user.AccessorType
					refGroup.Idps[iIdp].Accessors[iAcc].AccessorPublicKey = user.AccessorPublicKey
					refGroup.Idps[iIdp].Accessors[iAcc].Active = true
					foundAccessorInThisGroup = true
				}
			}
			if !foundAccessorInThisGroup {
				refGroup.Idps[iIdp].Accessors = append(refGroup.Idps[iIdp].Accessors, &accessor)
			}
			foundThisNodeID = true
			break
		}
	}
	if !foundThisNodeID {
		refGroup.Idps = append(refGroup.Idps, &idp)
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	if mode3 && minIdp > 0 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(user.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}

	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + user.AccessorID
	accessorToRefCodeValue := user.ReferenceGroupCode
	for _, identity := range user.NewIdentityList {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + identity.IdentityNamespace + keySeparator + identity.IdentityIdentifierHash
		identityToRefCodeValue := []byte(user.ReferenceGroupCode)
		app.state.Set([]byte(identityToRefCodeKey), []byte(identityToRefCodeValue))
	}
	app.state.Set([]byte(accessorToRefCodeKey), []byte(accessorToRefCodeValue))
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	var attributes []kv.Pair
	var attribute kv.Pair
	attribute.Key = []byte("reference_group_code")
	attribute.Value = []byte(user.ReferenceGroupCode)
	attributes = append(attributes, attribute)
	return app.ReturnDeliverTxLogWithAttributes(code.OK, "success", attributes)
}

func (app *ABCIApplication) checkRequest(requestID string, purpose string, minIdp int) types.ResponseDeliverTx {
	requestKey := requestKeyPrefix + keySeparator + requestID
	requestValue, err := app.state.GetVersioned([]byte(requestKey), app.state.Height, true)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if requestValue == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestValue), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	if request.Purpose != purpose {
		return app.ReturnDeliverTxLog(code.InvalidPurpose, "Request has a invalid purpose", "")
	}
	if request.UseCount > 0 {
		return app.ReturnDeliverTxLog(code.RequestIsAlreadyUsed, "Request is already used", "")
	}
	if !request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsNotClosed, "Request is not closed", "")
	}
	var acceptCount int
	acceptCount = 0
	for _, response := range request.ResponseList {
		if response.ValidIal != "true" {
			continue
		}
		if response.ValidSignature != "true" {
			continue
		}
		if response.Status == "accept" {
			acceptCount++
		}
	}
	if acceptCount >= minIdp {
		return app.ReturnDeliverTxLog(code.OK, "Request is completed", "")
	}
	return app.ReturnDeliverTxLog(code.RequestIsNotCompleted, "Request is not completed", "")
}

func (app *ABCIApplication) increaseRequestUseCount(requestID string) types.ResponseDeliverTx {
	requestKey := requestKeyPrefix + keySeparator + requestID
	requestValue, err := app.state.GetVersioned([]byte(requestKey), app.state.Height, true)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if requestValue == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestValue), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	request.UseCount = request.UseCount + 1
	requestProtobuf, err := utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	err = app.state.SetVersioned([]byte(requestKey), []byte(requestProtobuf))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) createIdpResponse(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateIdpResponse, Parameter: %s", param)
	var funcParam CreateIdpResponseParam
	err := json.Unmarshal([]byte(param), &funcParam)
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
		// Check IsClosed
		if request.Closed {
			return app.ReturnDeliverTxLog(code.RequestIsClosed, "Can't response a request that's closed", "")
		}
		// Check IsTimedOut
		if request.TimedOut {
			return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Can't response a request that's timed out", "")
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

	// Check duplicate before add
	chkDup := false
	for _, oldResponse := range request.ResponseList {
		if &response == oldResponse {
			chkDup = true
			break
		}
	}
	if chkDup == true {
		return app.ReturnDeliverTxLog(code.DuplicateResponse, "Duplicate Response", "")
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

func (app *ABCIApplication) updateIdentity(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateIdentity, Parameter: %s", param)
	var funcParam UpdateIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check IAL must less than Max IAL
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
	if funcParam.Ial > nodeDetail.MaxIal {
		return app.ReturnDeliverTxLog(code.IALError, "New IAL is greater than max IAL", "")
	}
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		return app.ReturnDeliverTxLog(code.GotRefGroupCodeAndIdentity, "Found reference group code and identity detail in parameter", "")
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	foundThisNodeID := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			foundThisNodeID = true
			break
		}
	}
	if foundThisNodeID == false {
		return app.ReturnDeliverTxLog(code.IdentityNotFoundInThisIdP, "Identity not found in this IdP", "")
	}
	for index, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			refGroup.Idps[index].Ial = funcParam.Ial
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	var attributes []kv.Pair
	var attribute kv.Pair
	attribute.Key = []byte("reference_group_code")
	attribute.Value = []byte(refGroupCode)
	attributes = append(attributes, attribute)
	return app.ReturnDeliverTxLogWithAttributes(code.OK, "success", attributes)
}

func (app *ABCIApplication) revokeIdentityAssociation(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RevokeIdentityAssociation, Parameter: %s", param)
	var funcParam RevokeIdentityAssociationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
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
	if !nodeDetail.Active {
		return app.ReturnDeliverTxLog(code.NodeIsNotActive, "Node is not active", "")
	}
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		return app.ReturnDeliverTxLog(code.GotRefGroupCodeAndIdentity, "Found reference group code and identity detail in parameter", "")
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	foundThisNodeID := false
	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			foundThisNodeID = true
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			break
		}
	}
	if foundThisNodeID == false {
		return app.ReturnDeliverTxLog(code.IdentityNotFoundInThisIdP, "Identity not found in this IdP", "")
	}
	if mode3 {
		minIdp := 1
		checkRequestResult := app.checkRequest(funcParam.RequestID, "RevokeIdentityAssociation", minIdp)
		if checkRequestResult.Code != code.OK {
			return checkRequestResult
		}
	}
	for iIdP, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			refGroup.Idps[iIdP].Active = false
			for iAcc := range idp.Accessors {
				refGroup.Idps[iIdP].Accessors[iAcc].Active = false
			}
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	if mode3 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	var attributes []kv.Pair
	var attribute kv.Pair
	attribute.Key = []byte("reference_group_code")
	attribute.Value = []byte(refGroupCode)
	attributes = append(attributes, attribute)
	return app.ReturnDeliverTxLogWithAttributes(code.OK, "success", attributes)
}

func (app *ABCIApplication) revokeAccessor(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RevokeAccessor, Parameter: %s", param)
	var funcParam RevokeAccessorParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// check node is active
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
	if !nodeDetail.Active {
		return app.ReturnDeliverTxLog(code.NodeIsNotActive, "Node is not active", "")
	}
	// check all accessor ID have the same ref group code
	firstRefGroup := ""
	for index, accsesorID := range funcParam.AccessorIDList {
		accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + accsesorID
		refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		if index == 0 {
			firstRefGroup = string(refGroupCodeFromDB)
		} else {
			if string(refGroupCodeFromDB) != firstRefGroup {
				return app.ReturnDeliverTxLog(code.AllAccessorMustHaveSameRefGroupCode, "All accessors must have same reference group code", "")
			}
		}
	}
	refGroupCode := firstRefGroup
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			accessorInIdP := make([]string, 0)
			activeAccessorCount := 0
			for _, accsesor := range idp.Accessors {
				accessorInIdP = append(accessorInIdP, accsesor.AccessorId)
				if accsesor.Active {
					activeAccessorCount++
				}
			}
			for _, accsesorID := range funcParam.AccessorIDList {
				if !contains(accsesorID, accessorInIdP) {
					return app.ReturnDeliverTxLog(code.AccessorNotFoundInThisIdP, "Accessor not found in this IdP", "")
				}
			}
			if activeAccessorCount-len(funcParam.AccessorIDList) < 1 {
				return app.ReturnDeliverTxLog(code.CannotRevokeAllAccessorsInThisIdP, "Cannot revoke all accessors in this IdP", "")
			}
		}
	}

	if mode3 {
		minIdp := 1
		checkRequestResult := app.checkRequest(funcParam.RequestID, "RevokeAccessor", minIdp)
		if checkRequestResult.Code != code.OK {
			return checkRequestResult
		}
	}

	for iIdP, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			for _, accsesorID := range funcParam.AccessorIDList {
				for iAcc, accsesor := range idp.Accessors {
					// app.logger.Debugf("Acces:%s", args)
					if accsesor.AccessorId == accsesorID {
						refGroup.Idps[iIdP].Accessors[iAcc].Active = false
						break
					}
				}
			}
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	if mode3 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}

	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	var attributes []kv.Pair
	var attribute kv.Pair
	attribute.Key = []byte("reference_group_code")
	attribute.Value = []byte(refGroupCode)
	attributes = append(attributes, attribute)
	return app.ReturnDeliverTxLogWithAttributes(code.OK, "success", attributes)
}

func (app *ABCIApplication) updateIdentityModeList(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateIdentityModeList, Parameter: %s", param)
	var funcParam UpdateIdentityModeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check IAL must less than Max IAL
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
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		return app.ReturnDeliverTxLog(code.GotRefGroupCodeAndIdentity, "Found reference group code and identity detail in parameter", "")
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	// Valid Mode
	var validMode = map[int32]bool{}
	allowedMode := app.GetAllowedModeFromStateDB("UpdateIdentityModeList", false)
	for _, mode := range allowedMode {
		validMode[mode] = true
	}
	var modeCount = map[int32]int{}
	for _, mode := range allowedMode {
		modeCount[mode] = 0
	}
	for _, mode := range funcParam.ModeList {
		if validMode[mode] {
			modeCount[mode] = modeCount[mode] + 1
		} else {
			return app.ReturnDeliverTxLog(code.InvalidMode, "Must be register identity on valid mode", "")
		}
	}
	funcParam.ModeList = make([]int32, 0)
	for mode, count := range modeCount {
		if count > 0 {
			funcParam.ModeList = append(funcParam.ModeList, mode)
		}
	}
	sort.Slice(funcParam.ModeList, func(i, j int) bool { return funcParam.ModeList[i] < funcParam.ModeList[j] })
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	foundThisNodeID := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			foundThisNodeID = true
			break
		}
	}
	if foundThisNodeID == false {
		return app.ReturnDeliverTxLog(code.IdentityNotFoundInThisIdP, "Identity not found in this IdP", "")
	}
	for index, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			// Check new mode list is higher than current mode list
			maxCurrentMode := MaxInt32(refGroup.Idps[index].Mode)
			maxNewMode := MaxInt32(funcParam.ModeList)
			if maxCurrentMode > maxNewMode {
				return app.ReturnDeliverTxLog(code.NewModeListMustBeHigherThanCurrentModeList, "New mode list must be higher than current mode list", "")
			}
			refGroup.Idps[index].Mode = funcParam.ModeList
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	var attributes []kv.Pair
	var attribute kv.Pair
	attribute.Key = []byte("reference_group_code")
	attribute.Value = []byte(refGroupCode)
	attributes = append(attributes, attribute)
	return app.ReturnDeliverTxLogWithAttributes(code.OK, "success", attributes)
}

func (app *ABCIApplication) addIdentity(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddIdentity, Parameter: %s", param)
	var funcParam AddIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
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
	user := funcParam
	// Check for identity_namespace and identity_identifier_hash. If exist, error.
	if user.ReferenceGroupCode == "" {
		return app.ReturnDeliverTxLog(code.RefGroupCodeCannotBeEmpty, "Please input reference group code", "")
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + user.ReferenceGroupCode
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var refGroup data.ReferenceGroup
	// If referenceGroupCode already existed, add new identity to group
	minIdp := 0
	if refGroupValue != nil {
		err := proto.Unmarshal(refGroupValue, &refGroup)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// If have at least one node active
		for _, idp := range refGroup.Idps {
			nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp.NodeId
			nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
			if err != nil {
				return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
			}
			if nodeDetailValue == nil {
				return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
			}
			var nodeDetail data.NodeDetail
			err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
			if nodeDetail.Active {
				minIdp = 1
				break
			}
		}
	}
	// Check number of Identifier in new list and old list in stateDB
	var namespaceCount = map[string]int{}
	var checkDuplicateNamespaceAndHash = map[string]int{}
	validNamespace := app.GetNamespaceMap(false)
	for _, identity := range user.NewIdentityList {
		if identity.IdentityNamespace == "" || identity.IdentityIdentifierHash == "" {
			return app.ReturnDeliverTxLog(code.IdentityCannotBeEmpty, "Please input identity detail", "")
		}
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + identity.IdentityNamespace + keySeparator + identity.IdentityIdentifierHash
		identityToRefCodeValue, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
		}
		if identityToRefCodeValue != nil {
			return app.ReturnDeliverTxLog(code.IdentityAlreadyExisted, "Identity already existed", "")
		}
		// check namespace is valid
		if !validNamespace[identity.IdentityNamespace] {
			return app.ReturnDeliverTxLog(code.InvalidNamespace, "Namespace is invalid", "")
		}
		namespaceCount[identity.IdentityNamespace] = namespaceCount[identity.IdentityNamespace] + 1
		checkDuplicateNamespaceAndHash[identity.IdentityNamespace+identity.IdentityIdentifierHash] = checkDuplicateNamespaceAndHash[identity.IdentityNamespace+identity.IdentityIdentifierHash] + 1
	}
	// Check duplicate count
	for _, count := range checkDuplicateNamespaceAndHash {
		if count > 1 {
			return app.ReturnDeliverTxLog(code.DuplicateIdentifier, "There are duplicate identifier", "")
		}
	}
	for _, identity := range refGroup.Identities {
		namespaceCount[identity.Namespace] = namespaceCount[identity.Namespace] + 1
	}
	allowedIdentifierCount := app.GetNamespaceAllowedIdentifierCountMap(false)
	for namespace, count := range namespaceCount {
		if count > allowedIdentifierCount[namespace] && allowedIdentifierCount[namespace] > 0 {
			return app.ReturnDeliverTxLog(code.IdentifierCountIsGreaterThanAllowedIdentifierCount, "Identifier count is greater than allowed identifier count", "")
		}
	}
	foundThisNodeID := false
	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			foundThisNodeID = true
			break
		}
	}
	if foundThisNodeID == false {
		return app.ReturnDeliverTxLog(code.IdentityNotFoundInThisIdP, "Identity not found in this IdP", "")
	}
	if mode3 {
		checkRequestResult := app.checkRequest(user.RequestID, "AddIdentity", minIdp)
		if checkRequestResult.Code != code.OK {
			return checkRequestResult
		}
	}
	for _, identity := range user.NewIdentityList {
		var newIdentity data.IdentityInRefGroup
		newIdentity.Namespace = identity.IdentityNamespace
		newIdentity.IdentifierHash = identity.IdentityIdentifierHash
		refGroup.Identities = append(refGroup.Identities, &newIdentity)
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	if mode3 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(user.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}
	for _, identity := range user.NewIdentityList {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + identity.IdentityNamespace + keySeparator + identity.IdentityIdentifierHash
		identityToRefCodeValue := []byte(user.ReferenceGroupCode)
		app.state.Set([]byte(identityToRefCodeKey), []byte(identityToRefCodeValue))
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	var attributes []kv.Pair
	var attribute kv.Pair
	attribute.Key = []byte("reference_group_code")
	attribute.Value = []byte(user.ReferenceGroupCode)
	attributes = append(attributes, attribute)
	return app.ReturnDeliverTxLogWithAttributes(code.OK, "success", attributes)
}

func (app *ABCIApplication) revokeAndAddAccessor(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RevokeAndAddAccessor, Parameter: %s", param)
	var funcParam RevokeAndAddAccessorParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// check node is active
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
	if !nodeDetail.Active {
		return app.ReturnDeliverTxLog(code.NodeIsNotActive, "Node is not active", "")
	}
	// Get ref group code from revoking accessor ID
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.RevokingAccessorID
	refGroupCode, err := app.state.Get([]byte(accessorToRefCodeKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupCode == nil {
		return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			accessorInIdP := make([]string, 0)
			activeAccessorCount := 0
			for _, accsesor := range idp.Accessors {
				accessorInIdP = append(accessorInIdP, accsesor.AccessorId)
				if accsesor.Active {
					activeAccessorCount++
				}
			}
			if !contains(funcParam.RevokingAccessorID, accessorInIdP) {
				return app.ReturnDeliverTxLog(code.AccessorNotFoundInThisIdP, "Accessor not found in this IdP", "")
			}
		}
	}
	if mode3 {
		minIdp := 1
		checkRequestResult := app.checkRequest(funcParam.RequestID, "RevokeAndAddAccessor", minIdp)
		if checkRequestResult.Code != code.OK {
			return checkRequestResult
		}
	}
	for iIdP, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			for iAcc, accsesor := range idp.Accessors {
				if accsesor.AccessorId == funcParam.RevokingAccessorID {
					refGroup.Idps[iIdP].Accessors[iAcc].Active = false
					break
				}
			}
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	// Check duplicate accessor ID
	accessorToRefCodeKey = accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if refGroupCodeFromDB != nil {
		return app.ReturnDeliverTxLog(code.DuplicateAccessorID, "Duplicate accessor ID", "")
	}
	foundThisNodeID := false
	mode3 = false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			foundThisNodeID = true
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			break
		}
	}
	if foundThisNodeID == false {
		return app.ReturnDeliverTxLog(code.IdentityNotFoundInThisIdP, "Identity not found in this IdP", "")
	}
	var accessor data.Accessor
	accessor.AccessorId = funcParam.AccessorID
	accessor.AccessorType = funcParam.AccessorType
	accessor.AccessorPublicKey = funcParam.AccessorPublicKey
	accessor.Active = true
	accessor.Owner = nodeID
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			idp.Accessors = append(idp.Accessors, &accessor)
			break
		}
	}
	if mode3 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	accessorToRefCodeKey = accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	accessorToRefCodeValue := refGroupCode
	app.state.Set([]byte(accessorToRefCodeKey), []byte(accessorToRefCodeValue))
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	var attributes []kv.Pair
	var attribute kv.Pair
	attribute.Key = []byte("reference_group_code")
	attribute.Value = []byte(refGroupCode)
	attributes = append(attributes, attribute)
	return app.ReturnDeliverTxLogWithAttributes(code.OK, "success", attributes)
}

func MaxInt32(v []int32) int32 {
	var m int32
	for i, e := range v {
		if i == 0 || e < m {
			m = e
		}
	}
	return m
}
