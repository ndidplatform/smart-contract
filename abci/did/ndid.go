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
	"AddNamespace":    true,
	"DeleteNamespace": true,
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
	app.SetStateDB([]byte(key), []byte(value))
	key = "NodeID" + "|" + funcParam.NodeID
	value = []byte(funcParam.PublicKey)
	app.SetStateDB([]byte(key), []byte(value))
	key = "MasterNDID"
	value = []byte(funcParam.PublicKey)
	app.SetStateDB([]byte(key), []byte(value))
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
		app.SetStateDB([]byte(key), []byte(value))
		key = "NodePublicKeyRole" + "|" + funcParam.PublicKey
		value = "Master" + funcParam.Role
		app.SetStateDB([]byte(key), []byte(value))
		createTokenAccount(funcParam.NodeID, app)

		// Add max_aal, min_ial when node is IdP
		if funcParam.Role == "IdP" {
			maxIalAalKey := "MaxIalAalNode" + "|" + funcParam.NodeID
			var maxIalAal MaxIalAal
			maxIalAal.MaxAal = funcParam.MaxAal
			maxIalAal.MaxIal = funcParam.MaxIal
			maxIalAalValue, err := json.Marshal(maxIalAal)
			if err != nil {
				return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
			}
			app.SetStateDB([]byte(maxIalAalKey), []byte(maxIalAalValue))
		}

		return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
	}
	return ReturnDeliverTxLog(code.CodeTypeError, "wrong role", "")
}

func addNamespace(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("AddNamespace")
	var funcParam Namespace
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "AllNamespace"
	chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces []Namespace

	if chkExists != nil {
		err = json.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
		}

		// Check duplicate namespace
		for _, namespace := range namespaces {
			if namespace.Namespace == funcParam.Namespace {
				return ReturnDeliverTxLog(code.CodeTypeError, "Duplicate namespace", "")
			}
		}
	}
	namespaces = append(namespaces, funcParam)
	value, err := json.Marshal(namespaces)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
}

func deleteNamespace(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("DeleteNamespace")
	var funcParam DeleteNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}

	key := "AllNamespace"
	chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces []Namespace

	if chkExists != nil {
		err = json.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
		}

		for index, namespace := range namespaces {
			if namespace.Namespace == funcParam.Namespace {
				namespaces = append(namespaces[:index], namespaces[index+1:]...)
				break
			}
		}

		value, err := json.Marshal(namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
		return ReturnDeliverTxLog(code.CodeTypeOK, "success", "")
	}

	return ReturnDeliverTxLog(code.CodeTypeOK, "Not found namespace", "")
}
