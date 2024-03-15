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

package app

import (
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strconv"
	"strings"

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	appTypes "github.com/ndidplatform/smart-contract/v9/abci/app/v1/types"
	"github.com/ndidplatform/smart-contract/v9/abci/code"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

var IsMethod = map[string]bool{
	"InitNDID":                         true,
	"RegisterNode":                     true,
	"AddNodeToken":                     true,
	"ReduceNodeToken":                  true,
	"SetNodeToken":                     true,
	"SetPriceFunc":                     true,
	"AddNamespace":                     true,
	"SetValidator":                     true,
	"AddService":                       true,
	"UpdateNodeByNDID":                 true,
	"UpdateService":                    true,
	"RegisterServiceDestinationByNDID": true,
	"DisableNode":                      true,
	"DisableNamespace":                 true,
	"DisableService":                   true,
	"DisableServiceDestinationByNDID":  true,
	"EnableNode":                       true,
	"EnableServiceDestinationByNDID":   true,
	"EnableNamespace":                  true,
	"EnableService":                    true,
	"AddErrorCode":                     true,
	"RemoveErrorCode":                  true,
	"RegisterIdentity":                 true,
	"AddAccessor":                      true,
	"CreateIdpResponse":                true,
	"UpdateIdentity":                   true,
	"CreateAsResponse":                 true,
	"RegisterServiceDestination":       true,
	"UpdateServiceDestination":         true,
	"CreateRequest":                    true,
	"SetMqAddresses":                   true,
	"UpdateNode":                       true,
	"CloseRequest":                     true,
	"TimeOutRequest":                   true,
	"SetDataReceived":                  true,
	"DisableServiceDestination":        true,
	"EnableServiceDestination":         true,
	"SetTimeOutBlockRegisterIdentity":  true,
	"AddNodeToProxyNode":               true,
	"UpdateNodeProxyNode":              true,
	"RemoveNodeFromProxyNode":          true,
	"RevokeAccessor":                   true,
	"SetInitData":                      true,
	"SetInitData_pb":                   true,
	"EndInit":                          true,
	"SetLastBlock":                     true,
	"RevokeIdentityAssociation":        true,
	"UpdateIdentityModeList":           true,
	"AddIdentity":                      true,
	"SetSupportedIALList":              true,
	"SetSupportedAALList":              true,
	"SetAllowedModeList":               true,
	"UpdateNamespace":                  true,
	"SetAllowedMinIalForRegisterIdentityAtFirstIdp":        true,
	"RevokeAndAddAccessor":                                 true,
	"SetServicePriceCeiling":                               true,
	"SetServicePriceMinEffectiveDatetimeDelay":             true,
	"SetServicePrice":                                      true,
	"CreateMessage":                                        true,
	"AddRequestType":                                       true,
	"RemoveRequestType":                                    true,
	"AddSuppressedIdentityModificationNotificationNode":    true,
	"RemoveSuppressedIdentityModificationNotificationNode": true,
	"AddAllowedNodeSupportedFeature":                       true,
	"RemoveAllowedNodeSupportedFeature":                    true,
}

func (app *ABCIApplication) isNDIDNode(node *data.NodeDetail) bool {
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleNdid {
		return true
	}

	return false
}

func (app *ABCIApplication) isNDIDNodeByNodeID(nodeID string, committedState bool) (bool, error) {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return app.isNDIDNode(&node), nil
}

func (app *ABCIApplication) isRPNode(node *data.NodeDetail) bool {
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleRp {
		return true
	}

	return false
}

func (app *ABCIApplication) isRPNodeByNodeID(nodeID string, committedState bool) (bool, error) {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return app.isRPNode(&node), nil
}

func (app *ABCIApplication) isIDPNode(node *data.NodeDetail) bool {
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleIdp && !node.IsIdpAgent {
		return true
	}

	return false
}

func (app *ABCIApplication) isIDPNodeByNodeID(nodeID string, committedState bool) (bool, error) {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return app.isIDPNode(&node), nil
}

func (app *ABCIApplication) isIDPAgentNode(node *data.NodeDetail) bool {
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleIdp && node.IsIdpAgent {
		return true
	}

	return false
}

func (app *ABCIApplication) isIDPAgentNodeByNodeID(nodeID string, committedState bool) (bool, error) {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return app.isIDPAgentNode(&node), nil
}

func (app *ABCIApplication) isIDPorIDPAgentNode(node *data.NodeDetail) bool {
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleIdp {
		return true
	}

	return false
}

func (app *ABCIApplication) isIDPorIDPAgentNodeByNodeID(nodeID string, committedState bool) (bool, error) {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return app.isIDPorIDPAgentNode(&node), nil
}

func (app *ABCIApplication) isASNode(node *data.NodeDetail) bool {
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleAs {
		return true
	}

	return false
}

func (app *ABCIApplication) isASNodeByNodeID(nodeID string, committedState bool) (bool, error) {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return app.isASNode(&node), nil
}

func (app *ABCIApplication) isIDPorRPNode(node *data.NodeDetail) bool {
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleRp {
		return true
	}
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleIdp && !node.IsIdpAgent {
		return true
	}

	return false
}

func (app *ABCIApplication) isIDPorRPNodeByNodeID(nodeID string, committedState bool) (bool, error) {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return app.isIDPorRPNode(&node), nil
}

func (app *ABCIApplication) isProxyNode(node *data.NodeDetail) bool {
	if appTypes.NodeRole(node.Role) == appTypes.NodeRoleProxy {
		return true
	}

	return false
}

func (app *ABCIApplication) isProxyNodeByNodeID(nodeID string, committedState bool) (bool, error) {
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(nodeDetailKey), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID not found",
		}
	}
	var node data.NodeDetail
	err = proto.Unmarshal(value, &node)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return app.isProxyNode(&node), nil
}

func createHash(message []byte, algorithm crypto.Hash) (hashResult []byte) {
	h := algorithm.New()
	h.Write(message)
	hashed := h.Sum(nil)

	return hashed
}

func verifySignature(
	method string,
	param []byte,
	chainID string,
	nonce []byte,
	signature []byte,
	publicKey string,
	algorithm appTypes.SignatureAlgorithm,
) (valid bool, err error) {
	publicKey = strings.Replace(publicKey, "\t", "", -1)
	block, _ := pem.Decode([]byte(publicKey))

	// Get key
	var callerPublicKey interface{}
	if strings.Contains(publicKey, "BEGIN RSA PUBLIC KEY") {
		callerPublicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
	} else {
		callerPublicKey, err = x509.ParsePKIXPublicKey(block.Bytes)
	}
	if err != nil {
		return false, err
	}

	// Build the message
	message := append([]byte(method), param...)
	message = append(message, []byte(chainID)...)
	message = append(message, []byte(nonce)...)

	switch pubKey := callerPublicKey.(type) {
	case *ecdsa.PublicKey:
		// Hash the message
		var hashAlgorithm crypto.Hash
		switch algorithm {
		case appTypes.SignatureAlgorithmECDSASHA256:
			hashAlgorithm = crypto.SHA256
		case appTypes.SignatureAlgorithmECDSASHA384:
			hashAlgorithm = crypto.SHA384
		}
		hashed := createHash(message, hashAlgorithm)

		return ecdsa.VerifyASN1(pubKey, hashed, signature), nil
	case *rsa.PublicKey:
		// Hash the message
		var hashAlgorithm crypto.Hash
		switch algorithm {
		case appTypes.SignatureAlgorithmRSAPSSSHA256,
			appTypes.SignatureAlgorithmRSAPKCS1V15SHA256:
			hashAlgorithm = crypto.SHA256
		case appTypes.SignatureAlgorithmRSAPSSSHA384,
			appTypes.SignatureAlgorithmRSAPKCS1V15SHA384:
			hashAlgorithm = crypto.SHA384
		case appTypes.SignatureAlgorithmRSAPSSSHA512,
			appTypes.SignatureAlgorithmRSAPKCS1V15SHA512:
			hashAlgorithm = crypto.SHA512
		}
		hashed := createHash(message, hashAlgorithm)

		switch algorithm {
		case appTypes.SignatureAlgorithmRSAPSSSHA256,
			appTypes.SignatureAlgorithmRSAPSSSHA384,
			appTypes.SignatureAlgorithmRSAPSSSHA512:
			err = rsa.VerifyPSS(pubKey, hashAlgorithm, hashed, signature, nil)
		case appTypes.SignatureAlgorithmRSAPKCS1V15SHA256,
			appTypes.SignatureAlgorithmRSAPKCS1V15SHA384,
			appTypes.SignatureAlgorithmRSAPKCS1V15SHA512:
			err = rsa.VerifyPKCS1v15(pubKey, hashAlgorithm, hashed, signature)
		}

		return err == nil, err
	case ed25519.PublicKey:
		valid := ed25519.Verify(pubKey, message, signature)
		return valid, nil
	}

	return false, nil
}

// ReturnCheckTx return types.ResponseDeliverTx
func ReturnCheckTx(code uint32, log string) types.ResponseCheckTx {
	return types.ResponseCheckTx{
		Code: code,
		Log:  log,
	}
}

func (app *ABCIApplication) getNodePublicKeyForSignatureVerification(
	method string,
	param []byte,
	nodeID string,
	committedState bool,
) (publicKeyPem string, algorithm appTypes.SignatureAlgorithm, retCode uint32, message string) {
	if method == "InitNDID" {
		publicKey, algorithmStr := getPublicKeyInitNDID(param)
		if publicKey == "" {
			return publicKey, "", code.CannotGetPublicKeyFromParam, "Can not get public key from parameter"
		}
		publicKeyPem = publicKey
		algorithm = appTypes.SignatureAlgorithm(algorithmStr)
	} else if method == "UpdateNode" {
		signingPublicKey := app.getSigningMasterPublicKeyFromNodeID(nodeID, committedState)
		if signingPublicKey == nil {
			return "", "", code.CannotGetMasterPublicKeyFromNodeID, "Can not get master public key from node ID"
		}
		publicKeyPem = signingPublicKey.PublicKey
		algorithm = appTypes.SignatureAlgorithm(signingPublicKey.Algorithm)
	} else {
		signingPublicKey := app.getSigningPublicKeyFromNodeID(nodeID, committedState)
		if signingPublicKey == nil {
			return "", "", code.CannotGetPublicKeyFromNodeID, "Can not get public key from node ID"
		}
		publicKeyPem = signingPublicKey.PublicKey
		algorithm = appTypes.SignatureAlgorithm(signingPublicKey.Algorithm)
	}
	return publicKeyPem, algorithm, code.OK, ""
}

func getPublicKeyInitNDID(param []byte) (string, string) {
	var funcParam InitNDIDParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return "", ""
	}
	return funcParam.SigningPublicKey, funcParam.SigningAlgorithm
}

func (app *ABCIApplication) getSigningMasterPublicKeyFromNodeID(nodeID string, committedState bool) *data.NodeKey {
	key := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		panic(err)
	}
	if value == nil {
		return nil
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return nil
	}
	return nodeDetail.SigningMasterPublicKey
}

func (app *ABCIApplication) getSigningPublicKeyFromNodeID(nodeID string, committedState bool) *data.NodeKey {
	key := nodeIDKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		panic(err)
	}
	if value == nil {
		return nil
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return nil
	}
	return nodeDetail.SigningPublicKey
}

func checkPubKeyForSigning(key string, algorithm appTypes.SignatureAlgorithm) (err error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return &ApplicationError{
			Code:    code.InvalidKeyFormat,
			Message: "Invalid key format. Cannot decode PEM.",
		}
	}
	var pub interface{}
	if strings.Contains(key, "BEGIN RSA PUBLIC KEY") {
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
	} else {
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	}
	if err != nil {
		return &ApplicationError{
			Code:    code.InvalidKeyFormat,
			Message: err.Error(),
		}
	}

	switch pubKey := pub.(type) {
	case *rsa.PublicKey:
		switch algorithm {
		case appTypes.SignatureAlgorithmRSAPSSSHA256,
			appTypes.SignatureAlgorithmRSAPSSSHA384,
			appTypes.SignatureAlgorithmRSAPSSSHA512,
			appTypes.SignatureAlgorithmRSAPKCS1V15SHA256,
			appTypes.SignatureAlgorithmRSAPKCS1V15SHA384,
			appTypes.SignatureAlgorithmRSAPKCS1V15SHA512:
			break
		default:
			return &ApplicationError{
				Code:    code.IncompatibleKeyAlgorithm,
				Message: "Incompatible key algorithm",
			}
		}

		if pubKey.N.BitLen() < 2048 {
			return &ApplicationError{
				Code:    code.RSAKeyLengthTooShort,
				Message: "RSA key length is too short. Must be at least 2048-bit.",
			}
		}
	case *ecdsa.PublicKey:
		// TODO: support secp256k1 curve
		switch algorithm {
		case appTypes.SignatureAlgorithmECDSASHA256:
			// secp256r1 / prime256v1
			if pubKey.Curve != elliptic.P256() {
				return &ApplicationError{
					Code:    code.UnsupportedSigningAlgorithm,
					Message: "unsupported signing algorithm",
				}
			}
		case appTypes.SignatureAlgorithmECDSASHA384:
			// secp384r1
			if pubKey.Curve != elliptic.P384() {
				return &ApplicationError{
					Code:    code.UnsupportedSigningAlgorithm,
					Message: "unsupported signing algorithm",
				}
			}
		default:
			return &ApplicationError{
				Code:    code.IncompatibleKeyAlgorithm,
				Message: "Incompatible key algorithm",
			}
		}

		return nil
	case ed25519.PublicKey:
		switch algorithm {
		case appTypes.SignatureAlgorithmEd25519:
			break
		default:
			return &ApplicationError{
				Code:    code.IncompatibleKeyAlgorithm,
				Message: "Incompatible key algorithm",
			}
		}

		return nil
	case *dsa.PublicKey:
		return &ApplicationError{
			Code:    code.UnsupportedKeyType,
			Message: "Unsupported key type",
		}
	default:
		return &ApplicationError{
			Code:    code.UnknownKeyType,
			Message: "Unknown key type",
		}
	}

	return nil
}

func checkPubKeyForEncryption(key string) (err error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return &ApplicationError{
			Code:    code.InvalidKeyFormat,
			Message: "Invalid key format. Cannot decode PEM.",
		}
	}
	var pub interface{}
	if strings.Contains(key, "BEGIN RSA PUBLIC KEY") {
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
	} else {
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	}
	if err != nil {
		return &ApplicationError{
			Code:    code.InvalidKeyFormat,
			Message: err.Error(),
		}
	}

	switch pubKey := pub.(type) {
	case *rsa.PublicKey:
		if pubKey.N.BitLen() < 2048 {
			return &ApplicationError{
				Code:    code.RSAKeyLengthTooShort,
				Message: "RSA key length is too short. Must be at least 2048-bit.",
			}
		}
	case *ecdsa.PublicKey, ed25519.PublicKey, *dsa.PublicKey:
		return &ApplicationError{
			Code:    code.UnsupportedKeyType,
			Message: "Unsupported key type",
		}
	default:
		return &ApplicationError{
			Code:    code.UnknownKeyType,
			Message: "Unknown key type",
		}
	}

	return nil
}

func checkAccessorPubKey(key string) (err error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return &ApplicationError{
			Code:    code.InvalidKeyFormat,
			Message: "Invalid key format. Cannot decode PEM.",
		}
	}
	var pub interface{}
	if strings.Contains(key, "BEGIN RSA PUBLIC KEY") {
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
	} else {
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	}
	if err != nil {
		return &ApplicationError{
			Code:    code.InvalidKeyFormat,
			Message: err.Error(),
		}
	}

	switch pubKey := pub.(type) {
	case *rsa.PublicKey:
		if pubKey.N.BitLen() < 2048 {
			return &ApplicationError{
				Code:    code.RSAKeyLengthTooShort,
				Message: "RSA key length is too short. Must be at least 2048-bit.",
			}
		}
	case *ecdsa.PublicKey, ed25519.PublicKey, *dsa.PublicKey:
		return &ApplicationError{
			Code:    code.UnsupportedKeyType,
			Message: "Unsupported key type",
		}
	default:
		return &ApplicationError{
			Code:    code.UnknownKeyType,
			Message: "Unknown key type",
		}
	}

	return nil
}

func (app *ABCIApplication) checkCanCreateTx(committedState bool) error {
	value, err := app.state.Get(initStateKeyBytes, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if string(value) == "" {
		return &ApplicationError{
			Code:    code.ChainIsNotInitialized,
			Message: "Chain is not initialized",
		}
	}
	if string(value) != "false" {
		return &ApplicationError{
			Code:    code.ChainIsNotInitialized,
			Message: "Chain is not initialized",
		}
	}

	return nil
}

func (app *ABCIApplication) checkLastBlock(committedState bool) error {
	value, err := app.state.Get(lastBlockKeyBytes, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if string(value) == "" {
		value = []byte("-1")
	}
	if string(value) == "-1" {
		return nil
	}
	lastBlock, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return &ApplicationError{
			Code:    code.ChainIsDisabled,
			Message: "Chain is disabled",
		}
	}
	if app.state.CurrentBlockHeight > lastBlock {
		return &ApplicationError{
			Code:    code.ChainIsDisabled,
			Message: "Chain is disabled",
		}
	}

	return nil
}

func (app *ABCIApplication) commonValidate(method string, param []byte, nonce []byte, signature []byte, nodeID string, committedState bool) error {
	// ---- Check current block <= last block ----
	if method != "SetLastBlock" {
		err := app.checkLastBlock(committedState)
		if err != nil {
			return err
		}
	}

	if method == "SetInitData" || method == "SetInitData_pb" {
		return nil
	}

	// ---- Check is in init state ----
	if method != "InitNDID" && method != "EndInit" {
		err := app.checkCanCreateTx(committedState)
		if err != nil {
			return err
		}
	}

	// If method is not 'InitNDID' then check node is active
	if method != "InitNDID" {
		// Get node detail by NodeID
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
		nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), committedState)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		// If node not found then return code.NodeIDNotFound
		if nodeDetailValue == nil {
			return &ApplicationError{
				Code:    code.NodeIDNotFound,
				Message: "Node ID not found",
			}
		}
		var nodeDetail data.NodeDetail
		err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			return &ApplicationError{
				Code:    code.UnmarshalError,
				Message: err.Error(),
			}
		}

		if !nodeDetail.Active {
			return &ApplicationError{
				Code:    code.NodeIsNotActive,
				Message: "Node is not active",
			}
		}

		// If node behind proxy then check proxy is active
		if nodeDetail.ProxyNodeId != "" {
			proxyNodeID := nodeDetail.ProxyNodeId
			// Get proxy node detail
			proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
			proxyNodeDetailValue, err := app.state.Get([]byte(proxyNodeDetailKey), committedState)
			if err != nil {
				return &ApplicationError{
					Code:    code.AppStateError,
					Message: err.Error(),
				}
			}
			if proxyNodeDetailValue == nil {
				return &ApplicationError{
					Code:    code.ProxyNodeNotFound,
					Message: "Proxy node not found",
				}
			}
			var proxyNode data.NodeDetail
			err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
			if err != nil {
				return &ApplicationError{
					Code:    code.UnmarshalError,
					Message: err.Error(),
				}
			}

			if !proxyNode.Active {
				return &ApplicationError{
					Code:    code.ProxyNodeIsNotActive,
					Message: "Proxy node is not active",
				}
			}
		}

		ndidNode := appTypes.NodeRole(nodeDetail.Role) == appTypes.NodeRoleNdid
		// check if node has enough token to execute a function
		if !ndidNode {
			needToken := app.getTokenPriceByFunc(method, committedState)
			nodeToken, err := app.getToken(nodeID, committedState)
			if err != nil {
				return &ApplicationError{
					Code:    code.TokenAccountNotFound,
					Message: err.Error(),
				}
			}
			if nodeToken < needToken {
				return &ApplicationError{
					Code:    code.TokenNotEnough,
					Message: "not enough token",
				}
			}
		}
	}

	return nil
}

// CheckTxRouter check if Tx is valid
// CheckTx must get committed state while DeliverTx must get uncommitted state
func (app *ABCIApplication) CheckTxRouter(method string, param []byte, nonce []byte, signature []byte, nodeID string, committedState bool) types.ResponseCheckTx {
	err := app.commonValidate(method, param, nonce, signature, nodeID, committedState)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return app.callCheckTx(method, param, nodeID)
}

func (app *ABCIApplication) callCheckTx(name string, param []byte, nodeID string) types.ResponseCheckTx {
	switch name {
	case "InitNDID":
		return app.initNDIDCheckTx(param, nodeID)
	case "SetInitData":
		return app.setInitDataCheckTx(param, nodeID)
	case "SetInitData_pb":
		return app.setInitData_pbCheckTx(param, nodeID)
	case "EndInit":
		return app.endInitCheckTx(param, nodeID)
	case "SetLastBlock":
		return app.setLastBlockCheckTx(param, nodeID)
	case "SetValidator":
		return app.setValidatorCheckTx(param, nodeID)
	case "SetPriceFunc":
		return app.setPriceFuncCheckTx(param, nodeID)

	case "RegisterNode":
		return app.registerNodeCheckTx(param, nodeID)
	case "UpdateNodeByNDID":
		return app.updateNodeByNDIDCheckTx(param, nodeID)
	case "DisableNode":
		return app.disableNodeCheckTx(param, nodeID)
	case "EnableNode":
		return app.enableNodeCheckTx(param, nodeID)
	case "AddNodeToProxyNode":
		return app.addNodeToProxyNodeCheckTx(param, nodeID)
	case "UpdateNodeProxyNode":
		return app.updateNodeProxyNodeCheckTx(param, nodeID)
	case "RemoveNodeFromProxyNode":
		return app.removeNodeFromProxyNodeCheckTx(param, nodeID)
	case "AddNodeToken":
		return app.addNodeTokenCheckTx(param, nodeID)
	case "ReduceNodeToken":
		return app.reduceNodeTokenCheckTx(param, nodeID)
	case "SetNodeToken":
		return app.setNodeTokenCheckTx(param, nodeID)
	case "AddAllowedNodeSupportedFeature":
		return app.addAllowedNodeSupportedFeatureCheckTx(param, nodeID)
	case "RemoveAllowedNodeSupportedFeature":
		return app.removeAllowedNodeSupportedFeatureCheckTx(param, nodeID)

	case "AddNamespace":
		return app.addNamespaceCheckTx(param, nodeID)
	case "DisableNamespace":
		return app.disableNamespaceCheckTx(param, nodeID)
	case "EnableNamespace":
		return app.enableNamespaceCheckTx(param, nodeID)
	case "UpdateNamespace":
		return app.updateNamespaceCheckTx(param, nodeID)
	case "SetAllowedMinIalForRegisterIdentityAtFirstIdp":
		return app.setAllowedMinIalForRegisterIdentityAtFirstIdpCheckTx(param, nodeID)
	case "SetTimeOutBlockRegisterIdentity":
		return app.setTimeOutBlockRegisterIdentityCheckTx(param, nodeID)
	case "AddSuppressedIdentityModificationNotificationNode":
		return app.addSuppressedIdentityModificationNotificationNodeCheckTx(param, nodeID)
	case "RemoveSuppressedIdentityModificationNotificationNode":
		return app.removeSuppressedIdentityModificationNotificationNodeCheckTx(param, nodeID)

	case "AddService":
		return app.addServiceCheckTx(param, nodeID)
	case "DisableService":
		return app.disableServiceCheckTx(param, nodeID)
	case "EnableService":
		return app.enableServiceCheckTx(param, nodeID)
	case "UpdateService":
		return app.updateServiceCheckTx(param, nodeID)
	case "RegisterServiceDestinationByNDID":
		return app.registerServiceDestinationByNDIDCheckTx(param, nodeID)
	case "DisableServiceDestinationByNDID":
		return app.disableServiceDestinationByNDIDCheckTx(param, nodeID)
	case "EnableServiceDestinationByNDID":
		return app.enableServiceDestinationByNDIDCheckTx(param, nodeID)
	case "SetServicePriceCeiling":
		return app.setServicePriceCeilingCheckTx(param, nodeID)
	case "SetServicePriceMinEffectiveDatetimeDelay":
		return app.setServicePriceMinEffectiveDatetimeDelayCheckTx(param, nodeID)

	case "SetSupportedIALList":
		return app.setSupportedIALListCheckTx(param, nodeID)
	case "SetSupportedAALList":
		return app.setSupportedAALListCheckTx(param, nodeID)
	case "SetAllowedModeList":
		return app.setAllowedModeListCheckTx(param, nodeID)
	case "AddRequestType":
		return app.addRequestTypeCheckTx(param, nodeID)
	case "RemoveRequestType":
		return app.removeRequestTypeCheckTx(param, nodeID)

	case "AddErrorCode":
		return app.addErrorCodeCheckTx(param, nodeID)
	case "RemoveErrorCode":
		return app.removeErrorCodeCheckTx(param, nodeID)

	case "UpdateNode":
		return app.updateNodeCheckTx(param, nodeID)
	case "SetMqAddresses":
		return app.setMqAddressesCheckTx(param, nodeID)

	case "RegisterIdentity":
		return app.registerIdentityCheckTx(param, nodeID)
	case "UpdateIdentity":
		return app.updateIdentityCheckTx(param, nodeID)
	case "AddIdentity":
		return app.addIdentityCheckTx(param, nodeID)
	case "RevokeIdentityAssociation":
		return app.revokeIdentityAssociationCheckTx(param, nodeID)
	case "AddAccessor":
		return app.addAccessorCheckTx(param, nodeID)
	case "RevokeAccessor":
		return app.revokeAccessorCheckTx(param, nodeID)
	case "RevokeAndAddAccessor":
		return app.revokeAndAddAccessorCheckTx(param, nodeID)
	case "UpdateIdentityModeList":
		return app.updateIdentityModeListCheckTx(param, nodeID)

	case "RegisterServiceDestination":
		return app.registerServiceDestinationCheckTx(param, nodeID)
	case "UpdateServiceDestination":
		return app.updateServiceDestinationCheckTx(param, nodeID)
	case "DisableServiceDestination":
		return app.disableServiceDestinationCheckTx(param, nodeID)
	case "EnableServiceDestination":
		return app.enableServiceDestinationCheckTx(param, nodeID)
	case "SetServicePrice":
		return app.setServicePriceCheckTx(param, nodeID)

	case "CreateRequest":
		return app.createRequestCheckTx(param, nodeID)
	case "CreateIdpResponse":
		return app.createIdpResponseCheckTx(param, nodeID)
	case "CreateAsResponse":
		return app.createAsResponseCheckTx(param, nodeID)
	case "SetDataReceived":
		return app.setDataReceivedCheckTx(param, nodeID)
	case "CloseRequest":
		return app.closeRequestCheckTx(param, nodeID)
	case "TimeOutRequest":
		return app.timeOutRequestCheckTx(param, nodeID)

	case "CreateMessage":
		return app.createMessageCheckTx(param, nodeID)

	default:
		return types.ResponseCheckTx{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}

func (app *ABCIApplication) isDuplicateNonce(nonce []byte, committedState bool) bool {
	nonceKey := append([]byte(nonceKeyPrefix+keySeparator), nonce...)
	hasNonce, err := app.state.Has(nonceKey, committedState)
	if err != nil {
		panic(err)
	}

	return hasNonce
}
