package did

import (
	"fmt"
	"reflect"

	"github.com/tendermint/abci/types"
)

// ReturnQuery return types.ResponseQuery
func ReturnQuery(value []byte, log string, height int64) types.ResponseQuery {
	fmt.Println(string(value))
	var res types.ResponseQuery
	res.Value = value
	res.Log = log
	res.Height = height
	return res
}

// QueryRouter is Pointer to function
func QueryRouter(method string, param string, app *DIDApplication) types.ResponseQuery {
	funcs := map[string]interface{}{
		"GetNodePublicKey":      getNodePublicKey,
		"GetMsqDestination":     getMsqDestination,
		"GetAccessorMethod":     getAccessorMethod,
		"GetRequest":            getRequest,
		"GetRequestDetail":      getRequestDetail,
		"GetServiceDestination": getServiceDestination,
		"GetMsqAddress":         getMsqAddress,
		"GetNodeToken":          getNodeToken,
		"GetPriceFunc":          getPriceFunc,
		"GetUsedTokenReport":    getUsedTokenReport,
	}
	value, _ := callQuery(funcs, method, param, app)
	return value[0].Interface().(types.ResponseQuery)
}

func callQuery(m map[string]interface{}, name string, param string, app *DIDApplication) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(app)
	result = f.Call(in)
	return
}
