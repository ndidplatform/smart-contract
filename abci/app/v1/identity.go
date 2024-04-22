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
	"sort"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type Identity struct {
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
}

type RegisterIdentityParam struct {
	ReferenceGroupCode string     `json:"reference_group_code"`
	NewIdentityList    []Identity `json:"new_identity_list"`
	Ial                float64    `json:"ial"`
	Lial               *bool      `json:"lial"`
	Laal               *bool      `json:"laal"`
	ModeList           []int32    `json:"mode_list"`
	AccessorID         string     `json:"accessor_id"`
	AccessorPublicKey  string     `json:"accessor_public_key"`
	AccessorType       string     `json:"accessor_type"`
	RequestID          string     `json:"request_id"`
}

func (app *ABCIApplication) validateRegisterIdentity(funcParam RegisterIdentityParam, callerNodeID string, committedState bool, checktx bool) error {
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

	ok := app.isIDPNode(&nodeDetail)
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP method",
		}
	}

	// stateless

	// Check for identity_namespace and identity_identifier_hash. If exist, error.
	if funcParam.ReferenceGroupCode == "" {
		return &ApplicationError{
			Code:    code.RefGroupCodeCannotBeEmpty,
			Message: "Reference group code is required",
		}
	}
	// Check accessor
	if funcParam.AccessorID == "" {
		return &ApplicationError{
			Code:    code.AccessorIDCannotBeEmpty,
			Message: "Accessor ID is required",
		}
	}
	if funcParam.AccessorPublicKey == "" {
		return &ApplicationError{
			Code:    code.AccessorPublicKeyCannotBeEmpty,
			Message: "Accessor public key is required",
		}
	}
	if funcParam.AccessorType == "" {
		return &ApplicationError{
			Code:    code.AccessorTypeCannotBeEmpty,
			Message: "Accessor type is required",
		}
	}

	var newIdentityNamespaceAndHash = map[string]bool{}
	for _, identity := range funcParam.NewIdentityList {
		if identity.IdentityNamespace == "" || identity.IdentityIdentifierHash == "" {
			return &ApplicationError{
				Code:    code.IdentityCannotBeEmpty,
				Message: "Identity detail is required",
			}
		}

		// Check for duplicates
		if _, ok := newIdentityNamespaceAndHash[identity.IdentityNamespace+identity.IdentityIdentifierHash]; ok {
			return &ApplicationError{
				Code:    code.DuplicateIdentifier,
				Message: "Duplicate identifiers",
			}
		}

		newIdentityNamespaceAndHash[identity.IdentityNamespace+identity.IdentityIdentifierHash] = true
	}

	err = checkAccessorPubKey(funcParam.AccessorPublicKey)
	if err != nil {
		return err
	}

	if checktx {
		return nil
	}

	// stateful

	// Validate user's ial is <= node's max_ial
	if funcParam.Ial > nodeDetail.MaxIal {
		return &ApplicationError{
			Code:    code.IALError,
			Message: "IAL must be less than or equals to registered node's max IAL",
		}
	}

	// Valid Mode
	var validMode = map[int32]bool{}
	allowedMode := app.GetAllowedModeFromStateDB("RegisterIdentity", committedState)
	for _, mode := range allowedMode {
		validMode[mode] = true
	}

	for _, mode := range funcParam.ModeList {
		if !validMode[mode] {
			return &ApplicationError{
				Code:    code.InvalidMode,
				Message: "Invalid mode for register identity",
			}
		}
	}
	modeMap := make(map[int32]struct{})
	for _, mode := range funcParam.ModeList {
		if _, ok := modeMap[mode]; !ok {
			modeMap[mode] = struct{}{}
		}
	}

	refGroupKey := refGroupCodeKeyPrefix + keySeparator + funcParam.ReferenceGroupCode
	refGroupValue, err := app.state.Get([]byte(refGroupKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}

	var refGroup data.ReferenceGroup

	mode3 := false
	if _, ok := modeMap[3]; ok {
		mode3 = true
	}

	minIdp := 0
	if refGroupValue != nil {
		err := proto.Unmarshal(refGroupValue, &refGroup)
		if err != nil {
			return &ApplicationError{
				Code:    code.UnmarshalError,
				Message: err.Error(),
			}
		}
		// If there's at least one node active
		for _, idp := range refGroup.Idps {
			nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp.NodeId
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
			err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				return &ApplicationError{
					Code:    code.UnmarshalError,
					Message: err.Error(),
				}
			}
			if nodeDetail.Active && idp.Active {
				minIdp = 1
				break
			}
		}
	}
	if mode3 && minIdp > 0 {
		err = app.checkRequestUsable(funcParam.RequestID, "RegisterIdentity", minIdp, committedState)
		if err != nil {
			return err
		}
	}

	// Check min_ial when RegisterIdentity when onboard as first IdP
	if refGroupValue == nil {
		if funcParam.Ial < app.GetAllowedMinIalForRegisterIdentityAtFirstIdpFromStateDB(committedState) {
			return &ApplicationError{
				Code:    code.IalMustBeGreaterOrEqualMinIal,
				Message: "IAL must be greater or equal min IAL when onboard as first IdP",
			}
		}
	}

	// Check number of Identifier in new list and old list in stateDB
	var namespaceCount = map[string]int{}
	validNamespace := app.GetNamespaceMap(false)
	for _, identity := range funcParam.NewIdentityList {
		if identity.IdentityNamespace == "" || identity.IdentityIdentifierHash == "" {
			return &ApplicationError{
				Code:    code.IdentityCannotBeEmpty,
				Message: "Identity detail is required",
			}
		}
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + identity.IdentityNamespace + keySeparator + identity.IdentityIdentifierHash
		identityToRefCodeValue, err := app.state.Get([]byte(identityToRefCodeKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if identityToRefCodeValue != nil {
			return &ApplicationError{
				Code:    code.IdentityAlreadyExists,
				Message: "Identity already exists",
			}
		}
		// check namespace is valid
		if !validNamespace[identity.IdentityNamespace] {
			return &ApplicationError{
				Code:    code.InvalidNamespace,
				Message: "Invalid namespace",
			}
		}
		namespaceCount[identity.IdentityNamespace] = namespaceCount[identity.IdentityNamespace] + 1
	}
	for _, identity := range refGroup.Identities {
		namespaceCount[identity.Namespace] = namespaceCount[identity.Namespace] + 1
	}
	allowedIdentifierCount := app.GetNamespaceAllowedIdentifierCountMap(committedState)
	for namespace, count := range namespaceCount {
		if count > allowedIdentifierCount[namespace] && allowedIdentifierCount[namespace] > 0 {
			return &ApplicationError{
				Code:    code.IdentifierCountIsGreaterThanAllowedIdentifierCount,
				Message: "Identifier count is greater than allowed identifier count",
			}
		}
	}

	// Check duplicate accessor ID
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupCodeFromDB != nil {
		// check if it is a case of reactivate accessor (with the same ref group code) at the same IdP
		reactivateAccessor := false

		refGroupKey := refGroupCodeKeyPrefix + keySeparator + funcParam.ReferenceGroupCode
		refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}

		if refGroupValue != nil {
			var refGroup data.ReferenceGroup
			err := proto.Unmarshal(refGroupValue, &refGroup)
			if err != nil {
				return &ApplicationError{
					Code:    code.UnmarshalError,
					Message: err.Error(),
				}
			}

			for _, idp := range refGroup.Idps {
				if idp.NodeId == callerNodeID {
					for _, accessor := range idp.Accessors {
						if accessor.AccessorId == funcParam.AccessorID {
							if !accessor.Active {
								reactivateAccessor = true
							}
						}
					}
				}
			}
		}

		if !reactivateAccessor {
			return &ApplicationError{
				Code:    code.DuplicateAccessorID,
				Message: "Duplicate accessor ID",
			}
		}
	}

	return nil
}

func (app *ABCIApplication) registerIdentityCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RegisterIdentityParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRegisterIdentity(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) registerIdentity(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RegisterIdentity, Parameter: %s", param)
	var funcParam RegisterIdentityParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRegisterIdentity(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	user := funcParam

	// remove duplicates
	modeMap := make(map[int32]struct{})
	modeListNoDuplicate := make([]int32, 0)
	for _, mode := range user.ModeList {
		if _, ok := modeMap[mode]; !ok {
			modeMap[mode] = struct{}{}
			modeListNoDuplicate = append(modeListNoDuplicate, mode)
		}
	}
	user.ModeList = modeListNoDuplicate
	// sort mode list
	sort.Slice(user.ModeList, func(i, j int) bool { return user.ModeList[i] < user.ModeList[j] })

	refGroupKey := refGroupCodeKeyPrefix + keySeparator + user.ReferenceGroupCode
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	var refGroup data.ReferenceGroup

	// If referenceGroupCode already exists, add new identity to group

	mode3 := false
	if _, ok := modeMap[3]; ok {
		mode3 = true
	}

	minIdp := 0
	if refGroupValue != nil {
		err := proto.Unmarshal(refGroupValue, &refGroup)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
		// If there's at least one node active
		for _, idp := range refGroup.Idps {
			nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp.NodeId
			nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
			if err != nil {
				return app.NewExecTxResult(code.AppStateError, err.Error(), "")
			}
			if nodeDetailValue == nil {
				return app.NewExecTxResult(code.NodeIDNotFound, "Node ID not found", "")
			}
			var nodeDetail data.NodeDetail
			err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
			}
			if nodeDetail.Active && idp.Active {
				minIdp = 1
				break
			}
		}
	}

	var accessor data.Accessor
	accessor.AccessorId = user.AccessorID
	accessor.AccessorType = user.AccessorType
	accessor.AccessorPublicKey = user.AccessorPublicKey
	accessor.Active = true
	accessor.Owner = callerNodeID
	accessor.CreationBlockHeight = app.state.CurrentBlockHeight
	accessor.CreationChainId = app.CurrentChain
	var idp data.IdPInRefGroup
	idp.NodeId = callerNodeID
	idp.Mode = append(idp.Mode, user.ModeList...)
	idp.Accessors = append(idp.Accessors, &accessor)
	idp.Ial = user.Ial
	if user.Lial != nil {
		idp.Lial = &wrapperspb.BoolValue{Value: *user.Lial}
	}
	if user.Laal != nil {
		idp.Laal = &wrapperspb.BoolValue{Value: *user.Laal}
	}
	idp.Active = true
	for _, identity := range user.NewIdentityList {
		var newIdentity data.IdentityInRefGroup
		newIdentity.Namespace = identity.IdentityNamespace
		newIdentity.IdentifierHash = identity.IdentityIdentifierHash
		refGroup.Identities = append(refGroup.Identities, &newIdentity)
	}
	foundThisNodeID := false
	for iIdp, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			refGroup.Idps[iIdp].Active = true
			refGroup.Idps[iIdp].Mode = funcParam.ModeList
			// should accessors be replaced instead?
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
			// update IAL, LIAL, LAAL
			refGroup.Idps[iIdp].Ial = user.Ial
			if user.Lial != nil {
				refGroup.Idps[iIdp].Lial = &wrapperspb.BoolValue{Value: *user.Lial}
			}
			if user.Laal != nil {
				refGroup.Idps[iIdp].Laal = &wrapperspb.BoolValue{Value: *user.Laal}
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
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
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

	var attributes []abcitypes.EventAttribute
	var attribute abcitypes.EventAttribute
	attribute.Key = "reference_group_code"
	attribute.Value = user.ReferenceGroupCode
	attributes = append(attributes, attribute)

	return app.NewExecTxResultWithAttributes(code.OK, "success", attributes)
}

type UpdateIdentityParam struct {
	ReferenceGroupCode     string   `json:"reference_group_code"`
	IdentityNamespace      string   `json:"identity_namespace"`
	IdentityIdentifierHash string   `json:"identity_identifier_hash"`
	Ial                    *float64 `json:"ial"`
	Lial                   *bool    `json:"lial"`
	Laal                   *bool    `json:"laal"`
}

func (app *ABCIApplication) validateUpdateIdentity(funcParam UpdateIdentityParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isIDPNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP method",
		}
	}

	// stateless

	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		return &ApplicationError{
			Code:    code.GotRefGroupCodeAndIdentity,
			Message: "Found reference group code and identity detail in parameter",
		}
	}

	if funcParam.Ial == nil && funcParam.Lial == nil && funcParam.Laal == nil {
		return &ApplicationError{
			Code:    code.NothingToUpdateIdentity,
			Message: "Nothing to update",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	if funcParam.Ial != nil {
		// Check IAL must less than Max IAL
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + callerNodeID
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
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
		if *funcParam.Ial > nodeDetail.MaxIal {
			return &ApplicationError{
				Code:    code.IALError,
				Message: "New IAL is greater than max IAL",
			}
		}
	}

	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeBytes, err := app.state.Get([]byte(identityToRefCodeKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if refGroupCodeBytes == nil {
			return &ApplicationError{
				Code:    code.RefGroupNotFound,
				Message: "Reference group not found",
			}
		}
		refGroupCode = string(refGroupCodeBytes)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupValue == nil {
		return &ApplicationError{
			Code:    code.RefGroupNotFound,
			Message: "Reference group not found",
		}
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	nodeIDToUpdateIndex := -1
	for index, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			nodeIDToUpdateIndex = index
			break
		}
	}
	if nodeIDToUpdateIndex < 0 {
		return &ApplicationError{
			Code:    code.IdentityNotFoundInThisIdP,
			Message: "Identity not found in this IdP",
		}
	}

	return nil
}

func (app *ABCIApplication) updateIdentityCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam UpdateIdentityParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateUpdateIdentity(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) updateIdentity(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("UpdateIdentity, Parameter: %s", param)
	var funcParam UpdateIdentityParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateUpdateIdentity(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeBytes, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.NewExecTxResult(code.AppStateError, err.Error(), "")
		}
		if refGroupCodeBytes == nil {
			return app.NewExecTxResult(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeBytes)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.NewExecTxResult(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}
	nodeIDToUpdateIndex := -1
	for index, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			nodeIDToUpdateIndex = index
			break
		}
	}

	if funcParam.Ial != nil {
		refGroup.Idps[nodeIDToUpdateIndex].Ial = *funcParam.Ial
	}

	if funcParam.Lial != nil {
		refGroup.Idps[nodeIDToUpdateIndex].Lial = &wrapperspb.BoolValue{Value: *funcParam.Lial}
	}

	if funcParam.Laal != nil {
		refGroup.Idps[nodeIDToUpdateIndex].Laal = &wrapperspb.BoolValue{Value: *funcParam.Laal}
	}

	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))

	var attributes []abcitypes.EventAttribute
	var attribute abcitypes.EventAttribute
	attribute.Key = "reference_group_code"
	attribute.Value = refGroupCode
	attributes = append(attributes, attribute)

	return app.NewExecTxResultWithAttributes(code.OK, "success", attributes)
}

type UpdateIdentityModeListParam struct {
	ReferenceGroupCode     string  `json:"reference_group_code"`
	IdentityNamespace      string  `json:"identity_namespace"`
	IdentityIdentifierHash string  `json:"identity_identifier_hash"`
	ModeList               []int32 `json:"mode_list"`
	RequestID              string  `json:"request_id"`
}

func (app *ABCIApplication) validateUpdateIdentityModeList(funcParam UpdateIdentityModeListParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isIDPNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP method",
		}
	}

	// stateless

	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		return &ApplicationError{
			Code:    code.GotRefGroupCodeAndIdentity,
			Message: "Found reference group code and identity detail in parameter",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if refGroupCodeFromDB == nil {
			return &ApplicationError{
				Code:    code.RefGroupNotFound,
				Message: "Reference group not found",
			}
		}
		refGroupCode = string(refGroupCodeFromDB)
	}

	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupValue == nil {
		return &ApplicationError{
			Code:    code.RefGroupNotFound,
			Message: "Reference group not found",
		}
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	foundThisNodeID := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			foundThisNodeID = true
			break
		}
	}
	if !foundThisNodeID {
		return &ApplicationError{
			Code:    code.IdentityNotFoundInThisIdP,
			Message: "Identity not found in this IdP",
		}
	}

	// Valid Mode
	var validMode = map[int32]bool{}
	allowedMode := app.GetAllowedModeFromStateDB("UpdateIdentityModeList", committedState)
	for _, mode := range allowedMode {
		validMode[mode] = true
	}
	for _, mode := range funcParam.ModeList {
		if !validMode[mode] {
			return &ApplicationError{
				Code:    code.InvalidMode,
				Message: "Must register identity on valid mode",
			}
		}
	}

	for index, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			// Check new mode list is higher than current mode list
			maxCurrentMode := MaxInt32(refGroup.Idps[index].Mode)
			maxNewMode := MaxInt32(funcParam.ModeList)
			if maxCurrentMode > maxNewMode {
				return &ApplicationError{
					Code:    code.NewModeListMustBeHigherThanCurrentModeList,
					Message: "New mode list must be higher than current mode list",
				}
			}
			break
		}
	}

	return nil
}

func (app *ABCIApplication) updateIdentityModeListCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam UpdateIdentityModeListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateUpdateIdentityModeList(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) updateIdentityModeList(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("UpdateIdentityModeList, Parameter: %s", param)
	var funcParam UpdateIdentityModeListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateUpdateIdentityModeList(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.NewExecTxResult(code.AppStateError, err.Error(), "")
		}
		if refGroupCodeFromDB == nil {
			return app.NewExecTxResult(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}

	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.NewExecTxResult(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	// remove duplicates
	modeMap := make(map[int32]struct{})
	modeListNoDuplicate := make([]int32, 0)
	for _, mode := range funcParam.ModeList {
		if _, ok := modeMap[mode]; !ok {
			modeMap[mode] = struct{}{}
			modeListNoDuplicate = append(modeListNoDuplicate, mode)
		}
	}
	funcParam.ModeList = modeListNoDuplicate
	// sort mode list
	sort.Slice(funcParam.ModeList, func(i, j int) bool { return funcParam.ModeList[i] < funcParam.ModeList[j] })

	for index, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			refGroup.Idps[index].Mode = funcParam.ModeList
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))

	var attributes []abcitypes.EventAttribute
	var attribute abcitypes.EventAttribute
	attribute.Key = "reference_group_code"
	attribute.Value = refGroupCode
	attributes = append(attributes, attribute)

	return app.NewExecTxResultWithAttributes(code.OK, "success", attributes)
}

type AddIdentityParam struct {
	ReferenceGroupCode string     `json:"reference_group_code"`
	NewIdentityList    []Identity `json:"new_identity_list"`
	RequestID          string     `json:"request_id"`
}

func (app *ABCIApplication) validateAddIdentity(funcParam AddIdentityParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isIDPNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP method",
		}
	}

	// stateless

	if funcParam.ReferenceGroupCode == "" {
		return &ApplicationError{
			Code:    code.RefGroupCodeCannotBeEmpty,
			Message: "Reference group code cannot be empty",
		}
	}

	var newIdentityNamespaceAndHash = map[string]bool{}
	for _, identity := range funcParam.NewIdentityList {
		if identity.IdentityNamespace == "" || identity.IdentityIdentifierHash == "" {
			return &ApplicationError{
				Code:    code.IdentityCannotBeEmpty,
				Message: "Please input identity detail",
			}
		}

		// Check for duplicates
		if _, ok := newIdentityNamespaceAndHash[identity.IdentityNamespace+identity.IdentityIdentifierHash]; ok {
			return &ApplicationError{
				Code:    code.DuplicateIdentifier,
				Message: "Duplicate identifiers",
			}
		}

		newIdentityNamespaceAndHash[identity.IdentityNamespace+identity.IdentityIdentifierHash] = true
	}

	if checktx {
		return nil
	}

	// stateful

	refGroupKey := refGroupCodeKeyPrefix + keySeparator + funcParam.ReferenceGroupCode
	refGroupValue, err := app.state.Get([]byte(refGroupKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupValue == nil {
		return &ApplicationError{
			Code:    code.RefGroupNotFound,
			Message: "Reference group not found",
		}
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	minIdp := 0
	// If have at least one node active
	for _, idp := range refGroup.Idps {
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp.NodeId
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
		err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			return &ApplicationError{
				Code:    code.UnmarshalError,
				Message: err.Error(),
			}
		}
		if nodeDetail.Active {
			minIdp = 1
			break
		}
	}

	// Check number of Identifier in new list and old list in stateDB
	var namespaceCount = map[string]int{}
	validNamespace := app.GetNamespaceMap(committedState)
	for _, identity := range funcParam.NewIdentityList {
		if identity.IdentityNamespace == "" || identity.IdentityIdentifierHash == "" {
			return &ApplicationError{
				Code:    code.IdentityCannotBeEmpty,
				Message: "Please input identity detail",
			}
		}
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + identity.IdentityNamespace + keySeparator + identity.IdentityIdentifierHash
		identityToRefCodeValue, err := app.state.Get([]byte(identityToRefCodeKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if identityToRefCodeValue != nil {
			return &ApplicationError{
				Code:    code.IdentityAlreadyExists,
				Message: "Identity already exists",
			}
		}
		// check namespace is valid
		if !validNamespace[identity.IdentityNamespace] {
			return &ApplicationError{
				Code:    code.InvalidNamespace,
				Message: "Invalid namespace",
			}
		}
		namespaceCount[identity.IdentityNamespace] = namespaceCount[identity.IdentityNamespace] + 1
	}
	for _, identity := range refGroup.Identities {
		namespaceCount[identity.Namespace] = namespaceCount[identity.Namespace] + 1
	}
	allowedIdentifierCount := app.GetNamespaceAllowedIdentifierCountMap(committedState)
	for namespace, count := range namespaceCount {
		if count > allowedIdentifierCount[namespace] && allowedIdentifierCount[namespace] > 0 {
			return &ApplicationError{
				Code:    code.IdentifierCountIsGreaterThanAllowedIdentifierCount,
				Message: "Identifier count is greater than allowed identifier count",
			}
		}
	}

	foundThisNodeID := false
	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
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
	if !foundThisNodeID {
		return &ApplicationError{
			Code:    code.IdentityNotFoundInThisIdP,
			Message: "Identity not found in this IdP",
		}
	}

	if mode3 {
		err = app.checkRequestUsable(funcParam.RequestID, "AddIdentity", minIdp, committedState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *ABCIApplication) addIdentityCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddIdentityParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddIdentity(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addIdentity(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddIdentity, Parameter: %s", param)
	var funcParam AddIdentityParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddIdentity(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	user := funcParam

	refGroupKey := refGroupCodeKeyPrefix + keySeparator + user.ReferenceGroupCode
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			break
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
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
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

	var attributes []abcitypes.EventAttribute
	var attribute abcitypes.EventAttribute
	attribute.Key = "reference_group_code"
	attribute.Value = user.ReferenceGroupCode
	attributes = append(attributes, attribute)

	return app.NewExecTxResultWithAttributes(code.OK, "success", attributes)
}

type RevokeIdentityAssociationParam struct {
	ReferenceGroupCode     string `json:"reference_group_code"`
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
	RequestID              string `json:"request_id"`
}

func (app *ABCIApplication) validateRevokeIdentityAssociation(funcParam RevokeIdentityAssociationParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isIDPNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP method",
		}
	}

	// stateless

	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		return &ApplicationError{
			Code:    code.GotRefGroupCodeAndIdentity,
			Message: "Found reference group code and identity detail in parameter",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if refGroupCodeFromDB == nil {
			return &ApplicationError{
				Code:    code.RefGroupNotFound,
				Message: "Reference group not found",
			}
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupValue == nil {
		return &ApplicationError{
			Code:    code.RefGroupNotFound,
			Message: "Reference group not found",
		}
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	foundThisNodeID := false
	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
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
	if !foundThisNodeID {
		return &ApplicationError{
			Code:    code.IdentityNotFoundInThisIdP,
			Message: "Identity not found in this IdP",
		}
	}

	if mode3 {
		minIdp := 1
		err = app.checkRequestUsable(funcParam.RequestID, "RevokeIdentityAssociation", minIdp, committedState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *ABCIApplication) revokeIdentityAssociationCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RevokeIdentityAssociationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRevokeIdentityAssociation(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) revokeIdentityAssociation(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RevokeIdentityAssociation, Parameter: %s", param)
	var funcParam RevokeIdentityAssociationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRevokeIdentityAssociation(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.NewExecTxResult(code.AppStateError, err.Error(), "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			break
		}
	}

	for iIdP, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			refGroup.Idps[iIdP].Active = false
			for iAcc := range idp.Accessors {
				refGroup.Idps[iIdP].Accessors[iAcc].Active = false
			}
			break
		}
	}

	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	if mode3 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))

	var attributes []abcitypes.EventAttribute
	var attribute abcitypes.EventAttribute
	attribute.Key = "reference_group_code"
	attribute.Value = refGroupCode
	attributes = append(attributes, attribute)

	return app.NewExecTxResultWithAttributes(code.OK, "success", attributes)
}

type AddAccessorParam struct {
	ReferenceGroupCode     string `json:"reference_group_code"`
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
	AccessorID             string `json:"accessor_id"`
	AccessorPublicKey      string `json:"accessor_public_key"`
	AccessorType           string `json:"accessor_type"`
	RequestID              string `json:"request_id"`
}

func (app *ABCIApplication) validateAddAccessor(funcParam AddAccessorParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isIDPNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP method",
		}
	}

	// stateless

	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		return &ApplicationError{
			Code:    code.GotRefGroupCodeAndIdentity,
			Message: "Found reference group code and identity detail in parameter",
		}
	}

	err = checkAccessorPubKey(funcParam.AccessorPublicKey)
	if err != nil {
		return err
	}

	if checktx {
		return nil
	}

	// stateful

	// Check duplicate accessor ID
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupCodeFromDB != nil {
		return &ApplicationError{
			Code:    code.DuplicateAccessorID,
			Message: "Duplicate accessor ID",
		}
	}

	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if refGroupCodeFromDB == nil {
			return &ApplicationError{
				Code:    code.RefGroupNotFound,
				Message: "Reference group not found",
			}
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupValue == nil {
		return &ApplicationError{
			Code:    code.RefGroupNotFound,
			Message: "Reference group not found",
		}
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	foundThisNodeID := false
	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
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
	if !foundThisNodeID {
		return &ApplicationError{
			Code:    code.IdentityNotFoundInThisIdP,
			Message: "Identity not found in this IdP",
		}
	}

	if mode3 {
		minIdp := 1
		err = app.checkRequestUsable(funcParam.RequestID, "AddAccessor", minIdp, committedState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *ABCIApplication) addAccessorCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddAccessorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddAccessor(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addAccessor(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddAccessor, Parameter: %s", param)
	var funcParam AddAccessorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddAccessor(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), false)
		if err != nil {
			return app.NewExecTxResult(code.AppStateError, err.Error(), "")
		}
		if refGroupCodeFromDB == nil {
			return app.NewExecTxResult(code.RefGroupNotFound, "Reference group not found", "")
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if refGroupValue == nil {
		return app.NewExecTxResult(code.RefGroupNotFound, "Reference group not found", "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			break
		}
	}

	var accessor data.Accessor
	accessor.AccessorId = funcParam.AccessorID
	accessor.AccessorType = funcParam.AccessorType
	accessor.AccessorPublicKey = funcParam.AccessorPublicKey
	accessor.Active = true
	accessor.Owner = callerNodeID
	accessor.CreationBlockHeight = app.state.CurrentBlockHeight
	accessor.CreationChainId = app.CurrentChain
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			idp.Accessors = append(idp.Accessors, &accessor)
			break
		}
	}
	refGroupValue, err = utils.ProtoDeterministicMarshal(&refGroup)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	if mode3 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}

	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	accessorToRefCodeValue := refGroupCode
	app.state.Set([]byte(accessorToRefCodeKey), []byte(accessorToRefCodeValue))
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))

	var attributes []abcitypes.EventAttribute
	var attribute abcitypes.EventAttribute
	attribute.Key = "reference_group_code"
	attribute.Value = refGroupCode
	attributes = append(attributes, attribute)

	return app.NewExecTxResultWithAttributes(code.OK, "success", attributes)
}

type RevokeAccessorParam struct {
	AccessorIDList []string `json:"accessor_id_list"`
	RequestID      string   `json:"request_id"`
}

func (app *ABCIApplication) validateRevokeAccessor(funcParam RevokeAccessorParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isIDPNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP method",
		}
	}

	// stateless

	// if len(funcParam.AccessorIDList) == 0 {
	// TODO: err
	// }

	if checktx {
		return nil
	}

	// stateful

	// check if all accessor IDs have the same ref group code
	var refGroupCode string
	for index, accessorID := range funcParam.AccessorIDList {
		accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + accessorID
		refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), false)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if refGroupCodeFromDB == nil {
			return &ApplicationError{
				Code:    code.RefGroupNotFound,
				Message: "Reference group not found",
			}
		}
		if index == 0 {
			refGroupCode = string(refGroupCodeFromDB)
		} else {
			if string(refGroupCodeFromDB) != refGroupCode {
				return &ApplicationError{
					Code:    code.AllAccessorMustHaveSameRefGroupCode,
					Message: "All accessors must have the same reference group code",
				}
			}
		}
	}

	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupValue == nil {
		return &ApplicationError{
			Code:    code.RefGroupNotFound,
			Message: "Reference group not found",
		}
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			accessorInIdP := make([]string, 0)
			activeAccessorCount := 0
			for _, accessor := range idp.Accessors {
				accessorInIdP = append(accessorInIdP, accessor.AccessorId)
				if accessor.Active {
					activeAccessorCount++
				}
			}
			for _, accessorID := range funcParam.AccessorIDList {
				if !contains(accessorID, accessorInIdP) {
					return &ApplicationError{
						Code:    code.AccessorNotFoundInThisIdP,
						Message: "Accessor not found in this IdP",
					}
				}
			}
			if activeAccessorCount-len(funcParam.AccessorIDList) < 1 {
				return &ApplicationError{
					Code:    code.CannotRevokeAllAccessorsInThisIdP,
					Message: "Cannot revoke all accessors in this IdP",
				}
			}
		}
	}

	if mode3 {
		minIdp := 1
		err = app.checkRequestUsable(funcParam.RequestID, "RevokeAccessor", minIdp, committedState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *ABCIApplication) revokeAccessorCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RevokeAccessorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRevokeAccessor(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) revokeAccessor(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RevokeAccessor, Parameter: %s", param)
	var funcParam RevokeAccessorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRevokeAccessor(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// get ref group code
	var refGroupCode string
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorIDList[0]
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	refGroupCode = string(refGroupCodeFromDB)

	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
		}
	}

	for iIdP, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			for _, accessorID := range funcParam.AccessorIDList {
				for iAcc, accessor := range idp.Accessors {
					// app.logger.Debugf("Acces:%s", args)
					if accessor.AccessorId == accessorID {
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
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	if mode3 {
		increaseRequestUseCountResult := app.increaseRequestUseCount(funcParam.RequestID)
		if increaseRequestUseCountResult.Code != code.OK {
			return increaseRequestUseCountResult
		}
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))

	var attributes []abcitypes.EventAttribute
	var attribute abcitypes.EventAttribute
	attribute.Key = "reference_group_code"
	attribute.Value = refGroupCode
	attributes = append(attributes, attribute)

	return app.NewExecTxResultWithAttributes(code.OK, "success", attributes)
}

type RevokeAndAddAccessorParam struct {
	RevokingAccessorID string `json:"revoking_accessor_id"`
	AccessorID         string `json:"accessor_id"`
	AccessorPublicKey  string `json:"accessor_public_key"`
	AccessorType       string `json:"accessor_type"`
	RequestID          string `json:"request_id"`
}

func (app *ABCIApplication) validateRevokeAndAddAccessor(funcParam RevokeAndAddAccessorParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isIDPNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallIdPMethod,
			Message: "This node does not have permission to call IdP method",
		}
	}

	// stateless

	err = checkAccessorPubKey(funcParam.AccessorPublicKey)
	if err != nil {
		return err
	}

	if checktx {
		return nil
	}

	// stateful

	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.RevokingAccessorID
	refGroupCode, err := app.state.Get([]byte(accessorToRefCodeKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupCode == nil {
		return &ApplicationError{
			Code:    code.RefGroupNotFound,
			Message: "Reference group not found",
		}
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupValue == nil {
		return &ApplicationError{
			Code:    code.RefGroupNotFound,
			Message: "Reference group not found",
		}
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	foundThisNodeID := false
	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			foundThisNodeID = true
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
			accessorInIdP := make([]string, 0)
			activeAccessorCount := 0
			for _, accessor := range idp.Accessors {
				accessorInIdP = append(accessorInIdP, accessor.AccessorId)
				if accessor.Active {
					activeAccessorCount++
				}
			}
			if !contains(funcParam.RevokingAccessorID, accessorInIdP) {
				return &ApplicationError{
					Code:    code.AccessorNotFoundInThisIdP,
					Message: "Accessor not found in this IdP",
				}
			}
		}
	}

	if !foundThisNodeID {
		return &ApplicationError{
			Code:    code.IdentityNotFoundInThisIdP,
			Message: "Identity not found in this IdP",
		}
	}

	if mode3 {
		minIdp := 1
		err = app.checkRequestUsable(funcParam.RequestID, "RevokeAndAddAccessor", minIdp, committedState)
		if err != nil {
			return err
		}
	}

	// check for add new accessor

	// Check duplicate accessor ID
	accessorToRefCodeKey = accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if refGroupCodeFromDB != nil {
		return &ApplicationError{
			Code:    code.DuplicateAccessorID,
			Message: "Duplicate accessor ID",
		}
	}

	return nil
}

func (app *ABCIApplication) revokeAndAddAccessorCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RevokeAndAddAccessorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRevokeAndAddAccessor(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) revokeAndAddAccessor(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RevokeAndAddAccessor, Parameter: %s", param)
	var funcParam RevokeAndAddAccessorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRevokeAndAddAccessor(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Get ref group code from revoking accessor ID
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.RevokingAccessorID
	refGroupCode, err := app.state.Get([]byte(accessorToRefCodeKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	mode3 := false
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			for _, mode := range idp.Mode {
				if mode == 3 {
					mode3 = true
					break
				}
			}
		}
	}

	for iIdP, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
			for iAcc, accessor := range idp.Accessors {
				if accessor.AccessorId == funcParam.RevokingAccessorID {
					refGroup.Idps[iIdP].Accessors[iAcc].Active = false
					break
				}
			}
			break
		}
	}

	// add new accessor

	var accessor data.Accessor
	accessor.AccessorId = funcParam.AccessorID
	accessor.AccessorType = funcParam.AccessorType
	accessor.AccessorPublicKey = funcParam.AccessorPublicKey
	accessor.Active = true
	accessor.Owner = callerNodeID
	accessor.CreationBlockHeight = app.state.CurrentBlockHeight
	accessor.CreationChainId = app.CurrentChain
	for _, idp := range refGroup.Idps {
		if idp.NodeId == callerNodeID {
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
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))
	accessorToRefCodeKey = accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	accessorToRefCodeValue := refGroupCode
	app.state.Set([]byte(accessorToRefCodeKey), []byte(accessorToRefCodeValue))
	app.state.Set([]byte(refGroupKey), []byte(refGroupValue))

	var attributes []abcitypes.EventAttribute
	var attribute abcitypes.EventAttribute
	attribute.Key = "reference_group_code"
	attribute.Value = string(refGroupCode)
	attributes = append(attributes, attribute)

	return app.NewExecTxResultWithAttributes(code.OK, "success", attributes)
}

type CheckExistingIdentityParam struct {
	ReferenceGroupCode     string `json:"reference_group_code"`
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
}

type CheckExistingIdentityResult struct {
	Exist bool `json:"exist"`
}

func (app *ABCIApplication) checkExistingIdentity(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("CheckExistingIdentity, Parameter: %s", param)
	var funcParam CheckExistingIdentityParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var result CheckExistingIdentityResult
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		return app.NewResponseQuery(returnValue, "Found reference group code and identity detail in parameter", app.state.Height)
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), true)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		if refGroupCodeFromDB == nil {
			returnValue, err := json.Marshal(result)
			if err != nil {
				return app.NewResponseQuery(nil, err.Error(), app.state.Height)
			}
			return app.NewResponseQuery(returnValue, "success", app.state.Height)
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		return app.NewResponseQuery(returnValue, "success", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		return app.NewResponseQuery(returnValue, "success", app.state.Height)
	}
	result.Exist = true
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetAccessorKeyParam struct {
	AccessorID string `json:"accessor_id"`
}

type GetAccessorKeyResult struct {
	AccessorPublicKey   string `json:"accessor_public_key"`
	AccessorType        string `json:"accessor_type"`
	Active              bool   `json:"active"`
	OwnerNodeID         string `json:"owner_node_id"`
	CreationBlockHeight int64  `json:"creation_block_height"`
	CreationChainID     string `json:"creation_chain_id"`
}

func (app *ABCIApplication) getAccessorKey(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetAccessorKey, Parameter: %s", param)
	var funcParam GetAccessorKeyParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAccessorKeyResult
	result.AccessorPublicKey = ""
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.AccessorPublicKey = accessor.AccessorPublicKey
				result.AccessorType = accessor.AccessorType
				result.Active = accessor.Active
				result.OwnerNodeID = accessor.Owner
				result.CreationBlockHeight = accessor.CreationBlockHeight
				result.CreationChainID = accessor.CreationChainId
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type CheckExistingAccessorIDParam struct {
	AccessorID string `json:"accessor_id"`
}

type CheckExistingResult struct {
	Exist bool `json:"exist"`
}

func (app *ABCIApplication) checkExistingAccessorID(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorID, Parameter: %s", param)
	var funcParam CheckExistingAccessorIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var result CheckExistingResult
	result.Exist = false
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.Exist = true
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetIdentityInfoParam struct {
	ReferenceGroupCode     string `json:"reference_group_code"`
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
	NodeID                 string `json:"node_id"`
}

type GetIdentityInfoResult struct {
	Ial      float64 `json:"ial"`
	Lial     *bool   `json:"lial"`
	Laal     *bool   `json:"laal"`
	ModeList []int32 `json:"mode_list"`
}

func (app *ABCIApplication) getIdentityInfo(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetIdentityInfo, Parameter: %s", param)
	var funcParam GetIdentityInfoParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var result GetIdentityInfoResult
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		return app.NewResponseQuery(returnValue, "Found reference group code and identity detail in parameter", app.state.Height)
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), true)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		if refGroupCodeFromDB == nil {
			returnValue, err := json.Marshal(result)
			if err != nil {
				return app.NewResponseQuery(nil, err.Error(), app.state.Height)
			}
			return app.NewResponseQuery(returnValue, "Reference group not found", app.state.Height)
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		return app.NewResponseQuery(returnValue, "Reference group not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		return app.NewResponseQuery(returnValue, "Reference group not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		if funcParam.NodeID == idp.NodeId && idp.Active {
			result.Ial = idp.Ial
			if idp.Lial != nil {
				result.Lial = &idp.Lial.Value
			}
			if idp.Laal != nil {
				result.Laal = &idp.Laal.Value
			}
			result.ModeList = idp.Mode
			break
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if result.Ial <= 0.0 {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetAccessorOwnerParam struct {
	AccessorID string `json:"accessor_id"`
}

type GetAccessorOwnerResult struct {
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) getAccessorOwner(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetAccessorOwner, Parameter: %s", param)
	var funcParam GetAccessorOwnerParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAccessorOwnerResult
	result.NodeID = ""
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.NodeID = idp.NodeId
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetReferenceGroupCodeParam struct {
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
}

type GetReferenceGroupCodeResult struct {
	ReferenceGroupCode string `json:"reference_group_code"`
}

func (app *ABCIApplication) GetReferenceGroupCode(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetReferenceGroupCode, Parameter: %s", param)
	var funcParam GetReferenceGroupCodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
	refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		refGroupCodeFromDB = []byte("")
	}
	var result GetReferenceGroupCodeResult
	result.ReferenceGroupCode = string(refGroupCodeFromDB)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if string(refGroupCodeFromDB) == "" {
		return app.NewResponseQuery(returnValue, "not found", app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

type GetReferenceGroupCodeByAccessorIDParam struct {
	AccessorID string `json:"accessor_id"`
}

func (app *ABCIApplication) GetReferenceGroupCodeByAccessorID(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetReferenceGroupCodeByAccessorID, Parameter: %s", param)
	var funcParam GetReferenceGroupCodeByAccessorIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		refGroupCodeFromDB = []byte("")
	}
	var result GetReferenceGroupCodeResult
	result.ReferenceGroupCode = string(refGroupCodeFromDB)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

//
// Identity related request operations
//

func (app *ABCIApplication) checkRequest(requestID string, purpose string, minIdp int) *abcitypes.ExecTxResult {
	requestKey := requestKeyPrefix + keySeparator + requestID
	requestValue, err := app.state.GetVersioned([]byte(requestKey), app.state.Height, true)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if requestValue == nil {
		return app.NewExecTxResult(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestValue), &request)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}
	if request.Purpose != purpose {
		return app.NewExecTxResult(code.InvalidPurpose, "Request has a invalid purpose", "")
	}
	if request.UseCount > 0 {
		return app.NewExecTxResult(code.RequestIsAlreadyUsed, "Request is already used", "")
	}
	if !request.Closed {
		return app.NewExecTxResult(code.RequestIsNotClosed, "Request is not closed", "")
	}
	var acceptCount int = 0
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
		return app.NewExecTxResult(code.OK, "Request is completed", "")
	}
	return app.NewExecTxResult(code.RequestIsNotCompleted, "Request is not completed", "")
}

func (app *ABCIApplication) checkRequestUsable(requestID string, purpose string, minIdp int, committedState bool) error {
	requestKey := requestKeyPrefix + keySeparator + requestID
	requestValue, err := app.state.GetVersioned([]byte(requestKey), app.state.Height, committedState)
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
	if request.Purpose != purpose {
		return &ApplicationError{
			Code:    code.InvalidPurpose,
			Message: "Request has a invalid purpose",
		}
	}
	if request.UseCount > 0 {
		return &ApplicationError{
			Code:    code.RequestIsAlreadyUsed,
			Message: "Request is already used",
		}
	}
	if !request.Closed {
		return &ApplicationError{
			Code:    code.RequestIsNotClosed,
			Message: "Request is not closed",
		}
	}
	var acceptCount int = 0
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
	if acceptCount < minIdp {
		return &ApplicationError{
			Code:    code.RequestIsNotCompleted,
			Message: "Request is not completed",
		}
	}
	return nil
}

func (app *ABCIApplication) increaseRequestUseCount(requestID string) *abcitypes.ExecTxResult {
	requestKey := requestKeyPrefix + keySeparator + requestID
	requestValue, err := app.state.GetVersioned([]byte(requestKey), app.state.Height, true)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if requestValue == nil {
		return app.NewExecTxResult(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request data.Request
	err = proto.Unmarshal([]byte(requestValue), &request)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}
	request.UseCount = request.UseCount + 1
	requestProtobuf, err := utils.ProtoDeterministicMarshal(&request)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	err = app.state.SetVersioned([]byte(requestKey), []byte(requestProtobuf))
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	return app.NewExecTxResult(code.OK, "success", "")
}

type GetAllowedMinIalForRegisterIdentityAtFirstIdpResult struct {
	MinIal float64 `json:"min_ial"`
}

func (app *ABCIApplication) GetAllowedMinIalForRegisterIdentityAtFirstIdp(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetAllowedMinIalForRegisterIdentityAtFirstIdp, Parameter: %s", param)
	var result GetAllowedMinIalForRegisterIdentityAtFirstIdpResult
	result.MinIal = app.GetAllowedMinIalForRegisterIdentityAtFirstIdpFromStateDB(true)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetAllowedMinIalForRegisterIdentityAtFirstIdpFromStateDB(committedState bool) float64 {
	allowedMinIalKey := "AllowedMinIalForRegisterIdentityAtFirstIdp"
	var allowedMinIal data.AllowedMinIalForRegisterIdentityAtFirstIdp
	allowedMinIalValue, err := app.state.Get([]byte(allowedMinIalKey), committedState)
	if err != nil {
		return 0
	}
	if allowedMinIalValue == nil {
		return 0
	}
	err = proto.Unmarshal(allowedMinIalValue, &allowedMinIal)
	if err != nil {
		return 0
	}
	return allowedMinIal.MinIal
}

//
