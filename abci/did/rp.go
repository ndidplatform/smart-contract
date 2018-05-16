package did

import (
	"encoding/json"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func createRequest(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("CreateRequest")
	var request Request
	err := json.Unmarshal([]byte(param), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	// set default value
	request.IsClosed = false
	request.IsTimedOut = false

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

func closeRequest(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("CloseRequest")
	var funcParam RequestIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.CodeTypeError, "Request ID not found", "")
	}

	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.CodeTypeError, "Can not close a timed out request", "")
	}

	request.IsClosed = true
	value, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))

	return ReturnDeliverTxLog(code.CodeTypeOK, "success", funcParam.RequestID)
}

func timeOutRequest(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("TimeOutRequest")
	var funcParam RequestIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.CodeTypeError, "Request ID not found", "")
	}

	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	if request.IsClosed {
		return ReturnDeliverTxLog(code.CodeTypeError, "Can not set time out a closed request", "")
	}

	request.IsTimedOut = true
	value, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))

	return ReturnDeliverTxLog(code.CodeTypeOK, "success", funcParam.RequestID)
}
