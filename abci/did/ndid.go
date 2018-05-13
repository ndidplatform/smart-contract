package did

import (
	"encoding/json"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

var isNDIDMethod = map[string]bool{
	"InitNDID":        true,
	"RegisterNode":    true,
	"AddNodeToken":    true,
	"ReduceNodeToken": true,
	"SetNodeToken":    true,
	"SetPriceFunc":    true,
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
