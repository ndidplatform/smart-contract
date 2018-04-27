package did

import (
	"fmt"
	"reflect"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

func checkTxInitNDID(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	if app.state.Owner == nil {
		return ReturnCheckTx(true)
	} else {
		return ReturnCheckTx(false)
	}
}

func checkTxTransferNDID(param string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	fmt.Println(param)
	return ReturnCheckTx(false)
}

func verifySignature(param string, signature string, publicKey string) (result bool, err error) {
	signatureJSON := []byte(`{"type":"ed25519","data":"` + signature + `"}`)
	infSignature, err := crypto.SignatureMapper.FromJSON(signatureJSON)
	if err != nil {
		return false, err
	}
	objSignature := infSignature.(crypto.SignatureEd25519)

	publicKeyJSON := []byte(`{"type":"ed25519","data":"` + publicKey + `"}`)
	infPublicKey, err := crypto.PubKeyMapper.FromJSON(publicKeyJSON)
	if err != nil {
		return false, err
	}
	objPublicKey := infPublicKey.(crypto.PubKeyEd25519)
	verifyResult := objPublicKey.VerifyBytes([]byte(param), objSignature.Wrap())
	return verifyResult, nil
}

// ReturnCheckTx return types.ResponseDeliverTx
func ReturnCheckTx(ok bool) types.ResponseCheckTx {
	if ok {
		return types.ResponseCheckTx{Code: code.CodeTypeOK}
	} else {
		return types.ResponseCheckTx{Code: code.CodeTypeUnauthorized}
	}
}

// CheckTxRouter is Pointer to function
func CheckTxRouter(method string, param string, signature string, publicKey string, app *DIDApplication) types.ResponseCheckTx {
	funcs := map[string]interface{}{
		"InitNDID":                   checkTxInitNDID,
		"TransferNDID":               checkTxTransferNDID,
		"AddNodePublicKey":           addNodePublicKey,
		"RegisterMsqDestination":     registerMsqDestination,
		"AddAccessorMethod":          addAccessorMethod,
		"CreateRequest":              createRequest,
		"CreateIdpResponse":          createIdpResponse,
		"SignData":                   signData,
		"RegisterServiceDestination": registerServiceDestination,
	}
	verifyResult, err := verifySignature(param, signature, publicKey)
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
