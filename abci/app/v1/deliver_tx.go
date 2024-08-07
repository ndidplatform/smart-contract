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
	abcitypes "github.com/cometbft/cometbft/abci/types"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
)

// app.NewExecTxResult return *abcitypes.ExecTxResult
func (app *ABCIApplication) NewExecTxResult(code uint32, log string, extraData string) *abcitypes.ExecTxResult {
	var attributes []abcitypes.EventAttribute
	if code == 0 {
		var attribute abcitypes.EventAttribute
		attribute.Key = "success"
		attribute.Value = "true"
		attributes = append(attributes, attribute)
	} else {
		var attribute abcitypes.EventAttribute
		attribute.Key = "success"
		attribute.Value = "false"
		attributes = append(attributes, attribute)
	}
	var events []abcitypes.Event
	event := abcitypes.Event{
		Type:       "did.result",
		Attributes: attributes,
	}
	events = append(events, event)
	return &abcitypes.ExecTxResult{
		Code:   code,
		Log:    log,
		Data:   []byte(extraData),
		Events: events,
	}
}

func (app *ABCIApplication) NewExecTxResultWithAttributes(code uint32, log string, additionalAttributes []abcitypes.EventAttribute) *abcitypes.ExecTxResult {
	var attributes []abcitypes.EventAttribute
	if code == 0 {
		var attribute abcitypes.EventAttribute
		attribute.Key = "success"
		attribute.Value = "true"
		attributes = append(attributes, attribute)
	} else {
		var attribute abcitypes.EventAttribute
		attribute.Key = "success"
		attribute.Value = "false"
		attributes = append(attributes, attribute)
	}
	attributes = append(attributes, additionalAttributes...)
	var events []abcitypes.Event
	event := abcitypes.Event{
		Type:       "did.result",
		Attributes: attributes,
	}
	events = append(events, event)
	return &abcitypes.ExecTxResult{
		Code:   code,
		Log:    log,
		Data:   []byte(""),
		Events: events,
	}
}

// DeliverTxRouter is Pointer to function
func (app *ABCIApplication) DeliverTxRouter(method string, param []byte, nonce []byte, signature []byte, nodeID string) *abcitypes.ExecTxResult {
	err := app.commonValidate(method, param, nonce, signature, nodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	result := app.callDeliverTx(method, param, nodeID)
	// ---- Burn token ----
	ndidNode, err := app.isNDIDNodeByNodeID(nodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}
	if !ndidNode && !regulatorMethod[method] {
		needToken := app.getTokenPriceByFunc(method, false)
		err := app.reduceToken(nodeID, needToken)
		if err != nil {
			if appErr, ok := err.(*ApplicationError); ok {
				return app.NewExecTxResult(appErr.Code, appErr.Message, "")
			}
			return app.NewExecTxResult(code.UnknownError, err.Error(), "")
		}
	}

	if mustCheckNodeSignature(method) {
		// Set used nonce to stateDB
		nonceKey := append([]byte(nonceKeyPrefix+keySeparator), nonce...)
		emptyValue := make([]byte, 0)
		app.state.Set(nonceKey, emptyValue)
		nonceStr := string(nonce)
		app.deliverTxNonceState[nonceStr] = []byte(nil)
	}

	return result
}

func (app *ABCIApplication) callDeliverTx(name string, param []byte, nodeID string) *abcitypes.ExecTxResult {
	switch name {
	case "InitNDID":
		return app.initNDID(param, nodeID)
	case "SetInitData":
		return app.setInitData(param, nodeID)
	case "SetInitData_pb":
		return app.setInitData_pb(param, nodeID)
	case "EndInit":
		return app.endInit(param, nodeID)
	case "SetLastBlock":
		return app.setLastBlock(param, nodeID)
	case "SetValidator":
		return app.setValidator(param, nodeID)
	case "SetPriceFunc":
		return app.setPriceFunc(param, nodeID)

	case "RegisterNode":
		return app.registerNode(param, nodeID)
	case "UpdateNodeByNDID":
		return app.updateNodeByNDID(param, nodeID)
	case "DisableNode":
		return app.disableNode(param, nodeID)
	case "EnableNode":
		return app.enableNode(param, nodeID)
	case "AddNodeToProxyNode":
		return app.addNodeToProxyNode(param, nodeID)
	case "UpdateNodeProxyNode":
		return app.updateNodeProxyNode(param, nodeID)
	case "RemoveNodeFromProxyNode":
		return app.removeNodeFromProxyNode(param, nodeID)
	case "AddNodeToken":
		return app.addNodeToken(param, nodeID)
	case "ReduceNodeToken":
		return app.reduceNodeToken(param, nodeID)
	case "SetNodeToken":
		return app.setNodeToken(param, nodeID)
	case "AddAllowedNodeSupportedFeature":
		return app.addAllowedNodeSupportedFeature(param, nodeID)
	case "RemoveAllowedNodeSupportedFeature":
		return app.removeAllowedNodeSupportedFeature(param, nodeID)

	case "AddNamespace":
		return app.addNamespace(param, nodeID)
	case "DisableNamespace":
		return app.disableNamespace(param, nodeID)
	case "EnableNamespace":
		return app.enableNamespace(param, nodeID)
	case "UpdateNamespace":
		return app.updateNamespace(param, nodeID)
	case "SetAllowedMinIalForRegisterIdentityAtFirstIdp":
		return app.setAllowedMinIalForRegisterIdentityAtFirstIdp(param, nodeID)
	case "SetTimeOutBlockRegisterIdentity":
		return app.setTimeOutBlockRegisterIdentity(param, nodeID)
	case "AddSuppressedIdentityModificationNotificationNode":
		return app.addSuppressedIdentityModificationNotificationNode(param, nodeID)
	case "RemoveSuppressedIdentityModificationNotificationNode":
		return app.removeSuppressedIdentityModificationNotificationNode(param, nodeID)

	case "AddService":
		return app.addService(param, nodeID)
	case "DisableService":
		return app.disableService(param, nodeID)
	case "EnableService":
		return app.enableService(param, nodeID)
	case "UpdateService":
		return app.updateService(param, nodeID)
	case "RegisterServiceDestinationByNDID":
		return app.registerServiceDestinationByNDID(param, nodeID)
	case "DisableServiceDestinationByNDID":
		return app.disableServiceDestinationByNDID(param, nodeID)
	case "EnableServiceDestinationByNDID":
		return app.enableServiceDestinationByNDID(param, nodeID)
	case "SetServicePriceCeiling":
		return app.setServicePriceCeiling(param, nodeID)
	case "SetServicePriceMinEffectiveDatetimeDelay":
		return app.setServicePriceMinEffectiveDatetimeDelay(param, nodeID)

	case "SetSupportedIALList":
		return app.setSupportedIALList(param, nodeID)
	case "SetSupportedAALList":
		return app.setSupportedAALList(param, nodeID)
	case "SetAllowedModeList":
		return app.setAllowedModeList(param, nodeID)
	case "AddRequestType":
		return app.addRequestType(param, nodeID)
	case "RemoveRequestType":
		return app.removeRequestType(param, nodeID)

	case "AddErrorCode":
		return app.addErrorCode(param, nodeID)
	case "RemoveErrorCode":
		return app.removeErrorCode(param, nodeID)

	case "UpdateNode":
		return app.updateNode(param, nodeID)
	case "SetMqAddresses":
		return app.setMqAddresses(param, nodeID)

	case "RegisterIdentity":
		return app.registerIdentity(param, nodeID)
	case "UpdateIdentity":
		return app.updateIdentity(param, nodeID)
	case "AddIdentity":
		return app.addIdentity(param, nodeID)
	case "RevokeIdentityAssociation":
		return app.revokeIdentityAssociation(param, nodeID)
	case "AddAccessor":
		return app.addAccessor(param, nodeID)
	case "RevokeAccessor":
		return app.revokeAccessor(param, nodeID)
	case "RevokeAndAddAccessor":
		return app.revokeAndAddAccessor(param, nodeID)
	case "UpdateIdentityModeList":
		return app.updateIdentityModeList(param, nodeID)

	case "RegisterServiceDestination":
		return app.registerServiceDestination(param, nodeID)
	case "UpdateServiceDestination":
		return app.updateServiceDestination(param, nodeID)
	case "DisableServiceDestination":
		return app.disableServiceDestination(param, nodeID)
	case "EnableServiceDestination":
		return app.enableServiceDestination(param, nodeID)
	case "SetServicePrice":
		return app.setServicePrice(param, nodeID)

	case "CreateRequest":
		return app.createRequest(param, nodeID)
	case "CreateIdpResponse":
		return app.createIdpResponse(param, nodeID)
	case "CreateAsResponse":
		return app.createAsResponse(param, nodeID)
	case "SetDataReceived":
		return app.setDataReceived(param, nodeID)
	case "CloseRequest":
		return app.closeRequest(param, nodeID)
	case "TimeOutRequest":
		return app.timeOutRequest(param, nodeID)

	case "CreateMessage":
		return app.createMessage(param, nodeID)

	default:
		return &abcitypes.ExecTxResult{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}
