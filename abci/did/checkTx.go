package did

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"reflect"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func checkTxInitNDID(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	if app.state.Owner == nil {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsNDID(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	if app.state.Owner != nil {
		owner := string(app.state.Owner)
		if owner == publicKey {
			return ReturnCheckTx(true)
		}
	}
	return ReturnCheckTx(false)
}

func checkIsIDP(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "IDP" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsRP(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "RP" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func checkIsAS(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	key := "NodePublicKeyRole" + "|" + publicKey
	value := app.state.db.Get(prefixKey([]byte(key)))
	if string(value) == "AS" {
		return ReturnCheckTx(true)
	}
	return ReturnCheckTx(false)
}

func verifySignature(param string, nonce string, signature string, publicKey string) (result bool, err error) {
	publicKey = strings.Replace(publicKey, "\t", "", -1)
	block, _ := pem.Decode([]byte(publicKey))
	senderPublicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
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

// CheckTxRouter is Pointer to function
func CheckTxRouter(method string, param string, nonce string, signature string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	funcs := map[string]interface{}{
		"InitNDID":                   checkTxInitNDID,
		"TransferNDID":               checkIsNDID,
		"RegisterNode":               checkIsNDID,
		"RegisterMsqDestination":     checkIsIDP,
		"AddAccessorMethod":          checkIsIDP,
		"CreateIdpResponse":          checkIsIDP,
		"SignData":                   checkIsAS,
		"RegisterServiceDestination": checkIsAS,
		"CreateRequest":              checkIsRP,
	}
	verifyResult, err := verifySignature(param, nonce, signature, publicKey)
	if err != nil || verifyResult == false {
		return ReturnCheckTx(false)
	}

	value, _ := callCheckTx(funcs, method, param, publicKey, app)
	return value[0].Interface().(types.ResponseCheckTx)
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
