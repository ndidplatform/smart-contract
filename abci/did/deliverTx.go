package did

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/tendermint/abci/example/code"
	"github.com/tendermint/abci/types"
)

func AddNodePublicKey(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("AddNodePublicKey")
	var nodePublicKey NodePublicKey
	err := json.Unmarshal([]byte(param), &nodePublicKey)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}

	key := "NodePublicKey" + "|" + nodePublicKey.NodeID
	value := nodePublicKey.PublicKey
	app.state.Size += 1
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	return ReturnDeliverTxLog("success")
}

func RegisterMsqDestination(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("RegisterMsqDestination")
	var funcParam RegisterMsqDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}

	for _, user := range funcParam.Users {
		key := "MsqDestination" + "|" + user.HashID
		chkExists := app.state.db.Get(prefixKey([]byte(key)))

		if chkExists != nil {
			var nodes []Node
			err = json.Unmarshal([]byte(chkExists), &nodes)
			if err != nil {
				return ReturnDeliverTxLog(err.Error())
			}

			newNode := Node{user.Ial, funcParam.NodeID}
			// Check duplicate before add
			chkDup := false
			for _, node := range nodes {
				if newNode == node {
					chkDup = true
					break
				}
			}

			if chkDup == false {
				nodes = append(nodes, newNode)
				value, err := json.Marshal(nodes)
				if err != nil {
					return ReturnDeliverTxLog(err.Error())
				}
				app.state.Size += 1
				app.state.db.Set(prefixKey([]byte(key)), []byte(value))
			}

		} else {
			var nodes []Node
			newNode := Node{user.Ial, funcParam.NodeID}
			nodes = append(nodes, newNode)
			value, err := json.Marshal(nodes)
			if err != nil {
				return ReturnDeliverTxLog(err.Error())
			}
			app.state.Size += 1
			app.state.db.Set(prefixKey([]byte(key)), []byte(value))
		}
	}

	return ReturnDeliverTxLog("success")
}

func AddAccessorMethod(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("AddAccessorMethod")
	var accessorMethod AccessorMethod
	err := json.Unmarshal([]byte(param), &accessorMethod)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}

	key := "AccessorMethod" + "|" + accessorMethod.AccessorID
	value, err := json.Marshal(accessorMethod)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	app.state.Size += 1
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	return ReturnDeliverTxLog("success")
}

func CreateRequest(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("CreateRequest")
	var request Request
	err := json.Unmarshal([]byte(param), &request)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}

	key := "Request" + "|" + request.RequestID
	value, err := json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	app.state.Size += 1
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	return ReturnDeliverTxLog("success")
}

func CreateIdpResponse(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("CreateIdpResponse")
	var response Response
	err := json.Unmarshal([]byte(param), &response)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}

	key := "Request" + "|" + response.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog("Request ID not found")
	} else {
		var request Request
		err := json.Unmarshal([]byte(value), &request)
		if err != nil {
			return ReturnDeliverTxLog(err.Error())
		}

		// Check duplicate before add
		chkDup := false
		for _, oldResponse := range request.Responses {
			if response == oldResponse {
				chkDup = true
				break
			}
		}

		if chkDup == false {
			request.Responses = append(request.Responses, response)
			value, err := json.Marshal(request)
			if err != nil {
				return ReturnDeliverTxLog(err.Error())
			}
			app.state.Size += 1
			app.state.db.Set(prefixKey([]byte(key)), []byte(value))

			// callback to RP
			uri := getEnv("CALLBACK_URI", "")
			if uri != "" {
				fmt.Println("CALLBACK_URI:" + uri)

				var callback Callback
				callback.RequestID = request.RequestID
				data, err := json.Marshal(callback)
				if err != nil {
					fmt.Println("error:", err)
					return ReturnDeliverTxLog(err.Error())
				}

				client := &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				}

				req, err := http.NewRequest("POST", uri, strings.NewReader(string(data)))
				if err != nil {
					return ReturnDeliverTxLog(err.Error())
				}
				req.Header.Set("Content-Type", "application/json")
				resp, _ := client.Do(req)
				fmt.Println(resp.Status)
			}

			return ReturnDeliverTxLog("success")
		} else {
			return ReturnDeliverTxLog("Response duplicate")
		}

	}
}

func ReturnDeliverTxLog(log string) types.ResponseDeliverTx {
	return types.ResponseDeliverTx{
		Code: code.CodeTypeOK,
		Log:  fmt.Sprintf(log)}
}

// Pointer to function
func DeliverTxRouter(method string, param string, app *DIDApplication) types.ResponseDeliverTx {
	funcs := map[string]interface{}{
		"AddNodePublicKey":       AddNodePublicKey,
		"RegisterMsqDestination": RegisterMsqDestination,
		"AddAccessorMethod":      AddAccessorMethod,
		"CreateRequest":          CreateRequest,
		"CreateIdpResponse":       CreateIdpResponse,
	}
	value, _ := CallDeliverTx(funcs, method, param, app)
	return value[0].Interface().(types.ResponseDeliverTx)
}

func CallDeliverTx(m map[string]interface{}, name string, param string, app *DIDApplication) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(app)
	result = f.Call(in)
	return
}
