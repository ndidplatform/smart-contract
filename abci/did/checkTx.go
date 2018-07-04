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
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"reflect"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/tendermint/abci/types"
)

var IsMethod = map[string]bool{
	"InitNDID":                         true,
	"RegisterNode":                     true,
	"AddNodeToken":                     true,
	"ReduceNodeToken":                  true,
	"SetNodeToken":                     true,
	"SetPriceFunc":                     true,
	"AddNamespace":                     true,
	"SetValidator":                     true,
	"AddService":                       true,
	"UpdateNodeByNDID":                 true,
	"UpdateService":                    true,
	"RegisterServiceDestinationByNDID": true,
	"UpdateServiceDestinationByNDID":   true,
	"DisableNode":                      true,
	"DisableNamespace":                 true,
	"DisableService":                   true,
	"DisableServiceDestinationByNDID":  true,
	"EnableNode":                       true,
	"EnableServiceDestinationByNDID":   true,
	"EnableNamespace":                  true,
	"EnableService":                    true,
	"RegisterMsqDestination":           true,
	"AddAccessorMethod":                true,
	"CreateIdpResponse":                true,
	"CreateIdentity":                   true,
	"UpdateIdentity":                   true,
	"DeclareIdentityProof":             true,
	"DisableMsqDestination":            true,
	"DisableAccessorMethod":            true,
	"EnableMsqDestination":             true,
	"EnableAccessorMethod":             true,
	"SignData":                         true,
	"RegisterServiceDestination":       true,
	"UpdateServiceDestination":         true,
	"CreateRequest":                    true,
	"RegisterMsqAddress":               true,
	"UpdateNode":                       true,
	"CloseRequest":                     true,
	"TimeOutRequest":                   true,
	"SetDataReceived":                  true,
}

func checkTxInitNDID(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	key := "MasterNDID"
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value == nil {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsMember(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "RP" ||
		string(value) == "IdP" ||
		string(value) == "AS" ||
		string(value) == "MasterRP" ||
		string(value) == "MasterIdP" ||
		string(value) == "MasterAS" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkTxRegisterMsqAddress(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var node NodeDetail
	err := json.Unmarshal([]byte(value), &node)
	if err != nil {
		return ReturnCheckTx(false)
	}

	if string(node.Role) == "RP" ||
		string(node.Role) == "IdP" ||
		string(node.Role) == "AS" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkNDID(param string, nodeID string, app *DIDApplication) bool {
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var node NodeDetail
	err := json.Unmarshal([]byte(value), &node)
	if err != nil {
		return false
	}
	if node.Role == "NDID" {
		return true
	}
	return false
}

func checkIdP(param string, nodeID string, app *DIDApplication) bool {
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var node NodeDetail
	err := json.Unmarshal([]byte(value), &node)
	if err != nil {
		return false
	}
	if node.Role == "IdP" {
		return true
	}
	return false
}

func checkAS(param string, nodeID string, app *DIDApplication) bool {
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var node NodeDetail
	err := json.Unmarshal([]byte(value), &node)
	if err != nil {
		return false
	}
	if node.Role == "AS" {
		return true
	}
	return false
}

func checkIdPorRP(param string, nodeID string, app *DIDApplication) bool {
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var node NodeDetail
	err := json.Unmarshal([]byte(value), &node)
	if err != nil {
		return false
	}
	if node.Role == "IdP" || node.Role == "RP" {
		return true
	}
	return false
}

func checkIsNDID(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	return ReturnCheckTx(checkNDID(param, nodeID, app))
}

func checkIsIDP(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	return ReturnCheckTx(checkIdP(param, nodeID, app))
}

func checkIsAS(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	return ReturnCheckTx(checkAS(param, nodeID, app))
}

func checkIsRPorIdP(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	return ReturnCheckTx(checkIdPorRP(param, nodeID, app))
}

func checkIsMasterNode(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "MasterIdP" ||
		string(value) == "MasterRP" ||
		string(value) == "MasterAS" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsOwnerRequest(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	var funcParam RequestIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnCheckTx(false)
	}

	// Check request is exist
	requestKey := "Request" + "|" + funcParam.RequestID
	_, requestValue := app.state.db.Get(prefixKey([]byte(requestKey)))

	if requestValue == nil {
		return types.ResponseCheckTx{Code: code.RequestIDNotFound, Log: "Request ID not found"}
	}

	key := "SpendGas" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	var reports []Report
	err = json.Unmarshal([]byte(value), &reports)
	if err != nil {
		return ReturnCheckTx(false)
	}

	for _, node := range reports {
		if node.Method == "CreateRequest" &&
			node.Data == funcParam.RequestID {
			return ReturnCheckTx(true)
		}
	}

	return ReturnCheckTx(false)
}

func verifySignature(param string, nonce string, signature string, publicKey string) (result bool, err error) {
	publicKey = strings.Replace(publicKey, "\t", "", -1)
	block, _ := pem.Decode([]byte(publicKey))
	senderPublicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	senderPublicKey := senderPublicKeyInterface.(*rsa.PublicKey)
	if err != nil {
		return false, err
	}
	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	PSSmessage := []byte(param + nonce)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	err = rsa.VerifyPKCS1v15(senderPublicKey, newhash, hashed, decodedSignature)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ReturnCheckTx return types.ResponseDeliverTx
func ReturnCheckTx(ok bool) types.ResponseCheckTx {
	if ok {
		return types.ResponseCheckTx{Code: code.OK}
	}
	return types.ResponseCheckTx{Code: code.Unauthorized}
}

func getPublicKeyInitNDID(param string) string {
	var funcParam InitNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ""
	}
	return funcParam.PublicKey
}

func getMasterPublicKeyFromNodeID(nodeID string, app *DIDApplication) string {
	key := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ""
		}
		return nodeDetail.MasterPublicKey
	}
	return ""
}

func getPublicKeyFromNodeID(nodeID string, app *DIDApplication) string {
	key := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ""
		}
		return nodeDetail.PublicKey
	}
	return ""
}

func getRoleFromNodeID(nodeID string, app *DIDApplication) string {
	key := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ""
		}
		return string(nodeDetail.Role)
	}
	return ""
}

var IsCheckOwnerRequestMethod = map[string]bool{
	"CloseRequest":    true,
	"TimeOutRequest":  true,
	"SetDataReceived": true,
}

var IsMasterKeyMethod = map[string]bool{
	"UpdateNode": true,
}

// CheckTxRouter is Pointer to function
func CheckTxRouter(method string, param string, nonce string, signature string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	funcs := map[string]interface{}{
		"InitNDID":                         checkTxInitNDID,
		"RegisterNode":                     checkIsNDID,
		"AddNodeToken":                     checkIsNDID,
		"ReduceNodeToken":                  checkIsNDID,
		"SetNodeToken":                     checkIsNDID,
		"SetPriceFunc":                     checkIsNDID,
		"AddNamespace":                     checkIsNDID,
		"SetValidator":                     checkIsNDID,
		"AddService":                       checkIsNDID,
		"UpdateNodeByNDID":                 checkIsNDID,
		"UpdateService":                    checkIsNDID,
		"RegisterServiceDestinationByNDID": checkIsNDID,
		"UpdateServiceDestinationByNDID":   checkIsNDID,
		"DisableNode":                      checkIsNDID,
		"DisableNamespace":                 checkIsNDID,
		"DisableService":                   checkIsNDID,
		"DisableServiceDestinationByNDID":  checkIsNDID,
		"EnableNode":                       checkIsNDID,
		"EnableServiceDestinationByNDID":   checkIsNDID,
		"EnableNamespace":                  checkIsNDID,
		"EnableService":                    checkIsNDID,
		"RegisterMsqDestination":           checkIsIDP,
		"AddAccessorMethod":                checkIsIDP,
		"CreateIdpResponse":                checkIsIDP,
		"CreateIdentity":                   checkIsIDP,
		"UpdateIdentity":                   checkIsIDP,
		"DeclareIdentityProof":             checkIsIDP,
		"DisableMsqDestination":            checkIsIDP,
		"DisableAccessorMethod":            checkIsIDP,
		"EnableMsqDestination":             checkIsIDP,
		"EnableAccessorMethod":             checkIsIDP,
		"SignData":                         checkIsAS,
		"RegisterServiceDestination":       checkIsAS,
		"UpdateServiceDestination":         checkIsAS,
		"CreateRequest":                    checkIsRPorIdP,
		"RegisterMsqAddress":               checkTxRegisterMsqAddress,
		"UpdateNode":                       checkIsMasterNode,
	}

	var publicKey string
	if method == "InitNDID" {
		publicKey = getPublicKeyInitNDID(param)
		if publicKey == "" {
			return ReturnCheckTx(false)
		}
	} else if method == "UpdateNode" {
		publicKey = getMasterPublicKeyFromNodeID(nodeID, app)
		if publicKey == "" {
			return ReturnCheckTx(false)
		}
	} else {
		publicKey = getPublicKeyFromNodeID(nodeID, app)
		if publicKey == "" {
			return ReturnCheckTx(false)
		}
	}

	verifyResult, err := verifySignature(param, nonce, signature, publicKey)
	if err != nil || verifyResult == false {
		return ReturnCheckTx(false)
	}

	var result types.ResponseCheckTx

	// special case checkIsOwnerRequest
	if IsCheckOwnerRequestMethod[method] {
		result = checkIsOwnerRequest(param, nodeID, app)
	} else if IsMasterKeyMethod[method] {
		// If verifyResult is true, return true
		return ReturnCheckTx(true)
	} else {
		value, _ := callCheckTx(funcs, method, param, nodeID, app)
		result = value[0].Interface().(types.ResponseCheckTx)
	}
	// check token for create Tx
	if result.Code == code.OK {
		if !checkNDID(nodeID, nodeID, app) && method != "InitNDID" {
			needToken := getTokenPriceByFunc(method, app)
			nodeToken, err := getToken(nodeID, app)
			if err != nil {
				result.Code = code.TokenAccountNotFound
				result.Log = "token account not found"
			}
			if nodeToken < needToken {
				result.Code = code.TokenNotEnough
				result.Log = "token not enough"
			}
		}
	}
	return result
}

func callCheckTx(m map[string]interface{}, name string, param string, nodeID string, app *DIDApplication) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 3)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(nodeID)
	in[2] = reflect.ValueOf(app)
	result = f.Call(in)
	return
}
