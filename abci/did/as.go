package did

import (
	"encoding/json"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func signData(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("SignData")
	var signData SignDataParam
	err := json.Unmarshal([]byte(param), &signData)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "Request" + "|" + signData.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check IsClosed
	if request.IsClosed {
		return ReturnDeliverTxLog(code.RequestIsClosed, "Request is closed", "")
	}

	// Check IsTimedOut
	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.RequestIsTimedOut, "Request is timed out", "")
	}

	key = "SignData" + "|" + signData.Signature
	value, err = json.Marshal(signData)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))

	key = "Request" + "|" + signData.RequestID
	request.SignDataCount++
	value, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))

	return ReturnDeliverTxLog(code.OK, "success", signData.RequestID)
}

func registerServiceDestination(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("RegisterServiceDestination")
	var funcParam RegisterServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Add Service Detail
	key := "Service" + "|" + funcParam.AsServiceID + "|" + funcParam.NodeID
	chkExists := app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID", "")
	} else {
		var service = Service{
			funcParam.ServiceName,
			funcParam.MinIal,
			funcParam.MinAal,
		}
		value, err := json.Marshal(service)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	}

	// Add ServiceDestination
	key = "ServiceDestination" + "|" + funcParam.AsServiceID
	chkExists = app.state.db.Get(prefixKey([]byte(key)))

	if chkExists != nil {
		var nodes GetServiceDestinationResult
		err := json.Unmarshal([]byte(chkExists), &nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		nodes.NodeID = append(nodes.NodeID, funcParam.NodeID)
		value, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	} else {
		var nodes GetServiceDestinationResult
		nodes.NodeID = append(nodes.NodeID, funcParam.NodeID)
		value, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
	}
	return ReturnDeliverTxLog(code.OK, "success", "")
}
