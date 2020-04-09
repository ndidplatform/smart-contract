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
 * Please contact info@napp.co.th for any further questions
 *
 */

package common

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/ndidplatform/smart-contract/v4/abci/app/v1"
	"github.com/ndidplatform/smart-contract/v4/test/data"
	"github.com/ndidplatform/smart-contract/v4/test/utils"
)

func SetMqAddresses(t *testing.T, nodeID, privK string, param app.SetMqAddressesParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "SetMqAddresses"
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

func TestSetMqAddresses(t *testing.T, nodeID, privK, ip string, port int64) {
	var mq app.MsqAddress
	mq.IP = ip
	mq.Port = port
	var param app.SetMqAddressesParam
	param.Addresses = make([]app.MsqAddress, 0)
	param.Addresses = append(param.Addresses, mq)
	SetMqAddresses(t, nodeID, privK, param)
}

func CreateRequest(t *testing.T, nodeID, privK string, param app.CreateRequestParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "CreateRequest"
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

func CloseRequest(t *testing.T, nodeID, privK string, param app.CloseRequestParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "CloseRequest"
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

func TestCreateRequest(t *testing.T, requestID string) {
	var nodeID string
	var privK string
	var param app.CreateRequestParam
	var datas []app.DataRequest
	param.RequestID = requestID
	switch requestID {
	case data.RequestID1.String():
		param.MinIdp = 0
		param.MinIal = 3
		param.MinAal = 3
		param.Timeout = 259200
		param.DataRequestList = datas
		param.MessageHash = "hash('Please allow...')"
		param.Mode = 3
		param.Purpose = "RegisterIdentity"
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case data.RequestID2.String():
		param.MinIdp = 1
		param.MinIal = 3
		param.MinAal = 3
		param.Timeout = 259200
		param.DataRequestList = datas
		param.MessageHash = "hash('Please allow...')"
		param.Mode = 3
		param.Purpose = "RegisterIdentity"
		param.IdPIDList = append(param.IdPIDList, data.IdP1)
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID3.String():
		param.MinIdp = 1
		param.MinIal = 3
		param.MinAal = 3
		param.Timeout = 259200
		param.DataRequestList = datas
		param.MessageHash = "hash('Please allow...')"
		param.Mode = 3
		param.Purpose = "AddAccessor"
		param.IdPIDList = append(param.IdPIDList, data.IdP1)
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID4.String():
		param.MinIdp = 1
		param.MinIal = 3
		param.MinAal = 3
		param.Timeout = 259200
		param.DataRequestList = datas
		param.MessageHash = "hash('Please allow...')"
		param.Mode = 3
		param.Purpose = "RevokeIdentityAssociation"
		param.IdPIDList = append(param.IdPIDList, data.IdP1)
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID5.String():
		param.MinIdp = 1
		param.MinIal = 3
		param.MinAal = 3
		param.Timeout = 259200
		param.DataRequestList = datas
		param.MessageHash = "hash('Please allow...')"
		param.Mode = 3
		param.Purpose = "RegisterIdentity"
		param.IdPIDList = append(param.IdPIDList, data.IdP1)
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID6.String():
		param.MinIdp = 1
		param.MinIal = 3
		param.MinAal = 3
		param.Timeout = 259200
		param.DataRequestList = datas
		param.MessageHash = "hash('Please allow...')"
		param.Mode = 3
		param.Purpose = "AddIdentity"
		param.IdPIDList = append(param.IdPIDList, data.IdP1)
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID7.String():
		param.MinIdp = 1
		param.MinIal = 2.3
		param.MinAal = 3
		param.Timeout = 259200
		param.DataRequestList = datas
		param.MessageHash = "hash('Please allow...')"
		param.Mode = 3
		param.Purpose = "RevokeAndAddAccessor"
		param.IdPIDList = append(param.IdPIDList, data.IdP2)
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	}
	CreateRequest(t, nodeID, privK, param)
}

func TestCloseRequest(t *testing.T, requestID string) {
	var nodeID string
	var privK string
	var param app.CloseRequestParam
	param.RequestID = requestID
	switch requestID {
	case data.RequestID1.String():
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case data.RequestID2.String():
		var res []app.ResponseValid
		var res1 app.ResponseValid
		res1.IdpID = data.IdP1
		tValue := true
		res1.ValidIal = &tValue
		res1.ValidSignature = &tValue
		res = append(res, res1)
		param.ResponseValidList = res
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID3.String():
		var res []app.ResponseValid
		var res1 app.ResponseValid
		res1.IdpID = data.IdP1
		tValue := true
		res1.ValidIal = &tValue
		res1.ValidSignature = &tValue
		res = append(res, res1)
		param.ResponseValidList = res
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID4.String():
		var res []app.ResponseValid
		var res1 app.ResponseValid
		res1.IdpID = data.IdP1
		tValue := true
		res1.ValidIal = &tValue
		res1.ValidSignature = &tValue
		res = append(res, res1)
		param.ResponseValidList = res
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID5.String():
		var res []app.ResponseValid
		var res1 app.ResponseValid
		res1.IdpID = data.IdP1
		tValue := true
		res1.ValidIal = &tValue
		res1.ValidSignature = &tValue
		res = append(res, res1)
		param.ResponseValidList = res
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID6.String():
		var res []app.ResponseValid
		var res1 app.ResponseValid
		res1.IdpID = data.IdP1
		tValue := true
		res1.ValidIal = &tValue
		res1.ValidSignature = &tValue
		res = append(res, res1)
		param.ResponseValidList = res
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case data.RequestID7.String():
		var res []app.ResponseValid
		var res1 app.ResponseValid
		res1.IdpID = data.IdP2
		tValue := true
		res1.ValidIal = &tValue
		res1.ValidSignature = &tValue
		res = append(res, res1)
		param.ResponseValidList = res
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	}
	CloseRequest(t, nodeID, privK, param)
}

func UpdateNode(t *testing.T, nodeID, privK string, param app.UpdateNodeParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "UpdateNode"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestUpdateNode(t *testing.T, caseID int64, expected string) {
	var nodeID string
	var privK string
	var param app.UpdateNodeParam
	switch caseID {
	case 1:
		idpKey2 := utils.GetPrivateKeyFromString(data.IdpPrivK2)
		idpPublicKeyBytes2, err := utils.GeneratePublicKey(&idpKey2.PublicKey)
		if err != nil {
			log.Fatal(err.Error())
		}
		param.PublicKey = string(idpPublicKeyBytes2)
		param.SupportedRequestMessageDataUrlTypeList = append(param.SupportedRequestMessageDataUrlTypeList, "text/plain")
		param.SupportedRequestMessageDataUrlTypeList = append(param.SupportedRequestMessageDataUrlTypeList, "application/pdf")
		nodeID = data.IdP1
		privK = data.AllMasterKey
	case 2:
		idpKey1 := utils.GetPrivateKeyFromString(data.IdpPrivK1)
		idpPublicKeyBytes1, err := utils.GeneratePublicKey(&idpKey1.PublicKey)
		if err != nil {
			log.Fatal(err.Error())
		}
		param.PublicKey = string(idpPublicKeyBytes1)
		param.SupportedRequestMessageDataUrlTypeList = append(param.SupportedRequestMessageDataUrlTypeList, "text/plain")
		nodeID = data.IdP1
		privK = data.AllMasterKey
	}
	UpdateNode(t, nodeID, privK, param, expected)
}
