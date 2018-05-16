package did

import (
	"encoding/json"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

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
	chk := false
	for _, oldResponse := range request.Responses {
		if response == oldResponse {
			chk = true
			break
		}
	}

	// Check min_idp
	if len(request.Responses) >= request.MinIdp {
		return ReturnDeliverTxLog(code.CodeTypeError, "Can't response a request that's complete response", "")
	}

	// Check IsClosed
	if request.IsClosed {
		return ReturnDeliverTxLog(code.CodeTypeError, "Can't response a request that's closed", "")
	}

	// Check IsTimedOut
	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.CodeTypeError, "Can't response a request that's timed out", "")
	}

	if chk == false {
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
