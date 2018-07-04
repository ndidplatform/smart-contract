/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package did

import (
	"encoding/json"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/tendermint/abci/types"
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

func getNodeMasterPublicKey(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodeMasterPublicKey, Parameter: %s", param)
	var funcParam GetNodeMasterPublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	key := "NodeID" + "|" + funcParam.NodeID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	var res GetNodeMasterPublicKeyResult
	res.MasterPublicKey = ""

	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
		res.MasterPublicKey = nodeDetail.MasterPublicKey
	}

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}

func getNodePublicKey(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodePublicKey, Parameter: %s", param)
	var funcParam GetNodePublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	key := "NodeID" + "|" + funcParam.NodeID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	var res GetNodePublicKeyResult
	res.PublicKey = ""

	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
		res.PublicKey = nodeDetail.PublicKey
	}

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}

func getNodeNameByNodeID(nodeID string, app *DIDApplication) string {
	key := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
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

func getIdpNodes(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetIdpNodes, Parameter: %s", param)
	var funcParam GetIdpNodesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var returnNodes GetIdpNodesResult
	returnNodes.Node = make([]MsqDestinationNode, 0)

	if funcParam.HashID == "" {
		// Get all IdP that's max_ial >= min_ial && max_aal >= min_aal
		idpsKey := "IdPList"
		_, idpsValue := app.state.db.GetVersioned(prefixKey([]byte(idpsKey)), height)
		var idpsList []string
		if idpsValue != nil {
			err := json.Unmarshal([]byte(idpsValue), &idpsList)
			if err != nil {
				return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
			}
			for _, idp := range idpsList {
				// check Max IAL
				maxIalAalKey := "MaxIalAalNode" + "|" + idp
				_, maxIalAalValue := app.state.db.GetVersioned(prefixKey([]byte(maxIalAalKey)), height)
				if maxIalAalValue != nil {
					var maxIalAal MaxIalAal
					err := json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
					if err != nil {
						return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
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
						// filter node is active
						nodeDetailKey := "NodeID" + "|" + idp
						_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
						if nodeDetailValue != nil {
							var nodeDetail NodeDetail
							err := json.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
							if err != nil {
								return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
							}
							if nodeDetail.Active {
								returnNodes.Node = append(returnNodes.Node, msqDesNode)
							}
						}
					}
				}
			}
		}
	} else {
		key := "MsqDestination" + "|" + funcParam.HashID
		_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

		if value != nil {
			var nodes []Node
			err = json.Unmarshal([]byte(value), &nodes)
			if err != nil {
				return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
			}

			for _, node := range nodes {
				if node.Ial >= funcParam.MinIal {
					// check Max IAL && AAL
					maxIalAalKey := "MaxIalAalNode" + "|" + node.NodeID
					_, maxIalAalValue := app.state.db.GetVersioned(prefixKey([]byte(maxIalAalKey)), height)

					if maxIalAalValue != nil {
						var maxIalAal MaxIalAal
						err := json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
						if err != nil {
							return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
						}
						if maxIalAal.MaxIal >= funcParam.MinIal &&
							maxIalAal.MaxAal >= funcParam.MinAal &&
							node.Active {
							nodeName := getNodeNameByNodeID(node.NodeID, app)
							var msqDesNode = MsqDestinationNode{
								node.NodeID,
								nodeName,
								maxIalAal.MaxIal,
								maxIalAal.MaxAal,
							}
							// filter node is active
							nodeDetailKey := "NodeID" + "|" + node.NodeID
							_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
							if nodeDetailValue != nil {
								var nodeDetail NodeDetail
								err := json.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
								if err != nil {
									return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
								}
								if nodeDetail.Active {
									returnNodes.Node = append(returnNodes.Node, msqDesNode)
								}
							}
						}
					}
				}
			}
		}
	}

	value, err := json.Marshal(returnNodes)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	// return ReturnQuery(value, "success", app.state.db.Version64(), app)
	if len(returnNodes.Node) > 0 {
		return ReturnQuery(value, "success", app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "not found", app.state.db.Version64(), app)
}

func getAsNodesByServiceId(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetAsNodesByServiceId, Parameter: %s", param)
	var funcParam GetAsNodesByServiceIdParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	key := "ServiceDestination" + "|" + funcParam.ServiceID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}

	// filter serive is active
	serviceKey := "Service" + "|" + funcParam.ServiceID
	_, serviceValue := app.state.db.Get(prefixKey([]byte(serviceKey)))
	if serviceValue != nil {
		var service ServiceDetail
		err = json.Unmarshal([]byte(serviceValue), &service)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
		if service.Active == false {
			var result GetAsNodesByServiceIdResult
			result.Node = make([]ASNode, 0)
			value, err := json.Marshal(result)
			if err != nil {
				return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
			}
			return ReturnQuery(value, "service is not active", app.state.db.Version64(), app)
		}
	} else {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}

	var storedData GetAsNodesByServiceIdResult
	err = json.Unmarshal([]byte(value), &storedData)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result GetAsNodesByServiceIdWithNameResult
	for index := range storedData.Node {
		var newRow = ASNodeResult{
			storedData.Node[index].ID,
			storedData.Node[index].Name,
			storedData.Node[index].MinIal,
			storedData.Node[index].MinAal,
		}

		// filter node is active
		nodeDetailKey := "NodeID" + "|" + storedData.Node[index].ID
		_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
		if nodeDetailValue != nil {
			var nodeDetail NodeDetail
			err := json.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
			if err != nil {
				return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
			}
			if nodeDetail.Active {
				result.Node = append(result.Node, newRow)
			}
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(resultJSON, "success", app.state.db.Version64(), app)
}

func getMsqAddress(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetMsqAddress, Parameter: %s", param)
	var funcParam GetMsqAddressParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	key := "MsqAddress" + "|" + funcParam.NodeID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}

func getCanAddAccessor(requestID string, app *DIDApplication) bool {
	result := false
	key := "Request" + "|" + requestID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
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

func getRequest(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetRequest, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	key := "Request" + "|" + funcParam.RequestID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}
	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var res GetRequestResult
	res.IsClosed = request.IsClosed
	res.IsTimedOut = request.IsTimedOut
	res.MessageHash = request.MessageHash
	res.Mode = request.Mode

	value, err = json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}

func getRequestDetail(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetRequestDetail, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	key := "Request" + "|" + funcParam.RequestID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}

	var result GetRequestDetailResult
	var request Request
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		value = []byte("")
		return ReturnQuery(value, err.Error(), app.state.db.Version64(), app)
	}
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		value = []byte("")
		return ReturnQuery(value, err.Error(), app.state.db.Version64(), app)
	}

	// Check Role, If it's IdP then Set set special = true
	ownerRole := getRoleFromNodeID(request.Owner, app)
	if string(ownerRole) == "IdP" {
		result.Special = true
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		value = []byte("")
		return ReturnQuery(value, err.Error(), app.state.db.Version64(), app)
	}

	return ReturnQuery(resultJSON, "success", app.state.db.Version64(), app)
}

func getNamespaceList(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetNamespaceList, Parameter: %s", param)
	key := "AllNamespace"
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}

	result := make([]Namespace, 0)
	// filter flag==true
	var namespaces []Namespace
	err := json.Unmarshal([]byte(value), &namespaces)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	for _, namespace := range namespaces {
		if namespace.Active {
			result = append(result, namespace)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
}

func getServiceDetail(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetServiceDetail, Parameter: %s", param)
	var funcParam GetServiceDetailParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	key := "Service" + "|" + funcParam.ServiceID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}

func updateNode(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNode, Parameter: %s", param)
	var funcParam UpdateNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "NodeID" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value != nil {
		var nodeDetail NodeDetail
		err := json.Unmarshal([]byte(value), &nodeDetail)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// update MasterPublicKey
		if funcParam.MasterPublicKey != "" {
			nodeDetail.MasterPublicKey = funcParam.MasterPublicKey
		}

		// update PublicKey
		if funcParam.PublicKey != "" {
			nodeDetail.PublicKey = funcParam.PublicKey
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

func checkExistingIdentity(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("CheckExistingIdentity, Parameter: %s", param)
	var funcParam CheckExistingIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result CheckExistingIdentityResult
	result.Exist = false

	key := "MsqDestination" + "|" + funcParam.HashID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value != nil {
		var nodes []Node
		err = json.Unmarshal([]byte(value), &nodes)
		if err == nil {
			result.Exist = true
		}
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
}

func getAccessorGroupID(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetAccessorGroupID, Parameter: %s", param)
	var funcParam GetAccessorGroupIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result GetAccessorGroupIDResult
	result.AccessorGroupID = ""

	key := "Accessor" + "|" + funcParam.AccessorID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

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
		return ReturnQuery(returnValue, "not found", app.state.db.Version64(), app)
	}

	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
}

func getAccessorKey(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetAccessorKey, Parameter: %s", param)
	var funcParam GetAccessorKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result GetAccessorKeyResult
	result.AccessorPublicKey = ""

	key := "Accessor" + "|" + funcParam.AccessorID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if value != nil {
		var accessor Accessor
		err = json.Unmarshal([]byte(value), &accessor)
		if err == nil {
			result.AccessorPublicKey = accessor.AccessorPublicKey
			result.Active = accessor.Active
		}
	}

	returnValue, err := json.Marshal(result)

	// If value == nil set log = "not found"
	if value == nil {
		return ReturnQuery(returnValue, "not found", app.state.db.Version64(), app)
	}

	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
}

func getServiceList(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetServiceList, Parameter: %s", param)
	key := "AllService"
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	if value == nil {
		result := make([]ServiceDetail, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}

	result := make([]ServiceDetail, 0)
	// filter flag==true
	var services []ServiceDetail
	err := json.Unmarshal([]byte(value), &services)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	for _, service := range services {
		if service.Active {
			result = append(result, service)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
}

func getServiceNameByServiceID(serviceID string, app *DIDApplication) string {
	key := "Service" + "|" + serviceID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
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

// func getNodeInfo(param string, app *DIDApplication, height int64) types.ResponseQuery{
// 	app.logger.Infof("GetNodeInfo, Parameter: %s", param)
// 	var result GetNodeInfoResult
// 	result.Version = app.Version
// 	value, err := json.Marshal(result)
// 	if err != nil {
// 		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
// 	}
// 	return ReturnQuery(value, "success", app.state.db.Version64(), app)
// }

func checkExistingAccessorID(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorID, Parameter: %s", param)
	var funcParam CheckExistingAccessorIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result CheckExistingResult
	result.Exist = false

	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	_, accessorValue := app.state.db.GetVersioned(prefixKey([]byte(accessorKey)), height)
	if accessorValue != nil {
		var accessor Accessor
		err = json.Unmarshal([]byte(accessorValue), &accessor)
		if err == nil {
			result.Exist = true
		}
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
}

func checkExistingAccessorGroupID(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorGroupID, Parameter: %s", param)
	var funcParam CheckExistingAccessorGroupIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result CheckExistingResult
	result.Exist = false

	accessorGroupKey := "AccessorGroup" + "|" + funcParam.AccessorGroupID
	_, accessorGroupValue := app.state.db.GetVersioned(prefixKey([]byte(accessorGroupKey)), height)
	if accessorGroupValue != nil {
		result.Exist = true
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
}

func getNodeInfo(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodeInfo, Parameter: %s", param)
	var funcParam GetNodeInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result GetNodeInfoResult

	nodeDetailKey := "NodeID" + "|" + funcParam.NodeID
	_, nodeDetailValue := app.state.db.GetVersioned(prefixKey([]byte(nodeDetailKey)), height)
	if nodeDetailValue != nil {
		var nodeDetail NodeDetail
		err = json.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
		result.MasterPublicKey = nodeDetail.MasterPublicKey
		result.PublicKey = nodeDetail.PublicKey
		result.NodeName = nodeDetail.NodeName
		result.Role = nodeDetail.Role
	}

	maxIalAalKey := "MaxIalAalNode" + "|" + funcParam.NodeID
	_, maxIalAalValue := app.state.db.GetVersioned(prefixKey([]byte(maxIalAalKey)), height)
	if maxIalAalValue != nil {
		var maxIalAal MaxIalAal
		err = json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
		result.MaxIal = maxIalAal.MaxIal
		result.MaxAal = maxIalAal.MaxAal
	}

	// publicKeyRoleKey := "NodePublicKeyRole" + "|" + result.PublicKey
	// _, role := app.state.db.GetVersioned(prefixKey([]byte(publicKeyRoleKey)), height)

	value, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}

func getIdentityInfo(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetIdentityInfo, Parameter: %s", param)
	var funcParam GetIdentityInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result GetIdentityInfoResult

	key := "MsqDestination" + "|" + funcParam.HashID
	_, chkExists := app.state.db.GetVersioned(prefixKey([]byte(key)), height)

	if chkExists != nil {
		var nodes []Node
		err = json.Unmarshal([]byte(chkExists), &nodes)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
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
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	if result.Ial > 0.0 {
		return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "not found", app.state.db.Version64(), app)
}

func getDataSignature(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetDataSignature, Parameter: %s", param)
	var funcParam GetDataSignatureParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	signDataKey := "SignData" + "|" + funcParam.NodeID + "|" + funcParam.ServiceID + "|" + funcParam.RequestID
	_, signDataValue := app.state.db.GetVersioned(prefixKey([]byte(signDataKey)), height)

	var result GetDataSignatureResult

	if signDataValue != nil {
		result.Signature = string(signDataValue)
	}

	returnValue, err := json.Marshal(result)
	if signDataValue != nil {
		return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "not found", app.state.db.Version64(), app)
}

func getIdentityProof(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetIdentityProof, Parameter: %s", param)
	var funcParam GetIdentityProofParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	identityProofKey := "IdentityProof" + "|" + funcParam.RequestID + "|" + funcParam.IdpID
	_, identityProofValue := app.state.db.GetVersioned(prefixKey([]byte(identityProofKey)), height)
	var result GetIdentityProofResult
	if identityProofValue != nil {
		result.IdentityProof = string(identityProofValue)
	}
	returnValue, err := json.Marshal(result)
	if identityProofValue != nil {
		return ReturnQuery(returnValue, "success", app.state.db.Version64(), app)
	}
	return ReturnQuery(returnValue, "not found", app.state.db.Version64(), app)
}

func getServicesByAsID(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetServicesByAsID, Parameter: %s", param)
	var funcParam GetServicesByAsIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	var result GetServicesByAsIDResult
	result.Services = make([]Service, 0)

	provideServiceKey := "ProvideService" + "|" + funcParam.AsID
	_, provideServiceValue := app.state.db.Get(prefixKey([]byte(provideServiceKey)))
	var services []Service
	if provideServiceValue != nil {
		err := json.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
	}

	nodeDetailKey := "NodeID" + "|" + funcParam.AsID
	_, nodeDetailValue := app.state.db.Get(prefixKey([]byte(nodeDetailKey)))
	var nodeDetail NodeDetail
	if nodeDetailValue != nil {
		err := json.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
		}
	}

	for _, provideService := range services {
		serviceKey := "Service" + "|" + provideService.ServiceID
		_, serviceValue := app.state.db.Get(prefixKey([]byte(serviceKey)))
		var service ServiceDetail
		if serviceValue != nil {
			err = json.Unmarshal([]byte(serviceValue), &service)
			if err != nil {
				return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
			}
		}
		if provideService.Active && nodeDetail.Active && service.Active {
			result.Services = append(result.Services, provideService)
		}
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}

	if len(result.Services) > 0 {
		return ReturnQuery(resultJSON, "success", app.state.db.Version64(), app)
	} else {
		return ReturnQuery(resultJSON, "not found", app.state.db.Version64(), app)
	}
}
