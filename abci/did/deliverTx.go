package did

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

// TODO: unit testing
func addNodePublicKey(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("AddNodePublicKey")
	var nodePublicKey NodePublicKey
	err := json.Unmarshal([]byte(param), &nodePublicKey)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "NodePublicKey" + "|" + nodePublicKey.NodeID
	value := nodePublicKey.PublicKey
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
}

func registerMsqDestination(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("RegisterMsqDestination")
	var funcParam RegisterMsqDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	for _, user := range funcParam.Users {
		key := "MsqDestination" + "|" + user.HashID
		chkExists := app.state.db.Get(prefixKey([]byte(key)))

		if chkExists != nil {
			var nodes []Node
			err = json.Unmarshal([]byte(chkExists), &nodes)
			if err != nil {
				return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
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
					return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
				}
				app.state.Size++
				app.state.db.Set(prefixKey([]byte(key)), []byte(value))
			}

		} else {
			var nodes []Node
			newNode := Node{user.Ial, funcParam.NodeID}
			nodes = append(nodes, newNode)
			value, err := json.Marshal(nodes)
			if err != nil {
				return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
			}
			app.state.Size++
			app.state.db.Set(prefixKey([]byte(key)), []byte(value))
		}
	}

	return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
}

func addAccessorMethod(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("AddAccessorMethod")
	var accessorMethod AccessorMethod
	err := json.Unmarshal([]byte(param), &accessorMethod)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "AccessorMethod" + "|" + accessorMethod.AccessorID
	value, err := json.Marshal(accessorMethod)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
}

func createRequest(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("CreateRequest")
	var request Request
	err := json.Unmarshal([]byte(param), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "Request" + "|" + request.RequestID
	value, err := json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	existValue := app.state.db.Get(prefixKey([]byte(key)))
	if existValue != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, "Duplicate Request ID", "")
	}

	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))

	return ReturnDeliverTxLog(code.CodeTypeOK, "success", request.RequestID)
}

func createIdpResponse(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("CreateIdpResponse")
	var response Response
	err := json.Unmarshal([]byte(param), &response)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "Request" + "|" + response.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.CodeTypeError, "Request ID not found", "")
	}
	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
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
			return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
		}
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))

		return ReturnDeliverTxLog(code.CodeTypeOK, "success", response.RequestID)
	}
	return ReturnDeliverTxLog(code.CodeTypeError, "Duplicate Response", "")
}

func signData(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("SignData")
	var signData SignDataParam
	err := json.Unmarshal([]byte(param), &signData)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "SignData" + "|" + signData.Signature
	value, err := json.Marshal(signData)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	return ReturnDeliverTxLog(code.CodeTypeOK, "success", signData.RequestID)
}

func registerServiceDestination(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("RegisterServiceDestination")
	var funcParam RegisterServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "ServiceDestination" + "|" + funcParam.AsServiceID
	chkExists := app.state.db.Get(prefixKey([]byte(key)))

	if chkExists != nil {
		var nodes GetServiceDestinationResult
		err := json.Unmarshal([]byte(chkExists), &nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
		}
		nodes.NodeID = append(nodes.NodeID, funcParam.NodeID)
		value, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
		}
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	} else {
		var nodes GetServiceDestinationResult
		nodes.NodeID = append(nodes.NodeID, funcParam.NodeID)
		value, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
		}
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	}
	return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
}

func initNDID(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("InitNDID")
	var funcParam InitNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}
	key := "NodePublicKeyRole" + "|" + funcParam.PublicKey
	value := []byte("MasterNDID")
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	key = "NodeID" + "|" + funcParam.NodeID
	value = []byte(funcParam.PublicKey)
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	key = "MasterNDID"
	value = []byte(funcParam.PublicKey)
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
}

func registerMsqAddress(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("RegisterMsqAddress")
	var funcParam RegisterMsqAddressParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}
	key := "MsqAddress" + "|" + funcParam.NodeID
	var msqAddress = MsqAddress{
		funcParam.IP,
		funcParam.Port,
	}
	value, err := json.Marshal(msqAddress)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
}

func registerNode(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("RegisterNode")
	var funcParam RegisterNode
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "NodeID" + "|" + funcParam.NodeID
	chkExists := app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, "Duplicate Node ID", "")
	}

	if funcParam.Role == "RP" ||
		funcParam.Role == "IdP" ||
		funcParam.Role == "AS" {
		key := "NodeID" + "|" + funcParam.NodeID
		value := funcParam.PublicKey
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
		key = "NodePublicKeyRole" + "|" + funcParam.PublicKey
		value = "Master" + funcParam.Role
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
		createTokenAccount(funcParam.NodeID, app)
		return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
	}
	return ReturnDeliverTxLog(code.CodeTypeError, "wrong role", "")
}

// ReturnDeliverTxLog return types.ResponseDeliverTx
func ReturnDeliverTxLog(code uint32, log string, extraData string) types.ResponseDeliverTx {
	return types.ResponseDeliverTx{
		Code: code,
		Log:  fmt.Sprintf(log),
		Data: []byte(extraData),
	}
}

var isNDIDMethod = map[string]bool{
	"InitNDID":        true,
	"RegisterNode":    true,
	"AddNodeToken":    true,
	"ReduceNodeToken": true,
	"SetNodeToken":    true,
	"SetPriceFunc":    true,
}

// DeliverTxRouter is Pointer to function
func DeliverTxRouter(method string, param string, nodeID string, app *DIDApplication) types.ResponseDeliverTx {
	funcs := map[string]interface{}{
		"InitNDID":                   initNDID,
		"RegisterNode":               registerNode,
		"RegisterMsqDestination":     registerMsqDestination,
		"AddAccessorMethod":          addAccessorMethod,
		"CreateRequest":              createRequest,
		"CreateIdpResponse":          createIdpResponse,
		"SignData":                   signData,
		"RegisterServiceDestination": registerServiceDestination,
		"RegisterMsqAddress":         registerMsqAddress,
		"AddNodeToken":               addNodeToken,
		"ReduceNodeToken":            reduceNodeToken,
		"SetNodeToken":               setNodeToken,
		"SetPriceFunc":               setPriceFunc,
	}
	value, _ := callDeliverTx(funcs, method, param, app)
	result := value[0].Interface().(types.ResponseDeliverTx)
	// ---- Burn token ----
	if result.Code == code.CodeTypeOK {
		if !isNDIDMethod[method] {
			needToken := getTokenPriceByFunc(method, app)
			err := reduceToken(nodeID, needToken, app)
			if err != nil {
				result.Code = code.CodeTypeError
				result.Log = err.Error()
				return result
			}
			// Write burn token report
			// only have result.Data in some method
			writeBurnTokenReport(nodeID, method, needToken, string(result.Data), app)
		}
	}
	return result
}

func callDeliverTx(m map[string]interface{}, name string, param string, app *DIDApplication) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(app)
	result = f.Call(in)
	return
}
