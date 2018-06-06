package did

import (
	"encoding/json"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func createRequest(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("CreateRequest")
	var request Request
	err := json.Unmarshal([]byte(param), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// set default value
	request.IsClosed = false
	request.IsTimedOut = false
	request.CanAddAccessor = false

	// set Owner
	request.Owner = nodeID

	// set Can add accossor
	ownerRole := getRoleFromNodeID(nodeID, app)
	if string(ownerRole) == "IdP" || string(ownerRole) == "MasterIdP" {
		request.CanAddAccessor = true
	}

	key := "Request" + "|" + request.RequestID
	value, err := json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	existValue := app.state.db.Get(prefixKey([]byte(key)))
	if existValue != nil {
		return ReturnDeliverTxLog(code.DuplicateRequestID, "Duplicate Request ID", "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", request.RequestID)
}

func closeRequest(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("CloseRequest")
	var funcParam RequestIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.RequestIsTimedOut, "Can not close a timed out request", "")
	}

	request.IsClosed = true
	value, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}

func timeOutRequest(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("TimeOutRequest")
	var funcParam RequestIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	if request.IsClosed {
		return ReturnDeliverTxLog(code.RequestIsClosed, "Can not set time out a closed request", "")
	}

	request.IsTimedOut = true
	value, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
}
