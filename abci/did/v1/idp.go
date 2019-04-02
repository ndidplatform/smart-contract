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
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/ndidplatform/smart-contract/abci/utils"
	"github.com/ndidplatform/smart-contract/protos/data"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

func (app *DIDApplication) AddAccessor(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddAccessor, Parameter: %s", param)
	var funcParam AccessorMethod
	err := json.Unmarshal([]byte(param), &funcParam)
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
		identityToRefCodeKey := "identityToRefCodeKey" + "|" + funcParam.IdentityNamespace + "|" + funcParam.IdentityIdentifierHash
		_, refGroupCodeFromDB := app.GetStateDB([]byte(identityToRefCodeKey))
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := "RefGroupCode" + "|" + string(refGroupCode)
	_, refGroupValue := app.GetStateDB([]byte(refGroupKey))
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
		increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
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
	accessorToRefCodeKey := "accessorToRefCodeKey" + "|" + funcParam.AccessorID
	accessorToRefCodeValue := refGroupCode
	app.SetStateDB([]byte(accessorToRefCodeKey), []byte(accessorToRefCodeValue))
	app.SetStateDB([]byte(refGroupKey), []byte(refGroupValue))
	var tags []cmn.KVPair
	var tag cmn.KVPair
	tag.Key = []byte("reference_group_code")
	tag.Value = []byte(refGroupCode)
	tags = append(tags, tag)
	return app.ReturnDeliverTxLogWitgTag(code.OK, "success", tags)
}

func (app *DIDApplication) registerIdentity(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterIdentity, Parameter: %s", param)
	var funcParam RegisterIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
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
	allowedMode := app.GetAllowedModeFromStateDB("RegisterIdentity")
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
	var namespaceCount = map[string]int{}
	for _, identity := range user.NewIdentityList {
		if identity.IdentityNamespace == "" || identity.IdentityIdentifierHash == "" {
			return app.ReturnDeliverTxLog(code.IdentityCannotBeEmpty, "Please input identity detail", "")
		}
		identityToRefCodeKey := "identityToRefCodeKey" + "|" + identity.IdentityNamespace + "|" + identity.IdentityIdentifierHash
		_, identityToRefCodeValue := app.GetStateDB([]byte(identityToRefCodeKey))
		if identityToRefCodeValue != nil {
			return app.ReturnDeliverTxLog(code.IdentityAlreadyExisted, "Identity already existed", "")
		}
		namespaceCount[identity.IdentityNamespace] = namespaceCount[identity.IdentityNamespace] + 1
	}
	for _, count := range namespaceCount {
		if count > 1 {
			return app.ReturnDeliverTxLog(code.DuplicatedNamespaceInIdentityList, "Namespace in identity list are duplicated", "")
		}
	}
	refGroupKey := "RefGroupCode" + "|" + user.ReferenceGroupCode
	_, refGroupValue := app.GetStateDB([]byte(refGroupKey))
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
			nodeDetailKey := "NodeID" + "|" + idp.NodeId
			_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
			if nodeDetailValue == nil {
				return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
			}
			var nodeDetail data.NodeDetail
			err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
			if nodeDetail.Active {
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
		increaseRequestUseCountResult := app.increaseRequestUseCount(user.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
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
	// Check duplicated namespace in ref group
	foundDuplicatedNamespace := false
	for _, identity := range user.NewIdentityList {
		for _, idenInRefGroup := range refGroup.Identities {
			if identity.IdentityNamespace == idenInRefGroup.Namespace {
				foundDuplicatedNamespace = true
			}
		}
	}
	if foundDuplicatedNamespace {
		return app.ReturnDeliverTxLog(code.DuplicatedNamespaceInIdentityList, "Namespace in identity list are duplicated", "")
	}
	for _, identity := range user.NewIdentityList {
		var newIdentity data.IdentityInRefGroup
		newIdentity.Namespace = identity.IdentityNamespace
		newIdentity.IdentifierHash = identity.IdentityIdentifierHash
		refGroup.Identities = append(refGroup.Identities, &newIdentity)
	}
	foundThisNodeID := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
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
	accessorToRefCodeKey := "accessorToRefCodeKey" + "|" + user.AccessorID
	accessorToRefCodeValue := user.ReferenceGroupCode
	for _, identity := range user.NewIdentityList {
		identityToRefCodeKey := "identityToRefCodeKey" + "|" + identity.IdentityNamespace + "|" + identity.IdentityIdentifierHash
		identityToRefCodeValue := []byte(user.ReferenceGroupCode)
		app.SetStateDB([]byte(identityToRefCodeKey), []byte(identityToRefCodeValue))
	}
	app.SetStateDB([]byte(accessorToRefCodeKey), []byte(accessorToRefCodeValue))
	app.SetStateDB([]byte(refGroupKey), []byte(refGroupValue))
	var tags []cmn.KVPair
	var tag cmn.KVPair
	tag.Key = []byte("reference_group_code")
	tag.Value = []byte(user.ReferenceGroupCode)
	tags = append(tags, tag)
	return app.ReturnDeliverTxLogWitgTag(code.OK, "success", tags)
}

func (app *DIDApplication) checkRequest(requestID string, purpose string, minIdp int) types.ResponseDeliverTx {
	requestKey := "Request" + "|" + requestID
	_, requestValue := app.GetCommittedVersionedStateDB([]byte(requestKey), app.state.Height)
	if requestValue == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err := proto.Unmarshal([]byte(requestValue), &request)
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

func (app *DIDApplication) increaseRequestUseCount(requestID string) types.ResponseDeliverTx {
	requestKey := "Request" + "|" + requestID
	_, requestValue := app.GetCommittedVersionedStateDB([]byte(requestKey), app.state.Height)
	if requestValue == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err := proto.Unmarshal([]byte(requestValue), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	request.UseCount = request.UseCount + 1
	requestProtobuf, err := utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetVersionedStateDB([]byte(requestKey), []byte(requestProtobuf))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) createIdpResponse(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateIdpResponse, Parameter: %s", param)
	var funcParam CreateIdpResponseParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := "Request" + "|" + funcParam.RequestID
	var response data.Response
	response.Ial = funcParam.Ial
	response.Aal = funcParam.Aal
	response.Status = funcParam.Status
	response.Signature = funcParam.Signature
	response.IdpId = nodeID
	_, value := app.GetVersionedStateDB([]byte(key), 0)
	if value == nil {
		return app.ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check duplicate before add
	chkDup := false
	for _, oldResponse := range request.ResponseList {
		if &response == oldResponse {
			chkDup = true
			break
		}
	}
	// Check AAL
	if request.MinAal > response.Aal {
		return app.ReturnDeliverTxLog(code.AALError, "Response's AAL is less than min AAL", "")
	}
	// Check IAL
	if request.MinIal > response.Ial {
		return app.ReturnDeliverTxLog(code.IALError, "Response's IAL is less than min IAL", "")
	}
	// Check AAL, IAL with MaxIalAal
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
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
	// Check min_idp
	if int64(len(request.ResponseList)) >= request.MinIdp {
		return app.ReturnDeliverTxLog(code.RequestIsCompleted, "Can't response a request that's complete response", "")
	}
	// Check IsClosed
	if request.Closed {
		return app.ReturnDeliverTxLog(code.RequestIsClosed, "Can't response a request that's closed", "")
	}
	// Check IsTimedOut
	if request.TimedOut {
		return app.ReturnDeliverTxLog(code.RequestIsTimedOut, "Can't response a request that's timed out", "")
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
	if chkDup == true {
		return app.ReturnDeliverTxLog(code.DuplicateResponse, "Duplicate Response", "")
	}
	request.ResponseList = append(request.ResponseList, &response)
	value, err = utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetVersionedStateDB([]byte(key), []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

func (app *DIDApplication) updateIdentity(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateIdentity, Parameter: %s", param)
	var funcParam UpdateIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check IAL must less than Max IAL
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
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
		identityToRefCodeKey := "identityToRefCodeKey" + "|" + funcParam.IdentityNamespace + "|" + funcParam.IdentityIdentifierHash
		_, refGroupCodeFromDB := app.GetStateDB([]byte(identityToRefCodeKey))
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := "RefGroupCode" + "|" + string(refGroupCode)
	_, refGroupValue := app.GetStateDB([]byte(refGroupKey))
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
	app.SetStateDB([]byte(refGroupKey), []byte(refGroupValue))
	var tags []cmn.KVPair
	var tag cmn.KVPair
	tag.Key = []byte("reference_group_code")
	tag.Value = []byte(refGroupCode)
	tags = append(tags, tag)
	return app.ReturnDeliverTxLogWitgTag(code.OK, "success", tags)
}

func (app *DIDApplication) revokeIdentityAssociation(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RevokeIdentityAssociation, Parameter: %s", param)
	var funcParam RevokeIdentityAssociationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
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
	minIdp := 1
	checkRequestResult := app.checkRequest(funcParam.RequestID, "RevokeIdentityAssociation", minIdp)
	if checkRequestResult.Code != code.OK {
		return checkRequestResult
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := "identityToRefCodeKey" + "|" + funcParam.IdentityNamespace + "|" + funcParam.IdentityIdentifierHash
		_, refGroupCodeFromDB := app.GetStateDB([]byte(identityToRefCodeKey))
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := "RefGroupCode" + "|" + string(refGroupCode)
	_, refGroupValue := app.GetStateDB([]byte(refGroupKey))
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
			refGroup.Idps[index].Active = false
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
	if increaseRequestUseCountResult.Code != code.OK {
		return increaseRequestUseCountResult
	}
	app.SetStateDB([]byte(refGroupKey), []byte(refGroupValue))
	var tags []cmn.KVPair
	var tag cmn.KVPair
	tag.Key = []byte("reference_group_code")
	tag.Value = []byte(refGroupCode)
	tags = append(tags, tag)
	return app.ReturnDeliverTxLogWitgTag(code.OK, "success", tags)
}

func (app *DIDApplication) revokeAccessor(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RevokeAccessor, Parameter: %s", param)
	var funcParam RevokeAccessorParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// check node is active
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
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
		accessorToRefCodeKey := "accessorToRefCodeKey" + "|" + accsesorID
		_, refGroupCodeFromDB := app.GetStateDB([]byte(accessorToRefCodeKey))
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
	refGroupKey := "RefGroupCode" + "|" + string(refGroupCode)
	_, refGroupValue := app.GetStateDB([]byte(refGroupKey))
	if refGroupValue == nil {
		return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for _, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			accessorInIdP := make([]string, 0)
			for _, accsesor := range idp.Accessors {
				accessorInIdP = append(accessorInIdP, accsesor.AccessorId)
			}
			for _, accsesorID := range funcParam.AccessorIDList {
				if !contains(accsesorID, accessorInIdP) {
					return app.ReturnDeliverTxLog(code.AccessorNotFoundInThisIdP, "Accessor not found in this IdP", "")
				}
				break
			}
		}
	}
	for iIdP, idp := range refGroup.Idps {
		if idp.NodeId == nodeID {
			for _, accsesorID := range funcParam.AccessorIDList {
				for iAcc, accsesor := range idp.Accessors {
					if accsesor.AccessorId == accsesorID {
						refGroup.Idps[iIdP].Accessors[iAcc].Active = false
					}
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
	increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
	if increaseRequestUseCountResult.Code != code.OK {
		return increaseRequestUseCountResult
	}
	app.SetStateDB([]byte(refGroupKey), []byte(refGroupValue))
	var tags []cmn.KVPair
	var tag cmn.KVPair
	tag.Key = []byte("reference_group_code")
	tag.Value = []byte(refGroupCode)
	tags = append(tags, tag)
	return app.ReturnDeliverTxLogWitgTag(code.OK, "success", tags)
}

func (app *DIDApplication) updateIdentityModeList(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateIdentityModeList, Parameter: %s", param)
	var funcParam UpdateIdentityModeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Check IAL must less than Max IAL
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
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
		identityToRefCodeKey := "identityToRefCodeKey" + "|" + funcParam.IdentityNamespace + "|" + funcParam.IdentityIdentifierHash
		_, refGroupCodeFromDB := app.GetStateDB([]byte(identityToRefCodeKey))
		if refGroupCodeFromDB == nil {
			return app.ReturnDeliverTxLog(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	// Valid Mode
	var validMode = map[int32]bool{}
	allowedMode := app.GetAllowedModeFromStateDB("RegisterIdentity")
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
	refGroupKey := "RefGroupCode" + "|" + string(refGroupCode)
	_, refGroupValue := app.GetStateDB([]byte(refGroupKey))
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
			refGroup.Idps[index].Mode = funcParam.ModeList
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(refGroupKey), []byte(refGroupValue))
	var tags []cmn.KVPair
	var tag cmn.KVPair
	tag.Key = []byte("reference_group_code")
	tag.Value = []byte(refGroupCode)
	tags = append(tags, tag)
	return app.ReturnDeliverTxLogWitgTag(code.OK, "success", tags)
}

func (app *DIDApplication) addIdentity(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddIdentity, Parameter: %s", param)
	var funcParam AddIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
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
	var namespaceCount = map[string]int{}
	for _, identity := range user.NewIdentityList {
		if identity.IdentityNamespace == "" || identity.IdentityIdentifierHash == "" {
			return app.ReturnDeliverTxLog(code.IdentityCannotBeEmpty, "Please input identity detail", "")
		}
		identityToRefCodeKey := "identityToRefCodeKey" + "|" + identity.IdentityNamespace + "|" + identity.IdentityIdentifierHash
		_, identityToRefCodeValue := app.GetStateDB([]byte(identityToRefCodeKey))
		if identityToRefCodeValue != nil {
			return app.ReturnDeliverTxLog(code.IdentityAlreadyExisted, "Identity already existed", "")
		}
		namespaceCount[identity.IdentityNamespace] = namespaceCount[identity.IdentityNamespace] + 1
	}
	for _, count := range namespaceCount {
		if count > 1 {
			return app.ReturnDeliverTxLog(code.DuplicatedNamespaceInIdentityList, "Namespace in identity list are duplicated", "")
		}
	}
	refGroupKey := "RefGroupCode" + "|" + user.ReferenceGroupCode
	_, refGroupValue := app.GetStateDB([]byte(refGroupKey))
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
			nodeDetailKey := "NodeID" + "|" + idp.NodeId
			_, nodeDetailValue := app.GetStateDB([]byte(nodeDetailKey))
			if nodeDetailValue == nil {
				return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
			}
			var nodeDetail data.NodeDetail
			err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}
			if nodeDetail.Active {
				minIdp = 1
				break
			}
		}
	}
	checkRequestResult := app.checkRequest(user.RequestID, "AddIdentity", minIdp)
	if checkRequestResult.Code != code.OK {
		return checkRequestResult
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
	// Check duplicated namespace in ref group
	foundDuplicatedNamespace := false
	for _, identity := range user.NewIdentityList {
		for _, idenInRefGroup := range refGroup.Identities {
			if identity.IdentityNamespace == idenInRefGroup.Namespace {
				foundDuplicatedNamespace = true
			}
		}
	}
	if foundDuplicatedNamespace {
		return app.ReturnDeliverTxLog(code.DuplicatedNamespaceInIdentityList, "Namespace in identity list are duplicated", "")
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
	increaseRequestUseCountResult := app.increaseRequestUseCount(user.RequestID)
	if increaseRequestUseCountResult.Code != code.OK {
		return increaseRequestUseCountResult
	}
	for _, identity := range user.NewIdentityList {
		identityToRefCodeKey := "identityToRefCodeKey" + "|" + identity.IdentityNamespace + "|" + identity.IdentityIdentifierHash
		identityToRefCodeValue := []byte(user.ReferenceGroupCode)
		app.SetStateDB([]byte(identityToRefCodeKey), []byte(identityToRefCodeValue))
	}
	app.SetStateDB([]byte(refGroupKey), []byte(refGroupValue))
	var tags []cmn.KVPair
	var tag cmn.KVPair
	tag.Key = []byte("reference_group_code")
	tag.Value = []byte(user.ReferenceGroupCode)
	tags = append(tags, tag)
	return app.ReturnDeliverTxLogWitgTag(code.OK, "success", tags)
}
