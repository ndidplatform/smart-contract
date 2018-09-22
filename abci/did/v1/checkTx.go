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
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/ndidplatform/smart-contract/protos/data"
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
	// "DisableMsqDestination":                 true,
	// "DisableAccessorMethod":                 true,
	// "EnableMsqDestination":                  true,
	// "EnableAccessorMethod":                  true,
	"SignData":                              true,
	"RegisterServiceDestination":            true,
	"UpdateServiceDestination":              true,
	"CreateRequest":                         true,
	"RegisterMsqAddress":                    true,
	"UpdateNode":                            true,
	"CloseRequest":                          true,
	"TimeOutRequest":                        true,
	"SetDataReceived":                       true,
	"DisableServiceDestination":             true,
	"EnableServiceDestination":              true,
	"ClearRegisterMsqDestinationTimeout":    true,
	"SetTimeOutBlockRegisterMsqDestination": true,
	"AddNodeToProxyNode":                    true,
	"UpdateNodeProxyNode":                   true,
	"RemoveNodeFromProxyNode":               true,
}

func checkTxInitNDID(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	key := "MasterNDID"
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value == nil {
		return ReturnCheckTx(code.OK, "")
	}
	// NDID node (first node of the network) is already existed
	return ReturnCheckTx(code.NDIDisAlreadyExisted, "NDID node is already existed")
}

func checkTxRegisterMsqAddress(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var node data.NodeDetail
	err := proto.Unmarshal(value, &node)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	if string(node.Role) == "RP" ||
		string(node.Role) == "IdP" ||
		string(node.Role) == "AS" ||
		string(node.Role) == "Proxy" {
		return ReturnCheckTx(code.OK, "")
	}
	return ReturnCheckTx(code.NoPermissionForRegisterMsqAddress, "This node does not have permission for register msq address")
}

func checkNDID(param string, nodeID string, app *DIDApplication) bool {
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var node data.NodeDetail
	err := proto.Unmarshal(value, &node)
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
	var node data.NodeDetail
	err := proto.Unmarshal(value, &node)
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
	var node data.NodeDetail
	err := proto.Unmarshal(value, &node)
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
	var node data.NodeDetail
	err := proto.Unmarshal(value, &node)
	if err != nil {
		return false
	}
	if node.Role == "IdP" || node.Role == "RP" {
		return true
	}
	return false
}

func checkIsNDID(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	ok := checkNDID(param, nodeID, app)
	if ok == false {
		return ReturnCheckTx(code.NoPermissionForCallNDIDMethod, "This node does not have permission for call NDID method")
	}
	return ReturnCheckTx(code.OK, "")
}

func checkIsIDP(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	ok := checkIdP(param, nodeID, app)
	if ok == false {
		return ReturnCheckTx(code.NoPermissionForCallIdPMethod, "This node does not have permission for call IdP method")
	}
	return ReturnCheckTx(code.OK, "")
}

func checkIsAS(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	ok := checkAS(param, nodeID, app)
	if ok == false {
		return ReturnCheckTx(code.NoPermissionForCallASMethod, "This node does not have permission for call AS method")
	}
	return ReturnCheckTx(code.OK, "")
}

func checkIsRPorIdP(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	ok := checkIdPorRP(param, nodeID, app)
	if ok == false {
		return ReturnCheckTx(code.NoPermissionForCallRPandIdPMethod, "This node does not have permission for call RP and IdP method")
	}
	return ReturnCheckTx(code.OK, "")
}

func checkIsOwnerRequest(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	var funcParam RequestIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	// Check request is exist
	requestKey := "Request" + "|" + funcParam.RequestID
	_, requestValue := app.state.db.Get(prefixKey([]byte(requestKey)))

	if requestValue == nil {
		return types.ResponseCheckTx{Code: code.RequestIDNotFound, Log: "Request ID not found"}
	}

	key := "SpendGas" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	var reports data.ReportList
	err = proto.Unmarshal([]byte(value), &reports)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	for _, node := range reports.Reports {
		if node.Method == "CreateRequest" &&
			node.Data == funcParam.RequestID {
			return ReturnCheckTx(code.OK, "")
		}
	}

	return ReturnCheckTx(code.NotOwnerOfRequest, "This node is not owner of request")
}

func verifySignature(param string, nonce []byte, signature []byte, publicKey string, method string) (result bool, err error) {
	publicKey = strings.Replace(publicKey, "\t", "", -1)
	block, _ := pem.Decode([]byte(publicKey))
	senderPublicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	senderPublicKey := senderPublicKeyInterface.(*rsa.PublicKey)
	if err != nil {
		return false, err
	}
	tempPSSmessage := append([]byte(method), []byte(param)...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	err = rsa.VerifyPKCS1v15(senderPublicKey, newhash, hashed, signature)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ReturnCheckTx return types.ResponseDeliverTx
func ReturnCheckTx(code uint32, log string) types.ResponseCheckTx {
	return types.ResponseCheckTx{
		Code: code,
		Log:  fmt.Sprintf(log),
	}
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
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal(value, &nodeDetail)
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
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal(value, &nodeDetail)
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
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal(value, &nodeDetail)
		if err != nil {
			return ""
		}
		return string(nodeDetail.Role)
	}
	return ""
}

func checkPubKey(key string) (returnCode uint32, log string) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return code.InvalidKeyFormat, "Invalid key format. Cannot decode PEM."
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return code.InvalidKeyFormat, err.Error()
	}

	switch pubKey := pub.(type) {
	case *rsa.PublicKey:
		if pubKey.N.BitLen() < 2048 {
			return code.RSAKeyLengthTooShort, "RSA key length is too short. Must be at least 2048-bit."
		}
	case *dsa.PublicKey, *ecdsa.PublicKey:
		return code.UnsupportedKeyType, "Unsupported key type. Only RSA is allowed."
	default:
		return code.UnknownKeyType, "Unknown key type. Only RSA is allowed."
	}
	return code.OK, ""
}

func checkNodePubKeys(param string) (returnCode uint32, log string) {
	var keys struct {
		MasterPublicKey string `json:"master_public_key"`
		PublicKey       string `json:"public_key"`
	}
	err := json.Unmarshal([]byte(param), &keys)
	if err != nil {
		return code.UnmarshalError, err.Error()
	}
	// Validate master public key format
	if keys.MasterPublicKey != "" {
		returnCode, log = checkPubKey(keys.MasterPublicKey)
		if returnCode != code.OK {
			return returnCode, log
		}
	}

	// Validate public key format
	if keys.PublicKey != "" {
		returnCode, log = checkPubKey(keys.PublicKey)
		if returnCode != code.OK {
			return returnCode, log
		}
	}
	return code.OK, ""
}

func checkAccessorPubKey(param string) (returnCode uint32, log string) {
	var key struct {
		AccessorPublicKey string `json:"accessor_public_key"`
	}
	err := json.Unmarshal([]byte(param), &key)
	if err != nil {
		return code.UnmarshalError, err.Error()
	}
	returnCode, log = checkPubKey(key.AccessorPublicKey)
	if returnCode != code.OK {
		return returnCode, log
	}
	return code.OK, ""
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
func CheckTxRouter(method string, param string, nonce []byte, signature []byte, nodeID string, app *DIDApplication) types.ResponseCheckTx {

	var publicKey string
	if method == "InitNDID" {
		publicKey = getPublicKeyInitNDID(param)
		if publicKey == "" {
			return ReturnCheckTx(code.CannotGetPublicKeyFromParam, "Can not get public key from parameter")
		}
	} else if method == "UpdateNode" {
		publicKey = getMasterPublicKeyFromNodeID(nodeID, app)
		if publicKey == "" {
			return ReturnCheckTx(code.CannotGetMasterPublicKeyFromNodeID, "Can not get master public key from node ID")
		}
	} else {
		publicKey = getPublicKeyFromNodeID(nodeID, app)
		if publicKey == "" {
			return ReturnCheckTx(code.CannotGetPublicKeyFromNodeID, "Can not get public key from node ID")
		}
	}

	// Check pub key
	if method == "InitNDID" || method == "RegisterNode" || method == "UpdateNode" {
		checkCode, log := checkNodePubKeys(param)
		if checkCode != code.OK {
			return ReturnCheckTx(checkCode, log)
		}
	} else if method == "CreateIdentity" || method == "AddAccessorMethod" {
		checkCode, log := checkAccessorPubKey(param)
		if checkCode != code.OK {
			return ReturnCheckTx(checkCode, log)
		}
	}

	// If method is not 'InitNDID' then check node is active
	if method != "InitNDID" {
		if !getActiveStatusByNodeID(nodeID, app) {
			return ReturnCheckTx(code.NodeIsNotActive, "Node is not active")
		}
	}

	verifyResult, err := verifySignature(param, nonce, signature, publicKey, method)
	if err != nil || verifyResult == false {
		return ReturnCheckTx(code.VerifySignatureError, err.Error())
	}

	var result types.ResponseCheckTx

	// special case checkIsOwnerRequest
	if IsCheckOwnerRequestMethod[method] {
		result = checkIsOwnerRequest(param, nodeID, app)
	} else if IsMasterKeyMethod[method] {
		// If verifyResult is true, return true
		return ReturnCheckTx(code.OK, "")
	} else {
		result = callCheckTx(method, param, nodeID, app)
	}
	// check token for create Tx
	if result.Code == code.OK {
		if !checkNDID(param, nodeID, app) && method != "InitNDID" {
			needToken := getTokenPriceByFunc(method, app, app.state.db.Version64())
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

func callCheckTx(name string, param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	switch name {
	case "InitNDID":
		return checkTxInitNDID(param, nodeID, app)
	case "RegisterNode",
		"AddNodeToken",
		"ReduceNodeToken",
		"SetNodeToken",
		"SetPriceFunc",
		"AddNamespace",
		"SetValidator",
		"AddService",
		"UpdateNodeByNDID",
		"UpdateService",
		"RegisterServiceDestinationByNDID",
		"DisableNode",
		"DisableNamespace",
		"DisableService",
		"DisableServiceDestinationByNDID",
		"EnableNode",
		"EnableServiceDestinationByNDID",
		"EnableNamespace",
		"EnableService",
		"SetTimeOutBlockRegisterMsqDestination",
		"AddNodeToProxyNode",
		"UpdateNodeProxyNode",
		"RemoveNodeFromProxyNode":
		return checkIsNDID(param, nodeID, app)
	case "RegisterMsqDestination",
		"AddAccessorMethod",
		"CreateIdpResponse",
		"CreateIdentity",
		"UpdateIdentity",
		"DeclareIdentityProof",
		// "DisableMsqDestination",
		// "DisableAccessorMethod",
		// "EnableMsqDestination",
		// "EnableAccessorMethod",
		"ClearRegisterMsqDestinationTimeout":
		return checkIsIDP(param, nodeID, app)
	case "SignData",
		"RegisterServiceDestination",
		"UpdateServiceDestination",
		"DisableServiceDestination",
		"EnableServiceDestination":
		return checkIsAS(param, nodeID, app)
	case "CreateRequest":
		return checkIsRPorIdP(param, nodeID, app)
	case "RegisterMsqAddress":
		return checkTxRegisterMsqAddress(param, nodeID, app)
	default:
		return types.ResponseCheckTx{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}

func getActiveStatusByNodeID(nodeID string, app *DIDApplication) bool {
	key := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal(value, &nodeDetail)
		if err != nil {
			return false
		}
		return nodeDetail.Active
	}
	return false
}

func checkIsProxyNode(nodeID string, app *DIDApplication) bool {
	nodeDetailKey := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var node data.NodeDetail
	err := proto.Unmarshal([]byte(value), &node)
	if err != nil {
		return false
	}
	if node.Role == "Proxy" {
		return true
	}
	return false
}
