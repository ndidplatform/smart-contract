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
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
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
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func registerNode(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("RegisterNode")
	var funcParam RegisterNode
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "NodeID" + "|" + funcParam.NodeID
	chkExists := app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicateNodeID, "Duplicate Node ID", "")
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
			fmt.Println("IDPIDP =====")
			maxIalAalKey := "MaxIalAalNode" + "|" + funcParam.NodeID
			var maxIalAal MaxIalAal
			maxIalAal.MaxAal = funcParam.MaxAal
			maxIalAal.MaxIal = funcParam.MaxIal
			maxIalAalValue, err := json.Marshal(maxIalAal)
			if err != nil {
				return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
			}
			app.SetStateDB([]byte(maxIalAalKey), []byte(maxIalAalValue))

			// Save all IdP's nodeID for GetMsqDestination
			idpsKey := "IdPList"
			idpsValue := app.state.db.Get(prefixKey([]byte(idpsKey)))
			var idpsList []string
			if idpsValue != nil {
				err := json.Unmarshal([]byte(idpsValue), &idpsList)
				if err != nil {
					return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
				}
			}
			idpsList = append(idpsList, funcParam.NodeID)
			idpsValue, err = json.Marshal(idpsList)
			if err != nil {
				return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
			}
			fmt.Println("IdPValue := " + string(idpsValue))
			app.SetStateDB([]byte(idpsKey), []byte(idpsValue))
		}

		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.WrongRole, "Wrong Role", "")
}

func addNamespace(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("AddNamespace")
	var funcParam Namespace
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "AllNamespace"
	chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces []Namespace

	if chkExists != nil {
		err = json.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// Check duplicate namespace
		for _, namespace := range namespaces {
			if namespace.Namespace == funcParam.Namespace {
				return ReturnDeliverTxLog(code.DuplicateNamespace, "Duplicate namespace", "")
			}
		}
	}
	namespaces = append(namespaces, funcParam)
	value, err := json.Marshal(namespaces)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func deleteNamespace(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("DeleteNamespace")
	var funcParam DeleteNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "AllNamespace"
	chkExists := app.state.db.Get(prefixKey([]byte(key)))

	var namespaces []Namespace

	if chkExists != nil {
		err = json.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index, namespace := range namespaces {
			if namespace.Namespace == funcParam.Namespace {
				namespaces = append(namespaces[:index], namespaces[index+1:]...)
				break
			}
		}

		value, err := json.Marshal(namespaces)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}

	return ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
}
