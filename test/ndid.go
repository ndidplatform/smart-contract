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
	"log"
	"testing"
	"time"

	"github.com/ndidplatform/smart-contract/abci/did/v1"
	"github.com/tendermint/tendermint/libs/common"
)

func InitNDID(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidpublicKeyBytes, err := generatePublicKey(&ndidKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var initNDIDparam did.InitNDIDParam
	initNDIDparam.NodeID = "NDID"
	initNDIDparam.PublicKey = string(ndidpublicKeyBytes)
	initNDIDparam.MasterPublicKey = string(ndidpublicKeyBytes)

	initNDIDparamJSON, err := json.Marshal(initNDIDparam)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(initNDIDparamJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "InitNDID"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	startTime := time.Now()
	result, _ := callTendermint([]byte(fnName), initNDIDparamJSON, []byte(nonce), signature, []byte(initNDIDparam.NodeID))
	resultObj, _ := result.(ResponseTx)
	if resultObj.Result.CheckTx.Log == "NDID node is already existed" {
		t.SkipNow()
	}
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	stopTime := time.Now()
	writeLog(fnName, (stopTime.UnixNano()-startTime.UnixNano())/int64(time.Millisecond))
	t.Logf("PASS: %s", fnName)
}

func RegisterNode(t *testing.T, param did.RegisterNode) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func SetTimeOutBlockRegisterMsqDestination(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	var param did.TimeOutBlockRegisterMsqDestination
	param.TimeOutBlock = 100
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

	fnName := "SetTimeOutBlockRegisterMsqDestination"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte("NDID"))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func AddNodeToken(t *testing.T, param did.AddNodeTokenParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "AddNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func ReduceNodeToken(t *testing.T, param did.ReduceNodeTokenParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "ReduceNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func SetNodeToken(t *testing.T, param did.SetNodeTokenParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "SetNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func AddService(t *testing.T, param did.AddServiceParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "AddService"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func DisableService(t *testing.T, param did.DisableServiceParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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

	fnName := "DisableService"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func EnableService(t *testing.T, param did.DisableServiceParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "EnableService"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func RegisterServiceDestinationByNDID(t *testing.T, param did.RegisterServiceDestinationByNDIDParam) {
	key := getPrivateKeyFromString(ndidPrivK)
	nodeID := []byte("NDID")
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "RegisterServiceDestinationByNDID"
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, nodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func UpdateService(t *testing.T, param did.UpdateServiceParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "UpdateService"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func SetPriceFunc(t *testing.T, param did.SetPriceFuncParam) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "SetPriceFunc"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func AddNamespace(t *testing.T, param did.Namespace) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	nodeID := "NDID"
	funcparamJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(funcparamJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "AddNamespace"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), funcparamJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func DisableNamespace(t *testing.T, param did.DisableNamespaceParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	nodeID := "NDID"
	funcparamJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(funcparamJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "DisableNamespace"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), funcparamJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func EnableNamespace(t *testing.T, param did.DisableNamespaceParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	nodeID := "NDID"
	funcparamJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(funcparamJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "EnableNamespace"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), funcparamJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func SetValidator(t *testing.T, param did.SetValidatorParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "SetValidator"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func UpdateNodeByNDID(t *testing.T, param did.UpdateNodeByNDIDParam) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "UpdateNodeByNDID"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func DisableNode(t *testing.T, param did.DisableNodeParam) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "DisableNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func EnableNode(t *testing.T, param did.DisableNodeParam) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	fnName := "EnableNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func DisableServiceDestinationByNDID(t *testing.T, param did.DisableServiceDestinationByNDIDParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "DisableServiceDestinationByNDID"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func EnableServiceDestinationByNDID(t *testing.T, param did.DisableServiceDestinationByNDIDParam) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"
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
	fnName := "EnableServiceDestinationByNDID"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func AddNodeToProxyNode(t *testing.T, param did.AddNodeToProxyNodeParam, expected string) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddNodeToProxyNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func UpdateNodeProxyNode(t *testing.T, param did.UpdateNodeProxyNodeParam, expected string) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "UpdateNodeProxyNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func RemoveNodeFromProxyNode(t *testing.T, param did.RemoveNodeFromProxyNode, expected string) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RemoveNodeFromProxyNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}
