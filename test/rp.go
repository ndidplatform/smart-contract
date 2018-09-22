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

func SetDataReceived(t *testing.T, param did.SetDataReceivedParam, expected string, nodeID string) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte(nodeID)
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	fnName := "SetDataReceived"
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func CloseRequest(t *testing.T, param did.CloseRequestParam, nodeID string) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte(nodeID)
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	fnName := "CloseRequest"
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TimeOutRequest(t *testing.T, param did.TimeOutRequestParam, nodeID string) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte(nodeID)
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	fnName := "TimeOutRequest"
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}
