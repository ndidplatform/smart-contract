package did

import (
	"encoding/json"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func registerMsqAddress(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("RegisterMsqAddress")
	var funcParam RegisterMsqAddressParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := "MsqAddress" + "|" + funcParam.NodeID
	var msqAddress = MsqAddress{
		funcParam.IP,
		funcParam.Port,
	}
	value, err := json.Marshal(msqAddress)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.SetStateDB([]byte(key), []byte(value))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func getNodePublicKey(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetNodePublicKey")
	var funcParam GetNodePublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "NodeID" + "|" + funcParam.NodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	var res GetNodePublicKeyResult
	res.PublicKey = ""

	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.Height)
		}
		res.PublicKey = nodeDetail.PublicKey
	}

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getNodeNameByNodeID(nodeID string, app *DIDApplication) string {
	key := "NodeID" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ""
		}
		return nodeDetail.NodeName
	}
	return ""
}

func getMsqDestination(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetMsqDestination")
	var funcParam GetMsqDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var returnNodes GetMsqDestinationResult

	if funcParam.HashID == "" {
		// Get all IdP that's max_ial >= min_ial && max_aal >= min_aal
		idpsKey := "IdPList"
		idpsValue := app.state.db.Get(prefixKey([]byte(idpsKey)))
		var idpsList []string
		if idpsValue != nil {
			err := json.Unmarshal([]byte(idpsValue), &idpsList)
			if err != nil {
				return ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, idp := range idpsList {
				// check Max IAL
				maxIalAalKey := "MaxIalAalNode" + "|" + idp
				maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
				if maxIalAalValue != nil {
					var maxIalAal MaxIalAal
					err := json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
					if err != nil {
						return ReturnQuery(nil, err.Error(), app.state.Height)
					}
					if maxIalAal.MaxIal >= funcParam.MinIal &&
						maxIalAal.MaxAal >= funcParam.MinAal {
						nodeName := getNodeNameByNodeID(idp, app)
						var msqDesNode = MsqDestinationNode{
							idp,
							nodeName,
						}
						returnNodes.Node = append(returnNodes.Node, msqDesNode)
					}
				}
			}
		}
	} else {
		key := "MsqDestination" + "|" + funcParam.HashID
		value := app.state.db.Get(prefixKey([]byte(key)))

		if value != nil {
			var nodes []Node
			err = json.Unmarshal([]byte(value), &nodes)
			if err != nil {
				return ReturnQuery(nil, err.Error(), app.state.Height)
			}

			for _, node := range nodes {
				if node.Ial >= funcParam.MinIal {
					// check Max IAL && AAL
					maxIalAalKey := "MaxIalAalNode" + "|" + node.NodeID
					maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
					if maxIalAalValue != nil {
						var maxIalAal MaxIalAal
						err := json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
						if err != nil {
							return ReturnQuery(nil, err.Error(), app.state.Height)
						}
						if maxIalAal.MaxIal >= funcParam.MinIal &&
							maxIalAal.MaxAal >= funcParam.MinAal {
							nodeName := getNodeNameByNodeID(node.NodeID, app)
							var msqDesNode = MsqDestinationNode{
								node.NodeID,
								nodeName,
							}
							returnNodes.Node = append(returnNodes.Node, msqDesNode)

						}
					}
				}
			}
		}
	}

	value, err := json.Marshal(returnNodes)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getServiceDestination(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetServiceDestination")
	var funcParam GetServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "ServiceDestination" + "|" + funcParam.AsServiceID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getMsqAddress(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetMsqAddress")
	var funcParam GetMsqAddressParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "MsqAddress" + "|" + funcParam.NodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getRequest(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetRequest")
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}

	asCount := 0
	// Get AS count
	for _, dataRequest := range request.DataRequestList {
		asCount = asCount + dataRequest.Count
	}

	status := "pending"
	acceptCount := 0
	rejectCount := 0
	for _, response := range request.Responses {
		if response.Status == "accept" {
			acceptCount++
		} else if response.Status == "reject" {
			rejectCount++
		}
	}

	if acceptCount > 0 {
		status = "confirmed"
	}

	if rejectCount > 0 {
		status = "rejected"
	}

	if acceptCount > 0 && rejectCount > 0 {
		status = "complicated"
	}

	if acceptCount >= request.MinIdp && request.SignDataCount >= asCount {
		status = "completed"
	}

	var res GetRequestResult
	res.Status = status
	res.IsClosed = request.IsClosed
	res.IsTimedOut = request.IsTimedOut
	res.MessageHash = request.MessageHash

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return ReturnQuery(value, "success", app.state.Height)
}

func getRequestDetail(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetRequestDetail")
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getAccessorMethod(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetAccessorMethod")
	var funcParam GetAccessorMethodParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "AccessorMethod" + "|" + funcParam.AccessorID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	var accessorMethod AccessorMethod
	err = json.Unmarshal([]byte(value), &accessorMethod)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var res GetAccessorMethodResult
	res.AccessorType = accessorMethod.AccessorType
	res.AccessorKey = accessorMethod.AccessorKey
	res.Commitment = accessorMethod.Commitment
	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getNamespaceList(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetNamespaceList")
	key := "AllNamespace"
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func getServiceDetail(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetServiceDetail")
	var funcParam GetServiceDetailParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := "Service" + "|" + funcParam.AsServiceID + "|" + funcParam.NodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}

func updateNode(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("UpdateNode")
	var funcParam UpdateNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "NodeID" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// update MasterPublicKey
		if funcParam.MasterPublicKey != "" {

			// set role old pubKey = ""
			publicKeyRoleKey := "NodePublicKeyRole" + "|" + nodeDetail.MasterPublicKey
			value := app.state.db.Get(prefixKey([]byte(key)))
			role := string(value)
			publicKeyRoleValue := ""
			app.SetStateDB([]byte(publicKeyRoleKey), []byte(publicKeyRoleValue))

			nodeDetail.MasterPublicKey = funcParam.MasterPublicKey
			// set role new pubKey
			publicKeyRoleKey = "NodePublicKeyRole" + "|" + funcParam.MasterPublicKey
			publicKeyRoleValue = role
			app.SetStateDB([]byte(publicKeyRoleKey), []byte(publicKeyRoleValue))
		}

		// update PublicKey
		if funcParam.PublicKey != "" {

			// set role old pubKey = ""
			publicKeyRoleKey := "NodePublicKeyRole" + "|" + nodeDetail.PublicKey
			value := app.state.db.Get(prefixKey([]byte(key)))
			role := string(value)
			publicKeyRoleValue := ""
			app.SetStateDB([]byte(publicKeyRoleKey), []byte(publicKeyRoleValue))

			nodeDetail.PublicKey = funcParam.PublicKey
			// set role new pubKey
			publicKeyRoleKey = "NodePublicKeyRole" + "|" + funcParam.PublicKey
			publicKeyRoleValue = role
			app.SetStateDB([]byte(publicKeyRoleKey), []byte(publicKeyRoleValue))
		}

		nodeDetailValue, err := json.Marshal(nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(nodeDetailValue))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
}
