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
	"UpdateValidator": true,
	"RegisterService": true,
	"DeleteService":   true,
}

func initNDID(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("InitNDID")
	var funcParam InitNDIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := "NodePublicKeyRole" + "|" + funcParam.PublicKey
	value := []byte("MasterNDID")
	app.SetStateDB([]byte(key), []byte(value))

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	// TODO: fix param InitNDID
	var nodeDetail = NodeDetail{
		funcParam.PublicKey,
		funcParam.PublicKey,
		"NDID",
	}
	nodeDetailValue, err := json.Marshal(nodeDetail)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))

	key = "MasterNDID"
	value = []byte(funcParam.PublicKey)
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func registerNode(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("RegisterNode")
	var funcParam RegisterNode
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "NodeID" + "|" + funcParam.NodeID
	// check Duplicate Node ID
	chkExists := app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicateNodeID, "Duplicate Node ID", "")
	}

	// check Duplicate Master Key
	key = "NodePublicKeyRole" + "|" + funcParam.MasterPublicKey
	chkExists = app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicatePublicKey, "Duplicate Public Key", "")
	}

	// check Duplicate Key
	key = "NodePublicKeyRole" + "|" + funcParam.PublicKey
	chkExists = app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicatePublicKey, "Duplicate Public Key", "")
	}

	if funcParam.Role == "RP" ||
		funcParam.Role == "IdP" ||
		funcParam.Role == "AS" {
		nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
		var nodeDetail = NodeDetail{
			funcParam.PublicKey,
			funcParam.MasterPublicKey,
			funcParam.NodeName,
		}
		nodeDetailValue, err := json.Marshal(nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(nodeDetailKey), []byte(nodeDetailValue))

		// Set master Role
		publicKeyRoleKey := "NodePublicKeyRole" + "|" + funcParam.MasterPublicKey
		publicKeyRoleValue := "Master" + funcParam.Role
		app.SetStateDB([]byte(publicKeyRoleKey), []byte(publicKeyRoleValue))

		// Set Role
		publicKeyRoleKey = "NodePublicKeyRole" + "|" + funcParam.PublicKey
		publicKeyRoleValue = funcParam.Role
		app.SetStateDB([]byte(publicKeyRoleKey), []byte(publicKeyRoleValue))

		createTokenAccount(funcParam.NodeID, app)

		// Add max_aal, min_ial when node is IdP
		if funcParam.Role == "IdP" {
			maxIalAalKey := "MaxIalAalNode" + "|" + funcParam.NodeID
			var maxIalAal MaxIalAal
			maxIalAal.MaxAal = funcParam.MaxAal
			maxIalAal.MaxIal = funcParam.MaxIal
			maxIalAalValue, err := json.Marshal(maxIalAal)
			if err != nil {
				return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
			}
			app.SetStateDB([]byte(maxIalAalKey), []byte(maxIalAalValue))

			// Save all IdP's nodeID for GetIdpNodes
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
			app.SetStateDB([]byte(idpsKey), []byte(idpsValue))
		}

		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.WrongRole, "Wrong Role", "")
}

func addNamespace(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
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

func deleteNamespace(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
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

func registerService(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("RegisterService")
	var funcParam RegisterServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.AsServiceID
	chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID", "")
	}

	// Add new service
	var service = Service{
		funcParam.ServiceName,
	}
	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(serviceKey), []byte(serviceJSON))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func deleteService(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("DeleteService")
	var funcParam DeleteServiceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	serviceKey := "Service" + "|" + funcParam.AsServiceID
	chkExists := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if chkExists == nil {
		return ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}
	app.DeleteStateDB([]byte(serviceKey))
	return ReturnDeliverTxLog(code.OK, "success", "")
}
