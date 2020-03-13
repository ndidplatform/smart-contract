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
 * Please contact info@co.th for any further questions
 *
 */

package ndid

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/ndidplatform/smart-contract/v4/abci/did/v1"
	"github.com/ndidplatform/smart-contract/v4/test/data"
	"github.com/ndidplatform/smart-contract/v4/test/query"
	"github.com/ndidplatform/smart-contract/v4/test/utils"
)

var ndidNodeID = "ndid"

func TestInitNDID(t *testing.T) {
	privKey := utils.GetPrivateKeyFromString(data.NdidPrivK)
	ndidpublicKeyBytes, err := utils.GeneratePublicKey(&privKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var param did.InitNDIDParam
	param.NodeID = ndidNodeID
	param.PublicKey = string(ndidpublicKeyBytes)
	param.MasterPublicKey = string(ndidpublicKeyBytes)
	InitNDID(t, ndidNodeID, data.NdidPrivK, param)
	IsInitEnded(t, false)
	EndInit(t, ndidNodeID, data.NdidPrivK, did.EndInitParam{})
	IsInitEnded(t, true)
}

func TestNDIDSetAllowedMinIalForRegisterIdentityAtFirstIdp(t *testing.T) {
	var param did.SetAllowedMinIalForRegisterIdentityAtFirstIdpParam
	param.MinIal = 2.3
	SetAllowedMinIalForRegisterIdentityAtFirstIdp(t, ndidNodeID, data.NdidPrivK, param)
}

func TestQueryGetAllowedMinIalForRegisterIdentityAtFirstIdp(t *testing.T) {
	var expected = `{"min_ial":2.3}`
	query.GetAllowedMinIalForRegisterIdentityAtFirstIdp(t, expected)
}

func InitNDID(t *testing.T, nodeID, privK string, param did.InitNDIDParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "InitNDID"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if resultObj.Result.CheckTx.Log == "NDID node is already existed" {
		t.SkipNow()
	}
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func EndInit(t *testing.T, nodeID, privK string, param did.EndInitParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "EndInit"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func IsInitEnded(t *testing.T, expected bool) {
	fnName := "IsInitEnded"
	var param did.IsInitEndedParam
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.IsInitEndedResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res.InitEnded; actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func SetAllowedMinIalForRegisterIdentityAtFirstIdp(t *testing.T, nodeID, privK string, param did.SetAllowedMinIalForRegisterIdentityAtFirstIdpParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "SetAllowedMinIalForRegisterIdentityAtFirstIdp"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func AddNamespace(t *testing.T, nodeID, privK string, param did.Namespace) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "AddNamespace"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddNamespace(t *testing.T, namespace string) {
	var param did.Namespace
	param.Namespace = namespace
	switch namespace {
	case data.UserNamespace1:
		param.Description = "Citizen ID"
		param.AllowedIdentifierCountInReferenceGroup = 1
		param.AllowedActiveIdentifierCountInReferenceGroup = 1
	case data.UserNamespace2:
		param.Description = "Passport"
	case data.UserNamespace3:
		param.Description = "Some ID"
	}
	AddNamespace(t, ndidNodeID, data.NdidPrivK, param)
}

func TestQueryGetNamespaceList(t *testing.T, expected string) {
	query.GetNamespaceList(t, expected)
}

func UpdateNamespace(t *testing.T, nodeID, privK string, param did.UpdateNamespaceParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "UpdateNamespace"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDUpdateNamespace(t *testing.T) {
	var param did.UpdateNamespaceParam
	param.Namespace = data.UserNamespace3
	param.AllowedIdentifierCountInReferenceGroup = 2
	param.AllowedActiveIdentifierCountInReferenceGroup = 2
	UpdateNamespace(t, ndidNodeID, data.NdidPrivK, param)
}

func RegisterNode(t *testing.T, nodeID, privK string, param did.RegisterNode) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "RegisterNode"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func SetNodeToken(t *testing.T, nodeID, privK string, param did.SetNodeTokenParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "SetNodeToken"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRegisterNode(t *testing.T, nodeID string) {
	masterKey := utils.GetPrivateKeyFromString(data.AllMasterKey)
	masterPublicKeyBytes, err := utils.GeneratePublicKey(&masterKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var param did.RegisterNode
	switch nodeID {
	case data.IdP1:
		privKey := utils.GetPrivateKeyFromString(data.IdpPrivK1)
		publicKeyBytes, err := utils.GeneratePublicKey(&privKey.PublicKey)
		if err != nil {
			log.Fatal(err.Error())
		}
		param.NodeID = nodeID
		param.PublicKey = string(publicKeyBytes)
		param.MasterPublicKey = string(masterPublicKeyBytes)
		param.NodeName = "IdP Number 1"
		param.Role = "IdP"
		param.MaxIal = 3.0
		param.MaxAal = 3.0
	case data.IdP2:
		privKey := utils.GetPrivateKeyFromString(data.IdpPrivK2)
		publicKeyBytes, err := utils.GeneratePublicKey(&privKey.PublicKey)
		if err != nil {
			log.Fatal(err.Error())
		}
		param.NodeID = nodeID
		param.PublicKey = string(publicKeyBytes)
		param.MasterPublicKey = string(masterPublicKeyBytes)
		param.NodeName = "IdP Number 2"
		param.Role = "IdP"
		param.MaxIal = 2.3
		param.MaxAal = 3.0
	case data.IdPAgent1:
		privKey := utils.GetPrivateKeyFromString(data.IdpPrivK3)
		publicKeyBytes, err := utils.GeneratePublicKey(&privKey.PublicKey)
		isIdPAgent := true
		if err != nil {
			log.Fatal(err.Error())
		}
		param.NodeID = nodeID
		param.PublicKey = string(publicKeyBytes)
		param.MasterPublicKey = string(masterPublicKeyBytes)
		param.NodeName = "IdP Agent 1"
		param.Role = "IdP"
		param.MaxIal = 2.3
		param.MaxAal = 3.0
		param.IsIdPAgent = &isIdPAgent
	case data.AS1:
		asKey := utils.GetPrivateKeyFromString(data.AsPrivK1)
		asPublicKeyBytes, err := utils.GeneratePublicKey(&asKey.PublicKey)
		if err != nil {
			log.Fatal(err.Error())
		}
		asMasterKey := utils.GetPrivateKeyFromString(data.AllMasterKey)
		asMasterPublicKeyBytes, err := utils.GeneratePublicKey(&asMasterKey.PublicKey)
		if err != nil {
			log.Fatal(err.Error())
		}
		param.NodeName = "AS1"
		param.NodeID = data.AS1
		param.PublicKey = string(asPublicKeyBytes)
		param.MasterPublicKey = string(asMasterPublicKeyBytes)
		param.Role = "AS"
	case data.AS2:
		asKey := utils.GetPrivateKeyFromString(data.AsPrivK2)
		asPublicKeyBytes, err := utils.GeneratePublicKey(&asKey.PublicKey)
		if err != nil {
			log.Fatal(err.Error())
		}
		asKey2 := utils.GetPrivateKeyFromString(data.AllMasterKey)
		asPublicKeyBytes2, err := utils.GeneratePublicKey(&asKey2.PublicKey)
		if err != nil {
			log.Fatal(err.Error())
		}
		param.NodeName = "AS2"
		param.NodeID = data.AS2
		param.PublicKey = string(asPublicKeyBytes)
		param.MasterPublicKey = string(asPublicKeyBytes2)
		param.Role = "AS"
	}
	RegisterNode(t, ndidNodeID, data.NdidPrivK, param)
}

func TestSetNodeToken(t *testing.T, nodeID string, amount float64) {
	var param did.SetNodeTokenParam
	param.NodeID = nodeID
	param.Amount = amount
	SetNodeToken(t, ndidNodeID, data.NdidPrivK, param)
}

func AddService(t *testing.T, nodeID, privK string, param did.AddServiceParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "AddService"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddService(t *testing.T, serviceID string) {
	var param did.AddServiceParam
	param.ServiceID = serviceID
	switch serviceID {
	case data.ServiceID1:
		param.ServiceName = "Bank statement"
		param.DataSchema = "DataSchema"
		param.DataSchemaVersion = "DataSchemaVersion"
	}
	AddService(t, ndidNodeID, data.NdidPrivK, param)
}

func RegisterServiceDestinationByNDID(t *testing.T, nodeID, privK string, param did.RegisterServiceDestinationByNDIDParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "RegisterServiceDestinationByNDID"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf(`PASS: %s, Expected log: "%s"`, fnName, expected)
}

func TestRegisterServiceDestinationByNDID(t *testing.T, caseID int64, expected string) {
	var param did.RegisterServiceDestinationByNDIDParam
	switch caseID {
	case 1:
		param.ServiceID = data.ServiceID1
		param.NodeID = "Invalid-node-ID"
	case 2:
		param.ServiceID = data.ServiceID1
		param.NodeID = data.IdP1
	case 3:
		param.ServiceID = data.ServiceID1
		param.NodeID = data.AS1
	case 4:
		param.ServiceID = data.ServiceID1
		param.NodeID = data.AS2
	}
	RegisterServiceDestinationByNDID(t, ndidNodeID, data.NdidPrivK, param, expected)
}

func AddErrorCode(t *testing.T, nodeID, privK string, param did.AddErrorCodeParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "AddErrorCode"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`param: %s`, paramJSON)
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddErrorCode(t *testing.T, errorCodeType string, errorCode string, description string, fatal bool, expected string) {
	param := did.AddErrorCodeParam{
		ErrorCode:   errorCode,
		Description: description,
		Fatal:       fatal,
		Type:        errorCodeType,
	}
	AddErrorCode(t, ndidNodeID, data.NdidPrivK, param, expected)
}

func RemoveErrorCode(t *testing.T, nodeID, privK string, param did.RemoveErrorCodeParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "RemoveErrorCode"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`param: %s`, paramJSON)
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRemoveErrorCode(t *testing.T, errorCodeType string, errorCode string, expected string) {
	param := did.RemoveErrorCodeParam{
		ErrorCode: errorCode,
		Type:      errorCodeType,
	}
	RemoveErrorCode(t, ndidNodeID, data.NdidPrivK, param, expected)
}
