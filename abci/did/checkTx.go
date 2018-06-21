/*
Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED 

This file is part of NDID software.

NDID is the free software: you can redistribute it and/or modify  it under the terms of the Affero GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or any later version.

NDID is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the Affero GNU General Public License for more details.

You should have received a copy of the Affero GNU General Public License along with the NDID source code.  If not, see https://www.gnu.org/licenses/agpl.txt.

please contact info@ndid.co.th for any further questions
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
	"github.com/tendermint/abci/types"
)

func checkTxInitNDID(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "MasterNDID"
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value == nil {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsMember(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
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

func checkTxRegisterMsqAddress(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "RP" ||
		string(value) == "IdP" ||
		string(value) == "AS" ||
		string(value) == "MasterRP" ||
		string(value) == "MasterIdP" ||
		string(value) == "MasterAS" {

		var funcParam RegisterMsqAddressParam
		err := json.Unmarshal([]byte(param), &funcParam)
		if err != nil {
			return ReturnCheckTx(false)
		}
		publicKeyFromStateDB := getPublicKeyFromNodeID(funcParam.NodeID, app)
		if publicKeyFromStateDB == "" {
			return ReturnCheckTx(false)
		}
		if publicKeyFromStateDB == publicKey {
			return ReturnCheckTx(true)
		}
		return ReturnCheckTx(false)
	}
	return ReturnCheckTx(false)
}

func checkNDID(param string, publicKey string, app *DIDApplication) bool {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "NDID" || string(value) == "MasterNDID" {
		return true
	}
	return false
}

func checkIsNDID(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	return ReturnCheckTx(checkNDID(param, publicKey, app))
}

func checkIsIDP(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "IdP" || string(value) == "MasterIdP" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsRP(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "RP" || string(value) == "MasterRP" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsRPorIdP(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "RP" || string(value) == "MasterRP" ||
		string(value) == "IdP" || string(value) == "MasterIdP" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsAS(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "AS" || string(value) == "MasterAS" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsMasterNode(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
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
	requestValue := app.state.db.Get(prefixKey([]byte(requestKey)))

	if requestValue == nil {
		return types.ResponseCheckTx{Code: code.RequestIDNotFound, Log: "Request ID not found"}
	}

	key := "SpendGas" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

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
	value := app.state.db.Get(prefixKey([]byte(key)))
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
	value := app.state.db.Get(prefixKey([]byte(key)))
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
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ""
		}
		roleKey := "NodePublicKeyRole" + "|" + nodeDetail.PublicKey
		roleValue := app.state.db.Get(prefixKey([]byte(roleKey)))
		return string(roleValue)
	}
	return ""
}

var IsCheckOwnerRequestMethod = map[string]bool{
	"CloseRequest":    true,
	"TimeOutRequest":  true,
	"SetDataReceived": true,
}

// CheckTxRouter is Pointer to function
func CheckTxRouter(method string, param string, nonce string, signature string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	funcs := map[string]interface{}{
		"InitNDID":                   checkTxInitNDID,
		"RegisterNode":               checkIsNDID,
		"RegisterMsqDestination":     checkIsIDP,
		"AddAccessorMethod":          checkIsIDP,
		"CreateIdpResponse":          checkIsIDP,
		"SignData":                   checkIsAS,
		"RegisterServiceDestination": checkIsAS,
		"CreateRequest":              checkIsRPorIdP,
		"RegisterMsqAddress":         checkTxRegisterMsqAddress,
		"AddNodeToken":               checkIsNDID,
		"ReduceNodeToken":            checkIsNDID,
		"SetNodeToken":               checkIsNDID,
		"SetPriceFunc":               checkIsNDID,
		"AddNamespace":               checkIsNDID,
		"DeleteNamespace":            checkIsNDID,
		"UpdateNode":                 checkIsMasterNode,
		"CreateIdentity":             checkIsIDP,
		"SetValidator":               checkIsNDID,
		"AddService":                 checkIsNDID,
		"DeleteService":              checkIsNDID,
		"UpdateNodeByNDID":           checkIsNDID,
		"UpdateIdentity":             checkIsIDP,
		"DeclareIdentityProof":       checkIsIDP,
		"UpdateServiceDestination":   checkIsAS,
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
	} else {
		value, _ := callCheckTx(funcs, method, param, publicKey, app)
		result = value[0].Interface().(types.ResponseCheckTx)
	}
	// check token for create Tx
	if result.Code == code.OK {
		if !checkNDID(nodeID, publicKey, app) && method != "InitNDID" {
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

func callCheckTx(m map[string]interface{}, name string, param string, publicKey string, app *DIDApplication) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 3)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(publicKey)
	in[2] = reflect.ValueOf(app)
	result = f.Call(in)
	return
}
