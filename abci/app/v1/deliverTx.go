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
	"fmt"

	"github.com/tendermint/tendermint/abci/types"
	kv "github.com/tendermint/tendermint/libs/kv"

	"github.com/ndidplatform/smart-contract/v5/abci/code"
)

// app.ReturnDeliverTxLog return types.ResponseDeliverTx
func (app *ABCIApplication) ReturnDeliverTxLog(code uint32, log string, extraData string) types.ResponseDeliverTx {
	var attributes []kv.Pair
	if code == 0 {
		var attribute kv.Pair
		attribute.Key = []byte("success")
		attribute.Value = []byte("true")
		attributes = append(attributes, attribute)
	} else {
		var attribute kv.Pair
		attribute.Key = []byte("success")
		attribute.Value = []byte("false")
		attributes = append(attributes, attribute)
	}
	var events []types.Event
	event := types.Event{
		Type:       "did.result",
		Attributes: attributes,
	}
	events = append(events, event)
	return types.ResponseDeliverTx{
		Code:   code,
		Log:    fmt.Sprintf(log),
		Data:   []byte(extraData),
		Events: events,
	}
}

func (app *ABCIApplication) ReturnDeliverTxLogWithAttributes(code uint32, log string, additionalAttributes []kv.Pair) types.ResponseDeliverTx {
	var attributes []kv.Pair
	if code == 0 {
		var attribute kv.Pair
		attribute.Key = []byte("success")
		attribute.Value = []byte("true")
		attributes = append(attributes, attribute)
	} else {
		var attribute kv.Pair
		attribute.Key = []byte("success")
		attribute.Value = []byte("false")
		attributes = append(attributes, attribute)
	}
	attributes = append(attributes, additionalAttributes...)
	var events []types.Event
	event := types.Event{
		Type:       "did.result",
		Attributes: attributes,
	}
	events = append(events, event)
	return types.ResponseDeliverTx{
		Code:   code,
		Log:    fmt.Sprintf(log),
		Data:   []byte(""),
		Events: events,
	}
}

// DeliverTxRouter is Pointer to function
func (app *ABCIApplication) DeliverTxRouter(method string, param string, nonce []byte, signature []byte, nodeID string) types.ResponseDeliverTx {
	// ---- check authorization ----
	checkTxResult := app.CheckTxRouter(method, param, nonce, signature, nodeID, false)
	if checkTxResult.Code != code.OK {
		if checkTxResult.Log != "" {
			return app.ReturnDeliverTxLog(checkTxResult.Code, checkTxResult.Log, "")
		}
		return app.ReturnDeliverTxLog(checkTxResult.Code, "Unauthorized", "")
	}

	result := app.callDeliverTx(method, param, nodeID)
	// ---- Burn token ----
	if !app.isNDIDNode(param, nodeID, false) && !isNDIDMethod[method] {
		needToken := app.getTokenPriceByFunc(method, false)
		errCode, errLog := app.reduceToken(nodeID, needToken)
		if errCode != code.OK {
			result.Code = errCode
			result.Log = errLog
		}
	}

	// Set used nonce to stateDB
	emptyValue := make([]byte, 0)
	app.state.Set([]byte(nonce), emptyValue)
	nonceStr := string(nonce)
	app.deliverTxNonceState[nonceStr] = []byte(nil)
	return result
}

func (app *ABCIApplication) callDeliverTx(name string, param string, nodeID string) types.ResponseDeliverTx {
	switch name {
	case "InitNDID":
		return app.initNDID(param, nodeID)
	case "RegisterNode":
		return app.registerNode(param, nodeID)
	case "RegisterIdentity":
		return app.registerIdentity(param, nodeID)
	case "AddAccessor":
		return app.AddAccessor(param, nodeID)
	case "CreateRequest":
		return app.createRequest(param, nodeID)
	case "CreateIdpResponse":
		return app.createIdpResponse(param, nodeID)
	case "CreateAsResponse":
		return app.createAsResponse(param, nodeID)
	case "RegisterServiceDestination":
		return app.registerServiceDestination(param, nodeID)
	case "SetMqAddresses":
		return app.setMqAddresses(param, nodeID)
	case "AddNodeToken":
		return app.addNodeToken(param, nodeID)
	case "ReduceNodeToken":
		return app.reduceNodeToken(param, nodeID)
	case "SetNodeToken":
		return app.setNodeToken(param, nodeID)
	case "SetPriceFunc":
		return app.setPriceFunc(param, nodeID)
	case "CloseRequest":
		return app.closeRequest(param, nodeID)
	case "TimeOutRequest":
		return app.timeOutRequest(param, nodeID)
	case "AddNamespace":
		return app.addNamespace(param, nodeID)
	case "UpdateNode":
		return app.updateNode(param, nodeID)
	case "SetValidator":
		return app.setValidator(param, nodeID)
	case "AddService":
		return app.addService(param, nodeID)
	case "SetDataReceived":
		return app.setDataReceived(param, nodeID)
	case "UpdateNodeByNDID":
		return app.updateNodeByNDID(param, nodeID)
	case "UpdateIdentity":
		return app.updateIdentity(param, nodeID)
	case "UpdateServiceDestination":
		return app.updateServiceDestination(param, nodeID)
	case "UpdateService":
		return app.updateService(param, nodeID)
	case "RegisterServiceDestinationByNDID":
		return app.registerServiceDestinationByNDID(param, nodeID)
	case "DisableNode":
		return app.disableNode(param, nodeID)
	case "DisableServiceDestinationByNDID":
		return app.disableServiceDestinationByNDID(param, nodeID)
	case "DisableNamespace":
		return app.disableNamespace(param, nodeID)
	case "DisableService":
		return app.disableService(param, nodeID)
	case "EnableNode":
		return app.enableNode(param, nodeID)
	case "EnableServiceDestinationByNDID":
		return app.enableServiceDestinationByNDID(param, nodeID)
	case "EnableNamespace":
		return app.enableNamespace(param, nodeID)
	case "EnableService":
		return app.enableService(param, nodeID)
	case "DisableServiceDestination":
		return app.disableServiceDestination(param, nodeID)
	case "EnableServiceDestination":
		return app.enableServiceDestination(param, nodeID)
	case "SetTimeOutBlockRegisterIdentity":
		return app.setTimeOutBlockRegisterIdentity(param, nodeID)
	case "AddNodeToProxyNode":
		return app.addNodeToProxyNode(param, nodeID)
	case "UpdateNodeProxyNode":
		return app.updateNodeProxyNode(param, nodeID)
	case "RemoveNodeFromProxyNode":
		return app.removeNodeFromProxyNode(param, nodeID)
	case "AddErrorCode":
		return app.addErrorCode(param, nodeID)
	case "RemoveErrorCode":
		return app.removeErrorCode(param, nodeID)
	case "SetInitData":
		return app.SetInitData(param, nodeID)
	case "EndInit":
		return app.EndInit(param, nodeID)
	case "SetLastBlock":
		return app.setLastBlock(param, nodeID)
	case "RevokeIdentityAssociation":
		return app.revokeIdentityAssociation(param, nodeID)
	case "RevokeAccessor":
		return app.revokeAccessor(param, nodeID)
	case "UpdateIdentityModeList":
		return app.updateIdentityModeList(param, nodeID)
	case "AddIdentity":
		return app.addIdentity(param, nodeID)
	case "SetAllowedModeList":
		return app.SetAllowedModeList(param, nodeID)
	case "UpdateNamespace":
		return app.updateNamespace(param, nodeID)
	case "SetAllowedMinIalForRegisterIdentityAtFirstIdp":
		return app.SetAllowedMinIalForRegisterIdentityAtFirstIdp(param, nodeID)
	case "RevokeAndAddAccessor":
		return app.revokeAndAddAccessor(param, nodeID)
	default:
		return types.ResponseDeliverTx{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}
