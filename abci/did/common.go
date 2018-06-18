package did

import (
	"encoding/json"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func registerMsqAddress(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterMsqAddress, Parameter: %s", param)
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

func getNodeMasterPublicKey(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetNodeMasterPublicKey, Parameter: %s", param)
	var funcParam GetNodeMasterPublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	key := "NodeID" + "|" + funcParam.NodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	var res GetNodeMasterPublicKeyResult
	res.MasterPublicKey = ""

	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.Height, app)
		}
		res.MasterPublicKey = nodeDetail.MasterPublicKey
	}

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(value, "success", app.state.Height, app)
}

func getNodePublicKey(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetNodePublicKey, Parameter: %s", param)
	var funcParam GetNodePublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	key := "NodeID" + "|" + funcParam.NodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	var res GetNodePublicKeyResult
	res.PublicKey = ""

	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.Height, app)
		}
		res.PublicKey = nodeDetail.PublicKey
	}

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(value, "success", app.state.Height, app)
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

func getIdpNodes(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetIdpNodes, Parameter: %s", param)
	var funcParam GetIdpNodesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var returnNodes GetIdpNodesResult
	returnNodes.Node = make([]MsqDestinationNode, 0)

	if funcParam.HashID == "" {
		// Get all IdP that's max_ial >= min_ial && max_aal >= min_aal
		idpsKey := "IdPList"
		idpsValue := app.state.db.Get(prefixKey([]byte(idpsKey)))
		var idpsList []string
		if idpsValue != nil {
			err := json.Unmarshal([]byte(idpsValue), &idpsList)
			if err != nil {
				return ReturnQuery(nil, err.Error(), app.state.Height, app)
			}
			for _, idp := range idpsList {
				// check Max IAL
				maxIalAalKey := "MaxIalAalNode" + "|" + idp
				maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
				if maxIalAalValue != nil {
					var maxIalAal MaxIalAal
					err := json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
					if err != nil {
						return ReturnQuery(nil, err.Error(), app.state.Height, app)
					}
					if maxIalAal.MaxIal >= funcParam.MinIal &&
						maxIalAal.MaxAal >= funcParam.MinAal {
						nodeName := getNodeNameByNodeID(idp, app)
						var msqDesNode = MsqDestinationNode{
							idp,
							nodeName,
							maxIalAal.MaxIal,
							maxIalAal.MaxAal,
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
				return ReturnQuery(nil, err.Error(), app.state.Height, app)
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
							return ReturnQuery(nil, err.Error(), app.state.Height, app)
						}
						if maxIalAal.MaxIal >= funcParam.MinIal &&
							maxIalAal.MaxAal >= funcParam.MinAal {
							nodeName := getNodeNameByNodeID(node.NodeID, app)
							var msqDesNode = MsqDestinationNode{
								node.NodeID,
								nodeName,
								maxIalAal.MaxIal,
								maxIalAal.MaxAal,
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
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	if len(returnNodes.Node) > 0 {
		return ReturnQuery(value, "success", app.state.Height, app)
	}
	return ReturnQuery(value, "not found", app.state.Height, app)
}

func getAsNodesByServiceId(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetAsNodesByServiceId, Parameter: %s", param)
	var funcParam GetAsNodesByServiceIdParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	key := "ServiceDestination" + "|" + funcParam.ServiceID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.Height, app)
		}
		return ReturnQuery(value, "not found", app.state.Height, app)
	}

	var storedData GetAsNodesByServiceIdResult
	err = json.Unmarshal([]byte(value), &storedData)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var result GetAsNodesByServiceIdWithNameResult
	for index := range storedData.Node {
		var newRow = ASNodeResult{
			storedData.Node[index].ID,
			storedData.Node[index].Name,
			storedData.Node[index].MinIal,
			storedData.Node[index].MinAal,
		}
		result.Node = append(result.Node, newRow)
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(resultJSON, "success", app.state.Height, app)
}

func getMsqAddress(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetMsqAddress, Parameter: %s", param)
	var funcParam GetMsqAddressParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	key := "MsqAddress" + "|" + funcParam.NodeID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height, app)
	}
	return ReturnQuery(value, "success", app.state.Height, app)
}

func getCanAddAccessor(requestID string, app *DIDApplication) bool {
	result := false
	key := "Request" + "|" + requestID
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var request Request
		err := json.Unmarshal([]byte(value), &request)
		if err == nil {
			if request.CanAddAccessor {
				result = true
			}
		}
	}
	return result
}

func getRequest(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetRequest, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height, app)
	}
	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	// Not derive status in ABCI
	// status := "pending"
	// acceptCount := 0
	// rejectCount := 0
	// for _, response := range request.Responses {
	// 	if response.Status == "accept" {
	// 		acceptCount++
	// 	} else if response.Status == "reject" {
	// 		rejectCount++
	// 	}
	// }

	// if acceptCount > 0 {
	// 	status = "confirmed"
	// }

	// if rejectCount > 0 {
	// 	status = "rejected"
	// }

	// if acceptCount > 0 && rejectCount > 0 {
	// 	status = "complicated"
	// }

	// // check AS's answer
	// checkAS := true
	// // Get AS count
	// for _, dataRequest := range request.DataRequestList {
	// 	if len(dataRequest.AnsweredAsIdList) < dataRequest.Count {
	// 		checkAS = false
	// 		break
	// 	}
	// }

	// if acceptCount >= request.MinIdp && checkAS {
	// 	status = "completed"
	// }

	var res GetRequestResult
	res.IsClosed = request.IsClosed
	res.IsTimedOut = request.IsTimedOut
	res.MessageHash = request.MessageHash
	res.Mode = request.Mode

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	return ReturnQuery(value, "success", app.state.Height, app)
}

func getRequestDetail(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetRequestDetail, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	key := "Request" + "|" + funcParam.RequestID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height, app)
	}

	// not get status
	// resultStatus := getRequest(param, app)
	// var requestResult GetRequestResult
	// err = json.Unmarshal([]byte(resultStatus.Value), &requestResult)
	// if err != nil {
	// 	value = []byte("")
	// 	return ReturnQuery(value, "not found", app.state.Height)
	// }

	var result GetRequestDetailResult
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		value = []byte("")
		return ReturnQuery(value, err.Error(), app.state.Height, app)
	}
	// not set status
	// result.Status = requestResult.Status

	resultJSON, err := json.Marshal(result)
	if err != nil {
		value = []byte("")
		return ReturnQuery(value, err.Error(), app.state.Height, app)
	}

	return ReturnQuery(resultJSON, "success", app.state.Height, app)
}

func getNamespaceList(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetNamespaceList, Parameter: %s", param)
	key := "AllNamespace"
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height, app)
	}
	return ReturnQuery(value, "success", app.state.Height, app)
}

func getServiceDetail(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetServiceDetail, Parameter: %s", param)
	var funcParam GetServiceDetailParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	key := "Service" + "|" + funcParam.ServiceID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.Height, app)
	}
	return ReturnQuery(value, "success", app.state.Height, app)
}

func updateNode(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNode, Parameter: %s", param)
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
			value := app.state.db.Get(prefixKey([]byte(publicKeyRoleKey)))
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
			value := app.state.db.Get(prefixKey([]byte(publicKeyRoleKey)))
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

func checkExistingIdentity(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("CheckExistingIdentity, Parameter: %s", param)
	var funcParam CheckExistingIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var result CheckExistingIdentityResult
	result.Exist = false

	key := "MsqDestination" + "|" + funcParam.HashID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value != nil {
		var nodes []Node
		err = json.Unmarshal([]byte(value), &nodes)
		if err == nil {
			result.Exist = true
		}
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(returnValue, "success", app.state.Height, app)
}

func getAccessorGroupID(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetAccessorGroupID, Parameter: %s", param)
	var funcParam GetAccessorGroupIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var result GetAccessorGroupIDResult
	result.AccessorGroupID = ""

	key := "Accessor" + "|" + funcParam.AccessorID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value != nil {
		var accessor Accessor
		err = json.Unmarshal([]byte(value), &accessor)
		if err == nil {
			result.AccessorGroupID = accessor.AccessorGroupID
		}
	}

	returnValue, err := json.Marshal(result)

	// If value == nil set log = "not found"
	if value == nil {
		return ReturnQuery(returnValue, "not found", app.state.Height, app)
	}

	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(returnValue, "success", app.state.Height, app)
}

func getAccessorKey(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetAccessorKey, Parameter: %s", param)
	var funcParam GetAccessorKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var result GetAccessorKeyResult
	result.AccessorPublicKey = ""

	key := "Accessor" + "|" + funcParam.AccessorID
	value := app.state.db.Get(prefixKey([]byte(key)))

	if value != nil {
		var accessor Accessor
		err = json.Unmarshal([]byte(value), &accessor)
		if err == nil {
			result.AccessorPublicKey = accessor.AccessorPublicKey
		}
	}

	returnValue, err := json.Marshal(result)

	// If value == nil set log = "not found"
	if value == nil {
		return ReturnQuery(returnValue, "not found", app.state.Height, app)
	}

	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(returnValue, "success", app.state.Height, app)
}

func getServiceList(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetServiceList, Parameter: %s", param)
	key := "AllService"
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value == nil {
		result := make([]ServiceDetail, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.Height, app)
		}
		return ReturnQuery(value, "not found", app.state.Height, app)
	}
	return ReturnQuery(value, "success", app.state.Height, app)
}

func getServiceNameByServiceID(serviceID string, app *DIDApplication) string {
	key := "Service" + "|" + serviceID
	value := app.state.db.Get(prefixKey([]byte(key)))
	var result ServiceDetail
	if value != nil {
		err := json.Unmarshal([]byte(value), &result)
		if err != nil {
			return ""
		}
		return result.ServiceName
	}
	return ""
}

// func getNodeInfo(param string, app *DIDApplication) types.ResponseQuery {
// 	app.logger.Infof("GetNodeInfo, Parameter: %s", param)
// 	var result GetNodeInfoResult
// 	result.Version = app.Version
// 	value, err := json.Marshal(result)
// 	if err != nil {
// 		return ReturnQuery(nil, err.Error(), app.state.Height, app)
// 	}
// 	return ReturnQuery(value, "success", app.state.Height, app)
// }

func checkExistingAccessorID(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorID, Parameter: %s", param)
	var funcParam CheckExistingAccessorIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var result CheckExistingResult
	result.Exist = false

	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	accessorValue := app.state.db.Get(prefixKey([]byte(accessorKey)))
	if accessorValue != nil {
		var accessor Accessor
		err = json.Unmarshal([]byte(accessorValue), &accessor)
		if err == nil {
			result.Exist = true
		}
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(returnValue, "success", app.state.Height, app)
}

func checkExistingAccessorGroupID(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorGroupID, Parameter: %s", param)
	var funcParam CheckExistingAccessorGroupIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var result CheckExistingResult
	result.Exist = false

	accessorGroupKey := "AccessorGroup" + "|" + funcParam.AccessorGroupID
	accessorGroupValue := app.state.db.Get(prefixKey([]byte(accessorGroupKey)))
	if accessorGroupValue != nil {
		result.Exist = true
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(returnValue, "success", app.state.Height, app)
}

func getNodeInfo(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetNodeInfo, Parameter: %s", param)
	var funcParam GetNodeInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var result GetNodeInfoResult

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	if nodeDetailValue != nil {
		var nodeDetail NodeDetail
		err = json.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.Height, app)
		}
		result.MasterPublicKey = nodeDetail.MasterPublicKey
		result.PublicKey = nodeDetail.PublicKey
		result.NodeName = nodeDetail.NodeName
	}

	maxIalAalKey := "MaxIalAalNode" + "|" + funcParam.NodeID
	maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
	if maxIalAalValue != nil {
		var maxIalAal MaxIalAal
		err = json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.Height, app)
		}
		result.MaxIal = maxIalAal.MaxIal
		result.MaxAal = maxIalAal.MaxAal
	}

	publicKeyRoleKey := "NodePublicKeyRole" + "|" + result.PublicKey
	role := app.state.db.Get(prefixKey([]byte(publicKeyRoleKey)))
	result.Role = string(role)

	value, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	return ReturnQuery(value, "success", app.state.Height, app)
}

func getIdentityInfo(param string, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("GetIdentityInfo, Parameter: %s", param)
	var funcParam GetIdentityInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	var result GetIdentityInfoResult

	key := "MsqDestination" + "|" + funcParam.HashID
	chkExists := app.state.db.Get(prefixKey([]byte(key)))

	if chkExists != nil {
		var nodes []Node
		err = json.Unmarshal([]byte(chkExists), &nodes)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.Height, app)
		}

		for _, node := range nodes {
			if node.NodeID == funcParam.NodeID {
				result.Ial = node.Ial
				break
			}
		}
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}

	if result.Ial > 0.0 {
		return ReturnQuery(returnValue, "success", app.state.Height, app)
	}
	return ReturnQuery(returnValue, "not found", app.state.Height, app)
}
