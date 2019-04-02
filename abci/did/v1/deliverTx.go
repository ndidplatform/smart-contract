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
	"encoding/base64"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// app.ReturnDeliverTxLog return types.ResponseDeliverTx
func (app *DIDApplication) ReturnDeliverTxLog(code uint32, log string, extraData string) types.ResponseDeliverTx {
	var tags []cmn.KVPair
	if code == 0 {
		var tag cmn.KVPair
		tag.Key = []byte("success")
		tag.Value = []byte("true")
		tags = append(tags, tag)
	} else {
		var tag cmn.KVPair
		tag.Key = []byte("success")
		tag.Value = []byte("false")
		tags = append(tags, tag)
	}
	return types.ResponseDeliverTx{
		Code: code,
		Log:  fmt.Sprintf(log),
		Data: []byte(extraData),
		Tags: tags,
	}
}

func (app *DIDApplication) ReturnDeliverTxLogWitgTag(code uint32, log string, specialTag []cmn.KVPair) types.ResponseDeliverTx {
	var tags []cmn.KVPair
	if code == 0 {
		var tag cmn.KVPair
		tag.Key = []byte("success")
		tag.Value = []byte("true")
		tags = append(tags, tag)
	} else {
		var tag cmn.KVPair
		tag.Key = []byte("success")
		tag.Value = []byte("false")
		tags = append(tags, tag)
	}
	tags = append(tags, specialTag...)
	return types.ResponseDeliverTx{
		Code: code,
		Log:  fmt.Sprintf(log),
		Data: []byte(""),
		Tags: tags,
	}
}

// DeliverTxRouter is Pointer to function
func (app *DIDApplication) DeliverTxRouter(method string, param string, nonce []byte, signature []byte, nodeID string) types.ResponseDeliverTx {
	// ---- check authorization ----
	checkTxResult := app.CheckTxRouter(method, param, nonce, signature, nodeID)
	if checkTxResult.Code != code.OK {
		if checkTxResult.Log != "" {
			return app.ReturnDeliverTxLog(checkTxResult.Code, checkTxResult.Log, "")
		}
		return app.ReturnDeliverTxLog(checkTxResult.Code, "Unauthorized", "")
	}

	result := app.callDeliverTx(method, param, nodeID)
	// ---- Burn token ----
	if !app.checkNDID(param, nodeID) && !isNDIDMethod[method] {
		needToken := app.getTokenPriceByFunc(method)
		errCode, errLog := app.reduceToken(nodeID, needToken)
		if errCode != code.OK {
			result.Code = errCode
			result.Log = errLog
		}
	}

	// Set used nonce to stateDB
	emptyValue := make([]byte, 0)
	app.SetStateDB([]byte(nonce), emptyValue)
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	app.deliverTxNonceState[nonceBase64] = []byte(nil)
	return result
}

func (app *DIDApplication) callDeliverTx(name string, param string, nodeID string) types.ResponseDeliverTx {
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
	case "SignData":
		return app.signData(param, nodeID)
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
	case "SetAllowedIdentifierCountForNamespace":
		return app.SetAllowedIdentifierCountForNamespace(param, nodeID)
	default:
		return types.ResponseDeliverTx{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}
