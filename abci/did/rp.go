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
