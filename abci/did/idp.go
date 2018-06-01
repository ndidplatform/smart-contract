package did

import (
	"encoding/json"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func createIdentity(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("CreateIdentity")
	var funcParam CreateIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	var accessor = Accessor{
		funcParam.AccessorType,
		funcParam.AccessorPublicKey,
		funcParam.AccessorGroupID,
	}

	accessorJSON, err := json.Marshal(accessor)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	accessorGroupKey := "AccessorGroup" + "|" + funcParam.AccessorGroupID
	accessorGroup := funcParam.AccessorGroupID

	// Check duplicate accessor_id
	chkAccessorKeyExists := app.state.db.Get(prefixKey([]byte(accessorKey)))
	if chkAccessorKeyExists != nil {
		return ReturnDeliverTxLog(code.DuplicateAccessorID, "Duplicate Accessor ID", "")
	}

	// Check duplicate accessor_group_id
	chkAccessorGroupKeyExists := app.state.db.Get(prefixKey([]byte(accessorGroupKey)))
	if chkAccessorGroupKeyExists != nil {
		return ReturnDeliverTxLog(code.DuplicateAccessorGroupID, "Duplicate Accessor Group ID", "")
	}

	app.SetStateDB([]byte(accessorKey), []byte(accessorJSON))
	app.SetStateDB([]byte(accessorGroupKey), []byte(accessorGroup))

	return ReturnDeliverTxLog(code.OK, "success", "")
}

func setCanAddAccessorToFalse(requestID string, app *DIDApplication) {
	key := "Request" + "|" + requestID
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var request Request
		err := json.Unmarshal([]byte(value), &request)
		if err == nil {
			request.CanAddAccessor = false
			value, err := json.Marshal(request)
			if err == nil {
				app.SetStateDB([]byte(key), []byte(value))
			}
		}
	}
}

func addAccessorMethod(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("AddAccessorMethod")
	var funcParam AccessorMethod
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// AccessorGroupID: must already exist
	accessorGroupKey := "AccessorGroup" + "|" + funcParam.AccessorGroupID
	chkAccessorGroupKeyExists := app.state.db.Get(prefixKey([]byte(accessorGroupKey)))
	if chkAccessorGroupKeyExists == nil {
		return ReturnDeliverTxLog(code.AccessorGroupIDNotFound, "Accessor Group ID not found", "")
	}

	// AccessorID: must not duplicate
	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	chkAccessorKeyExists := app.state.db.Get(prefixKey([]byte(accessorKey)))
	if chkAccessorKeyExists != nil {
		return ReturnDeliverTxLog(code.DuplicateAccessorID, "Duplicate Accessor ID", "")
	}

	// Request must be completed, can be used only once, special type
	var getRequestparam GetRequestParam
	getRequestparam.RequestID = funcParam.RequestID
	getRequestparamJSON, err := json.Marshal(getRequestparam)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	var request = getRequest(string(getRequestparamJSON), app)
	var requestResult GetRequestResult
	err = json.Unmarshal([]byte(request.Value), &requestResult)
	if err != nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	if requestResult.Status != "completed" {
		return ReturnDeliverTxLog(code.RequestIsNotCompleted, "Request is not completed", "")
	}
	// check special type of Request && set can used only once
	canAddAccessor := getCanAddAccessor(funcParam.RequestID, app)
	if canAddAccessor != true {
		return ReturnDeliverTxLog(code.RequestIsNotSpecial, "Request is not special", "")
	}
	setCanAddAccessorToFalse(funcParam.RequestID, app)

	var accessor = Accessor{
		funcParam.AccessorType,
		funcParam.AccessorPublicKey,
		funcParam.AccessorGroupID,
	}

	accessorJSON, err := json.Marshal(accessor)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(accessorKey), []byte(accessorJSON))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func registerMsqDestination(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("RegisterMsqDestination")
	var funcParam RegisterMsqDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	maxIalAalKey := "MaxIalAalNode" + "|" + funcParam.NodeID
	maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
	if maxIalAalValue != nil {
		var maxIalAal MaxIalAal
		err := json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// Validate user's ial is <= node's max_ial
		for _, user := range funcParam.Users {
			if user.Ial > maxIalAal.MaxIal {
				return ReturnDeliverTxLog(code.IALError, "IAL must be less than or equals to registered node's MAX IAL", "")
			}
		}
	}

	// If validate passed then add Msq Destination
	for _, user := range funcParam.Users {
		key := "MsqDestination" + "|" + user.HashID
		chkExists := app.state.db.Get(prefixKey([]byte(key)))

		if chkExists != nil {
			var nodes []Node
			err = json.Unmarshal([]byte(chkExists), &nodes)
			if err != nil {
				return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
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
					return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
				}
				app.SetStateDB([]byte(key), []byte(value))
			}

		} else {
			var nodes []Node
			newNode := Node{user.Ial, funcParam.NodeID}
			nodes = append(nodes, newNode)
			value, err := json.Marshal(nodes)
			if err != nil {
				return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
			}
			app.SetStateDB([]byte(key), []byte(value))
		}
	}

	return ReturnDeliverTxLog(code.OK, "success", "")
}

func createIdpResponse(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("CreateIdpResponse")
	var response Response
	err := json.Unmarshal([]byte(param), &response)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "Request" + "|" + response.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check duplicate before add
	chk := false
	for _, oldResponse := range request.Responses {
		if response == oldResponse {
			chk = true
			break
		}
	}

	// Check AAL
	if request.MinAal > response.Aal {
		return ReturnDeliverTxLog(code.AALError, "Response's AAL is less than min AAL", "")
	}

	// Check IAL
	if request.MinIal > response.Ial {
		return ReturnDeliverTxLog(code.IALError, "Response's IAL is less than min IAL", "")
	}

	// Check min_idp
	if len(request.Responses) >= request.MinIdp {
		return ReturnDeliverTxLog(code.RequestIsCompleted, "Can't response a request that's complete response", "")
	}

	// Check IsClosed
	if request.IsClosed {
		return ReturnDeliverTxLog(code.RequestIsClosed, "Can't response a request that's closed", "")
	}

	// Check IsTimedOut
	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.RequestIsTimedOut, "Can't response a request that's timed out", "")
	}

	if chk == false {
		request.Responses = append(request.Responses, response)
		value, err := json.Marshal(request)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
		return ReturnDeliverTxLog(code.OK, "success", response.RequestID)
	}
	return ReturnDeliverTxLog(code.DuplicateResponse, "Duplicate Response", "")
}
