package did

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/tendermint/abci/types"
)

func getNodePublicKey(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetNodePublicKey")
	var funcParam GetNodePublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "Node" + "|" + funcParam.NodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("[]")
		return ReturnQuery(value, "not found", app.state.Height)
	}

	var node RegisterNode
	err = json.Unmarshal(value, &node)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var res GetNodePublicKeyPesult
	res.PublicKey = string(value)
	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getMsqDestination(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetMsqDestination")
	var funcParam GetMsqDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "MsqDestination" + "|" + funcParam.HashID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("[]")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	var nodes []Node
	err = json.Unmarshal([]byte(value), &nodes)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var returnNodes GetMsqDestinationResult
	for _, node := range nodes {
		if node.Ial >= funcParam.MinIal {
			returnNodes.NodeID = append(returnNodes.NodeID, node.NodeID)
		}
	}

	value, err = json.Marshal(returnNodes)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getAccessorMethod(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetAccessorMethod")
	var funcParam GetAccessorMethodParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "AccessorMethod" + "|" + funcParam.AccessorID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	var accessorMethod AccessorMethod
	err = json.Unmarshal([]byte(value), &accessorMethod)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var res GetAccessorMethodResult
	res.AccessorType = accessorMethod.AccessorType
	res.AccessorKey = accessorMethod.AccessorKey
	res.Commitment = accessorMethod.Commitment
	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getRequest(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetRequest")
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}

	status := "pending"
	acceptCount := 0
	for _, response := range request.Responses {
		if response.Status == "accept" {
			acceptCount++
		} else if response.Status == "reject" {
			status = "rejected"
			break
		}
	}

	if acceptCount >= request.MinIdp {
		status = "completed"
	}

	var res GetRequestResult
	res.Status = status
	res.MessageHash = request.MessageHash

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return ReturnQuery(value, "success", app.state.Height)
}

func getRequestDetail(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetRequestDetail")
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getServiceDestination(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetServiceDestination")
	var funcParam GetServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "ServiceDestination" + "|" + funcParam.AsID + "|" + funcParam.AsServiceID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

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
