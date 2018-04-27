package did

import (
	"reflect"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func checkTxInitNDID(param string, app *DIDApplication) types.ResponseCheckTx {
	if app.state.Owner == nil {
		return ReturnCheckTx(true)
	} else {
		return ReturnCheckTx(false)
	}
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
func CheckTxRouter(method string, param string, app *DIDApplication) types.ResponseCheckTx {
	funcs := map[string]interface{}{
		"InitNDID":                   checkTxInitNDID,
		"AddNodePublicKey":           addNodePublicKey,
		"RegisterMsqDestination":     registerMsqDestination,
		"AddAccessorMethod":          addAccessorMethod,
		"CreateRequest":              createRequest,
		"CreateIdpResponse":          createIdpResponse,
		"SignData":                   signData,
		"RegisterServiceDestination": registerServiceDestination,
	}
	value, _ := callCheckTx(funcs, method, param, app)
	return value[0].Interface().(types.ResponseCheckTx)
}

func callCheckTx(m map[string]interface{}, name string, param string, app *DIDApplication) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(app)
	result = f.Call(in)
	return
}
