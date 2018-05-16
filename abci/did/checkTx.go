package did

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"reflect"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func checkTxInitNDID(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "MasterNDID"
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value == nil {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsMember(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "RP" ||
		string(value) == "IdP" ||
		string(value) == "AS" ||
		string(value) == "MasterRP" ||
		string(value) == "MasterIdP" ||
		string(value) == "MasterAS" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkTxRegisterMsqAddress(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "RP" ||
		string(value) == "IdP" ||
		string(value) == "AS" ||
		string(value) == "MasterRP" ||
		string(value) == "MasterIdP" ||
		string(value) == "MasterAS" {

		var funcParam RegisterMsqAddressParam
		err := json.Unmarshal([]byte(param), &funcParam)
		if err != nil {
			return ReturnCheckTx(false)
		}
		publicKeyFromStateDB := getPublicKeyFromNodeID(funcParam.NodeID, app)
		if publicKeyFromStateDB == "" {
			return ReturnCheckTx(false)
		}
		if publicKeyFromStateDB == publicKey {
			return ReturnCheckTx(true)
		}
		return ReturnCheckTx(false)
	}
	return ReturnCheckTx(false)
}

func checkNDID(param string, publicKey string, app *DIDApplication) bool {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "NDID" || string(value) == "MasterNDID" {
		return true
	}
	return false
}

func checkIsNDID(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	return ReturnCheckTx(checkNDID(param, publicKey, app))
}

func checkIsIDP(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "IdP" || string(value) == "MasterIdP" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsRP(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "RP" || string(value) == "MasterRP" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsAS(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "AS" || string(value) == "MasterAS" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsOwnerRequest(param string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	var funcParam RequestIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnCheckTx(false)
	}

	key := "SpendGas" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	var reports []Report
	err = json.Unmarshal([]byte(value), &reports)
	if err != nil {
		return ReturnCheckTx(false)
	}

	for _, node := range reports {
		if node.Method == "CreateRequest" &&
			node.Data == funcParam.RequestID {
			return ReturnCheckTx(true)
		}
	}

	return ReturnCheckTx(false)
}

func verifySignature(param string, nonce string, signature string, publicKey string) (result bool, err error) {
	publicKey = strings.Replace(publicKey, "\t", "", -1)
	block, _ := pem.Decode([]byte(publicKey))
	senderPublicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	senderPublicKey := senderPublicKeyInterface.(*rsa.PublicKey)
	if err != nil {
		return false, err
	}
	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	PSSmessage := []byte(param + nonce)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	err = rsa.VerifyPKCS1v15(senderPublicKey, newhash, hashed, decodedSignature)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ReturnCheckTx return types.ResponseDeliverTx
func ReturnCheckTx(ok bool) types.ResponseCheckTx {
	if ok {
		return types.ResponseCheckTx{Code: code.CodeTypeOK}
	}
	return types.ResponseCheckTx{Code: code.CodeTypeUnauthorized}
}

func getPublicKeyInitNDID(param string) string {
	var funcParam InitNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ""
	}
	return funcParam.PublicKey
}

func getPublicKeyFromNodeID(nodeID string, app *DIDApplication) string {
	key := "NodeID" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		return string(value)
	}
	return ""
}

var IsCheckOwnerRequestMethod = map[string]bool{
	"CloseRequest":   true,
	"TimeOutRequest": true,
}

// CheckTxRouter is Pointer to function
func CheckTxRouter(method string, param string, nonce string, signature string, nodeID string, app *DIDApplication) types.ResponseCheckTx {
	funcs := map[string]interface{}{
		"InitNDID":                   checkTxInitNDID,
		"RegisterNode":               checkIsNDID,
		"RegisterMsqDestination":     checkIsIDP,
		"AddAccessorMethod":          checkIsIDP,
		"CreateIdpResponse":          checkIsIDP,
		"SignData":                   checkIsAS,
		"RegisterServiceDestination": checkIsAS,
		"CreateRequest":              checkIsRP,
		"RegisterMsqAddress":         checkTxRegisterMsqAddress,
		"AddNodeToken":               checkIsNDID,
		"ReduceNodeToken":            checkIsNDID,
		"SetNodeToken":               checkIsNDID,
		"SetPriceFunc":               checkIsNDID,
	}

	var publicKey string
	if method == "InitNDID" {
		publicKey = getPublicKeyInitNDID(param)
		if publicKey == "" {
			return ReturnCheckTx(false)
		}
	} else {
		publicKey = getPublicKeyFromNodeID(nodeID, app)
		if publicKey == "" {
			return ReturnCheckTx(false)
		}
	}

	verifyResult, err := verifySignature(param, nonce, signature, publicKey)
	if err != nil || verifyResult == false {
		return ReturnCheckTx(false)
	}

	var result types.ResponseCheckTx

	// special case checkIsOwnerRequest
	if IsCheckOwnerRequestMethod[method] {
		result = checkIsOwnerRequest(param, nodeID, app)
	} else {
		value, _ := callCheckTx(funcs, method, param, publicKey, app)
		result = value[0].Interface().(types.ResponseCheckTx)
	}
	// check token for create Tx
	if result.Code == code.CodeTypeOK {
		if !checkNDID(nodeID, publicKey, app) && method != "InitNDID" {
			needToken := getTokenPriceByFunc(method, app)
			nodeToken, err := getToken(nodeID, app)
			if err != nil {
				result.Code = code.CodeTypeUnauthorized
			}
			if nodeToken < needToken {
				result.Code = code.CodeTypeUnauthorized
				result.Log = "token not enough"
			}
		}
	}
	return result
}

func callCheckTx(m map[string]interface{}, name string, param string, publicKey string, app *DIDApplication) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 3)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(publicKey)
	in[2] = reflect.ValueOf(app)
	result = f.Call(in)
	return
}
