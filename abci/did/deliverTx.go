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
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// ReturnDeliverTxLog return types.ResponseDeliverTx
func ReturnDeliverTxLog(code uint32, log string, extraData string) types.ResponseDeliverTx {
	var tags []cmn.KVPair
	if code == 0 {
		tags = []cmn.KVPair{
			{[]byte("success"), []byte("true")},
		}
	} else {
		tags = []cmn.KVPair{
			{[]byte("success"), []byte("false")},
		}
	}
	return types.ResponseDeliverTx{
		Code: code,
		Log:  fmt.Sprintf(log),
		Data: []byte(extraData),
		Tags: tags,
	}
}

// DeliverTxRouter is Pointer to function
func DeliverTxRouter(method string, param string, nonce string, signature string, nodeID string, app *DIDApplication) types.ResponseDeliverTx {
	// ---- check authorization ----
	checkTxResult := CheckTxRouter(method, param, nonce, signature, nodeID, app)
	if checkTxResult.Code != code.OK {
		if checkTxResult.Log != "" {
			return ReturnDeliverTxLog(checkTxResult.Code, checkTxResult.Log, "")
		}
		return ReturnDeliverTxLog(checkTxResult.Code, "Unauthorized", "")
	}

	result := callDeliverTx(method, param, app, nodeID)
	// ---- Burn token ----
	if result.Code == code.OK {
		if !isNDIDMethod[method] {
			needToken := getTokenPriceByFunc(method, app, app.state.db.Version64())
			err := reduceToken(nodeID, needToken, app)
			if err != nil {
				result.Code = code.TokenAccountNotFound
				result.Log = err.Error()
				return result
			}
			// Write burn token report
			// only have result.Data in some method
			writeBurnTokenReport(nodeID, method, needToken, string(result.Data), app)
		}
	}
	return result
}

func callDeliverTx(name string, param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	switch name {
	case "InitNDID":
		return initNDID(param, app, nodeID)
	case "RegisterNode":
		return registerNode(param, app, nodeID)
	case "RegisterMsqDestination":
		return registerMsqDestination(param, app, nodeID)
	case "AddAccessorMethod":
		return addAccessorMethod(param, app, nodeID)
	case "CreateRequest":
		return createRequest(param, app, nodeID)
	case "CreateIdpResponse":
		return createIdpResponse(param, app, nodeID)
	case "SignData":
		return signData(param, app, nodeID)
	case "RegisterServiceDestination":
		return registerServiceDestination(param, app, nodeID)
	case "RegisterMsqAddress":
		return registerMsqAddress(param, app, nodeID)
	case "AddNodeToken":
		return addNodeToken(param, app, nodeID)
	case "ReduceNodeToken":
		return reduceNodeToken(param, app, nodeID)
	case "SetNodeToken":
		return setNodeToken(param, app, nodeID)
	case "SetPriceFunc":
		return setPriceFunc(param, app, nodeID)
	case "CloseRequest":
		return closeRequest(param, app, nodeID)
	case "TimeOutRequest":
		return timeOutRequest(param, app, nodeID)
	case "AddNamespace":
		return addNamespace(param, app, nodeID)
	case "UpdateNode":
		return updateNode(param, app, nodeID)
	case "CreateIdentity":
		return createIdentity(param, app, nodeID)
	case "SetValidator":
		return setValidator(param, app, nodeID)
	case "AddService":
		return addService(param, app, nodeID)
	case "SetDataReceived":
		return setDataReceived(param, app, nodeID)
	case "UpdateNodeByNDID":
		return updateNodeByNDID(param, app, nodeID)
	case "UpdateIdentity":
		return updateIdentity(param, app, nodeID)
	case "DeclareIdentityProof":
		return declareIdentityProof(param, app, nodeID)
	case "UpdateServiceDestination":
		return updateServiceDestination(param, app, nodeID)
	case "UpdateService":
		return updateService(param, app, nodeID)
	case "RegisterServiceDestinationByNDID":
		return registerServiceDestinationByNDID(param, app, nodeID)
	case "DisableMsqDestination":
		return disableMsqDestination(param, app, nodeID)
	case "DisableAccessorMethod":
		return disableAccessorMethod(param, app, nodeID)
	case "DisableNode":
		return disableNode(param, app, nodeID)
	case "DisableServiceDestinationByNDID":
		return disableServiceDestinationByNDID(param, app, nodeID)
	case "DisableNamespace":
		return disableNamespace(param, app, nodeID)
	case "DisableService":
		return disableService(param, app, nodeID)
	case "EnableMsqDestination":
		return enableMsqDestination(param, app, nodeID)
	case "EnableAccessorMethod":
		return enableAccessorMethod(param, app, nodeID)
	case "EnableNode":
		return enableNode(param, app, nodeID)
	case "EnableServiceDestinationByNDID":
		return enableServiceDestinationByNDID(param, app, nodeID)
	case "EnableNamespace":
		return enableNamespace(param, app, nodeID)
	case "EnableService":
		return enableService(param, app, nodeID)
	case "DisableServiceDestination":
		return disableServiceDestination(param, app, nodeID)
	case "EnableServiceDestination":
		return enableServiceDestination(param, app, nodeID)
	default:
		return types.ResponseDeliverTx{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}
