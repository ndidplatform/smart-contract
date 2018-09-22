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

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
	"github.com/tendermint/tendermint/libs/common"
)

func RegisterServiceDestination(t *testing.T, param did.RegisterServiceDestinationParam, priveKFile string, nodeID string, expected string) {
	asKey := getPrivateKeyFromString(priveKFile)
	asNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	fnName := "RegisterServiceDestination"
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func UpdateServiceDestination(t *testing.T, param did.UpdateServiceDestinationParam, nodeID string) {
	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	fnName := "UpdateServiceDestination"
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func SignData(t *testing.T, param did.SignDataParam, expected string, nodeID string) {
	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "SignData"
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func DisableServiceDestination(t *testing.T, param did.DisableServiceDestinationParam, nodeID string) {
	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	fnName := "DisableServiceDestination"
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func EnableServiceDestination(t *testing.T, param did.DisableServiceDestinationParam, nodeID string) {
	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte(nodeID)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	fnName := "EnableServiceDestination"
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}
