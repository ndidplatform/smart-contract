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

package as

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ndidplatform/smart-contract/v9/abci/app/v1"
	"github.com/ndidplatform/smart-contract/v9/test/data"
	"github.com/ndidplatform/smart-contract/v9/test/utils"
)

func RegisterServiceDestination(t *testing.T, nodeID, privK string, param app.RegisterServiceDestinationParam, expected string, expectResultFrom string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "RegisterServiceDestination"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	var actual string
	if expectResultFrom == "CheckTx" {
		actual = resultObj.Result.CheckTx.Log
	} else {
		actual = resultObj.Result.TxResult.Log
	}
	if actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRegisterServiceDestination(t *testing.T, caseID int64, expected string, expectResultFrom string) {
	var nodeID string
	var privK string
	var param app.RegisterServiceDestinationParam
	switch caseID {
	case 1:
		param.ServiceID = data.ServiceID1
		param.MinAal = 1.1
		param.MinIal = 1.2
		param.SupportedNamespaceList = append(param.SupportedNamespaceList, data.UserNamespace1)
		nodeID = data.AS1
		privK = data.AsPrivK1
	case 2:
		param.ServiceID = data.ServiceID1
		param.MinAal = 1.1
		param.MinIal = 1.2
		param.SupportedNamespaceList = append(param.SupportedNamespaceList, data.UserNamespace1)
		nodeID = data.AS2
		privK = data.AsPrivK2
	}
	RegisterServiceDestination(t, nodeID, privK, param, expected, expectResultFrom)
}

func UpdateServiceDestination(t *testing.T, nodeID, privK string, param app.UpdateServiceDestinationParam, expected string, expectResultFrom string) {
	privKey := utils.GetPrivateKeyFromString(privK)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "UpdateServiceDestination"
	nonce, signature := utils.CreateSignatureAndNonce(fnName, paramJSON, privKey)
	result, _ := utils.CreateTxn([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	var actual string
	if expectResultFrom == "CheckTx" {
		actual = resultObj.Result.CheckTx.Log
	} else {
		actual = resultObj.Result.TxResult.Log
	}
	if actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestUpdateServiceDestination(t *testing.T, caseID int64, expected string, expectResultFrom string) {
	var nodeID string
	var privK string
	var param app.UpdateServiceDestinationParam
	switch caseID {
	case 1:
		param.ServiceID = data.ServiceID1
		param.MinAal = 1.4
		param.MinIal = 1.5
		param.SupportedNamespaceList = append(param.SupportedNamespaceList, data.UserNamespace2)
		nodeID = data.AS1
		privK = data.AsPrivK1
	}
	UpdateServiceDestination(t, nodeID, privK, param, expected, expectResultFrom)
}
