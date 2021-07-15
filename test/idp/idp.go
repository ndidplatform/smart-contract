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

package idp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ndidplatform/smart-contract/v6/abci/app/v1"
	"github.com/ndidplatform/smart-contract/v6/test/data"
	"github.com/ndidplatform/smart-contract/v6/test/utils"
)

func RegisterIdentity(t *testing.T, nodeID, privK string, param app.RegisterIdentityParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "RegisterIdentity"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf(`PASS: %s, Expected log: "%s"`, fnName, expected)
}

func TestRegisterIdentity(t *testing.T, caseID int64, expected string) {
	h1 := sha256.New()
	h2 := sha256.New()
	var nodeID string
	var privK string
	var param app.RegisterIdentityParam
	switch caseID {
	case 1:
		h1.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h1.Sum(nil)
		param.ReferenceGroupCode = ""
		var identity app.Identity
		identity.IdentityNamespace = data.UserNamespace1
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.Ial = 3
		param.ModeList = append(param.ModeList, 2)
		param.AccessorID = data.AccessorID1.String()
		param.AccessorPublicKey = data.AccessorPubKey1
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID1.String()
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case 2:
		h1.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h1.Sum(nil)
		h2.Write([]byte(data.UserNamespace1 + data.UserID2))
		userHash2 := h2.Sum(nil)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		var identity app.Identity
		identity.IdentityNamespace = data.UserNamespace1
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		var identity2 app.Identity
		identity2.IdentityNamespace = data.UserNamespace1
		identity2.IdentityIdentifierHash = hex.EncodeToString(userHash2)
		param.NewIdentityList = append(param.NewIdentityList, identity2)
		param.Ial = 3
		param.ModeList = append(param.ModeList, 2)
		param.AccessorID = data.AccessorID1.String()
		param.AccessorPublicKey = data.AccessorPubKey1
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID1.String()
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case 3:
		h1.Write([]byte("Inlavid Namespace" + data.UserID1))
		userHash := h1.Sum(nil)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		var identity app.Identity
		identity.IdentityNamespace = "Inlavid Namespace"
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.Ial = 3
		param.ModeList = append(param.ModeList, 2)
		param.AccessorID = data.AccessorID1.String()
		param.AccessorPublicKey = data.AccessorPubKey1
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID1.String()
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case 4:
		h1.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h1.Sum(nil)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		var identity app.Identity
		identity.IdentityNamespace = data.UserNamespace1
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.Ial = 3
		param.ModeList = append(param.ModeList, 2)
		param.AccessorID = data.AccessorID1.String()
		param.AccessorPublicKey = data.AccessorPubKey1
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID1.String()
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case 5:
		h1.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h1.Sum(nil)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		var identity app.Identity
		identity.IdentityNamespace = data.UserNamespace1
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.Ial = 2.3
		param.ModeList = append(param.ModeList, 2)
		param.AccessorID = data.AccessorID2.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID2.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case 6:
		h1.Write([]byte(data.UserNamespace1 + data.UserID2))
		userHash := h1.Sum(nil)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		var identity app.Identity
		identity.IdentityNamespace = data.UserNamespace1
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.Ial = 2.3
		param.ModeList = append(param.ModeList, 2)
		param.AccessorID = data.AccessorID2.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID2.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case 7:
		h1.Write([]byte(data.UserNamespace1 + data.UserID2))
		userHash := h1.Sum(nil)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		var identity app.Identity
		identity.IdentityNamespace = data.UserNamespace1
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.Ial = 2.3
		param.ModeList = append(param.ModeList, 2)
		param.AccessorID = data.AccessorID2.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID2.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case 8:
		h1.Write([]byte(data.UserNamespace2 + data.UserID1))
		userHash := h1.Sum(nil)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		var identity app.Identity
		identity.IdentityNamespace = data.UserNamespace2
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.Ial = 2.3
		param.ModeList = append(param.ModeList, 2)
		param.AccessorID = data.AccessorID2.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID2.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case 9:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.Ial = 2.3
		param.ModeList = append(param.ModeList, 2)
		param.ModeList = append(param.ModeList, 3)
		param.AccessorID = data.AccessorID2.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID5.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	}
	RegisterIdentity(t, nodeID, privK, param, expected)
}

func CreateIdpResponse(t *testing.T, nodeID, privK string, param app.CreateIdpResponseParam) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "CreateIdpResponse"
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

func TestCreateIdpResponse(t *testing.T, requestID string) {
	var nodeID string
	var privK string
	var param app.CreateIdpResponseParam
	param.RequestID = requestID
	switch requestID {
	case data.RequestID2.String():
		param.Aal = 3
		param.Ial = 3
		param.Signature = "signature"
		param.Status = "accept"
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case data.RequestID3.String():
		param.Aal = 3
		param.Ial = 3
		param.Signature = "signature"
		param.Status = "accept"
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case data.RequestID4.String():
		param.Aal = 3
		param.Ial = 3
		param.Signature = "signature"
		param.Status = "accept"
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case data.RequestID5.String():
		param.Aal = 3
		param.Ial = 3
		param.Signature = "signature"
		param.Status = "accept"
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case data.RequestID6.String():
		param.Aal = 3
		param.Ial = 3
		param.Signature = "signature"
		param.Status = "accept"
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	case data.RequestID7.String():
		param.Aal = 3
		param.Ial = 2.3
		param.Signature = "signature"
		param.Status = "accept"
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	}
	CreateIdpResponse(t, nodeID, privK, param)
}

func AddAccessor(t *testing.T, nodeID, privK string, param app.AddAccessorParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "AddAccessor"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf(`PASS: %s, Expected log: "%s"`, fnName, expected)
}

func TestAddAccessor(t *testing.T, caseID int64, expected string) {
	var nodeID string
	var privK string
	var param app.AddAccessorParam
	switch caseID {
	case 1:
		h := sha256.New()
		h.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h.Sum(nil)
		param.IdentityNamespace = data.UserNamespace1
		param.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.AccessorID = data.AccessorID3.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID1.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case 2:
		param.AccessorID = data.AccessorID3.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID1.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	case 3:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.AccessorID = data.AccessorID3.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID3.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	}
	AddAccessor(t, nodeID, privK, param, expected)
}

func UpdateIdentity(t *testing.T, nodeID, privK string, param app.UpdateIdentityParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "UpdateIdentity"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestUpdateIdentity(t *testing.T, caseID int64, expected string) {
	var nodeID string
	var privK string
	var param app.UpdateIdentityParam
	switch caseID {
	case 1:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		ial := 2.3
		param.Ial = &ial
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	}
	UpdateIdentity(t, nodeID, privK, param, expected)
}

func RevokeIdentityAssociation(t *testing.T, nodeID, privK string, param app.RevokeIdentityAssociationParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "RevokeIdentityAssociation"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRevokeIdentityAssociation(t *testing.T, caseID int64, expected string) {
	var nodeID string
	var privK string
	var param app.RevokeIdentityAssociationParam
	switch caseID {
	case 1:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.RequestID = data.RequestID4.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	}
	RevokeIdentityAssociation(t, nodeID, privK, param, expected)
}

func UpdateIdentityModeList(t *testing.T, nodeID, privK string, param app.UpdateIdentityModeListParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "UpdateIdentityModeList"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestUpdateIdentityModeList(t *testing.T, caseID int64, expected string) {
	var nodeID string
	var privK string
	var param app.UpdateIdentityModeListParam
	switch caseID {
	case 1:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.ModeList = append(param.ModeList, 2)
		param.ModeList = append(param.ModeList, 3)
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	}
	UpdateIdentityModeList(t, nodeID, privK, param, expected)
}

func AddIdentity(t *testing.T, nodeID, privK string, param app.AddIdentityParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "AddIdentity"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddIdentity(t *testing.T, caseID int64, expected string) {
	var nodeID string
	var privK string
	var param app.AddIdentityParam
	switch caseID {
	case 1:
		h := sha256.New()
		h.Write([]byte(data.UserNamespace3 + data.UserID1))
		userHash := h.Sum(nil)
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		var identity app.Identity
		identity.IdentityNamespace = data.UserNamespace3
		identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NewIdentityList = append(param.NewIdentityList, identity)
		param.RequestID = data.RequestID6.String()
		nodeID = data.IdP2
		privK = data.IdpPrivK2
	}
	AddIdentity(t, nodeID, privK, param, expected)
}

func RevokeAndAddAccessor(t *testing.T, nodeID, privK string, param app.RevokeAndAddAccessorParam, expected string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "RevokeAndAddAccessor"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRevokeAndAddAccessor(t *testing.T, caseID int64, expected string) {
	var nodeID string
	var privK string
	var param app.RevokeAndAddAccessorParam
	switch caseID {
	case 1:
		param.RevokingAccessorID = data.AccessorID1.String()
		param.AccessorID = data.AccessorID5.String()
		param.AccessorPublicKey = data.AccessorPubKey2
		param.AccessorType = "RSA2048"
		param.RequestID = data.RequestID7.String()
		nodeID = data.IdP1
		privK = data.IdpPrivK1
	}
	RevokeAndAddAccessor(t, nodeID, privK, param, expected)
}
