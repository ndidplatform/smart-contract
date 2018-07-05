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
	"reflect"

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
	funcs := map[string]interface{}{
		"InitNDID":                         initNDID,
		"RegisterNode":                     registerNode,
		"RegisterMsqDestination":           registerMsqDestination,
		"AddAccessorMethod":                addAccessorMethod,
		"CreateRequest":                    createRequest,
		"CreateIdpResponse":                createIdpResponse,
		"SignData":                         signData,
		"RegisterServiceDestination":       registerServiceDestination,
		"RegisterMsqAddress":               registerMsqAddress,
		"AddNodeToken":                     addNodeToken,
		"ReduceNodeToken":                  reduceNodeToken,
		"SetNodeToken":                     setNodeToken,
		"SetPriceFunc":                     setPriceFunc,
		"CloseRequest":                     closeRequest,
		"TimeOutRequest":                   timeOutRequest,
		"AddNamespace":                     addNamespace,
		"UpdateNode":                       updateNode,
		"CreateIdentity":                   createIdentity,
		"SetValidator":                     setValidator,
		"AddService":                       addService,
		"SetDataReceived":                  setDataReceived,
		"UpdateNodeByNDID":                 updateNodeByNDID,
		"UpdateIdentity":                   updateIdentity,
		"DeclareIdentityProof":             declareIdentityProof,
		"UpdateServiceDestination":         updateServiceDestination,
		"UpdateService":                    updateService,
		"RegisterServiceDestinationByNDID": registerServiceDestinationByNDID,
		"UpdateServiceDestinationByNDID":   updateServiceDestinationByNDID,
		"DisableMsqDestination":            disableMsqDestination,
		"DisableAccessorMethod":            disableAccessorMethod,
		"DisableNode":                      disableNode,
		"DisableServiceDestinationByNDID":  disableServiceDestinationByNDID,
		"DisableNamespace":                 disableNamespace,
		"DisableService":                   disableService,
		"EnableMsqDestination":             enableMsqDestination,
		"EnableAccessorMethod":             enableAccessorMethod,
		"EnableNode":                       enableNode,
		"EnableServiceDestinationByNDID":   enableServiceDestinationByNDID,
		"EnableNamespace":                  enableNamespace,
		"EnableService":                    enableService,
	}

	// ---- check authorization ----
	checkTxResult := CheckTxRouter(method, param, nonce, signature, nodeID, app)
	if checkTxResult.Code != code.OK {
		return ReturnDeliverTxLog(checkTxResult.Code, "Unauthorized", "")
	}

	value, _ := callDeliverTx(funcs, method, param, app, nodeID)
	result := value[0].Interface().(types.ResponseDeliverTx)
	// ---- Burn token ----
	if result.Code == code.OK {
		if !isNDIDMethod[method] {
			needToken := getTokenPriceByFunc(method, app)
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

func callDeliverTx(m map[string]interface{}, name string, param string, app *DIDApplication, nodeID string) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 3)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(app)
	in[2] = reflect.ValueOf(nodeID)
	result = f.Call(in)
	return
}
