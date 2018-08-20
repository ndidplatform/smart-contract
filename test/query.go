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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
)

func GetNodePublicKey(t *testing.T, param did.GetNodePublicKeyParam, expected string) {
	fnName := "GetNodePublicKey"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetNodePublicKeyResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := res.PublicKey; actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetNodeMasterPublicKey(t *testing.T, param did.GetNodePublicKeyParam, expected string) {
	fnName := "GetNodeMasterPublicKey"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetNodeMasterPublicKeyResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := res.MasterPublicKey; actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetNodeToken(t *testing.T, param did.GetNodeTokenParam, expected did.GetNodeTokenResult) {
	fnName := "GetNodeToken"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetNodeTokenExpectString(t *testing.T, param did.GetNodeTokenParam, expected string) {
	fnName := "GetNodeToken"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetIdpNodes(t *testing.T, param did.GetIdpNodesParam, expected []did.MsqDestinationNode) {
	fnName := "GetIdpNodes"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetIdpNodesResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res.Node; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetIdpNodesExpectString(t *testing.T, param did.GetIdpNodesParam, expected string) {
	fnName := "GetIdpNodes"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetMsqAddress(t *testing.T, param did.GetMsqAddressParam, expected did.MsqAddress) {
	fnName := "GetMsqAddress"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.MsqAddress
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetMsqAddressExpectString(t *testing.T, param did.GetMsqAddressParam, expected string) {
	fnName := "GetMsqAddress"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.MsqAddress
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetServiceDetail(t *testing.T, param did.GetServiceDetailParam, expected did.ServiceDetail) {
	fnName := "GetServiceDetail"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.ServiceDetail
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetServiceDetailExpectString(t *testing.T, param did.GetServiceDetailParam, expected string) {
	fnName := "GetServiceDetail"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.ServiceDetail
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetAsNodesByServiceId(t *testing.T, param did.GetAsNodesByServiceIdParam, expected string) {
	fnName := "GetAsNodesByServiceId"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetAsNodesByServiceIdResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetIdentityProof(t *testing.T, param did.GetIdentityProofParam, expected did.GetIdentityProofResult) {
	fnName := "GetIdentityProof"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetIdentityProofResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetIdentityProofExpectString(t *testing.T, param did.GetIdentityProofParam, expected string) {
	fnName := "GetIdentityProof"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetIdentityProofResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetPriceFunc(t *testing.T, param did.GetPriceFuncParam, expected did.GetPriceFuncResult) {
	fnName := "GetPriceFunc"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetPriceFuncResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetUsedTokenReport(t *testing.T, param did.GetUsedTokenReportParam, expectedString string) {
	fnName := "GetUsedTokenReport"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res []did.Report
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected []did.Report
	json.Unmarshal([]byte(expectedString), &expected)
	if resultObj.Result.Response.Log == expectedString {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetRequestDetail(t *testing.T, param did.GetRequestParam, expected string) {
	fnName := "GetRequestDetail"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetRequest(t *testing.T, param did.GetRequestParam, expected did.GetRequestResult) {
	fnName := "GetRequest"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetRequestResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetRequestExpectString(t *testing.T, param did.GetRequestParam, expected string) {
	fnName := "GetRequest"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetRequestResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetNamespaceList(t *testing.T, expected []did.Namespace) {
	fnName := "GetNamespaceList"
	paramJSON := []byte("")
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res []did.Namespace
	err := json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetNamespaceListExpectString(t *testing.T, expected string) {
	fnName := "GetNamespaceList"
	paramJSON := []byte("")
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func CheckExistingIdentity(t *testing.T, param did.CheckExistingIdentityParam, expected string) {
	fnName := "CheckExistingIdentity"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetAccessorGroupID(t *testing.T, param did.GetAccessorGroupIDParam, expected string) {
	fnName := "GetAccessorGroupID"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetAccessorKey(t *testing.T, param did.GetAccessorGroupIDParam, expected string) {
	fnName := "GetAccessorKey"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetServiceList(t *testing.T, expected []did.ServiceDetail) {
	fnName := "GetServiceList"
	paramJSON := []byte("")
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res []did.ServiceDetail
	err := json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetNodeInfo(t *testing.T, param did.GetNodeInfoParam, expected string) {
	fnName := "GetNodeInfo"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func CheckExistingAccessorID(t *testing.T, param did.CheckExistingAccessorIDParam, expected string) {
	fnName := "CheckExistingAccessorID"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func CheckExistingAccessorGroupID(t *testing.T, param did.CheckExistingAccessorGroupIDParam, expected string) {
	fnName := "CheckExistingAccessorGroupID"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetIdentityInfo(t *testing.T, param did.GetIdentityInfoParam, expected string) {
	fnName := "GetIdentityInfo"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetServicesByAsID(t *testing.T, param did.GetServicesByAsIDParam, expected string) {
	fnName := "GetServicesByAsID"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func GetDataSignature(t *testing.T, param did.GetDataSignatureParam, expected string) {
	fnName := "GetDataSignature"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}
