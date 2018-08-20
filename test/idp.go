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

package test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
	"github.com/tendermint/tendermint/libs/common"
)

func RegisterMsqDestination(t *testing.T, param did.RegisterMsqDestinationParam, privKeyFile string, nodeID string, expected string) {
	idpKey := getPrivateKeyFromString(privKeyFile)
	idpNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "RegisterMsqDestination"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func DeclareIdentityProof(t *testing.T, param did.DeclareIdentityProofParam, privKeyFile string, nodeID string) {
	idpKey := getPrivateKeyFromString(privKeyFile)
	idpNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "DeclareIdentityProof"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func CreateIdpResponse(t *testing.T, param did.CreateIdpResponseParam, privKeyFile string, nodeID string) {
	idpKey := getPrivateKeyFromString(privKeyFile)
	idpNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "CreateIdpResponse"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func CreateIdentity(t *testing.T, param did.CreateIdentityParam) {
	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "CreateIdentity"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func AddAccessorMethod(t *testing.T, param did.AccessorMethod) {
	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "AddAccessorMethod"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func ClearRegisterMsqDestinationTimeout(t *testing.T, param did.ClearRegisterMsqDestinationTimeoutParam, privKeyFile string, nodeID string) {
	idpKey := getPrivateKeyFromString(privKeyFile)
	idpNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "ClearRegisterMsqDestinationTimeout"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func UpdateIdentity(t *testing.T, param did.UpdateIdentityParam) {
	idpKey := getPrivateKeyFromString(idpPrivK2)
	idpNodeID := []byte("IdP1")
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "UpdateIdentity"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}
