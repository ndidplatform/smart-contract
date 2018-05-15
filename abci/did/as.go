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
