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

package query

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/ndidplatform/smart-contract/v6/abci/app/v1"
	"github.com/ndidplatform/smart-contract/v6/test/data"
	"github.com/ndidplatform/smart-contract/v6/test/utils"
)

func GetAllowedMinIalForRegisterIdentityAtFirstIdp(t *testing.T, expected string) {
	fnName := "GetAllowedMinIalForRegisterIdentityAtFirstIdp"
	result, _ := utils.Query([]byte(fnName), []byte(""))
	resultObj, _ := result.(utils.ResponseQuery)
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

func GetNamespaceList(t *testing.T, expected string) {
	fnName := "GetNamespaceList"
	paramJSON := []byte("")
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func CheckExistingIdentity(t *testing.T, param app.CheckExistingIdentityParam, expected string) {
	fnName := "CheckExistingIdentity"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryCheckExistingIdentity(t *testing.T, namespace, userID, expected string) {
	h := sha256.New()
	h.Write([]byte(namespace + userID))
	userHash := h.Sum(nil)
	var param app.CheckExistingIdentityParam
	param.IdentityNamespace = namespace
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	CheckExistingIdentity(t, param, expected)
}

func GetIdentityInfo(t *testing.T, param app.GetIdentityInfoParam, expected string) {
	fnName := "GetIdentityInfo"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s, Param: %s", fnName, paramJSON)
}

func TestGetIdentityInfo(t *testing.T, caseID int64, expected string) {
	var param app.GetIdentityInfoParam
	switch caseID {
	case 1:
		h := sha256.New()
		h.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h.Sum(nil)
		param.IdentityNamespace = data.UserNamespace1
		param.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NodeID = data.IdP1
	case 2:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.NodeID = data.IdP1
	case 3:
		h := sha256.New()
		h.Write([]byte(data.UserNamespace3 + data.UserID1))
		userHash := h.Sum(nil)
		param.IdentityNamespace = data.UserNamespace3
		param.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.NodeID = data.IdP1
	}
	GetIdentityInfo(t, param, expected)
}

func GetIdpNodes(t *testing.T, param app.GetIdpNodesParam, expected string) {
	fnName := "GetIdpNodes"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); actual != expected {
		t.Errorf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestGetIdpNodes(t *testing.T, caseID int64, expected string) {
	h := sha256.New()
	var param app.GetIdpNodesParam
	switch caseID {
	case 1:
		param.MinIal = 3
		param.MinAal = 3
	case 2:
		h.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h.Sum(nil)
		param.IdentityNamespace = data.UserNamespace1
		param.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.MinIal = 3
		param.MinAal = 3
		param.ModeList = append(param.ModeList, 3)
	case 3:
		h.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h.Sum(nil)
		param.IdentityNamespace = data.UserNamespace1
		param.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.MinIal = 3
		param.MinAal = 3
	case 4:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.MinIal = 3
		param.MinAal = 3
	case 5:
		h.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h.Sum(nil)
		param.IdentityNamespace = data.UserNamespace1
		param.IdentityIdentifierHash = hex.EncodeToString(userHash)
		param.MinIal = 2.3
		param.MinAal = 3
	}
	GetIdpNodes(t, param, expected)
}

func GetIdpNodesInfo(t *testing.T, param app.GetIdpNodesParam, expected string) {
	fnName := "GetIdpNodesInfo"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestGetIdpNodesInfo(t *testing.T, caseID int64, expected string) {
	var param app.GetIdpNodesParam
	switch caseID {
	case 1:
		param.MinIal = 3
		param.MinAal = 3
	case 2:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.MinIal = 3
		param.MinAal = 3
	case 3:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.MinIal = 2.3
		param.MinAal = 3
	case 4:
		param.ReferenceGroupCode = data.ReferenceGroupCode1.String()
		param.MinIal = 2.3
		param.MinAal = 3
		param.SupportedRequestMessageDataUrlTypeList = append(param.SupportedRequestMessageDataUrlTypeList, "text/plain")
		param.SupportedRequestMessageDataUrlTypeList = append(param.SupportedRequestMessageDataUrlTypeList, "application/pdf")
	}
	GetIdpNodesInfo(t, param, expected)
}

func GetReferenceGroupCodeByAccessorID(t *testing.T, param app.GetReferenceGroupCodeByAccessorIDParam, expected string) {
	fnName := "GetReferenceGroupCodeByAccessorID"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
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

func TestGetReferenceGroupCodeByAccessorID(t *testing.T, accessorID, expected string) {
	var param app.GetReferenceGroupCodeByAccessorIDParam
	param.AccessorID = accessorID
	GetReferenceGroupCodeByAccessorID(t, param, expected)
}

func GetReferenceGroupCode(t *testing.T, param app.GetReferenceGroupCodeParam, expected string) {
	fnName := "GetReferenceGroupCode"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
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

func TestGetReferenceGroupCode(t *testing.T, caseID int64, expected string) {
	var param app.GetReferenceGroupCodeParam
	switch caseID {
	case 1:
		h := sha256.New()
		h.Write([]byte(data.UserNamespace1 + data.UserID1))
		userHash := h.Sum(nil)
		param.IdentityNamespace = data.UserNamespace1
		param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	}
	GetReferenceGroupCode(t, param, expected)
}

func GetAsNodesByServiceId(t *testing.T, param app.GetAsNodesByServiceIdParam, expected string) {
	fnName := "GetAsNodesByServiceId"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res app.GetAsNodesByServiceIdResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestGetAsNodesByServiceId(t *testing.T, caseID int64, expected string) {
	var param app.GetAsNodesByServiceIdParam
	switch caseID {
	case 1:
		param.ServiceID = data.ServiceID1
	}
	GetAsNodesByServiceId(t, param, expected)
}

func GetAsNodesInfoByServiceId(t *testing.T, param app.GetAsNodesByServiceIdParam, expected string) {
	fnName := "GetAsNodesInfoByServiceId"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res app.GetAsNodesByServiceIdResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resultObj.Result.Response.Log == expected {
		t.Logf("PASS: %s", fnName)
		return
	}
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %s\nActual: %s", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestGetAsNodesInfoByServiceId(t *testing.T, caseID int64, expected string) {
	var param app.GetAsNodesByServiceIdParam
	switch caseID {
	case 1:
		param.ServiceID = data.ServiceID1
	}
	GetAsNodesInfoByServiceId(t, param, expected)
}

func GetServicesByAsID(t *testing.T, param app.GetServicesByAsIDParam, expected string) {
	fnName := "GetServicesByAsID"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
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

func TestGetServicesByAsID(t *testing.T, caseID int64, expected string) {
	var param app.GetServicesByAsIDParam
	switch caseID {
	case 1:
		param.AsID = data.AS1
	}
	GetServicesByAsID(t, param, expected)
}

func GetAccessorKey(t *testing.T, param app.GetAccessorGroupIDParam, expected string) {
	fnName := "GetAccessorKey"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
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

func TestGetAccessorKey(t *testing.T, accessorID, expected string) {
	var param app.GetAccessorGroupIDParam
	param.AccessorID = accessorID
	GetAccessorKey(t, param, expected)
}

func GetAllowedModeList(t *testing.T, param app.GetAllowedModeListParam, expected string) {
	fnName := "GetAllowedModeList"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
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

func TestGetAllowedModeList(t *testing.T, purpose, expected string) {
	var param app.GetAllowedModeListParam
	param.Purpose = purpose
	GetAllowedModeList(t, param, expected)
}

func GetNodeInfo(t *testing.T, param app.GetNodeInfoParam, expected string) {
	fnName := "GetNodeInfo"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
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

func TestGetNodeInfo(t *testing.T, nodeID, expected string) {
	var param app.GetNodeInfoParam
	param.NodeID = nodeID
	GetNodeInfo(t, param, expected)
}

func getErrorCodeList(t *testing.T, param app.GetErrorCodeListParam, expected string) {
	fnName := "GetErrorCodeList"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := utils.Query([]byte(fnName), paramJSON)
	resultObj, _ := result.(utils.ResponseQuery)
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

func TestGetErrorCodeList(t *testing.T, errorCodeType string, expected string) {
	var param app.GetErrorCodeListParam
	param.Type = errorCodeType
	getErrorCodeList(t, param, expected)
}
