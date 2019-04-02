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
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"
	"testing"

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
	uuid "github.com/satori/go.uuid"
)

var RP1 = RandStringRunes(20)
var IdP1 = RandStringRunes(20)
var IdP2 = RandStringRunes(20)
var IdP4 = RandStringRunes(20)
var IdP5 = RandStringRunes(20)
var IdP6 = RandStringRunes(20)
var AS1 = RandStringRunes(20)
var AS2 = RandStringRunes(20)
var Proxy1 = RandStringRunes(20)
var Proxy2 = RandStringRunes(20)

var IdP6BehindProxy1 = RandStringRunes(20)
var AS3BehindProxy1 = RandStringRunes(20)

var serviceID1 = RandStringRunes(20)
var serviceID2 = RandStringRunes(20)

var requestID1 = uuid.NewV4()
var requestID2 = uuid.NewV4()
var requestID3 = uuid.NewV4()
var requestID4 = uuid.NewV4()
var requestID5 = uuid.NewV4()
var requestID6 = uuid.NewV4()
var namespaceID1 = RandStringRunes(20)
var namespaceID2 = RandStringRunes(20)
var accessorID1 = uuid.NewV4()
var accessorID2 = uuid.NewV4()
var accessorID3 = uuid.NewV4()

// var accessorGroupID1 = uuid.NewV4()

var serviceID3 = RandStringRunes(20)
var serviceID4 = RandStringRunes(20)
var serviceID5 = RandStringRunes(20)

var serviceID6 = RandStringRunes(20)

var userID1 = RandStringRunes(20)
var userID2 = RandStringRunes(20)

var referenceGroupCode1 = uuid.NewV4()

func TestInitNDID(t *testing.T) {
	InitNDID(t)
	IsInitEnded(t, false)
	EndInit(t)
	IsInitEnded(t, true)
}

func TestRegisterNodeIDP1(t *testing.T) {
	idpKey := getPrivateKeyFromString(idpPrivK)
	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	idpMasterKey := getPrivateKeyFromString(allMasterKey)
	idpMasterPublicKeyBytes, err := generatePublicKey(&idpMasterKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var param did.RegisterNode
	param.NodeID = IdP1
	param.PublicKey = string(idpPublicKeyBytes)
	param.MasterPublicKey = string(idpMasterPublicKeyBytes)
	param.NodeName = "IdP Number 1"
	param.Role = "IdP"
	param.MaxIal = 3.0
	param.MaxAal = 3.0
	RegisterNode(t, param)
}

func TestSetNodeTokenIDP1(t *testing.T) {
	var param = did.SetNodeTokenParam{
		IdP1,
		100.0,
	}
	SetNodeToken(t, param)
}

func TestRegisterNodeIDP2(t *testing.T) {
	idpKey := getPrivateKeyFromString(idpPrivK2)
	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	idpMasterKey := getPrivateKeyFromString(allMasterKey)
	idpMasterPublicKeyBytes, err := generatePublicKey(&idpMasterKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var param did.RegisterNode
	param.NodeID = IdP2
	param.PublicKey = string(idpPublicKeyBytes)
	param.MasterPublicKey = string(idpMasterPublicKeyBytes)
	param.NodeName = "IdP Number 2"
	param.Role = "IdP"
	param.MaxIal = 2.3
	param.MaxAal = 3.0
	RegisterNode(t, param)
}

func TestSetNodeTokenIDP2(t *testing.T) {
	var param = did.SetNodeTokenParam{
		IdP2,
		100.0,
	}
	SetNodeToken(t, param)
}

func TestIdP1SetMqAddresses(t *testing.T) {
	var mq did.MsqAddress
	mq.IP = "192.168.3.99"
	mq.Port = 8000
	var param did.SetMqAddressesParam
	param.Addresses = make([]did.MsqAddress, 0)
	param.Addresses = append(param.Addresses, mq)
	SetMqAddresses(t, param, idpPrivK, IdP1)
}

func TestIdP2SetMqAddresses(t *testing.T) {
	var mq did.MsqAddress
	mq.IP = "192.168.3.100"
	mq.Port = 8000
	var param did.SetMqAddressesParam
	param.Addresses = make([]did.MsqAddress, 0)
	param.Addresses = append(param.Addresses, mq)
	SetMqAddresses(t, param, idpPrivK2, IdP2)
}

func TestQueryCheckExistingIdentityBeforeRegister(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	var param did.CheckExistingIdentityParam
	param.IdentityNamespace = userNamespace
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	var expected = `{"exist":false}`
	CheckExistingIdentity(t, param, expected)
}

func TestIdP1CreateRequestForRegisterIdentity(t *testing.T) {
	var datas []did.DataRequest
	var param did.Request
	param.RequestID = requestID1.String()
	param.MinIdp = 0
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3
	param.Purpose = "RegisterIdentity"
	CreateRequest(t, param, idpPrivK, IdP1)
}

func TestIdP1CloseRequestForRegisterIdentity(t *testing.T) {
	var res []did.ResponseValid
	var param = did.CloseRequestParam{
		requestID1.String(),
		res,
	}
	CloseRequestByIdP(t, param, idpPrivK, IdP1)
}

func TestIdP1RegisterIdentityWithOutRefGroupCode(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	var user did.RegisterIdentityParam
	user.ReferenceGroupCode = ""
	var identity did.Identity
	identity.IdentityNamespace = userNamespace
	identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
	user.NewIdentityList = append(user.NewIdentityList, identity)
	user.Ial = 3
	user.ModeList = append(user.ModeList, 2)
	user.AccessorID = accessorID1.String()
	user.AccessorPublicKey = accessorPubKey1
	user.AccessorType = "RSA2048"
	user.RequestID = requestID1.String()
	var param = user
	RegisterIdentity(t, param, idpPrivK, IdP1, "Please input reference group code")
}

func TestIdP1RegisterIdentityWithDuplicateNamespace(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	h2 := sha256.New()
	h2.Write([]byte(userNamespace + userID2))
	userHash2 := h.Sum(nil)
	var user did.RegisterIdentityParam
	user.ReferenceGroupCode = referenceGroupCode1.String()
	var identity did.Identity
	identity.IdentityNamespace = userNamespace
	identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
	user.NewIdentityList = append(user.NewIdentityList, identity)
	var identity2 did.Identity
	identity2.IdentityNamespace = userNamespace
	identity2.IdentityIdentifierHash = hex.EncodeToString(userHash2)
	user.NewIdentityList = append(user.NewIdentityList, identity2)
	user.Ial = 3
	user.ModeList = append(user.ModeList, 2)
	user.AccessorID = accessorID1.String()
	user.AccessorPublicKey = accessorPubKey1
	user.AccessorType = "RSA2048"
	user.RequestID = requestID1.String()
	var param = user
	RegisterIdentity(t, param, idpPrivK, IdP1, "Namespace in identity list are duplicated")
}

func TestIdP1RegisterIdentity(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	var user did.RegisterIdentityParam
	user.ReferenceGroupCode = referenceGroupCode1.String()
	var identity did.Identity
	identity.IdentityNamespace = userNamespace
	identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
	user.NewIdentityList = append(user.NewIdentityList, identity)
	user.Ial = 3
	user.ModeList = append(user.ModeList, 2)
	user.AccessorID = accessorID1.String()
	user.AccessorPublicKey = accessorPubKey1
	user.AccessorType = "RSA2048"
	user.RequestID = requestID1.String()
	var param = user
	RegisterIdentity(t, param, idpPrivK, IdP1, "success")
}

func TestQueryGetIdentityInfoByIdentity(t *testing.T) {
	var param did.GetIdentityInfoParam
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	param.IdentityNamespace = userNamespace
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	param.NodeID = IdP1
	expected := `{"ial":3,"mode_list":[2]}`
	GetIdentityInfo(t, param, expected)
}

func TestQueryGetIdentityInfoByRefGroupCode(t *testing.T) {
	var param did.GetIdentityInfoParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.NodeID = IdP1
	expected := `{"ial":3,"mode_list":[2]}`
	GetIdentityInfo(t, param, expected)
}

func TestQueryCheckExistingIdentityAfterRegister(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	var param did.CheckExistingIdentityParam
	param.IdentityNamespace = userNamespace
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	var expected = `{"exist":true}`
	CheckExistingIdentity(t, param, expected)
}

func TestQueryGetIdpNodesForMode1(t *testing.T) {
	var param did.GetIdpNodesParam
	param.MinIal = 3
	param.MinAal = 3
	var expected = `{"node":[{"node_id":"` + IdP1 + `","node_name":"IdP Number 1","max_ial":3,"max_aal":3,"mode_list":[1]}]}`
	GetIdpNodesExpectString(t, param, expected)
}

func TestQueryGetIdpNodesInfoForMode1(t *testing.T) {
	var param did.GetIdpNodesParam
	param.MinIal = 3
	param.MinAal = 3
	var expected = `{"node":[{"node_id":"` + IdP1 + `","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"mode_list":[1]}]}`
	GetIdpNodesInfo(t, param, expected)
}

func TestQueryGetIdpNodes1ByIdentity(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	var param did.GetIdpNodesParam
	param.IdentityNamespace = userNamespace
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	param.MinIal = 3
	param.MinAal = 3
	var expected = `{"node":[{"node_id":"` + IdP1 + `","node_name":"IdP Number 1","max_ial":3,"max_aal":3,"mode_list":[2]}]}`
	GetIdpNodesExpectString(t, param, expected)
}

func TestQueryGetIdpNodes1ByRefGroupCode(t *testing.T) {
	var param did.GetIdpNodesParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.MinIal = 3
	param.MinAal = 3
	var expected = `{"node":[{"node_id":"` + IdP1 + `","node_name":"IdP Number 1","max_ial":3,"max_aal":3,"mode_list":[2]}]}`
	GetIdpNodesExpectString(t, param, expected)
}

func TestQueryGetIdpNodesInfoByRefGroupCode(t *testing.T) {
	var param did.GetIdpNodesParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.MinIal = 3
	param.MinAal = 3
	var expected = `{"node":[{"node_id":"` + IdP1 + `","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"mode_list":[2]}]}`
	GetIdpNodesInfo(t, param, expected)
}

func TestIdP2CreateRequestForRegisterIdentity(t *testing.T) {
	var datas []did.DataRequest
	var param did.Request
	param.RequestID = requestID2.String()
	param.MinIdp = 1
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3
	param.Purpose = "RegisterIdentity"
	param.IdPIDList = append(param.IdPIDList, IdP1)
	CreateRequest(t, param, idpPrivK2, IdP2)
}

func TestIdPCreateIdpResponse(t *testing.T) {
	var param did.CreateIdpResponseParam
	param.Aal = 3
	param.Ial = 3
	param.RequestID = requestID2.String()
	param.Signature = "signature"
	param.Status = "accept"
	CreateIdpResponse(t, param, idpPrivK, IdP1)
}

func TestIdP2CloseRequestForRegisterIdentity(t *testing.T) {
	var res []did.ResponseValid
	var res1 did.ResponseValid
	res1.IdpID = IdP1
	tValue := true
	res1.ValidIal = &tValue
	res1.ValidSignature = &tValue
	res = append(res, res1)
	var param = did.CloseRequestParam{
		requestID2.String(),
		res,
	}
	CloseRequestByIdP(t, param, idpPrivK2, IdP2)
}

func TestIdP2RegisterIdentityToExistedRefGroupExpectError(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	var user did.RegisterIdentityParam
	user.ReferenceGroupCode = referenceGroupCode1.String()
	var identity did.Identity
	identity.IdentityNamespace = userNamespace
	identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
	user.NewIdentityList = append(user.NewIdentityList, identity)
	user.Ial = 2.3
	user.ModeList = append(user.ModeList, 2)
	user.AccessorID = accessorID2.String()
	user.AccessorPublicKey = accessorPubKey2
	user.AccessorType = "RSA2048"
	user.RequestID = requestID2.String()
	var param = user
	RegisterIdentity(t, param, idpPrivK2, IdP2, "Identity already existed")
}

func TestIdP2RegisterIdentityWithDuplicatedNamespaceToExistedRefGroupExpectError(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID2))
	userHash := h.Sum(nil)
	var user did.RegisterIdentityParam
	user.ReferenceGroupCode = referenceGroupCode1.String()
	var identity did.Identity
	identity.IdentityNamespace = userNamespace
	identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
	user.NewIdentityList = append(user.NewIdentityList, identity)
	user.Ial = 2.3
	user.ModeList = append(user.ModeList, 2)
	user.AccessorID = accessorID2.String()
	user.AccessorPublicKey = accessorPubKey2
	user.AccessorType = "RSA2048"
	user.RequestID = requestID2.String()
	var param = user
	RegisterIdentity(t, param, idpPrivK2, IdP2, "Namespace in identity list are duplicated")
}

func TestIdP2RegisterIdentityToExistedRefGroup(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace2 + userID1))
	userHash := h.Sum(nil)
	var user did.RegisterIdentityParam
	user.ReferenceGroupCode = referenceGroupCode1.String()
	var identity did.Identity
	identity.IdentityNamespace = userNamespace2
	identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
	user.NewIdentityList = append(user.NewIdentityList, identity)
	user.Ial = 2.3
	user.ModeList = append(user.ModeList, 2)
	user.AccessorID = accessorID2.String()
	user.AccessorPublicKey = accessorPubKey2
	user.AccessorType = "RSA2048"
	user.RequestID = requestID2.String()
	var param = user
	RegisterIdentity(t, param, idpPrivK2, IdP2, "success")
}

func TestQueryGetIdpNodes1ByIdentity2(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	var param did.GetIdpNodesParam
	param.IdentityNamespace = userNamespace
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	param.MinIal = 2.3
	param.MinAal = 3
	var expected = `{"node":[{"node_id":"` + IdP1 + `","node_name":"IdP Number 1","max_ial":3,"max_aal":3,"mode_list":[2]},{"node_id":"` + IdP2 + `","node_name":"IdP Number 2","max_ial":2.3,"max_aal":3,"mode_list":[2]}]}`
	GetIdpNodesExpectString(t, param, expected)
}

func TestQueryGetReferenceGroupCodeByAccessorID1(t *testing.T) {
	var param did.GetReferenceGroupCodeByAccessorIDParam
	param.AccessorID = accessorID3.String()
	var expected = `{"reference_group_code":""}`
	GetReferenceGroupCodeByAccessorIDExpectString(t, param, expected)
}
func TestQueryGetReferenceGroupCodeByAccessorID2(t *testing.T) {
	var param did.GetReferenceGroupCodeByAccessorIDParam
	param.AccessorID = accessorID1.String()
	var expected = `{"reference_group_code":"` + referenceGroupCode1.String() + `"}`
	GetReferenceGroupCodeByAccessorIDExpectString(t, param, expected)
}

func TestIdPAddAccessorMethodWithInvalidParameter1(t *testing.T) {
	var param did.AccessorMethod
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	param.IdentityNamespace = userNamespace
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.AccessorID = accessorID3.String()
	param.AccessorPublicKey = accessorPubKey2
	param.AccessorType = "RSA2048"
	param.RequestID = requestID1.String()
	AddAccessorMethod(t, param, idpPrivK2, IdP2, "Found reference group code and identity detail in parameter")
}

func TestIdPAddAccessorMethodWithInvalidParameter2(t *testing.T) {
	var param did.AccessorMethod
	param.AccessorID = accessorID3.String()
	param.AccessorPublicKey = accessorPubKey2
	param.AccessorType = "RSA2048"
	param.RequestID = requestID1.String()
	AddAccessorMethod(t, param, idpPrivK2, IdP2, "Reference group not found")
}

func TestIdP2CreateRequestForAddAccessorMethod(t *testing.T) {
	var datas []did.DataRequest
	var param did.Request
	param.RequestID = requestID3.String()
	param.MinIdp = 1
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3
	param.Purpose = "AddAccessorMethod"
	param.IdPIDList = append(param.IdPIDList, IdP1)
	CreateRequest(t, param, idpPrivK2, IdP2)
}

func TestIdP1CreateIdpResponse(t *testing.T) {
	var param did.CreateIdpResponseParam
	param.Aal = 3
	param.Ial = 3
	param.RequestID = requestID3.String()
	param.Signature = "signature"
	param.Status = "accept"
	CreateIdpResponse(t, param, idpPrivK, IdP1)
}

func TestIdP2CloseRequestForrequestID3(t *testing.T) {
	var res []did.ResponseValid
	var res1 did.ResponseValid
	res1.IdpID = IdP1
	tValue := true
	res1.ValidIal = &tValue
	res1.ValidSignature = &tValue
	res = append(res, res1)
	var param = did.CloseRequestParam{
		requestID3.String(),
		res,
	}
	CloseRequestByIdP(t, param, idpPrivK2, IdP2)
}

func TestIdPAddAccessorMethod(t *testing.T) {
	var param did.AccessorMethod
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.AccessorID = accessorID3.String()
	param.AccessorPublicKey = accessorPubKey2
	param.AccessorType = "RSA2048"
	param.RequestID = requestID3.String()
	AddAccessorMethod(t, param, idpPrivK2, IdP2, "success")
}

func TestQueryGetReferenceGroupCodeByAccessorID3(t *testing.T) {
	var param did.GetReferenceGroupCodeByAccessorIDParam
	param.AccessorID = accessorID3.String()
	var expected = `{"reference_group_code":"` + referenceGroupCode1.String() + `"}`
	GetReferenceGroupCodeByAccessorIDExpectString(t, param, expected)
}
func TestQueryGetReferenceGroupCode1(t *testing.T) {
	var param did.GetReferenceGroupCodeParam
	h := sha256.New()
	h.Write([]byte(userNamespace + userID1))
	userHash := h.Sum(nil)
	param.IdentityNamespace = userNamespace
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	var expected = `{"reference_group_code":"` + referenceGroupCode1.String() + `"}`
	GetReferenceGroupCodeExpectString(t, param, expected)
}

func TestNDIDAddService(t *testing.T) {
	var param did.AddServiceParam
	param.ServiceID = serviceID1
	param.ServiceName = "Bank statement"
	param.DataSchema = "DataSchema"
	param.DataSchemaVersion = "DataSchemaVersion"
	AddService(t, param)
}

func TestRegisterNodeAS(t *testing.T) {
	asKey := getPrivateKeyFromString(asPrivK)
	asPublicKeyBytes, err := generatePublicKey(&asKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	asMasterKey := getPrivateKeyFromString(allMasterKey)
	asMasterPublicKeyBytes, err := generatePublicKey(&asMasterKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var param did.RegisterNode
	param.NodeName = "AS1"
	param.NodeID = AS1
	param.PublicKey = string(asPublicKeyBytes)
	param.MasterPublicKey = string(asMasterPublicKeyBytes)
	param.Role = "AS"
	RegisterNode(t, param)
}

func TestSetNodeTokenAS1(t *testing.T) {
	var param = did.SetNodeTokenParam{
		AS1,
		100.0,
	}
	SetNodeToken(t, param)
}

func TestAS1SetMqAddresses(t *testing.T) {
	var mq did.MsqAddress
	mq.IP = "192.168.3.102"
	mq.Port = 8000
	var param did.SetMqAddressesParam
	param.Addresses = make([]did.MsqAddress, 0)
	param.Addresses = append(param.Addresses, mq)
	SetMqAddresses(t, param, asPrivK, AS1)
}

func TestRegisterServiceDestinationByNDIDForInvalidNodeID(t *testing.T) {
	var param = did.RegisterServiceDestinationByNDIDParam{
		serviceID1,
		"Invalid-node-ID",
	}
	RegisterServiceDestinationByNDIDExpectedString(t, param, "Node ID not found")
}

func TestRegisterServiceDestinationByNDIDForInvalidRole(t *testing.T) {
	var param = did.RegisterServiceDestinationByNDIDParam{
		serviceID1,
		IdP1,
	}
	RegisterServiceDestinationByNDIDExpectedString(t, param, "Role of node ID is not AS")
}

func TestASRegisterServiceDestinationByNDIDForAS1(t *testing.T) {
	var param = did.RegisterServiceDestinationByNDIDParam{
		serviceID1,
		AS1,
	}
	RegisterServiceDestinationByNDID(t, param)
}

func TestASRegisterServiceDestination(t *testing.T) {
	var param did.RegisterServiceDestinationParam
	param.ServiceID = serviceID1
	param.MinAal = 1.1
	param.MinIal = 1.2
	param.AcceptedNamespaceList = append(param.AcceptedNamespaceList, userNamespace)
	RegisterServiceDestination(t, param, asPrivK, AS1, "success")
}

func TestQueryGetAsNodesByServiceId(t *testing.T) {
	var param did.GetAsNodesByServiceIdParam
	param.ServiceID = serviceID1
	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":1.2,"min_aal":1.1,"accepted_namespace_list":["cid"]}]}`
	GetAsNodesByServiceId(t, param, expected)
}

func TestQueryGetAsNodesInfoByServiceId(t *testing.T) {
	var param did.GetAsNodesByServiceIdParam
	param.ServiceID = serviceID1
	var expected = `{"node":[{"node_id":"` + AS1 + `","name":"AS1","min_ial":1.2,"min_aal":1.1,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.102","port":8000}],"accepted_namespace_list":["cid"]}]}`
	GetAsNodesInfoByServiceId(t, param, expected)
}

func TestAS1UpdateServiceDestination(t *testing.T) {
	var param did.UpdateServiceDestinationParam
	param.ServiceID = serviceID1
	param.MinAal = 1.4
	param.MinIal = 1.5
	param.AcceptedNamespaceList = append(param.AcceptedNamespaceList, userNamespace2)
	UpdateServiceDestination(t, param, AS1)
}

func TestQueryGetAsNodesByServiceIdAfterUpdated(t *testing.T) {
	var param did.GetAsNodesByServiceIdParam
	param.ServiceID = serviceID1
	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":1.5,"min_aal":1.4,"accepted_namespace_list":["passport"]}]}`
	GetAsNodesByServiceId(t, param, expected)
}

func TestQueryGetServicesByAsID(t *testing.T) {
	var param = did.GetServicesByAsIDParam{
		AS1,
	}
	var expected = `{"services":[{"service_id":"` + serviceID1 + `","min_ial":1.5,"min_aal":1.4,"active":true,"suspended":false,"accepted_namespace_list":["cid"]}]}`
	GetServicesByAsID(t, param, expected)
}

func TestQueryGetIdentityInfoByRefGroupCodeBeforeUpdate(t *testing.T) {
	var param did.GetIdentityInfoParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.NodeID = IdP1
	expected := `{"ial":3,"mode_list":[2]}`
	GetIdentityInfo(t, param, expected)
}

func TestIdP1UpdateIdentity(t *testing.T) {
	var param did.UpdateIdentityParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.Ial = 2.3
	UpdateIdentity(t, param, idpPrivK, IdP1)
}

func TestQueryGetIdentityInfoByRefGroupCodeAfterUpdate(t *testing.T) {
	var param did.GetIdentityInfoParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.NodeID = IdP1
	expected := `{"ial":2.3,"mode_list":[2]}`
	GetIdentityInfo(t, param, expected)
}

func TestQueryGetIdpNodesInfoByRefGroupCodeBeforeRevokeIdentityAssociation(t *testing.T) {
	var param did.GetIdpNodesParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.MinIal = 2.3
	param.MinAal = 3
	var expected = `{"node":[{"node_id":"` + IdP1 + `","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"mode_list":[2]},{"node_id":"` + IdP2 + `","name":"IdP Number 2","max_ial":2.3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.100","port":8000}],"mode_list":[2]}]}`
	GetIdpNodesInfo(t, param, expected)
}

func TestIdP2CreateRequestForRevokeIdentityAssociation(t *testing.T) {
	var datas []did.DataRequest
	var param did.Request
	param.RequestID = requestID4.String()
	param.MinIdp = 1
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3
	param.Purpose = "RevokeIdentityAssociation"
	param.IdPIDList = append(param.IdPIDList, IdP1)
	CreateRequest(t, param, idpPrivK2, IdP2)
}

func TestIdP1CreateIdpResponseForRevokeIdentityAssociation(t *testing.T) {
	var param did.CreateIdpResponseParam
	param.Aal = 3
	param.Ial = 3
	param.RequestID = requestID4.String()
	param.Signature = "signature"
	param.Status = "accept"
	CreateIdpResponse(t, param, idpPrivK, IdP1)
}

func TestIdP2CloseRequestForRevokeIdentityAssociation(t *testing.T) {
	var res []did.ResponseValid
	var res1 did.ResponseValid
	res1.IdpID = IdP1
	tValue := true
	res1.ValidIal = &tValue
	res1.ValidSignature = &tValue
	res = append(res, res1)
	var param = did.CloseRequestParam{
		requestID4.String(),
		res,
	}
	CloseRequestByIdP(t, param, idpPrivK2, IdP2)
}

func TestIdP2RevokeIdentityAssociation(t *testing.T) {
	var param did.RevokeIdentityAssociationParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.RequestID = requestID4.String()
	RevokeIdentityAssociation(t, param, idpPrivK2, IdP2)
}

func TestQueryGetIdpNodesInfoByRefGroupCodeAfterRevokeIdentityAssociation(t *testing.T) {
	var param did.GetIdpNodesParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.MinIal = 2.3
	param.MinAal = 3
	var expected = `{"node":[{"node_id":"` + IdP1 + `","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"mode_list":[2]}]}`
	GetIdpNodesInfo(t, param, expected)
}

func TestQueryGetAccessorKey1(t *testing.T) {
	var param = did.GetAccessorGroupIDParam{
		accessorID1.String(),
	}
	var expected = `{"accessor_public_key":"` + strings.Replace(accessorPubKey1, "\n", "\\n", -1) + `","active":true}`
	GetAccessorKey(t, param, expected)
}

func TestQueryGetAccessorKey2(t *testing.T) {
	var param = did.GetAccessorGroupIDParam{
		accessorID2.String(),
	}
	var expected = `{"accessor_public_key":"` + strings.Replace(accessorPubKey2, "\n", "\\n", -1) + `","active":true}`
	GetAccessorKey(t, param, expected)
}

func TestQueryGetAllowedModeListExpectString(t *testing.T) {
	var param did.GetAllowedModeListParam
	var expected = `{"allowed_mode_list":[1,2,3]}`
	GetAllowedModeListExpectString(t, param, expected)
}

func TestQueryGetAccessorKey2BeforeRevoke(t *testing.T) {
	var param = did.GetAccessorGroupIDParam{
		accessorID2.String(),
	}
	var expected = `{"accessor_public_key":"` + strings.Replace(accessorPubKey2, "\n", "\\n", -1) + `","active":true}`
	GetAccessorKey(t, param, expected)
}

func TestIdP2CreateRequestForRevokeAccessor(t *testing.T) {
	var datas []did.DataRequest
	var param did.Request
	param.RequestID = requestID5.String()
	param.MinIdp = 1
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3
	param.Purpose = "RevokeAccessor"
	param.IdPIDList = append(param.IdPIDList, IdP1)
	CreateRequest(t, param, idpPrivK2, IdP2)
}

func TestIdP1CreateIdpResponseForRevokeAccessor(t *testing.T) {
	var param did.CreateIdpResponseParam
	param.Aal = 3
	param.Ial = 3
	param.RequestID = requestID5.String()
	param.Signature = "signature"
	param.Status = "accept"
	CreateIdpResponse(t, param, idpPrivK, IdP1)
}

func TestIdP2CloseRequestForRevokeAccessor(t *testing.T) {
	var res []did.ResponseValid
	var res1 did.ResponseValid
	res1.IdpID = IdP1
	tValue := true
	res1.ValidIal = &tValue
	res1.ValidSignature = &tValue
	res = append(res, res1)
	var param = did.CloseRequestParam{
		requestID5.String(),
		res,
	}
	CloseRequestByIdP(t, param, idpPrivK2, IdP2)
}

func TestIdP2RevokeAccessor(t *testing.T) {
	var param did.RevokeAccessorParam
	param.RequestID = requestID5.String()
	param.AccessorIDList = append(param.AccessorIDList, accessorID2.String())
	RevokeAccessor(t, param, idpPrivK2, IdP2, "success")
}

func TestQueryGetAccessorKey2AfterRevoke(t *testing.T) {
	var param = did.GetAccessorGroupIDParam{
		accessorID2.String(),
	}
	var expected = `{"accessor_public_key":"` + strings.Replace(accessorPubKey2, "\n", "\\n", -1) + `","active":false}`
	GetAccessorKey(t, param, expected)
}

func TestQueryGetIdentityInfoBeforeUpdateIdentityModeList(t *testing.T) {
	var param did.GetIdentityInfoParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.NodeID = IdP1
	expected := `{"ial":2.3,"mode_list":[2]}`
	GetIdentityInfo(t, param, expected)
}

func TestIdP1UpdateIdentityModeList(t *testing.T) {
	var param did.UpdateIdentityModeListParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.ModeList = append(param.ModeList, 2)
	param.ModeList = append(param.ModeList, 3)
	UpdateIdentityModeList(t, param, idpPrivK, IdP1)
}

func TestQueryGetIdentityInfoAfterUpdateIdentityModeList(t *testing.T) {
	var param did.GetIdentityInfoParam
	param.ReferenceGroupCode = referenceGroupCode1.String()
	param.NodeID = IdP1
	expected := `{"ial":2.3,"mode_list":[2,3]}`
	GetIdentityInfo(t, param, expected)
}

func TestIdP2CreateRequestForAddIdentity(t *testing.T) {
	var datas []did.DataRequest
	var param did.Request
	param.RequestID = requestID6.String()
	param.MinIdp = 1
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3
	param.Purpose = "AddIdentity"
	param.IdPIDList = append(param.IdPIDList, IdP1)
	CreateRequest(t, param, idpPrivK2, IdP2)
}

func TestIdP2CreateIdpResponseForAddIdentity(t *testing.T) {
	var param did.CreateIdpResponseParam
	param.Aal = 3
	param.Ial = 3
	param.RequestID = requestID6.String()
	param.Signature = "signature"
	param.Status = "accept"
	CreateIdpResponse(t, param, idpPrivK, IdP1)
}

func TestIdP2CloseRequestForAddIdentity(t *testing.T) {
	var res []did.ResponseValid
	var res1 did.ResponseValid
	res1.IdpID = IdP1
	tValue := true
	res1.ValidIal = &tValue
	res1.ValidSignature = &tValue
	res = append(res, res1)
	var param = did.CloseRequestParam{
		requestID6.String(),
		res,
	}
	CloseRequestByIdP(t, param, idpPrivK2, IdP2)
}

func TestIdP2AddIdentity(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(userNamespace3 + userID1))
	userHash := h.Sum(nil)
	var user did.AddIdentityParam
	user.ReferenceGroupCode = referenceGroupCode1.String()
	var identity did.Identity
	identity.IdentityNamespace = userNamespace3
	identity.IdentityIdentifierHash = hex.EncodeToString(userHash)
	user.NewIdentityList = append(user.NewIdentityList, identity)
	user.RequestID = requestID6.String()
	var param = user
	AddIdentity(t, param, idpPrivK2, IdP2, "success")
}

func TestQueryGetIdentityInfoByIdentityAfterAddIdentity(t *testing.T) {
	var param did.GetIdentityInfoParam
	h := sha256.New()
	h.Write([]byte(userNamespace3 + userID1))
	userHash := h.Sum(nil)
	param.IdentityNamespace = userNamespace3
	param.IdentityIdentifierHash = hex.EncodeToString(userHash)
	param.NodeID = IdP1
	expected := `{"ial":2.3,"mode_list":[2,3]}`
	GetIdentityInfo(t, param, expected)
}

func TestQueryGetAllowedModeListBeforeSet(t *testing.T) {
	var param did.GetAllowedModeListParam
	param.Purpose = "RegisterIdentity"
	var expected = `{"allowed_mode_list":[2,3]}`
	GetAllowedModeListExpectString(t, param, expected)
}

func TestNDIDSetAllowedModeList(t *testing.T) {
	var param did.SetAllowedModeListParam
	param.Purpose = "RegisterIdentity"
	param.AllowedModeList = append(param.AllowedModeList, 3)
	SetAllowedModeList(t, param)
}

func TestQueryGetAllowedModeListAfterSet(t *testing.T) {
	var param did.GetAllowedModeListParam
	param.Purpose = "RegisterIdentity"
	var expected = `{"allowed_mode_list":[3]}`
	GetAllowedModeListExpectString(t, param, expected)
}

// ---  Old test ---

// func TestInitNDID(t *testing.T) {
// 	InitNDID(t)
// 	IsInitEnded(t, false)
// 	createRequestInInitState(t)
// 	EndInit(t)
// 	IsInitEnded(t, true)
// }

// func createRequestInInitState(t *testing.T) {
// 	var datas []did.DataRequest
// 	var data1 did.DataRequest
// 	data1.ServiceID = serviceID1
// 	data1.Count = 1
// 	data1.RequestParamsHash = "hash"
// 	datas = append(datas, data1)
// 	var param did.Request
// 	param.RequestID = requestID4.String()
// 	param.MinIdp = 1
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	param.Timeout = 259200
// 	param.DataRequestList = datas
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	CreateRequestExpectLog(t, param, rpPrivK, RP1, "Chain is not initialized")
// }

// func TestSetLastBlock1(t *testing.T) {
// 	var param did.SetLastBlockParam
// 	param.BlockHeight = 0
// 	SetLastBlock(t, param)
// 	time.Sleep(2 * time.Second)
// }

// func TestCreateRequestAferSetLastBlock(t *testing.T) {
// 	var datas []did.DataRequest
// 	var data1 did.DataRequest
// 	data1.ServiceID = serviceID1
// 	data1.Count = 1
// 	data1.RequestParamsHash = "hash"
// 	datas = append(datas, data1)
// 	var param did.Request
// 	param.RequestID = requestID4.String()
// 	param.MinIdp = 1
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	param.Timeout = 259200
// 	param.DataRequestList = datas
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	CreateRequestExpectLog(t, param, rpPrivK, RP1, "Chain is disabled")
// }

// func TestSetLastBlock2(t *testing.T) {
// 	var param did.SetLastBlockParam
// 	param.BlockHeight = -1
// 	SetLastBlock(t, param)
// }

// func TestInitData(t *testing.T) {
// 	var param = did.SetPriceFuncParam{
// 		"CreateRequest",
// 		1,
// 	}
// 	SetPriceFunc(t, param)
// }

// func TestRegisterNodeRP(t *testing.T) {
// 	rpKey := getPrivateKeyFromString(rpPrivK)
// 	rpPublicKeyBytes, err := generatePublicKey(&rpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	rpKey2 := getPrivateKeyFromString(allMasterKey)
// 	rpPublicKeyBytes2, err := generatePublicKey(&rpKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	var param did.RegisterNode
// 	param.NodeID = RP1
// 	param.PublicKey = string(rpPublicKeyBytes)
// 	param.MasterPublicKey = string(rpPublicKeyBytes2)
// 	param.Role = "RP"
// 	param.NodeName = "Node RP 1"

// 	RegisterNode(t, param)
// }

// func TestRegisterNodeIDP(t *testing.T) {
// 	idpKey := getPrivateKeyFromString(idpPrivK)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	idpKey2 := getPrivateKeyFromString(allMasterKey)
// 	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	var param did.RegisterNode
// 	param.NodeID = IdP1
// 	param.PublicKey = string(idpPublicKeyBytes)
// 	param.MasterPublicKey = string(idpPublicKeyBytes2)
// 	param.NodeName = "IdP Number 1 from ..."
// 	param.Role = "IdP"
// 	param.MaxIal = 3.0
// 	param.MaxAal = 3.0

// 	RegisterNode(t, param)
// }

// func TestRegisterNodeIDP10(t *testing.T) {
// 	idpKey := getPrivateKeyFromString(idpPrivK)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	idpKey2 := getPrivateKeyFromString(allMasterKey)
// 	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	var param did.RegisterNode
// 	param.NodeID = IdP10
// 	param.PublicKey = string(idpPublicKeyBytes)
// 	param.MasterPublicKey = string(idpPublicKeyBytes2)
// 	param.NodeName = "IdP Number 1 from ..."
// 	param.Role = "IdP"
// 	param.MaxIal = 3.0
// 	param.MaxAal = 3.0

// 	RegisterNode(t, param)
// }

// func TestRegisterNodeAS(t *testing.T) {
// 	asKey := getPrivateKeyFromString(asPrivK)
// 	asPublicKeyBytes, err := generatePublicKey(&asKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	asKey2 := getPrivateKeyFromString(allMasterKey)
// 	asPublicKeyBytes2, err := generatePublicKey(&asKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	var param did.RegisterNode
// 	param.NodeName = "AS1"
// 	param.NodeID = AS1
// 	param.PublicKey = string(asPublicKeyBytes)
// 	param.MasterPublicKey = string(asPublicKeyBytes2)
// 	param.Role = "AS"

// 	RegisterNode(t, param)
// }

// func TestNDIDSetTimeOutBlockRegisterIdentity(t *testing.T) {
// 	SetTimeOutBlockRegisterIdentity(t)
// }

// func TestQueryGetNodePublicKeyRP(t *testing.T) {
// 	var param = did.GetNodePublicKeyParam{
// 		RP1,
// 	}
// 	rpKey := getPrivateKeyFromString(rpPrivK)
// 	rpPublicKeyBytes, err := generatePublicKey(&rpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	GetNodePublicKey(t, param, string(rpPublicKeyBytes))
// }

// func TestQueryGetNodeMasterPublicKeyRP(t *testing.T) {
// 	var param = did.GetNodePublicKeyParam{
// 		RP1,
// 	}
// 	rpKey := getPrivateKeyFromString(allMasterKey)
// 	rpPublicKeyBytes, err := generatePublicKey(&rpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	GetNodeMasterPublicKey(t, param, string(rpPublicKeyBytes))
// }

// func TestQueryGetNodePublicKeyIdP(t *testing.T) {
// 	var param = did.GetNodePublicKeyParam{
// 		IdP1,
// 	}
// 	idpKey := getPrivateKeyFromString(idpPrivK)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	GetNodePublicKey(t, param, string(idpPublicKeyBytes))
// }

// func TestQueryGetNodePublicKeyAS(t *testing.T) {
// 	var param = did.GetNodePublicKeyParam{
// 		AS1,
// 	}
// 	asKey := getPrivateKeyFromString(asPrivK)
// 	asPublicKeyBytes, err := generatePublicKey(&asKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	GetNodePublicKey(t, param, string(asPublicKeyBytes))
// }

// func TestAddNodeTokenRP(t *testing.T) {
// 	var param = did.AddNodeTokenParam{
// 		RP1,
// 		111.11,
// 	}
// 	AddNodeToken(t, param)
// }

// func TestAddNodeTokenIdP(t *testing.T) {
// 	var param = did.AddNodeTokenParam{
// 		IdP1,
// 		222.22,
// 	}
// 	AddNodeToken(t, param)
// }

// func TestAddNodeTokenIdP10(t *testing.T) {
// 	var param = did.AddNodeTokenParam{
// 		IdP10,
// 		222.22,
// 	}
// 	AddNodeToken(t, param)
// }

// func TestAddNodeTokenAS(t *testing.T) {
// 	var param = did.AddNodeTokenParam{
// 		AS1,
// 		333.33,
// 	}
// 	AddNodeToken(t, param)
// }

// func TestQueryGetNodeTokenRP(t *testing.T) {
// 	var param = did.GetNodeTokenParam{
// 		RP1,
// 	}
// 	var expected = did.GetNodeTokenResult{
// 		111.11,
// 	}
// 	GetNodeToken(t, param, expected)
// }

// func TestReduceNodeTokenRP(t *testing.T) {
// 	var param = did.ReduceNodeTokenParam{
// 		RP1,
// 		61.11,
// 	}
// 	ReduceNodeToken(t, param)
// }

// func TestQueryGetNodeTokenRPAfterReduce(t *testing.T) {
// 	var param = did.GetNodeTokenParam{
// 		RP1,
// 	}
// 	var expected = did.GetNodeTokenResult{
// 		50.0,
// 	}
// 	GetNodeToken(t, param, expected)
// }

// func TestSetNodeTokenRP(t *testing.T) {
// 	var param = did.SetNodeTokenParam{
// 		RP1,
// 		100.0,
// 	}
// 	SetNodeToken(t, param)
// }

// func TestQueryGetNodeTokenRPAfterSetToken(t *testing.T) {
// 	var param = did.GetNodeTokenParam{
// 		RP1,
// 	}
// 	var expected = did.GetNodeTokenResult{
// 		100.0,
// 	}
// 	GetNodeToken(t, param, expected)
// }

// func TestNDIDAddService(t *testing.T) {
// 	var param did.AddServiceParam
// 	param.ServiceID = serviceID1
// 	param.ServiceName = "Bank statement"
// 	param.DataSchema = "DataSchema"
// 	param.DataSchemaVersion = "DataSchemaVersion"
// 	AddService(t, param)
// }

// func TestNDIDAddServiceAgain(t *testing.T) {
// 	var param did.AddServiceParam
// 	param.ServiceID = serviceID2
// 	param.ServiceName = "Bank statement"
// 	param.DataSchema = "DataSchema"
// 	param.DataSchemaVersion = "DataSchemaVersion"
// 	AddService(t, param)
// }

// func TestNDIDDisableService(t *testing.T) {
// 	var param = did.DisableServiceParam{
// 		serviceID2,
// 	}
// 	DisableService(t, param)
// }

// func TestIdPRegisterIdentity(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var users []did.User
// 	var user = did.User{
// 		hex.EncodeToString(userHash),
// 		3,
// 		true,
// 	}
// 	users = append(users, user)
// 	var param = did.RegisterIdentityParam{
// 		users,
// 	}
// 	RegisterIdentity(t, param, idpPrivK, IdP1, "success")
// }

// func TestDisableOldIdPNode1(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var param did.GetIdpNodesParam
// 	param.HashID = hex.EncodeToString(userHash)
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	idps := GetIdpNodesForDisable(t, param)
// 	for _, idp := range idps {
// 		if idp.ID != IdP1 {
// 			var param did.DisableNodeParam
// 			param.NodeID = idp.ID
// 			DisableNode(t, param)
// 		}
// 	}

// }

// func TestQueryGetMqAddressesBeforeRegister(t *testing.T) {
// 	var param = did.GetMqAddressesParam{
// 		IdP1,
// 	}
// 	var expected []did.MsqAddress
// 	GetMqAddresses(t, param, expected)
// }

// func TestIdPSetMqAddresses(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.99"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, idpPrivK, IdP1)
// }

// func TestQueryGetIdpNodesInfo1(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var param did.GetIdpNodesParam
// 	param.HashID = hex.EncodeToString(userHash)
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	var expected = `{"node":[{"node_id":"` + IdP1 + `","name":"IdP Number 1 from ...","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}]}]}`
// 	GetIdpNodesInfo(t, param, expected)
// }

// func TestQueryGetIdpNodes(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var param did.GetIdpNodesParam
// 	param.HashID = hex.EncodeToString(userHash)
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	var expected = []did.MsqDestinationNode{
// 		{
// 			IdP1,
// 			"IdP Number 1 from ...",
// 			3.0,
// 			3.0,
// 		},
// 	}
// 	GetIdpNodes(t, param, expected)
// }

// func TestQueryGetMqAddresses(t *testing.T) {
// 	var param = did.GetMqAddressesParam{
// 		IdP1,
// 	}
// 	var expected []did.MsqAddress
// 	var msq did.MsqAddress
// 	msq.IP = "192.168.3.99"
// 	msq.Port = 8000
// 	expected = append(expected, msq)
// 	GetMqAddresses(t, param, expected)
// }

// func TestRegisterServiceDestinationByNDIDForInvalidNodeID(t *testing.T) {
// 	var param = did.RegisterServiceDestinationByNDIDParam{
// 		serviceID1,
// 		"Invalid-node-ID",
// 	}
// 	RegisterServiceDestinationByNDIDExpectedString(t, param, "Node ID not found")
// }

// func TestRegisterServiceDestinationByNDIDForInvalidRole(t *testing.T) {
// 	var param = did.RegisterServiceDestinationByNDIDParam{
// 		serviceID1,
// 		IdP1,
// 	}
// 	RegisterServiceDestinationByNDIDExpectedString(t, param, "Role of node ID is not AS")
// }

// func TestASRegisterServiceDestinationByNDIDForAS1(t *testing.T) {
// 	var param = did.RegisterServiceDestinationByNDIDParam{
// 		serviceID1,
// 		AS1,
// 	}
// 	RegisterServiceDestinationByNDID(t, param)
// }

// func TestASRegisterServiceDestination(t *testing.T) {
// 	var param = did.RegisterServiceDestinationParam{
// 		serviceID1,
// 		1.1,
// 		1.2,
// 	}
// 	RegisterServiceDestination(t, param, asPrivK, AS1, "success")
// }

// func TestASRegisterServiceDestination2(t *testing.T) {
// 	var param = did.RegisterServiceDestinationParam{
// 		serviceID1,
// 		1.1,
// 		1.2,
// 	}
// 	RegisterServiceDestination(t, param, asPrivK, AS1, "Duplicate service ID in provide service list")
// }

// func TestQueryGetServiceDetail1(t *testing.T) {
// 	var param = did.GetServiceDetailParam{
// 		serviceID1,
// 	}
// 	var expected = did.ServiceDetail{
// 		serviceID1,
// 		"Bank statement",
// 		"DataSchema",
// 		"DataSchemaVersion",
// 		true,
// 	}
// 	GetServiceDetail(t, param, expected)
// }

// func TestNDIDUpdateService(t *testing.T) {
// 	var param did.UpdateServiceParam
// 	param.ServiceID = serviceID1
// 	param.ServiceName = "Bank statement (ย้อนหลัง 3 เดือน)"
// 	param.DataSchemaVersion = "DataSchemaVersion2"
// 	UpdateService(t, param)
// }

// func TestQueryGetServiceDetail2(t *testing.T) {
// 	var param = did.GetServiceDetailParam{
// 		serviceID1,
// 	}
// 	var expected = did.ServiceDetail{
// 		serviceID1,
// 		"Bank statement (ย้อนหลัง 3 เดือน)",
// 		"DataSchema",
// 		"DataSchemaVersion2",
// 		true,
// 	}
// 	GetServiceDetail(t, param, expected)
// }
// func TestASUpdateServiceDestination(t *testing.T) {
// 	var param = did.UpdateServiceDestinationParam{
// 		serviceID1,
// 		1.4,
// 		1.5,
// 	}
// 	UpdateServiceDestination(t, param, AS1)
// }

// func TestQueryGetAsNodesByServiceId(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID1
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":1.4,"min_aal":1.5}]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestAS1SetMqAddresses(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.102"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, asPrivK, AS1)
// }

// func TestQueryGetAsNodesInfoByServiceId(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID1
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","name":"AS1","min_ial":1.4,"min_aal":1.5,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.102","port":8000}]}]}`
// 	GetAsNodesInfoByServiceId(t, param, expected)
// }

// func TestRPCreateRequest(t *testing.T) {
// 	var datas []did.DataRequest
// 	var data1 did.DataRequest
// 	data1.ServiceID = serviceID1
// 	data1.Count = 1
// 	data1.RequestParamsHash = "hash"
// 	data1.As = append(data1.As, AS1)
// 	datas = append(datas, data1)
// 	var param did.Request
// 	param.RequestID = requestID1.String()
// 	param.MinIdp = 1
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	param.Timeout = 259200
// 	param.IdPIDList = append(param.IdPIDList, IdP1)
// 	// param.IdPIDList = append(param.IdPIDList, IdP2)
// 	param.DataRequestList = datas
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	CreateRequest(t, param, rpPrivK, RP1)
// }

// func TestCreateRequestWithDuplicateNonce(t *testing.T) {
// 	var datas []did.DataRequest
// 	var data1 did.DataRequest
// 	data1.ServiceID = serviceID1
// 	data1.Count = 1
// 	data1.RequestParamsHash = "hash"
// 	datas = append(datas, data1)
// 	var param did.Request
// 	param.RequestID = requestID6.String()
// 	param.MinIdp = 1
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	param.Timeout = 259200
// 	param.DataRequestList = datas
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	var param2 did.Request
// 	param2.RequestID = requestID6.String()
// 	param2.MinIdp = 1
// 	param2.MinIal = 3
// 	param2.MinAal = 3
// 	param2.Timeout = 259200
// 	param2.DataRequestList = datas
// 	param2.MessageHash = "hashhashhashhashhashhash"
// 	param2.Mode = 3
// 	CreateRequestWithDuplicateNonce(t, param, param2, rpPrivK, RP1, "success")
// }

// func TestQueryGetNodeTokenRPAfterCreatRequest(t *testing.T) {
// 	var param = did.GetNodeTokenParam{
// 		RP1,
// 	}
// 	var expected = did.GetNodeTokenResult{
// 		98.0,
// 	}
// 	GetNodeToken(t, param, expected)
// }

// func TestIdPCreateIdpResponse(t *testing.T) {
// 	var param = did.CreateIdpResponseParam{
// 		requestID1.String(),
// 		3,
// 		3,
// 		"accept",
// 		"signature",
// 		"Magic",
// 		"Magic",
// 	}
// 	CreateIdpResponse(t, param, idpPrivK, IdP1)
// }

// func TestASSignData(t *testing.T) {
// 	var param = did.SignDataParam{
// 		serviceID1,
// 		requestID1.String(),
// 		"sign(data,asKey)",
// 	}
// 	SignData(t, param, "success", AS1)
// }

// func TestASSignData2(t *testing.T) {
// 	var param = did.SignDataParam{
// 		serviceID1,
// 		requestID1.String(),
// 		"sign(data,asKey)",
// 	}
// 	SignData(t, param, "Duplicate AS ID in answered AS list", AS1)
// }

// func TestRPSetDataReceived(t *testing.T) {
// 	var param = did.SetDataReceivedParam{
// 		requestID1.String(),
// 		serviceID1,
// 		AS1,
// 	}
// 	SetDataReceived(t, param, "success", RP1)
// }

// func TestRPSetDataReceived2(t *testing.T) {
// 	var param = did.SetDataReceivedParam{
// 		requestID1.String(),
// 		serviceID1,
// 		AS1,
// 	}
// 	SetDataReceived(t, param, "Duplicate AS ID in data request", RP1)
// }

// func TestIdPCreateRequestSpecial(t *testing.T) {
// 	var datas []did.DataRequest
// 	var param did.Request
// 	param.RequestID = requestID2.String()
// 	param.MinIdp = 1
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	param.Timeout = 259200
// 	param.DataRequestList = datas
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	param.Purpose = "AddAccessor"
// 	param.IdPIDList = append(param.IdPIDList, IdP1)
// 	CreateRequest(t, param, idpPrivK, IdP10)
// }

// func TestNDIDSetPrice(t *testing.T) {
// 	var param = did.SetPriceFuncParam{
// 		"CreateRequest",
// 		9.99,
// 	}
// 	SetPriceFunc(t, param)
// }

// func TestNDIDGetPrice(t *testing.T) {
// 	var param = did.GetPriceFuncParam{
// 		"CreateRequest",
// 	}
// 	var expected = did.GetPriceFuncResult{
// 		9.99,
// 	}
// 	GetPriceFunc(t, param, expected)
// }

// // func TestReportGetUsedTokenRP(t *testing.T) {
// // 	expectedString := `[{"method":"CreateRequest","price":1,"data":"` + requestID1.String() + `"},{"method":"SetDataReceived","price":1,"data":"` + requestID1.String() + `"}]`
// // 	var param = did.GetUsedTokenReportParam{
// // 		RP1,
// // 	}
// // 	GetUsedTokenReport(t, param, expectedString)
// // }

// // func TestReportGetUsedTokenAS(t *testing.T) {
// // 	var param = did.GetUsedTokenReportParam{
// // 		AS1,
// // 	}
// // 	expectedString := `[{"method":"RegisterServiceDestination","price":1,"data":""},{"method":"UpdateServiceDestination","price":1,"data":""},{"method":"SetMqAddresses","price":1,"data":""},{"method":"SignData","price":1,"data":"` + requestID1.String() + `"}]`
// // 	GetUsedTokenReport(t, param, expectedString)
// // }

// func TestQueryGetRequestClosed(t *testing.T) {
// 	var param = did.GetRequestParam{
// 		requestID1.String(),
// 	}
// 	var expected = did.GetRequestResult{
// 		true,
// 		false,
// 		"hash('Please allow...')",
// 		3,
// 	}
// 	GetRequest(t, param, expected)
// }

// func TestCreateRequest(t *testing.T) {
// 	var datas []did.DataRequest
// 	var data1 did.DataRequest
// 	data1.ServiceID = serviceID1
// 	data1.As = []string{
// 		AS1,
// 	}
// 	data1.Count = 2
// 	data1.RequestParamsHash = "hash"

// 	var data2 did.DataRequest
// 	data2.ServiceID = "credit"
// 	data2.As = []string{
// 		AS1,
// 	}
// 	data2.Count = 2
// 	data2.RequestParamsHash = "hash"
// 	datas = append(datas, data1)
// 	datas = append(datas, data2)
// 	var param did.Request
// 	param.RequestID = requestID3.String()
// 	param.MinIdp = 1
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	param.Timeout = 259200
// 	param.IdPIDList = append(param.IdPIDList, IdP1)
// 	// param.IdPIDList = append(param.IdPIDList, IdP2)
// 	param.DataRequestList = datas
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	CreateRequest(t, param, rpPrivK, RP1)
// }

// func TestIdPCreateIdpResponse2(t *testing.T) {
// 	var param = did.CreateIdpResponseParam{
// 		requestID3.String(),
// 		3,
// 		3,
// 		"accept",
// 		"signature",
// 		"Magic",
// 		"Magic",
// 	}
// 	CreateIdpResponse(t, param, idpPrivK, IdP1)
// }

// func TestQueryGetRequestTimedOut(t *testing.T) {
// 	var param = did.GetRequestParam{
// 		requestID3.String(),
// 	}
// 	var expected = did.GetRequestResult{
// 		false,
// 		true,
// 		"hash('Please allow...')",
// 		3,
// 	}
// 	GetRequest(t, param, expected)
// }

// func TestDisableOldNamespace(t *testing.T) {
// 	namespaces := GetNamespaceListForDisable(t)
// 	for _, namespace := range namespaces {
// 		var param did.DisableNamespaceParam
// 		param.Namespace = namespace.Namespace
// 		DisableNamespace(t, param)
// 	}

// }

// func TestAddNamespaceCID(t *testing.T) {
// 	var param did.Namespace
// 	param.Namespace = namespaceID1
// 	param.Description = "Citizen ID"
// 	AddNamespace(t, param)
// }

// func TestAddNamespaceTel(t *testing.T) {
// 	var param did.Namespace
// 	param.Namespace = namespaceID2
// 	param.Description = "Tel number"
// 	AddNamespace(t, param)
// }

// func TestDisableNamespace(t *testing.T) {
// 	var param did.DisableNamespaceParam
// 	param.Namespace = namespaceID2
// 	DisableNamespace(t, param)
// }

// func TestQueryGetNamespaceList(t *testing.T) {
// 	var expected = []did.Namespace{
// 		did.Namespace{
// 			namespaceID1,
// 			"Citizen ID",
// 			true,
// 		},
// 	}
// 	GetNamespaceList(t, expected)
// }

// func TestIdPRegisterAccessor(t *testing.T) {
// 	var param = did.RegisterAccessorParam{
// 		accessorID1.String(),
// 		"accessor_type",
// 		accessorPubKey,
// 		accessorGroupID1.String(),
// 	}
// 	RegisterAccessor(t, param, IdP1)
// }

// func TestQueryGetAccessorsInAccessorGroupInvalidIdP(t *testing.T) {
// 	var param did.GetAccessorsInAccessorGroupParam
// 	param.AccessorGroupID = accessorGroupID1.String()
// 	param.IdpID = IdP2
// 	expected := string(`{"accessor_list":[]}`)
// 	GetAccessorsInAccessorGroup(t, param, expected)
// }

// func TestQueryGetAccessorsInAccessorGroup(t *testing.T) {
// 	var param did.GetAccessorsInAccessorGroupParam
// 	param.AccessorGroupID = accessorGroupID1.String()
// 	param.IdpID = IdP1
// 	expected := string(`{"accessor_list":["` + accessorID1.String() + `"]}`)
// 	GetAccessorsInAccessorGroup(t, param, expected)
// }

// func TestIdPAddAccessorMethod(t *testing.T) {
// 	var param = did.AccessorMethod{
// 		accessorID2.String(),
// 		"accessor_type_2",
// 		accessorPubKey2,
// 		accessorGroupID1.String(),
// 		requestID2.String(),
// 	}
// 	AddAccessorMethod(t, param, IdP10, true)
// }

// func TestIdPAddAccessorMethod2(t *testing.T) {
// 	var param = did.AccessorMethod{
// 		accessorID3.String(),
// 		"accessor_type_2",
// 		accessorPubKey2,
// 		accessorGroupID1.String(),
// 		requestID2.String(),
// 	}
// 	AddAccessorMethod(t, param, IdP10, false)
// }

// func TestQueryGetAccessorsInAccessorGroup_IdP1(t *testing.T) {
// 	var param did.GetAccessorsInAccessorGroupParam
// 	param.AccessorGroupID = accessorGroupID1.String()
// 	param.IdpID = IdP1
// 	expected := string(`{"accessor_list":["` + accessorID1.String() + `"]}`)
// 	GetAccessorsInAccessorGroup(t, param, expected)
// }

// func TestQueryGetAccessorsInAccessorGroup_IdP10(t *testing.T) {
// 	var param did.GetAccessorsInAccessorGroupParam
// 	param.AccessorGroupID = accessorGroupID1.String()
// 	param.IdpID = IdP10
// 	expected := string(`{"accessor_list":["` + accessorID2.String() + `"]}`)
// 	GetAccessorsInAccessorGroup(t, param, expected)
// }

// func TestQueryGetAccessorsInAccessorGroup_WithOut_IdP_ID(t *testing.T) {
// 	var param did.GetAccessorsInAccessorGroupParam
// 	param.AccessorGroupID = accessorGroupID1.String()
// 	expected := string(`{"accessor_list":["` + accessorID1.String() + `","` + accessorID2.String() + `"]}`)
// 	GetAccessorsInAccessorGroup(t, param, expected)
// }

// func TestIdP10CreateRequestSpecial(t *testing.T) {
// 	var datas []did.DataRequest
// 	var param did.Request
// 	param.RequestID = requestID5.String()
// 	param.MinIdp = 1
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	param.Timeout = 259200
// 	param.DataRequestList = datas
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	param.Purpose = "RevokeAccessor"
// 	param.IdPIDList = append(param.IdPIDList, IdP10)
// 	CreateRequest(t, param, idpPrivK, IdP10)
// }

// func TestIdP10CreateIdpResponse(t *testing.T) {
// 	var param = did.CreateIdpResponseParam{
// 		requestID5.String(),
// 		3,
// 		3,
// 		"accept",
// 		"signature",
// 		"Magic",
// 		"Magic",
// 	}
// 	CreateIdpResponse(t, param, idpPrivK, IdP10)
// }

// func TestIdP10RevokeAccessorMethod(t *testing.T) {
// 	var param did.RevokeAccessorMethodParam
// 	param.RequestID = requestID5.String()
// 	param.AccessorIDList = append(param.AccessorIDList, accessorID2.String())
// 	RevokeAccessorMethod(t, param, IdP10, "success")
// }

// func TestQueryGetAccessorsInAccessorGroup_WithOut_IdP_ID_After_Removed(t *testing.T) {
// 	var param did.GetAccessorsInAccessorGroupParam
// 	param.AccessorGroupID = accessorGroupID1.String()
// 	expected := string(`{"accessor_list":["` + accessorID1.String() + `"]}`)
// 	GetAccessorsInAccessorGroup(t, param, expected)
// }

// func TestQueryGetAccessorKeyAfterRevoke(t *testing.T) {
// 	var param = did.GetAccessorGroupIDParam{
// 		accessorID2.String(),
// 	}
// 	var expected = `{"accessor_public_key":"` + strings.Replace(accessorPubKey2, "\n", "\\n", -1) + `","active":false}`
// 	GetAccessorKey(t, param, expected)
// }

// func TestIdP1ClearRegisterIdentityTimeout(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)

// 	var param = did.ClearRegisterIdentityTimeoutParam{
// 		hex.EncodeToString(userHash),
// 	}
// 	ClearRegisterIdentityTimeout(t, param, idpPrivK, IdP1)
// }

// func TestQueryCheckExistingIdentity(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var param = did.CheckExistingIdentityParam{
// 		hex.EncodeToString(userHash),
// 	}
// 	var expected = `{"exist":true}`
// 	CheckExistingIdentity(t, param, expected)
// }

// func TestQueryGetAccessorGroupID(t *testing.T) {
// 	var param = did.GetAccessorGroupIDParam{
// 		accessorID2.String(),
// 	}
// 	var expected = `{"accessor_group_id":"` + accessorGroupID1.String() + `"}`
// 	GetAccessorGroupID(t, param, expected)
// }

// func TestQueryGetAccessorOwner(t *testing.T) {
// 	var param = did.GetAccessorOwnerParam{
// 		accessorID2.String(),
// 	}
// 	var expected = `{"node_id":"` + IdP10 + `"}`
// 	GetAccessorOwner(t, param, expected)
// }

// func TestQueryGetAccessorKey(t *testing.T) {
// 	var param = did.GetAccessorGroupIDParam{
// 		accessorID1.String(),
// 	}
// 	var expected = `{"accessor_public_key":"` + strings.Replace(accessorPubKey, "\n", "\\n", -1) + `","active":true}`
// 	GetAccessorKey(t, param, expected)
// }

// func TestDisableOldIdPNode2(t *testing.T) {
// 	var param did.GetIdpNodesParam
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	idps := GetIdpNodesForDisable(t, param)
// 	for _, idp := range idps {
// 		if idp.ID != IdP1 {
// 			var param did.DisableNodeParam
// 			param.NodeID = idp.ID
// 			DisableNode(t, param)
// 		}
// 	}
// }

// func TestRegisterNodeIDP2(t *testing.T) {
// 	idpKey := getPrivateKeyFromString(idpPrivK3)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param did.RegisterNode
// 	param.NodeID = IdP2
// 	param.PublicKey = string(idpPublicKeyBytes)
// 	param.Role = "IdP"
// 	param.MaxIal = 3.0
// 	param.MaxAal = 3.0
// 	RegisterNode(t, param)
// }

// func TestQueryGetIdpNodes2(t *testing.T) {
// 	var param did.GetIdpNodesParam
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	var expected = []did.MsqDestinationNode{
// 		{
// 			IdP1,
// 			"IdP Number 1 from ...",
// 			3.0,
// 			3.0,
// 		},
// 		{
// 			IdP2,
// 			"",
// 			3.0,
// 			3.0,
// 		},
// 	}
// 	GetIdpNodes(t, param, expected)
// }

// func TestIdPUpdateNode(t *testing.T) {
// 	idpKey2 := getPrivateKeyFromString(idpPrivK2)
// 	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param = did.UpdateNodeParam{
// 		string(idpPublicKeyBytes2),
// 		"",
// 	}
// 	UpdateNode(t, param, allMasterKey, IdP1)
// }

// func TestSetValidator(t *testing.T) {
// 	var param did.SetValidatorParam
// 	param.PublicKey = getValidatorPubkey(1)
// 	param.Power = 100
// 	SetValidator(t, param)
// }

// func TestSetValidator2(t *testing.T) {
// 	var param did.SetValidatorParam
// 	param.PublicKey = getValidatorPubkey(2)
// 	param.Power = 1
// 	SetValidator(t, param)
// }

// func TestSetValidator3(t *testing.T) {
// 	var param did.SetValidatorParam
// 	param.PublicKey = getValidatorPubkey(3)
// 	param.Power = 1
// 	SetValidator(t, param)
// }

// func TestDisableOldService(t *testing.T) {
// 	services := GetServiceListForDisable(t)
// 	for _, service := range services {
// 		if service.ServiceID != serviceID1 {
// 			var param = did.DisableServiceParam{
// 				service.ServiceID,
// 			}
// 			DisableService(t, param)
// 		}
// 	}
// }

// func TestQueryGetServiceList(t *testing.T) {
// 	var expected = `[{"service_id":"` + serviceID1 + `","service_name":"Bank statement (ย้อนหลัง 3 เดือน)","active":true}]`
// 	GetServiceList(t, expected)
// }

// func TestUpdateNodeByNDID(t *testing.T) {
// 	var param did.UpdateNodeByNDIDParam
// 	param.NodeID = IdP1
// 	param.MaxIal = 2.3
// 	param.MaxAal = 2.4
// 	UpdateNodeByNDID(t, param)
// }

// func TestUpdateNodeRPByNDID(t *testing.T) {
// 	var param did.UpdateNodeByNDIDParam
// 	param.NodeID = RP1
// 	param.NodeName = "Node RP 1 edited"
// 	UpdateNodeByNDID(t, param)
// }

// func TestQueryGetNodeInfo(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = IdP1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 1 from ...","role":"IdP","max_ial":2.3,"max_aal":2.4,"mq":[{"ip":"192.168.3.99","port":8000}]}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestQueryCheckExistingAccessorID(t *testing.T) {
// 	var param did.CheckExistingAccessorIDParam
// 	param.AccessorID = accessorID1.String()
// 	expected := `{"exist":true}`
// 	CheckExistingAccessorID(t, param, expected)
// }

// func TestQueryCheckExistingAccessorGroupID(t *testing.T) {
// 	var param did.CheckExistingAccessorGroupIDParam
// 	param.AccessorGroupID = accessorGroupID1.String()
// 	expected := `{"exist":true}`
// 	CheckExistingAccessorGroupID(t, param, expected)
// }

// func TestIdPUpdateIdentity(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var param did.UpdateIdentityParam
// 	param.HashID = hex.EncodeToString(userHash)
// 	param.Ial = 2.2
// 	UpdateIdentity(t, param, IdP1)
// }

// func TestQueryGetIdentityInfo(t *testing.T) {
// 	var param did.GetIdentityInfoParam
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	param.NodeID = IdP1
// 	param.HashID = hex.EncodeToString(userHash)
// 	expected := `{"ial":2.2}`
// 	GetIdentityInfo(t, param, expected)
// }

// func TestQueryGetDataSignature(t *testing.T) {
// 	var param did.GetDataSignatureParam
// 	param.NodeID = AS1
// 	param.RequestID = requestID1.String()
// 	param.ServiceID = serviceID1
// 	expected := `{"signature":"sign(data,asKey)"}`
// 	GetDataSignature(t, param, expected)
// }

// func TestDisableOldIdPNode3(t *testing.T) {
// 	var param did.GetIdpNodesParam
// 	param.HashID = ""
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	idps := GetIdpNodesForDisable(t, param)
// 	for _, idp := range idps {
// 		if idp.ID != IdP1 && idp.ID != IdP4 {
// 			var param did.DisableNodeParam
// 			param.NodeID = idp.ID
// 			DisableNode(t, param)
// 		}
// 	}
// }

// func TestRegisterNodeIDP4(t *testing.T) {
// 	idpKey := getPrivateKeyFromString(idpPrivK4)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	idpKey2 := getPrivateKeyFromString(allMasterKey)
// 	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param did.RegisterNode
// 	param.NodeID = IdP4
// 	param.PublicKey = string(idpPublicKeyBytes)
// 	param.MasterPublicKey = string(idpPublicKeyBytes2)
// 	param.NodeName = "IdP Number 4 from ..."
// 	param.Role = "IdP"
// 	param.MaxIal = 3.0
// 	param.MaxAal = 3.0
// 	RegisterNode(t, param)
// }

// func TestRegisterNodeIDP5(t *testing.T) {
// 	idpKey := getPrivateKeyFromString(idpPrivK5)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	idpKey2 := getPrivateKeyFromString(allMasterKey)
// 	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param did.RegisterNode
// 	param.NodeID = IdP5
// 	param.PublicKey = string(idpPublicKeyBytes)
// 	param.MasterPublicKey = string(idpPublicKeyBytes2)
// 	param.NodeName = "IdP Number 5 from ..."
// 	param.Role = "IdP"
// 	param.MaxIal = 3.0
// 	param.MaxAal = 3.0
// 	RegisterNode(t, param)
// }

// func TestSetNodeTokenIDP4(t *testing.T) {
// 	var param = did.SetNodeTokenParam{
// 		IdP4,
// 		100.0,
// 	}
// 	SetNodeToken(t, param)
// }

// func TestSetNodeTokenIDP5(t *testing.T) {
// 	var param = did.SetNodeTokenParam{
// 		IdP5,
// 		100.0,
// 	}
// 	SetNodeToken(t, param)
// }

// func TestIdPUpdateNode4(t *testing.T) {
// 	idpKey2 := getPrivateKeyFromString(idpPrivK5)
// 	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param = did.UpdateNodeParam{
// 		string(idpPublicKeyBytes2),
// 		"",
// 	}
// 	UpdateNode(t, param, allMasterKey, IdP4)
// }

// func TestIdPUpdateNode5(t *testing.T) {
// 	idpKey2 := getPrivateKeyFromString(idpPrivK4)
// 	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param = did.UpdateNodeParam{
// 		string(idpPublicKeyBytes2),
// 		string(idpPublicKeyBytes2),
// 	}
// 	UpdateNode(t, param, allMasterKey, IdP5)
// }

// func TestIdP4SetMqAddresses(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.99"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, idpPrivK5, IdP4)
// }

// func TestIdP5SetMqAddresses(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.99"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, idpPrivK4, IdP5)
// }

// func TestQueryGetNodeInfoIdP4(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = IdP4
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu9+CK/vznpXtAUC0QhuJ\ngYKCfMMBiIgVcp2A+e+SsKvv6ESQ72R8K6nQAhH2MGtnj3ScLI0tMwCtgotWCEGi\nyUXKXLVTiqAqtwflCUVuxCDVuvOm3GQCxvwzE34jEgbGZ33G3tV7uKTtifhoJzVY\nD+WkZVslBhaBgQCUewCX4zkCCTYC5VEhkr7K8HGEr6n1eBOO5VORCkrHKYoZK7eu\nNjyWvWYyVN07F8K0RhgIF9Xsa6Tiu1Yf8zuyJ/awR6U4Nw+oTkvRpx64+caBNYgR\n4n8peg9ZJeTAwV49o1ymx34pPjHUgSdpyhZX4i3z9ji+o7KbNkA/O0l+3doMuH1e\nxwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 4 from ...","role":"IdP","max_ial":3,"max_aal":3,"mq":[{"ip":"192.168.3.99","port":8000}]}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestQueryGetNodeInfoIdP5(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = IdP5
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApbxaA5aKnkpnV7+dMW5x\n7iEINouvjhQ8gl6+8A6ApiVbYIzJCCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaH\nOk1FfX6tk5i5uGWifRQHueXhXjR8HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7X\nHdh9Mqz+XAkq4T6vtdaocduarNG6ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3T\nHkM9noP2AYEH2MBHCzuu+bifCwurOBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jd\nAmbbqMyGpAsF+WjW4V2rcTMFZOoYFYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP\n2wIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApbxaA5aKnkpnV7+dMW5x\n7iEINouvjhQ8gl6+8A6ApiVbYIzJCCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaH\nOk1FfX6tk5i5uGWifRQHueXhXjR8HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7X\nHdh9Mqz+XAkq4T6vtdaocduarNG6ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3T\nHkM9noP2AYEH2MBHCzuu+bifCwurOBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jd\nAmbbqMyGpAsF+WjW4V2rcTMFZOoYFYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP\n2wIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 5 from ...","role":"IdP","max_ial":3,"max_aal":3,"mq":[{"ip":"192.168.3.99","port":8000}]}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestIdP4RegisterIdentity0(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID2))
// 	userHash := h.Sum(nil)
// 	var users []did.User
// 	var user = did.User{
// 		hex.EncodeToString(userHash),
// 		3,
// 		true,
// 	}
// 	users = append(users, user)
// 	var param = did.RegisterIdentityParam{
// 		users,
// 	}
// 	RegisterIdentity(t, param, idpPrivK5, IdP4, "success")
// }

// func TestIdP4RegisterIdentity11(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID2))
// 	userHash := h.Sum(nil)
// 	var users []did.User
// 	var user = did.User{
// 		hex.EncodeToString(userHash),
// 		3,
// 		true,
// 	}
// 	users = append(users, user)
// 	var param = did.RegisterIdentityParam{
// 		users,
// 	}
// 	RegisterIdentity(t, param, idpPrivK5, IdP4, "This node is not first IdP")
// }

// func TestIdP4ClearRegisterIdentityTimeout(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID2))
// 	userHash := h.Sum(nil)
// 	var param = did.ClearRegisterIdentityTimeoutParam{
// 		hex.EncodeToString(userHash),
// 	}
// 	ClearRegisterIdentityTimeout(t, param, idpPrivK5, IdP4)
// }

// func TestIdP4RegisterIdentity12(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID2))
// 	userHash := h.Sum(nil)
// 	var users []did.User
// 	var user = did.User{
// 		hex.EncodeToString(userHash),
// 		3,
// 		true,
// 	}
// 	users = append(users, user)
// 	var param = did.RegisterIdentityParam{
// 		users,
// 	}
// 	RegisterIdentity(t, param, idpPrivK5, IdP4, "success")
// }

// func TestIdP4RegisterIdentity2(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var users []did.User
// 	var user = did.User{
// 		hex.EncodeToString(userHash),
// 		3,
// 		false,
// 	}
// 	users = append(users, user)
// 	var param = did.RegisterIdentityParam{
// 		users,
// 	}
// 	RegisterIdentity(t, param, idpPrivK5, IdP4, "success")
// }

// func TestQueryGetIdpNodes3(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var param did.GetIdpNodesParam
// 	param.HashID = hex.EncodeToString(userHash)
// 	param.MinIal = 1
// 	param.MinAal = 1
// 	var expected = `{"node":[{"node_id":"` + IdP1 + `","node_name":"IdP Number 1 from ...","max_ial":2.3,"max_aal":2.4},{"node_id":"` + IdP4 + `","node_name":"IdP Number 4 from ...","max_ial":3,"max_aal":3}]}`
// 	GetIdpNodesExpectString(t, param, expected)
// }

// func TestRegisterNodeAS2(t *testing.T) {
// 	asKey := getPrivateKeyFromString(asPrivK2)
// 	asPublicKeyBytes, err := generatePublicKey(&asKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	asKey2 := getPrivateKeyFromString(allMasterKey)
// 	asPublicKeyBytes2, err := generatePublicKey(&asKey2.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param did.RegisterNode
// 	param.NodeName = AS2
// 	param.NodeID = AS2
// 	param.PublicKey = string(asPublicKeyBytes)
// 	param.MasterPublicKey = string(asPublicKeyBytes2)
// 	param.Role = "AS"
// 	RegisterNode(t, param)
// }

// func TestSetNodeTokenAS2(t *testing.T) {
// 	var param = did.SetNodeTokenParam{
// 		AS2,
// 		100.0,
// 	}
// 	SetNodeToken(t, param)
// }

// func TestASRegisterServiceDestinationByNDID(t *testing.T) {
// 	var param = did.RegisterServiceDestinationByNDIDParam{
// 		serviceID1,
// 		AS2,
// 	}
// 	RegisterServiceDestinationByNDID(t, param)
// }

// func TestAS2RegisterServiceDestination(t *testing.T) {
// 	var param = did.RegisterServiceDestinationParam{
// 		serviceID1,
// 		2.8,
// 		2.9,
// 	}
// 	RegisterServiceDestination(t, param, asPrivK2, AS2, "success")
// }

// func TestQueryGetAsNodesByServiceId2(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID1
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":1.4,"min_aal":1.5},{"node_id":"` + AS2 + `","node_name":"` + AS2 + `","min_ial":2.8,"min_aal":2.9}]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestDisableNode(t *testing.T) {
// 	var param did.DisableNodeParam
// 	param.NodeID = IdP1
// 	DisableNode(t, param)
// }

// func TestDisableNode2(t *testing.T) {
// 	var param did.DisableNodeParam
// 	param.NodeID = AS2
// 	DisableNode(t, param)
// }

// func TestQueryGetAsNodesByServiceId3(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID1
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":1.4,"min_aal":1.5}]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestNDIDDisableService2(t *testing.T) {
// 	var param = did.DisableServiceParam{
// 		serviceID1,
// 	}
// 	DisableService(t, param)
// }
// func TestQueryGetAsNodesByServiceId4(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID1
// 	var expected = `{"node":[]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestNDIDAddService3(t *testing.T) {
// 	var param did.AddServiceParam
// 	param.ServiceID = serviceID3
// 	param.ServiceName = "Bank statement"
// 	param.DataSchema = "DataSchema"
// 	param.DataSchemaVersion = "DataSchemaVersion"
// 	AddService(t, param)
// }

// func TestNDIDAddService4(t *testing.T) {
// 	var param did.AddServiceParam
// 	param.ServiceID = serviceID4
// 	param.ServiceName = "Bank statement"
// 	param.DataSchema = "DataSchema"
// 	param.DataSchemaVersion = "DataSchemaVersion"
// 	AddService(t, param)
// }

// func TestNDIDAddService5(t *testing.T) {
// 	var param did.AddServiceParam
// 	param.ServiceID = serviceID5
// 	param.ServiceName = "Bank statement"
// 	param.DataSchema = "DataSchema"
// 	param.DataSchemaVersion = "DataSchemaVersion"
// 	AddService(t, param)
// }

// func TestASRegisterServiceDestinationByNDID3(t *testing.T) {
// 	var param = did.RegisterServiceDestinationByNDIDParam{
// 		serviceID3,
// 		AS1,
// 	}
// 	RegisterServiceDestinationByNDID(t, param)
// }

// func TestASRegisterServiceDestinationByNDID4(t *testing.T) {
// 	var param = did.RegisterServiceDestinationByNDIDParam{
// 		serviceID4,
// 		AS1,
// 	}
// 	RegisterServiceDestinationByNDID(t, param)
// }

// func TestASRegisterServiceDestinationByNDID5(t *testing.T) {
// 	var param = did.RegisterServiceDestinationByNDIDParam{
// 		serviceID5,
// 		AS1,
// 	}
// 	RegisterServiceDestinationByNDID(t, param)
// }

// func TestAS1RegisterServiceDestinationBankStatement1(t *testing.T) {
// 	var param = did.RegisterServiceDestinationParam{
// 		serviceID3,
// 		2.8,
// 		2.9,
// 	}
// 	RegisterServiceDestination(t, param, asPrivK, AS1, "success")
// }

// func TestAS1RegisterServiceDestinationBankStatement2(t *testing.T) {
// 	var param = did.RegisterServiceDestinationParam{
// 		serviceID4,
// 		2.2,
// 		2.2,
// 	}
// 	RegisterServiceDestination(t, param, asPrivK, AS1, "success")
// }

// func TestAS1RegisterServiceDestinationBankStatement3(t *testing.T) {
// 	var param = did.RegisterServiceDestinationParam{
// 		serviceID5,
// 		3.3,
// 		3.3,
// 	}
// 	RegisterServiceDestination(t, param, asPrivK, AS1, "success")
// }

// func TestASUpdateServiceDestination2(t *testing.T) {
// 	var param = did.UpdateServiceDestinationParam{
// 		serviceID3,
// 		1.1,
// 		1.1,
// 	}
// 	UpdateServiceDestination(t, param, AS1)
// }

// func TestQueryGetServicesByAsID(t *testing.T) {
// 	var param = did.GetServicesByAsIDParam{
// 		AS1,
// 	}
// 	var expected = `{"services":[{"service_id":"` + serviceID3 + `","min_ial":1.1,"min_aal":1.1,"active":true,"suspended":false},{"service_id":"` + serviceID4 + `","min_ial":2.2,"min_aal":2.2,"active":true,"suspended":false},{"service_id":"` + serviceID5 + `","min_ial":3.3,"min_aal":3.3,"active":true,"suspended":false}]}`
// 	GetServicesByAsID(t, param, expected)
// }

// func TestNDIDDisableService3(t *testing.T) {
// 	var param = did.DisableServiceParam{
// 		serviceID3,
// 	}
// 	DisableService(t, param)
// }

// func TestNDIDDisableServiceDestinationByNDIDInvalidNodeID(t *testing.T) {
// 	var param = did.DisableServiceDestinationByNDIDParam{
// 		serviceID4,
// 		"Invalid node ID",
// 	}
// 	DisableServiceDestinationByNDIDExpectedString(t, param, "Node ID not found")
// }

// func TestNDIDDisableServiceDestinationByNDIDInvalidRole(t *testing.T) {
// 	var param = did.DisableServiceDestinationByNDIDParam{
// 		serviceID4,
// 		IdP1,
// 	}
// 	DisableServiceDestinationByNDIDExpectedString(t, param, "Role of node ID is not AS")
// }

// func TestNDIDDisableServiceDestinationByNDID(t *testing.T) {
// 	var param = did.DisableServiceDestinationByNDIDParam{
// 		serviceID4,
// 		AS1,
// 	}
// 	DisableServiceDestinationByNDID(t, param)
// }

// func TestQueryGetAsNodesByServiceID(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID4
// 	var expected = `{"node":[]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestQueryGetServicesByAsID2(t *testing.T) {
// 	var param = did.GetServicesByAsIDParam{
// 		AS1,
// 	}
// 	var expected = `{"services":[{"service_id":"` + serviceID4 + `","min_ial":2.2,"min_aal":2.2,"active":true,"suspended":true},{"service_id":"` + serviceID5 + `","min_ial":3.3,"min_aal":3.3,"active":true,"suspended":false}]}`
// 	GetServicesByAsID(t, param, expected)
// }

// func TestQueryGetIdpNodes6(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var param did.GetIdpNodesParam
// 	param.HashID = hex.EncodeToString(userHash)
// 	param.MinIal = 1
// 	param.MinAal = 1
// 	var expected = `{"node":[{"node_id":"` + IdP4 + `","node_name":"IdP Number 4 from ...","max_ial":3,"max_aal":3}]}`
// 	GetIdpNodesExpectString(t, param, expected)
// }

// func TestQueryGetAccessorKey3(t *testing.T) {
// 	var param = did.GetAccessorGroupIDParam{
// 		accessorID1.String(),
// 	}
// 	var expected = `{"accessor_public_key":"` + strings.Replace(accessorPubKey, "\n", "\\n", -1) + `","active":true}`
// 	GetAccessorKey(t, param, expected)
// }

// func TestEnableNode(t *testing.T) {
// 	var param did.DisableNodeParam
// 	param.NodeID = IdP1
// 	EnableNode(t, param)
// }

// func TestQueryGetIdpNodes7(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + userID))
// 	userHash := h.Sum(nil)
// 	var param did.GetIdpNodesParam
// 	param.HashID = hex.EncodeToString(userHash)
// 	param.MinIal = 1
// 	param.MinAal = 1
// 	var expected = `{"node":[{"node_id":"` + IdP1 + `","node_name":"IdP Number 1 from ...","max_ial":2.3,"max_aal":2.4},{"node_id":"` + IdP4 + `","node_name":"IdP Number 4 from ...","max_ial":3,"max_aal":3}]}`
// 	GetIdpNodesExpectString(t, param, expected)
// }

// func TestNDIDEnableServiceDestinationByNDIDInvalidNodeID(t *testing.T) {
// 	var param = did.DisableServiceDestinationByNDIDParam{
// 		serviceID4,
// 		"Invalid node ID",
// 	}
// 	EnableServiceDestinationByNDIDExpectedString(t, param, "Node ID not found")
// }

// func TestNDIDEnableServiceDestinationByNDIDInvalidRole(t *testing.T) {
// 	var param = did.DisableServiceDestinationByNDIDParam{
// 		serviceID4,
// 		IdP1,
// 	}
// 	EnableServiceDestinationByNDIDExpectedString(t, param, "Role of node ID is not AS")
// }

// func TestNDIDEnableServiceDestinationByNDID(t *testing.T) {
// 	var param = did.DisableServiceDestinationByNDIDParam{
// 		serviceID4,
// 		AS1,
// 	}
// 	EnableServiceDestinationByNDID(t, param)
// }

// func TestQueryGetAsNodesByServiceIDAfterEnable(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID4
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":2.2,"min_aal":2.2}]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestQueryGetServicesByAsID3(t *testing.T) {
// 	var param = did.GetServicesByAsIDParam{
// 		AS1,
// 	}
// 	var expected = `{"services":[{"service_id":"` + serviceID4 + `","min_ial":2.2,"min_aal":2.2,"active":true,"suspended":false},{"service_id":"` + serviceID5 + `","min_ial":3.3,"min_aal":3.3,"active":true,"suspended":false}]}`
// 	GetServicesByAsID(t, param, expected)
// }

// func TestEnableNamespace(t *testing.T) {
// 	var param did.DisableNamespaceParam
// 	param.Namespace = namespaceID2
// 	EnableNamespace(t, param)
// }

// func TestQueryGetNamespaceList2(t *testing.T) {
// 	expected := `[{"namespace":"` + namespaceID1 + `","description":"Citizen ID","active":true},{"namespace":"` + namespaceID2 + `","description":"Tel number","active":true}]`
// 	GetNamespaceListExpectString(t, expected)
// }

// func TestNDIDEnableService(t *testing.T) {
// 	var param = did.DisableServiceParam{
// 		serviceID1,
// 	}
// 	EnableService(t, param)
// }

// func TestQueryGetAsNodesByServiceId6(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID1
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":1.4,"min_aal":1.5}]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestQueryGetNodeNotFound(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = "123123"
// 	expected := `{}`
// 	GetNodeInfo(t, param, expected)
// }

// func TestRP1SetMqAddresses(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.99"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, rpPrivK, RP1)
// }

// func TestQueryGetNodeInfoRP1(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = RP1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwCB4UBzQcnd6GAzPgbt9\nj2idW23qKZrsvldPNifmOPLfLlMusv4EcyJf4L42/aQbTn1rVSu1blGkuCK+oRlK\nWmZEWh3xv9qrwCwov9Jme/KOE98zOMB10/xwnYotPadV0de80wGvKT7OlBlGulQR\nRhhgENNCPSxdUlozrPhrzGstXDr9zTYQoR3UD/7Ntmew3mnXvKj/8+U48hw913Xn\n6btBP3Uqg2OurXDGdrWciWgIMDEGyk65NOc8FOGa4AjYXzyi9TqOIfmysWhzKzU+\nfLysZQo10DfznnQN3w9+pI+20j2zB6ggpL75RjZKYgHU49pbvjF/eOSTOg9o5HwX\n0wIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"Node RP 1 edited","role":"RP","mq":[{"ip":"192.168.3.99","port":8000}]}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestASDisableServiceDestination(t *testing.T) {
// 	var param = did.DisableServiceDestinationParam{
// 		serviceID1,
// 	}
// 	DisableServiceDestination(t, param, AS1)
// }

// func TestQueryGetAsNodesByServiceId7(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID1
// 	var expected = `{"node":[]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestASEnableServiceDestination(t *testing.T) {
// 	var param = did.DisableServiceDestinationParam{
// 		serviceID1,
// 	}
// 	EnableServiceDestination(t, param, AS1)
// }

// func TestQueryGetAsNodesByServiceId8(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID1
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":1.4,"min_aal":1.5}]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestDisableNodeRP(t *testing.T) {
// 	var param did.DisableNodeParam
// 	param.NodeID = RP1
// 	DisableNode(t, param)
// }

// func TestRPCreateRequestAferDisableNode(t *testing.T) {
// 	var datas []did.DataRequest
// 	var data1 did.DataRequest
// 	data1.ServiceID = serviceID1
// 	data1.Count = 1
// 	data1.RequestParamsHash = "hash"
// 	datas = append(datas, data1)
// 	var param did.Request
// 	param.RequestID = requestID4.String()
// 	param.MinIdp = 1
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	param.Timeout = 259200
// 	param.DataRequestList = datas
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	fmt.Println(param.RequestID)
// 	CreateRequestExpectLog(t, param, rpPrivK, RP1, "Node is not active")
// }

// func TestEnableNodeRP1(t *testing.T) {
// 	var param did.DisableNodeParam
// 	param.NodeID = RP1
// 	EnableNode(t, param)
// }

// func TestRPCreateRequestAferEnableNodeWithDisabledIdP(t *testing.T) {
// 	var datas []did.DataRequest
// 	var data1 did.DataRequest
// 	data1.ServiceID = serviceID1
// 	data1.Count = 1
// 	data1.RequestParamsHash = "hash"
// 	data1.As = append(data1.As, AS1)
// 	datas = append(datas, data1)
// 	var param did.Request
// 	param.RequestID = requestID4.String()
// 	param.MinIdp = 1
// 	param.MinIal = 1
// 	param.MinAal = 1
// 	param.Timeout = 259200
// 	param.DataRequestList = datas
// 	param.IdPIDList = append(param.IdPIDList, IdP1)
// 	param.IdPIDList = append(param.IdPIDList, IdP2)
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	CreateRequestExpectLogDeliverTx(t, param, rpPrivK, RP1, "Node ID in IdP list is not active")
// }

// func TestRPCreateRequestAferEnableNode(t *testing.T) {
// 	var datas []did.DataRequest
// 	var data1 did.DataRequest
// 	data1.ServiceID = serviceID1
// 	data1.Count = 1
// 	data1.RequestParamsHash = "hash"
// 	data1.As = append(data1.As, AS1)
// 	datas = append(datas, data1)
// 	var param did.Request
// 	param.RequestID = requestID4.String()
// 	param.MinIdp = 1
// 	param.MinIal = 1
// 	param.MinAal = 1
// 	param.Timeout = 259200
// 	param.DataRequestList = datas
// 	param.IdPIDList = append(param.IdPIDList, IdP1)
// 	// param.IdPIDList = append(param.IdPIDList, IdP2)
// 	param.MessageHash = "hash('Please allow...')"
// 	param.Mode = 3
// 	CreateRequest(t, param, rpPrivK, RP1)
// }

// func TestIdPCreateIdpResponseNewRequest(t *testing.T) {
// 	var param = did.CreateIdpResponseParam{
// 		requestID4.String(),
// 		2,
// 		2,
// 		"accept",
// 		"signature",
// 		"Magic",
// 		"Magic",
// 	}
// 	CreateIdpResponse(t, param, idpPrivK2, IdP1)
// }

// func TestNDIDDisableServiceDestinationByNDIDForTest(t *testing.T) {
// 	var param = did.DisableServiceDestinationByNDIDParam{
// 		serviceID1,
// 		AS1,
// 	}
// 	DisableServiceDestinationByNDID(t, param)
// }

// func TestASSignDataForNewRequest(t *testing.T) {
// 	var param = did.SignDataParam{
// 		serviceID1,
// 		requestID4.String(),
// 		"sign(data,asKey)",
// 	}
// 	SignData(t, param, "Service destination is not approved by NDID", AS1)
// }

// func TestNDIDEnableServiceDestinationByNDIDForTest(t *testing.T) {
// 	var param = did.DisableServiceDestinationByNDIDParam{
// 		serviceID1,
// 		AS1,
// 	}
// 	EnableServiceDestinationByNDID(t, param)
// }

// func TestASDisableServiceDestination2(t *testing.T) {
// 	var param = did.DisableServiceDestinationParam{
// 		serviceID1,
// 	}
// 	DisableServiceDestination(t, param, AS1)
// }

// func TestASSignDataForNewRequest1(t *testing.T) {
// 	var param = did.SignDataParam{
// 		serviceID1,
// 		requestID4.String(),
// 		"sign(data,asKey)",
// 	}
// 	SignData(t, param, "Service destination is not active", AS1)
// }
// func TestNDIDDisableServiceForTest(t *testing.T) {
// 	var param = did.DisableServiceParam{
// 		serviceID1,
// 	}
// 	DisableService(t, param)
// }

// func TestASSignDataForNewRequest2(t *testing.T) {
// 	var param = did.SignDataParam{
// 		serviceID1,
// 		requestID4.String(),
// 		"sign(data,asKey)",
// 	}
// 	SignData(t, param, "Service is not active", AS1)
// }

// // Test invalid value
// func TestQueryGetNodePublicKeyInvalid(t *testing.T) {
// 	var param = did.GetNodePublicKeyParam{
// 		"RP10000",
// 	}
// 	expected := "not found"
// 	GetNodePublicKey(t, param, expected)
// }

// func TestQueryGetNodeMasterPublicKeyInvalid(t *testing.T) {
// 	var param = did.GetNodePublicKeyParam{
// 		"RP10000",
// 	}
// 	expected := "not found"
// 	GetNodeMasterPublicKey(t, param, expected)
// }

// func TestQueryGetIdpNodesInvalid(t *testing.T) {
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + "invalid user"))
// 	userHash := h.Sum(nil)
// 	var param did.GetIdpNodesParam
// 	param.HashID = hex.EncodeToString(userHash)
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	expected := "not found"
// 	GetIdpNodesExpectString(t, param, expected)
// }

// func TestQueryGetRequestInvalid(t *testing.T) {
// 	var param = did.GetRequestParam{
// 		"ef6f4c9c-818b-42b8-8904-3d97c-invalid",
// 	}
// 	expected := "not found"
// 	GetRequestExpectString(t, param, expected)
// }

// func TestQueryGetRequestDetailInvalid(t *testing.T) {
// 	var param = did.GetRequestParam{
// 		"ef6f4c9c-818b-42b8-8904-3d97c-invalid",
// 	}
// 	expected := "not found"
// 	GetRequestDetail(t, param, expected)
// }

// func TestQueryGetAsNodesByServiceIdInvalid(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = "statement-invalid"
// 	expected := "not found"
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestQueryGetMqAddressesInvalid(t *testing.T) {
// 	var param = did.GetMqAddressesParam{
// 		"IdP1-Invalid",
// 	}
// 	expected := "not found"
// 	GetMqAddressesExpectString(t, param, expected)
// }

// func TestQueryGetNodeTokenInvalid(t *testing.T) {
// 	var param = did.GetNodeTokenParam{
// 		"RP1-Invalid",
// 	}
// 	expected := "not found"
// 	GetNodeTokenExpectString(t, param, expected)
// }

// // func TestReportGetUsedTokenInvalid(t *testing.T) {
// // 	var param = did.GetUsedTokenReportParam{
// // 		"RP1-Invalid",
// // 	}
// // 	expected := "not found"
// // 	GetUsedTokenReport(t, param, expected)
// // }

// func TestQueryGetServiceDetailInvalid(t *testing.T) {
// 	var param = did.GetServiceDetailParam{
// 		"statement-invalid",
// 	}
// 	expected := "not found"
// 	GetServiceDetailExpectString(t, param, expected)
// }

// func TestQueryGetAccessorGroupIDInvalid(t *testing.T) {
// 	var param = did.GetAccessorGroupIDParam{
// 		"accessor_id_2-Invalid",
// 	}
// 	expected := "not found"
// 	GetAccessorGroupID(t, param, expected)
// }

// func TestQueryGetAccessorKeyInvalid(t *testing.T) {
// 	var param = did.GetAccessorGroupIDParam{
// 		"accessor_id-Invalid",
// 	}
// 	expected := "not found"
// 	GetAccessorKey(t, param, expected)
// }

// func TestQueryGetNodeInfoInvalid(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = "IdP1-Invalid"
// 	expected := "not found"
// 	GetNodeInfo(t, param, expected)
// }

// func TestQueryGetIdentityInfoInvalid(t *testing.T) {
// 	var param did.GetIdentityInfoParam
// 	h := sha256.New()
// 	h.Write([]byte(userNamespace + "Invalid user"))
// 	userHash := h.Sum(nil)
// 	param.NodeID = IdP1
// 	param.HashID = hex.EncodeToString(userHash)
// 	expected := "not found"
// 	GetIdentityInfo(t, param, expected)
// }

// func TestQueryGetDataSignatureInvalid(t *testing.T) {
// 	var param did.GetDataSignatureParam
// 	param.NodeID = "AS1-Invalid"
// 	param.RequestID = requestID1.String()
// 	param.ServiceID = serviceID1
// 	expected := "not found"
// 	GetDataSignature(t, param, expected)
// }

// func TestQueryGetServicesByAsIDInvalid(t *testing.T) {
// 	var param = did.GetServicesByAsIDParam{
// 		"AS1-Invalid",
// 	}
// 	expected := "not found"
// 	GetServicesByAsID(t, param, expected)
// }

// func TestQueryGetAsNodesByServiceIdBeforeUpdateNodeName(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID4
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"AS1","min_ial":2.2,"min_aal":2.2}]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestUpdateNodeAS1ByNDID(t *testing.T) {
// 	var param did.UpdateNodeByNDIDParam
// 	param.NodeID = AS1
// 	param.NodeName = "UpdatedName_AS1"
// 	UpdateNodeByNDID(t, param)
// }

// func TestQueryGetAsNodesByServiceIdAfterUpdateNodeName(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID4
// 	var expected = `{"node":[{"node_id":"` + AS1 + `","node_name":"UpdatedName_AS1","min_ial":2.2,"min_aal":2.2}]}`
// 	GetAsNodesByServiceId(t, param, expected)
// }

// func TestUpdateNodeNDID(t *testing.T) {
// 	ndidKey := getPrivateKeyFromString(ndidPrivK)
// 	ndidpublicKeyBytes, err := generatePublicKey(&ndidKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param = did.UpdateNodeParam{
// 		string(ndidpublicKeyBytes),
// 		"",
// 	}
// 	UpdateNode(t, param, ndidPrivK, "NDID")
// }

// func TestQueryGetIdpNodesInfo2(t *testing.T) {
// 	var param did.GetIdpNodesParam
// 	param.HashID = ""
// 	param.MinIal = 3
// 	param.MinAal = 3
// 	var expected = `{"node":[{"node_id":"` + IdP4 + `","name":"IdP Number 4 from ...","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu9+CK/vznpXtAUC0QhuJ\ngYKCfMMBiIgVcp2A+e+SsKvv6ESQ72R8K6nQAhH2MGtnj3ScLI0tMwCtgotWCEGi\nyUXKXLVTiqAqtwflCUVuxCDVuvOm3GQCxvwzE34jEgbGZ33G3tV7uKTtifhoJzVY\nD+WkZVslBhaBgQCUewCX4zkCCTYC5VEhkr7K8HGEr6n1eBOO5VORCkrHKYoZK7eu\nNjyWvWYyVN07F8K0RhgIF9Xsa6Tiu1Yf8zuyJ/awR6U4Nw+oTkvRpx64+caBNYgR\n4n8peg9ZJeTAwV49o1ymx34pPjHUgSdpyhZX4i3z9ji+o7KbNkA/O0l+3doMuH1e\nxwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}]},{"node_id":"` + IdP5 + `","name":"IdP Number 5 from ...","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApbxaA5aKnkpnV7+dMW5x\n7iEINouvjhQ8gl6+8A6ApiVbYIzJCCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaH\nOk1FfX6tk5i5uGWifRQHueXhXjR8HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7X\nHdh9Mqz+XAkq4T6vtdaocduarNG6ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3T\nHkM9noP2AYEH2MBHCzuu+bifCwurOBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jd\nAmbbqMyGpAsF+WjW4V2rcTMFZOoYFYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP\n2wIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}]}]}`
// 	GetIdpNodesInfo(t, param, expected)
// }

// func TestQueryGetNodeInfoInvalidNodeID(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = "InvalidNodeID"
// 	expected := string(`{}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestQueryGetIdpNodesInfo3(t *testing.T) {
// 	param := make(map[string]interface{})
// 	param["min_ial"] = 3
// 	param["min_aal"] = 3
// 	nodeIDList := make([]string, 0)
// 	nodeIDList = append(nodeIDList, IdP5)
// 	param["node_id_list"] = nodeIDList
// 	jsonStr, err := json.Marshal(param)
// 	if err != nil {
// 		panic(err)
// 	}
// 	var expected = `{"node":[{"node_id":"` + IdP5 + `","name":"IdP Number 5 from ...","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApbxaA5aKnkpnV7+dMW5x\n7iEINouvjhQ8gl6+8A6ApiVbYIzJCCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaH\nOk1FfX6tk5i5uGWifRQHueXhXjR8HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7X\nHdh9Mqz+XAkq4T6vtdaocduarNG6ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3T\nHkM9noP2AYEH2MBHCzuu+bifCwurOBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jd\nAmbbqMyGpAsF+WjW4V2rcTMFZOoYFYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP\n2wIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}]}]}`
// 	GetIdpNodesInfoParamJSON(t, string(jsonStr), expected)
// }

// func TestRegisterProxyNode(t *testing.T) {
// 	idpKey := getPrivateKeyFromString(idpPrivK)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param did.RegisterNode
// 	param.NodeID = Proxy1
// 	param.NodeName = "Proxy1"
// 	param.PublicKey = string(idpPublicKeyBytes)
// 	param.MasterPublicKey = string(idpPublicKeyBytes)
// 	param.Role = "Proxy"
// 	RegisterNode(t, param)
// }
// func TestSetNodeTokenProxy1(t *testing.T) {
// 	var param = did.SetNodeTokenParam{
// 		Proxy1,
// 		100.0,
// 	}
// 	SetNodeToken(t, param)
// }

// func TestQueryGetNodeInfoProxy1BeforeRegisterMsq(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = Proxy1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"Proxy1","role":"Proxy","mq":null}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestRegisterIdP6BehindProxy1(t *testing.T) {
// 	idpKey := getPrivateKeyFromString(idpPrivK)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param did.RegisterNode
// 	param.NodeID = IdP6BehindProxy1
// 	param.NodeName = "IdP6BehindProxy1"
// 	param.PublicKey = string(idpPublicKeyBytes)
// 	param.MasterPublicKey = string(idpPublicKeyBytes)
// 	param.Role = "IdP"
// 	param.MaxIal = 3
// 	param.MaxAal = 3
// 	RegisterNode(t, param)
// }

// func TestSetNodeTokenIdP6BehindProxy1(t *testing.T) {
// 	var param = did.SetNodeTokenParam{
// 		IdP6BehindProxy1,
// 		100.0,
// 	}
// 	SetNodeToken(t, param)
// }

// func TestSetMqAddressesIdP6BehindProxy1_After_Register_Node(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.102"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, idpPrivK, IdP6BehindProxy1)
// }

// func TestAddNodeToProxyNodeProxy_Invalid(t *testing.T) {
// 	var param = did.AddNodeToProxyNodeParam{
// 		IdP6BehindProxy1,
// 		"Invalid-Proxy",
// 		"KEY_ON_PROXY",
// 	}
// 	AddNodeToProxyNode(t, param, "Proxy node ID not found")
// }
// func TestAddNodeToProxyNodeProxy1(t *testing.T) {
// 	var param = did.AddNodeToProxyNodeParam{
// 		IdP6BehindProxy1,
// 		Proxy1,
// 		"KEY_ON_PROXY",
// 	}
// 	AddNodeToProxyNode(t, param, "success")
// }

// func TestAddNodeToProxyNodeProxy2(t *testing.T) {
// 	var param = did.AddNodeToProxyNodeParam{
// 		IdP6BehindProxy1,
// 		Proxy1,
// 		"KEY_ON_PROXY",
// 	}
// 	AddNodeToProxyNode(t, param, "This node ID is already associated with a proxy node")
// }

// func TestAddNodeToProxyNodeProxy1Proxy1(t *testing.T) {
// 	var param = did.AddNodeToProxyNodeParam{
// 		Proxy1,
// 		Proxy1,
// 		"KEY_ON_PROXY",
// 	}
// 	AddNodeToProxyNode(t, param, "This node ID is an ID of a proxy node")
// }

// func TestQueryGetNodeInfoIdP6BehindProxy1BeforeProxyRegisterMsq(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = IdP6BehindProxy1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP6BehindProxy1","role":"IdP","max_ial":3,"max_aal":3,"proxy":{"node_id":"` + Proxy1 + `","node_name":"Proxy1","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":null,"config":"KEY_ON_PROXY"}}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestSetMqAddressesProxy1(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.99"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, idpPrivK, Proxy1)
// }

// func TestQueryGetNodeInfoProxy1(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = Proxy1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"Proxy1","role":"Proxy","mq":[{"ip":"192.168.3.99","port":8000}]}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestQueryGetNodeInfoIdP6BehindProxy1(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = IdP6BehindProxy1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP6BehindProxy1","role":"IdP","max_ial":3,"max_aal":3,"proxy":{"node_id":"` + Proxy1 + `","node_name":"Proxy1","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"config":"KEY_ON_PROXY"}}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestQueryGetGetNodesBehindProxyNode1(t *testing.T) {
// 	var param did.GetNodesBehindProxyNodeParam
// 	param.ProxyNodeID = Proxy1
// 	expected := string(`{"nodes":[{"node_id":"` + IdP6BehindProxy1 + `","node_name":"IdP6BehindProxy1","role":"IdP","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","max_ial":3,"max_aal":3,"config":"KEY_ON_PROXY"}]}`)
// 	GetNodesBehindProxyNode(t, param, expected)
// }

// func TestQueryGetIdpNodesInfo4(t *testing.T) {
// 	param := make(map[string]interface{})
// 	param["min_ial"] = 3
// 	param["min_aal"] = 3
// 	jsonStr, err := json.Marshal(param)
// 	if err != nil {
// 		panic(err)
// 	}
// 	var expected = `{"node":[{"node_id":"` + IdP4 + `","name":"IdP Number 4 from ...","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu9+CK/vznpXtAUC0QhuJ\ngYKCfMMBiIgVcp2A+e+SsKvv6ESQ72R8K6nQAhH2MGtnj3ScLI0tMwCtgotWCEGi\nyUXKXLVTiqAqtwflCUVuxCDVuvOm3GQCxvwzE34jEgbGZ33G3tV7uKTtifhoJzVY\nD+WkZVslBhaBgQCUewCX4zkCCTYC5VEhkr7K8HGEr6n1eBOO5VORCkrHKYoZK7eu\nNjyWvWYyVN07F8K0RhgIF9Xsa6Tiu1Yf8zuyJ/awR6U4Nw+oTkvRpx64+caBNYgR\n4n8peg9ZJeTAwV49o1ymx34pPjHUgSdpyhZX4i3z9ji+o7KbNkA/O0l+3doMuH1e\nxwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}]},{"node_id":"` + IdP5 + `","name":"IdP Number 5 from ...","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApbxaA5aKnkpnV7+dMW5x\n7iEINouvjhQ8gl6+8A6ApiVbYIzJCCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaH\nOk1FfX6tk5i5uGWifRQHueXhXjR8HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7X\nHdh9Mqz+XAkq4T6vtdaocduarNG6ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3T\nHkM9noP2AYEH2MBHCzuu+bifCwurOBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jd\nAmbbqMyGpAsF+WjW4V2rcTMFZOoYFYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP\n2wIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}]},{"node_id":"` + IdP6BehindProxy1 + `","name":"IdP6BehindProxy1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","proxy":{"node_id":"` + Proxy1 + `","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"config":"KEY_ON_PROXY"}}]}`
// 	GetIdpNodesInfoParamJSON(t, string(jsonStr), expected)
// }

// func TestDisableNodeProxy1(t *testing.T) {
// 	var param did.DisableNodeParam
// 	param.NodeID = Proxy1
// 	DisableNode(t, param)
// }

// func TestQueryGetIdpNodesInfo4AfterDisableProxy(t *testing.T) {
// 	param := make(map[string]interface{})
// 	param["min_ial"] = 3
// 	param["min_aal"] = 3
// 	jsonStr, err := json.Marshal(param)
// 	if err != nil {
// 		panic(err)
// 	}
// 	var expected = `{"node":[{"node_id":"` + IdP4 + `","name":"IdP Number 4 from ...","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu9+CK/vznpXtAUC0QhuJ\ngYKCfMMBiIgVcp2A+e+SsKvv6ESQ72R8K6nQAhH2MGtnj3ScLI0tMwCtgotWCEGi\nyUXKXLVTiqAqtwflCUVuxCDVuvOm3GQCxvwzE34jEgbGZ33G3tV7uKTtifhoJzVY\nD+WkZVslBhaBgQCUewCX4zkCCTYC5VEhkr7K8HGEr6n1eBOO5VORCkrHKYoZK7eu\nNjyWvWYyVN07F8K0RhgIF9Xsa6Tiu1Yf8zuyJ/awR6U4Nw+oTkvRpx64+caBNYgR\n4n8peg9ZJeTAwV49o1ymx34pPjHUgSdpyhZX4i3z9ji+o7KbNkA/O0l+3doMuH1e\nxwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}]},{"node_id":"` + IdP5 + `","name":"IdP Number 5 from ...","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApbxaA5aKnkpnV7+dMW5x\n7iEINouvjhQ8gl6+8A6ApiVbYIzJCCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaH\nOk1FfX6tk5i5uGWifRQHueXhXjR8HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7X\nHdh9Mqz+XAkq4T6vtdaocduarNG6ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3T\nHkM9noP2AYEH2MBHCzuu+bifCwurOBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jd\nAmbbqMyGpAsF+WjW4V2rcTMFZOoYFYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP\n2wIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}]}]}`
// 	GetIdpNodesInfoParamJSON(t, string(jsonStr), expected)
// }

// func TestEnableNodeProxy1(t *testing.T) {
// 	var param did.DisableNodeParam
// 	param.NodeID = Proxy1
// 	EnableNode(t, param)
// }

// func TestRegisterAS3BehindProxy1(t *testing.T) {
// 	asKey := getPrivateKeyFromString(asPrivK)
// 	asPublicKeyBytes, err := generatePublicKey(&asKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param did.RegisterNode
// 	param.NodeID = AS3BehindProxy1
// 	param.NodeName = "AS3BehindProxy1"
// 	param.PublicKey = string(asPublicKeyBytes)
// 	param.MasterPublicKey = string(asPublicKeyBytes)
// 	param.Role = "AS"
// 	RegisterNode(t, param)
// }

// func TestSetNodeTokenAS3BehindProxy1(t *testing.T) {
// 	var param = did.SetNodeTokenParam{
// 		AS3BehindProxy1,
// 		100.0,
// 	}
// 	SetNodeToken(t, param)
// }

// func TestAddNodeToProxyNodeAS3BehindProxy1(t *testing.T) {
// 	var param = did.AddNodeToProxyNodeParam{
// 		AS3BehindProxy1,
// 		Proxy1,
// 		"KEY_ON_PROXY",
// 	}
// 	AddNodeToProxyNode(t, param, "success")
// }

// func TestQueryGetNodeInfoAS3BehindProxy1(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = AS3BehindProxy1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"AS3BehindProxy1","role":"AS","proxy":{"node_id":"` + Proxy1 + `","node_name":"Proxy1","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"config":"KEY_ON_PROXY"}}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestQueryGetGetNodesBehindProxyNode2(t *testing.T) {
// 	var param did.GetNodesBehindProxyNodeParam
// 	param.ProxyNodeID = Proxy1
// 	expected := string(`{"nodes":[{"node_id":"` + IdP6BehindProxy1 + `","node_name":"IdP6BehindProxy1","role":"IdP","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","max_ial":3,"max_aal":3,"config":"KEY_ON_PROXY"},{"node_id":"` + AS3BehindProxy1 + `","node_name":"AS3BehindProxy1","role":"AS","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","config":"KEY_ON_PROXY"}]}`)
// 	GetNodesBehindProxyNode(t, param, expected)
// }

// func TestNDIDAddServiceserviceID6(t *testing.T) {
// 	var param did.AddServiceParam
// 	param.ServiceID = serviceID6
// 	param.ServiceName = "Service 6"
// 	param.DataSchema = "DataSchema"
// 	param.DataSchemaVersion = "DataSchemaVersion"
// 	AddService(t, param)
// }

// func TestASRegisterServiceDestinationByNDIDForserviceID6(t *testing.T) {
// 	var param = did.RegisterServiceDestinationByNDIDParam{
// 		serviceID6,
// 		AS3BehindProxy1,
// 	}
// 	RegisterServiceDestinationByNDID(t, param)
// }

// func TestASRegisterServiceDestinationserviceID6(t *testing.T) {
// 	var param = did.RegisterServiceDestinationParam{
// 		serviceID6,
// 		1.1,
// 		1.2,
// 	}
// 	RegisterServiceDestination(t, param, asPrivK, AS3BehindProxy1, "success")
// }

// func TestQueryGetAsNodesInfoByServiceIdWithProxy(t *testing.T) {
// 	var param did.GetAsNodesByServiceIdParam
// 	param.ServiceID = serviceID6
// 	var expected = `{"node":[{"node_id":"` + AS3BehindProxy1 + `","name":"AS3BehindProxy1","min_ial":1.1,"min_aal":1.2,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","proxy":{"node_id":"` + Proxy1 + `","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"config":"KEY_ON_PROXY"}}]}`
// 	GetAsNodesInfoByServiceId(t, param, expected)
// }

// func TestRegisterProxyNodeProxy2(t *testing.T) {
// 	idpKey := getPrivateKeyFromString(idpPrivK)
// 	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var param did.RegisterNode
// 	param.NodeID = Proxy2
// 	param.NodeName = "Proxy2"
// 	param.PublicKey = string(idpPublicKeyBytes)
// 	param.MasterPublicKey = string(idpPublicKeyBytes)
// 	param.Role = "Proxy"
// 	RegisterNode(t, param)
// }
// func TestSetNodeTokenProxy2(t *testing.T) {
// 	var param = did.SetNodeTokenParam{
// 		Proxy2,
// 		100.0,
// 	}
// 	SetNodeToken(t, param)
// }

// func TestSetMqAddressesProxy2(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.99"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, idpPrivK, Proxy2)
// }

// func TestUpdateNodeProxyNodeProxy2(t *testing.T) {
// 	var param = did.UpdateNodeProxyNodeParam{
// 		IdP6BehindProxy1,
// 		Proxy2,
// 		"KEY_ON_PROXY",
// 	}
// 	UpdateNodeProxyNode(t, param, "success")
// }

// func TestQueryGetGetNodesBehindProxyNode3(t *testing.T) {
// 	var param did.GetNodesBehindProxyNodeParam
// 	param.ProxyNodeID = Proxy1
// 	expected := string(`{"nodes":[{"node_id":"` + AS3BehindProxy1 + `","node_name":"AS3BehindProxy1","role":"AS","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","config":"KEY_ON_PROXY"}]}`)
// 	GetNodesBehindProxyNode(t, param, expected)
// }
// func TestQueryGetGetNodesBehindProxyNode4(t *testing.T) {
// 	var param did.GetNodesBehindProxyNodeParam
// 	param.ProxyNodeID = Proxy2
// 	expected := string(`{"nodes":[{"node_id":"` + IdP6BehindProxy1 + `","node_name":"IdP6BehindProxy1","role":"IdP","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","max_ial":3,"max_aal":3,"config":"KEY_ON_PROXY"}]}`)
// 	GetNodesBehindProxyNode(t, param, expected)
// }

// func TestUpdateNodeProxyNodeProxy2_2(t *testing.T) {
// 	var param = did.UpdateNodeProxyNodeParam{
// 		IdP6BehindProxy1,
// 		Proxy2,
// 		"KEY_ON_NODE",
// 	}
// 	UpdateNodeProxyNode(t, param, "success")
// }

// func TestQueryGetGetNodesBehindProxyNode4_2(t *testing.T) {
// 	var param did.GetNodesBehindProxyNodeParam
// 	param.ProxyNodeID = Proxy2
// 	expected := string(`{"nodes":[{"node_id":"` + IdP6BehindProxy1 + `","node_name":"IdP6BehindProxy1","role":"IdP","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","max_ial":3,"max_aal":3,"config":"KEY_ON_NODE"}]}`)
// 	GetNodesBehindProxyNode(t, param, expected)
// }

// func TestUpdateNodeProxyNodeProxy2_3(t *testing.T) {
// 	var param = did.UpdateNodeProxyNodeParam{
// 		IdP6BehindProxy1,
// 		Proxy2,
// 		"KEY_ON_PROXY",
// 	}
// 	UpdateNodeProxyNode(t, param, "success")
// }

// func TestUpdateNodeProxyNodeProxy2_InvalidProxy(t *testing.T) {
// 	var param = did.UpdateNodeProxyNodeParam{
// 		IdP6BehindProxy1,
// 		"Invalid-Proxy",
// 		"KEY_ON_PROXY",
// 	}
// 	UpdateNodeProxyNode(t, param, "Proxy node ID not found")
// }

// func TestQueryGetNodeInfoIdP6BehindProxy2(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = IdP6BehindProxy1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP6BehindProxy1","role":"IdP","max_ial":3,"max_aal":3,"proxy":{"node_id":"` + Proxy2 + `","node_name":"Proxy2","public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"config":"KEY_ON_PROXY"}}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestRemoveNodeFromProxyNode1(t *testing.T) {
// 	var param = did.RemoveNodeFromProxyNode{
// 		IdP6BehindProxy1,
// 	}
// 	RemoveNodeFromProxyNode(t, param, "success")
// }

// func TestQueryGetNodeInfoIdP6BehindProxy3_After_Remove_From_Proxy(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = IdP6BehindProxy1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP6BehindProxy1","role":"IdP","max_ial":3,"max_aal":3,"mq":null}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestSetMqAddressesIdP6BehindProxy1(t *testing.T) {
// 	var mq did.MsqAddress
// 	mq.IP = "192.168.3.102"
// 	mq.Port = 8000
// 	var param did.SetMqAddressesParam
// 	param.Addresses = make([]did.MsqAddress, 0)
// 	param.Addresses = append(param.Addresses, mq)
// 	SetMqAddresses(t, param, idpPrivK, IdP6BehindProxy1)
// }

// func TestQueryGetGetNodesBehindProxyNode5(t *testing.T) {
// 	var param did.GetNodesBehindProxyNodeParam
// 	param.ProxyNodeID = Proxy2
// 	expected := string(`{"nodes":[]}`)
// 	GetNodesBehindProxyNode(t, param, expected)
// }

// func TestQueryGetNodeInfoIdP6BehindProxy3(t *testing.T) {
// 	var param did.GetNodeInfoParam
// 	param.NodeID = IdP6BehindProxy1
// 	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP6BehindProxy1","role":"IdP","max_ial":3,"max_aal":3,"mq":[{"ip":"192.168.3.102","port":8000}]}`)
// 	GetNodeInfo(t, param, expected)
// }

// func TestQueryGetNodeIDListAll(t *testing.T) {
// 	var param did.GetNodeIDListParam
// 	expected := string(`{"node_id_list":["` + RP1 + `","` + IdP1 + `","` + AS1 + `","` + IdP4 + `","` + IdP5 + `","` + Proxy1 + `","` + IdP6BehindProxy1 + `","` + AS3BehindProxy1 + `","` + Proxy2 + `"]}`)
// 	GetNodeIDList(t, param, expected)
// }

// func TestQueryGetNodeIDListRP(t *testing.T) {
// 	var param did.GetNodeIDListParam
// 	param.Role = "RP"
// 	expected := string(`{"node_id_list":["` + RP1 + `"]}`)
// 	GetNodeIDList(t, param, expected)
// }

// func TestQueryGetNodeIDListIdP(t *testing.T) {
// 	var param did.GetNodeIDListParam
// 	param.Role = "IdP"
// 	expected := string(`{"node_id_list":["` + IdP1 + `","` + IdP4 + `","` + IdP5 + `","` + IdP6BehindProxy1 + `"]}`)
// 	GetNodeIDList(t, param, expected)
// }

// func TestQueryGetNodeIDListAS(t *testing.T) {
// 	var param did.GetNodeIDListParam
// 	param.Role = "AS"
// 	expected := string(`{"node_id_list":["` + AS1 + `","` + AS3BehindProxy1 + `"]}`)
// 	GetNodeIDList(t, param, expected)
// }

// func TestDisableAllNode(t *testing.T) {
// 	var param did.GetNodeIDListParam
// 	allNode := GetNodeIDListForDisable(t, param)
// 	for _, nodeID := range allNode {
// 		var param did.DisableNodeParam
// 		param.NodeID = nodeID
// 		DisableNode(t, param)
// 	}
// }

// func TestGetChainHistory(t *testing.T) {
// 	GetChainHistory(t)
// }

// // func TestSetValidator0(t *testing.T) {
// // 	var param did.SetValidatorParam
// // 	param.PublicKey = getValidatorPubkey(3)
// // 	param.Power = 0
// // 	SetValidator(t, param)
// // }
