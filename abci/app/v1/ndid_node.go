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
	"strconv"
	"strings"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	appTypes "github.com/ndidplatform/smart-contract/v9/abci/app/v1/types"
	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type RegisterNodeParam struct {
	NodeID                 string   `json:"node_id"`
	SigningPublicKey       string   `json:"signing_public_key"`
	SigningAlgorithm       string   `json:"signing_algorithm"`
	SigningMasterPublicKey string   `json:"signing_master_public_key"`
	SigningMasterAlgorithm string   `json:"signing_master_algorithm"`
	EncryptionPublicKey    string   `json:"encryption_public_key"`
	EncryptionAlgorithm    string   `json:"encryption_algorithm"`
	NodeName               string   `json:"node_name"`
	Role                   string   `json:"role"`
	MaxIal                 float64  `json:"max_ial"` // IdP only attribute
	MaxAal                 float64  `json:"max_aal"` // IdP only attribute
	SupportedFeatureList   []string `json:"supported_feature_list"`
	IsIdPAgent             *bool    `json:"agent"` // IdP only attribute
	UseWhitelist           *bool    `json:"node_id_whitelist_active"`
	Whitelist              []string `json:"node_id_whitelist"`
}

func (app *ABCIApplication) validateRegisterNode(funcParam RegisterNodeParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	// Validate master public key format
	err = checkPubKeyForSigning(
		funcParam.SigningMasterPublicKey,
		appTypes.SignatureAlgorithm(funcParam.SigningMasterAlgorithm),
	)
	if err != nil {
		return err
	}

	// Validate public key format
	err = checkPubKeyForSigning(
		funcParam.SigningPublicKey,
		appTypes.SignatureAlgorithm(funcParam.SigningAlgorithm),
	)
	if err != nil {
		return err
	}

	// Validate encryption public key format
	err = checkPubKeyForEncryption(funcParam.EncryptionPublicKey)
	if err != nil {
		return err
	}

	key := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	// check Duplicate Node ID
	chkExists, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if chkExists != nil {
		return &ApplicationError{
			Code:    code.DuplicateNodeID,
			Message: "Duplicate Node ID",
		}
	}

	// check role is valid
	if !(strings.EqualFold(funcParam.Role, string(appTypes.NodeRoleRp)) ||
		strings.EqualFold(funcParam.Role, string(appTypes.NodeRoleIdp)) ||
		strings.EqualFold(funcParam.Role, string(appTypes.NodeRoleAs)) ||
		strings.EqualFold(funcParam.Role, string(appTypes.NodeRoleProxy))) {
		return &ApplicationError{
			Code:    code.InvalidNodeRole,
			Message: "Invalid node role",
		}
	}

	// check if supported feature are valid/allowed
	for _, supportedFeature := range funcParam.SupportedFeatureList {
		key := nodeSupportedFeatureKeyPrefix + keySeparator + supportedFeature
		exists, err := app.state.Has([]byte(key), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if !exists {
			return &ApplicationError{
				Code:    code.NodeSupportedFeatureDoesNotExist,
				Message: "invalid node supported feature",
			}
		}
	}

	// if node is Idp or rp, set use_whitelist and whitelist
	if strings.EqualFold(funcParam.Role, string(appTypes.NodeRoleRp)) ||
		strings.EqualFold(funcParam.Role, string(appTypes.NodeRoleIdp)) {
		if funcParam.Whitelist != nil {
			// check if all node in whitelist exists
			for _, whitelistNode := range funcParam.Whitelist {
				whitelistKey := nodeIDKeyPrefix + keySeparator + whitelistNode
				hasWhitelistKey, err := app.state.Has([]byte(whitelistKey), committedState)
				if err != nil {
					return &ApplicationError{
						Code:    code.AppStateError,
						Message: err.Error(),
					}
				}
				if !hasWhitelistKey {
					return &ApplicationError{
						Code:    code.NodeIDNotFound,
						Message: "Whitelist node not exist",
					}
				}
			}
		}
	}

	return nil
}

func (app *ABCIApplication) registerNodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RegisterNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRegisterNode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) registerNode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RegisterNode, Parameter: %s", param)
	var funcParam RegisterNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRegisterNode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// create node detail
	var nodeDetail data.NodeDetail
	nodeDetail.SigningPublicKey = &data.NodeKey{
		PublicKey:           funcParam.SigningPublicKey,
		Algorithm:           funcParam.SigningAlgorithm,
		Version:             1,
		CreationBlockHeight: app.state.CurrentBlockHeight,
		CreationChainId:     app.CurrentChain,
		Active:              true,
	}
	// create key version history
	nodeKeyKey :=
		nodeKeyKeyPrefix + keySeparator +
			"signing" + keySeparator +
			funcParam.NodeID + keySeparator +
			strconv.FormatInt(nodeDetail.SigningPublicKey.Version, 10)
	nodeKeyValue, err := utils.ProtoDeterministicMarshal(nodeDetail.SigningPublicKey)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

	nodeDetail.SigningMasterPublicKey = &data.NodeKey{
		PublicKey:           funcParam.SigningMasterPublicKey,
		Algorithm:           funcParam.SigningMasterAlgorithm,
		Version:             1,
		CreationBlockHeight: app.state.CurrentBlockHeight,
		CreationChainId:     app.CurrentChain,
		Active:              true,
	}
	// create key version history
	nodeKeyKey =
		nodeKeyKeyPrefix + keySeparator +
			"signing_master" + keySeparator +
			funcParam.NodeID + keySeparator +
			strconv.FormatInt(nodeDetail.SigningMasterPublicKey.Version, 10)
	nodeKeyValue, err = utils.ProtoDeterministicMarshal(nodeDetail.SigningMasterPublicKey)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

	nodeDetail.EncryptionPublicKey = &data.NodeKey{
		PublicKey:           funcParam.EncryptionPublicKey,
		Algorithm:           funcParam.EncryptionAlgorithm,
		Version:             1,
		CreationBlockHeight: app.state.CurrentBlockHeight,
		CreationChainId:     app.CurrentChain,
		Active:              true,
	}
	// create key version history
	nodeKeyKey =
		nodeKeyKeyPrefix + keySeparator +
			"encryption" + keySeparator +
			funcParam.NodeID + keySeparator +
			strconv.FormatInt(nodeDetail.EncryptionPublicKey.Version, 10)
	nodeKeyValue, err = utils.ProtoDeterministicMarshal(nodeDetail.EncryptionPublicKey)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeKeyKey), []byte(nodeKeyValue))

	nodeDetail.NodeName = funcParam.NodeName

	switch strings.ToLower(funcParam.Role) {
	case "rp":
		nodeDetail.Role = string(appTypes.NodeRoleRp)
	case "idp":
		nodeDetail.Role = string(appTypes.NodeRoleIdp)
	case "as":
		nodeDetail.Role = string(appTypes.NodeRoleAs)
	case "proxy":
		nodeDetail.Role = string(appTypes.NodeRoleProxy)
	}

	nodeDetail.Active = true

	nodeDetail.SupportedFeatureList = funcParam.SupportedFeatureList

	// if node is IdP, set max_aal, min_ial, on_the_fly_support, is_idp_agent, and supported_request_message_type_list
	if appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleIdp {
		nodeDetail.MaxAal = funcParam.MaxAal
		nodeDetail.MaxIal = funcParam.MaxIal
		nodeDetail.IsIdpAgent = funcParam.IsIdPAgent != nil && *funcParam.IsIdPAgent
		nodeDetail.SupportedRequestMessageDataUrlTypeList = make([]string, 0)
	}
	// if node is Idp or rp, set use_whitelist and whitelist
	if appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleIdp ||
		appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleRp {
		if funcParam.UseWhitelist != nil && *funcParam.UseWhitelist {
			nodeDetail.UseWhitelist = true
		} else {
			nodeDetail.UseWhitelist = false
		}
		if funcParam.Whitelist != nil {
			nodeDetail.Whitelist = funcParam.Whitelist
		} else {
			nodeDetail.Whitelist = []string{}
		}
	}

	// if node is IdP, add node id to IdPList
	if appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleIdp {
		var idpsList data.IdPList
		idpsKey := "IdPList"
		idpsValue, err := app.state.Get([]byte(idpsKey), false)
		if err != nil {
			return app.NewExecTxResult(code.AppStateError, err.Error(), "")
		}
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
			}
		}
		idpsList.NodeId = append(idpsList.NodeId, funcParam.NodeID)
		idpsListByte, err := utils.ProtoDeterministicMarshal(&idpsList)
		if err != nil {
			return app.NewExecTxResult(code.MarshalError, err.Error(), "")
		}
		app.state.Set(idpListKeyBytes, []byte(idpsListByte))
	}

	// if node is rp, add node id to rpList
	if appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleRp {
		var rpsList data.RPList
		rpsKey := "rpList"
		rpsValue, err := app.state.Get([]byte(rpsKey), false)
		if err != nil {
			return app.NewExecTxResult(code.AppStateError, err.Error(), "")
		}
		if rpsValue != nil {
			err := proto.Unmarshal(rpsValue, &rpsList)
			if err != nil {
				return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
			}
		}
		rpsList.NodeId = append(rpsList.NodeId, funcParam.NodeID)
		rpsListByte, err := utils.ProtoDeterministicMarshal(&rpsList)
		if err != nil {
			return app.NewExecTxResult(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(rpsKey), []byte(rpsListByte))
	}

	// if node is as, add node id to asList
	if appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleAs {
		var asList data.ASList
		asKey := "asList"
		asValue, err := app.state.Get([]byte(asKey), false)
		if err != nil {
			return app.NewExecTxResult(code.AppStateError, err.Error(), "")
		}
		if asValue != nil {
			err := proto.Unmarshal(asValue, &asList)
			if err != nil {
				return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
			}
		}
		asList.NodeId = append(asList.NodeId, funcParam.NodeID)
		asListByte, err := utils.ProtoDeterministicMarshal(&asList)
		if err != nil {
			return app.NewExecTxResult(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(asKey), []byte(asListByte))
	}

	var allList data.AllList
	allKey := "allList"
	allValue, err := app.state.Get([]byte(allKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	if allValue != nil {
		err := proto.Unmarshal(allValue, &allList)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	}
	allList.NodeId = append(allList.NodeId, funcParam.NodeID)
	allListByte, err := utils.ProtoDeterministicMarshal(&allList)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(allKey), []byte(allListByte))

	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.createTokenAccount(funcParam.NodeID)

	return app.NewExecTxResult(code.OK, "success", "")
}

type UpdateNodeByNDIDParam struct {
	NodeID               string   `json:"node_id"`
	MaxIal               float64  `json:"max_ial"`
	MaxAal               float64  `json:"max_aal"`
	SupportedFeatureList []string `json:"supported_feature_list"`
	NodeName             string   `json:"node_name"`
	IsIdPAgent           *bool    `json:"agent"`
	UseWhitelist         *bool    `json:"node_id_whitelist_active"`
	Whitelist            []string `json:"node_id_whitelist"`
}

func (app *ABCIApplication) validateUpdateNodeByNDID(funcParam UpdateNodeByNDIDParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
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
	var node data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &node)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	if funcParam.SupportedFeatureList != nil {
		// check if supported feature are valid/allowed
		for _, supportedFeature := range funcParam.SupportedFeatureList {
			key := nodeSupportedFeatureKeyPrefix + keySeparator + supportedFeature
			exists, err := app.state.Has([]byte(key), committedState)
			if err != nil {
				return &ApplicationError{
					Code:    code.AppStateError,
					Message: err.Error(),
				}
			}
			if !exists {
				return &ApplicationError{
					Code:    code.NodeSupportedFeatureDoesNotExist,
					Message: "invalid node supported feature",
				}
			}
		}
	}

	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleIdp ||
		appTypes.NodeRole(node.Role) == appTypes.NodeRoleRp {
		if funcParam.Whitelist != nil {
			// check if all node in whitelist exists
			for _, whitelistNode := range funcParam.Whitelist {
				whitelistKey := nodeIDKeyPrefix + keySeparator + whitelistNode
				hasWhitelistKey, err := app.state.Has([]byte(whitelistKey), committedState)
				if err != nil {
					return &ApplicationError{
						Code:    code.AppStateError,
						Message: err.Error(),
					}
				}
				if !hasWhitelistKey {
					return &ApplicationError{
						Code:    code.NodeIDNotFound,
						Message: "Whitelist node does not exist",
					}
				}
			}
			node.Whitelist = funcParam.Whitelist
		}
	}

	return nil
}

func (app *ABCIApplication) updateNodeByNDIDCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam UpdateNodeByNDIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateUpdateNodeByNDID(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) updateNodeByNDID(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("UpdateNodeByNDID, Parameter: %s", param)
	var funcParam UpdateNodeByNDIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateUpdateNodeByNDID(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var node data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &node)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}
	// Selective update
	if funcParam.NodeName != "" {
		node.NodeName = funcParam.NodeName
	}
	if funcParam.SupportedFeatureList != nil {
		node.SupportedFeatureList = funcParam.SupportedFeatureList
	}
	// If node is IdP then update max_ial, max_aal and is_idp_agent
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleIdp {
		if funcParam.MaxIal > 0 {
			node.MaxIal = funcParam.MaxIal
		}
		if funcParam.MaxAal > 0 {
			node.MaxAal = funcParam.MaxAal
		}
		if funcParam.IsIdPAgent != nil {
			node.IsIdpAgent = *funcParam.IsIdPAgent
		}
	}
	// If node is Idp or rp, update use_whitelist and whitelist
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleIdp ||
		appTypes.NodeRole(node.Role) == appTypes.NodeRoleRp {
		if funcParam.UseWhitelist != nil {
			node.UseWhitelist = *funcParam.UseWhitelist
		}
		if funcParam.Whitelist != nil {
			node.Whitelist = funcParam.Whitelist
		}
	}
	nodeDetailValue, err = utils.ProtoDeterministicMarshal(&node)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailValue))

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableNodeParam struct {
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateDisableNode(funcParam DisableNodeParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
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

	return nil
}

func (app *ABCIApplication) disableNodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableNode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableNode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableNode, Parameter: %s", param)
	var funcParam DisableNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableNode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}
	nodeDetail.Active = false
	nodeDetailValue, err = utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailValue))

	return app.NewExecTxResult(code.OK, "success", "")
}

type EnableNodeParam struct {
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateEnableNode(funcParam EnableNodeParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
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

	return nil
}

func (app *ABCIApplication) enableNodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableNode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableNode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableNode, Parameter: %s", param)
	var funcParam EnableNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableNode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}
	nodeDetail.Active = true
	nodeDetailValue, err = utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailValue))

	return app.NewExecTxResult(code.OK, "success", "")
}

type AddNodeToProxyNodeParam struct {
	NodeID      string `json:"node_id"`
	ProxyNodeID string `json:"proxy_node_id"`
	Config      string `json:"config"`
}

func (app *ABCIApplication) validateAddNodeToProxyNode(funcParam AddNodeToProxyNodeParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	// Check already associated with a proxy
	if nodeDetail.ProxyNodeId != "" {
		return &ApplicationError{
			Code:    code.NodeIDIsAlreadyAssociatedWithProxyNode,
			Message: "This node ID is already associated with a proxy node",
		}
	}
	// Check is not proxy node
	nodeProxyNode, err := app.isProxyNodeByNodeID(funcParam.NodeID, committedState)
	if err != nil {
		return err
	}
	if nodeProxyNode {
		return &ApplicationError{
			Code:    code.NodeIDisProxyNode,
			Message: "This node ID is an ID of a proxy node",
		}
	}
	// Check ProxyNodeID is proxy node
	proxyNodeProxyNode, err := app.isProxyNodeByNodeID(funcParam.ProxyNodeID, committedState)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok && appErr.Code == code.NodeIDNotFound {
			return &ApplicationError{
				Code:    code.ProxyNodeNotFound,
				Message: "Proxy node ID not found",
			}
		}
		return err
	}
	if !proxyNodeProxyNode {
		return &ApplicationError{
			Code:    code.ProxyNodeNotFound,
			Message: "Proxy node ID not found",
		}
	}

	return nil
}

func (app *ABCIApplication) addNodeToProxyNodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddNodeToProxyNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddNodeToProxyNode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addNodeToProxyNode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddNodeToProxyNode, Parameter: %s", param)
	var funcParam AddNodeToProxyNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddNodeToProxyNode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	behindProxyNodeKey := behindProxyNodeKeyPrefix + keySeparator + funcParam.ProxyNodeID
	behindProxyNodeValue, err := app.state.Get([]byte(behindProxyNodeKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var nodes data.BehindNodeList
	if behindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	} else {
		nodes.Nodes = make([]string, 0)
	}

	// Set proxy node ID and proxy config
	nodeDetail.ProxyNodeId = funcParam.ProxyNodeID
	nodeDetail.ProxyConfig = funcParam.Config

	nodes.Nodes = append(nodes.Nodes, funcParam.NodeID)
	behindProxyNodeValue, err = utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	// Delete mq address
	msqAddres := make([]*data.MQ, 0)
	nodeDetail.Mq = msqAddres
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set([]byte(behindProxyNodeKey), []byte(behindProxyNodeValue))

	return app.NewExecTxResult(code.OK, "success", "")
}

type UpdateNodeProxyNodeParam struct {
	NodeID      string `json:"node_id"`
	ProxyNodeID string `json:"proxy_node_id"`
	Config      string `json:"config"`
}

func (app *ABCIApplication) validateUpdateNodeProxyNode(funcParam UpdateNodeProxyNodeParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	// Check already associated with a proxy
	if nodeDetail.ProxyNodeId == "" {
		return &ApplicationError{
			Code:    code.NodeIDHasNotBeenAssociatedWithProxyNode,
			Message: "This node has not been associated with a proxy node",
		}
	}
	if funcParam.ProxyNodeID != "" {
		// Check ProxyNodeID is proxy node
		proxyNodeProxyNode, err := app.isProxyNodeByNodeID(funcParam.ProxyNodeID, committedState)
		if err != nil {
			if appErr, ok := err.(*ApplicationError); ok && appErr.Code == code.NodeIDNotFound {
				return &ApplicationError{
					Code:    code.ProxyNodeNotFound,
					Message: "Proxy node ID not found",
				}
			}
			return err
		}
		if !proxyNodeProxyNode {
			return &ApplicationError{
				Code:    code.ProxyNodeNotFound,
				Message: "Proxy node ID not found",
			}
		}
	}

	return nil
}

func (app *ABCIApplication) updateNodeProxyNodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam UpdateNodeProxyNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateUpdateNodeProxyNode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) updateNodeProxyNode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("UpdateNodeProxyNode, Parameter: %s", param)
	var funcParam UpdateNodeProxyNodeParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateUpdateNodeProxyNode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	behindProxyNodeKey := behindProxyNodeKeyPrefix + keySeparator + nodeDetail.ProxyNodeId
	behindProxyNodeValue, err := app.state.Get([]byte(behindProxyNodeKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var nodes data.BehindNodeList
	if behindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	} else {
		nodes.Nodes = make([]string, 0)
	}

	newBehindProxyNodeKey := behindProxyNodeKeyPrefix + keySeparator + funcParam.ProxyNodeID
	newBehindProxyNodeValue, err := app.state.Get([]byte(newBehindProxyNodeKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var newProxyNodes data.BehindNodeList
	if newBehindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(newBehindProxyNodeValue), &newProxyNodes)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	} else {
		newProxyNodes.Nodes = make([]string, 0)
	}
	if funcParam.ProxyNodeID != "" {
		if nodeDetail.ProxyNodeId != funcParam.ProxyNodeID {
			// Delete from old proxy list
			for i, node := range nodes.Nodes {
				if node == funcParam.NodeID {
					copy(nodes.Nodes[i:], nodes.Nodes[i+1:])
					nodes.Nodes[len(nodes.Nodes)-1] = ""
					nodes.Nodes = nodes.Nodes[:len(nodes.Nodes)-1]
				}
			}
			// Add to new proxy list
			newProxyNodes.Nodes = append(newProxyNodes.Nodes, funcParam.NodeID)
		}
		nodeDetail.ProxyNodeId = funcParam.ProxyNodeID
	}
	if funcParam.Config != "" {
		nodeDetail.ProxyConfig = funcParam.Config
	}
	behindProxyNodeValue, err = utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	newBehindProxyNodeValue, err = utils.ProtoDeterministicMarshal(&newProxyNodes)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set([]byte(behindProxyNodeKey), []byte(behindProxyNodeValue))
	app.state.Set([]byte(newBehindProxyNodeKey), []byte(newBehindProxyNodeValue))

	return app.NewExecTxResult(code.OK, "success", "")
}

type RemoveNodeFromProxyNode struct {
	NodeID string `json:"node_id"`
}

func (app *ABCIApplication) validateRemoveNodeFromProxyNode(funcParam RemoveNodeFromProxyNode, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	// If node not found then return code.NodeIDNotFound
	if nodeDetailValue == nil {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	// Check is not proxy node
	nodeProxyNode, err := app.isProxyNodeByNodeID(funcParam.NodeID, committedState)
	if err != nil {
		return err
	}
	if nodeProxyNode {
		return &ApplicationError{
			Code:    code.NodeIDisProxyNode,
			Message: "This node ID is an ID of a proxy node",
		}
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	// Check already associated with a proxy
	if nodeDetail.ProxyNodeId == "" {
		return &ApplicationError{
			Code:    code.NodeIDHasNotBeenAssociatedWithProxyNode,
			Message: "This node has not been associated with a proxy node",
		}
	}

	return nil
}

func (app *ABCIApplication) removeNodeFromProxyNodeCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RemoveNodeFromProxyNode
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRemoveNodeFromProxyNode(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) removeNodeFromProxyNode(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RemoveNodeFromProxyNode, Parameter: %s", param)
	var funcParam RemoveNodeFromProxyNode
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRemoveNodeFromProxyNode(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Get node detail by NodeID
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	// Unmarshal node detail
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	behindProxyNodeKey := behindProxyNodeKeyPrefix + keySeparator + nodeDetail.ProxyNodeId
	behindProxyNodeValue, err := app.state.Get([]byte(behindProxyNodeKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var nodes data.BehindNodeList
	if behindProxyNodeValue != nil {
		err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
		// Delete from old proxy list
		for i, node := range nodes.Nodes {
			if node == funcParam.NodeID {
				copy(nodes.Nodes[i:], nodes.Nodes[i+1:])
				nodes.Nodes[len(nodes.Nodes)-1] = ""
				nodes.Nodes = nodes.Nodes[:len(nodes.Nodes)-1]
			}
		}
	} else {
		nodes.Nodes = make([]string, 0)
	}
	// Delete node proxy ID and proxy config
	nodeDetail.ProxyNodeId = ""
	nodeDetail.ProxyConfig = ""
	behindProxyNodeValue, err = utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	app.state.Set([]byte(behindProxyNodeKey), []byte(behindProxyNodeValue))

	return app.NewExecTxResult(code.OK, "success", "")
}
