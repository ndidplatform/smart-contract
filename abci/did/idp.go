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

func createIdentity(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateIdentity, Parameter: %s", param)
	var funcParam CreateIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	var accessor = Accessor{
		funcParam.AccessorType,
		funcParam.AccessorPublicKey,
		funcParam.AccessorGroupID,
		true,
		nodeID,
	}

	accessorJSON, err := json.Marshal(accessor)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	accessorGroupKey := "AccessorGroup" + "|" + funcParam.AccessorGroupID
	accessorGroup := funcParam.AccessorGroupID

	// Check duplicate accessor_id
	_, chkAccessorKeyExists := app.state.db.Get(prefixKey([]byte(accessorKey)))
	if chkAccessorKeyExists != nil {
		return ReturnDeliverTxLog(code.DuplicateAccessorID, "Duplicate Accessor ID", "")
	}

	// Check duplicate accessor_group_id
	_, chkAccessorGroupKeyExists := app.state.db.Get(prefixKey([]byte(accessorGroupKey)))
	if chkAccessorGroupKeyExists != nil {
		return ReturnDeliverTxLog(code.DuplicateAccessorGroupID, "Duplicate Accessor Group ID", "")
	}

	app.SetStateDB([]byte(accessorKey), []byte(accessorJSON))
	app.SetStateDB([]byte(accessorGroupKey), []byte(accessorGroup))

	return ReturnDeliverTxLog(code.OK, "success", "")
}

func setCanAddAccessorToFalse(requestID string, app *DIDApplication) {
	key := "Request" + "|" + requestID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		var request Request
		err := json.Unmarshal([]byte(value), &request)
		if err == nil {
			request.CanAddAccessor = false
			value, err := json.Marshal(request)
			if err == nil {
				app.SetStateDB([]byte(key), []byte(value))
			}
		}
	}
}

func addAccessorMethod(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddAccessorMethod, Parameter: %s", param)
	var funcParam AccessorMethod
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// AccessorGroupID: must already exist
	accessorGroupKey := "AccessorGroup" + "|" + funcParam.AccessorGroupID
	_, chkAccessorGroupKeyExists := app.state.db.Get(prefixKey([]byte(accessorGroupKey)))
	if chkAccessorGroupKeyExists == nil {
		return ReturnDeliverTxLog(code.AccessorGroupIDNotFound, "Accessor Group ID not found", "")
	}

	// AccessorID: must not duplicate
	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	_, chkAccessorKeyExists := app.state.db.Get(prefixKey([]byte(accessorKey)))
	if chkAccessorKeyExists != nil {
		return ReturnDeliverTxLog(code.DuplicateAccessorID, "Duplicate Accessor ID", "")
	}

	// Request must be completed, can be used only once, special type
	var getRequestparam GetRequestParam
	getRequestparam.RequestID = funcParam.RequestID
	getRequestparamJSON, err := json.Marshal(getRequestparam)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	var request = getRequest(string(getRequestparamJSON), app, app.state.db.Version64())
	var requestDetail = getRequestDetail(string(getRequestparamJSON), app, app.state.db.Version64())
	var requestResult GetRequestResult
	var requestDetailResult GetRequestDetailResult
	err = json.Unmarshal([]byte(request.Value), &requestResult)
	if err != nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	err = json.Unmarshal([]byte(requestDetail.Value), &requestDetailResult)
	if err != nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}

	// Check accept result >= min_idp
	acceptCount := 0
	for _, response := range requestDetailResult.Responses {
		if response.Status == "accept" {
			acceptCount++
		}
	}
	if acceptCount < requestDetailResult.MinIdp {
		return ReturnDeliverTxLog(code.RequestIsNotCompleted, "Request is not completed", "")
	}

	if requestDetailResult.Mode != 3 {
		return ReturnDeliverTxLog(code.InvalidMode, "Onboard request must be mode 3", "")
	}

	if requestDetailResult.MinIdp < 1 {
		return ReturnDeliverTxLog(code.InvalidMinIdp, "Onboard request min_idp must be at least 1", "")
	}
	// check special type of Request && set can used only once
	canAddAccessor := getCanAddAccessor(funcParam.RequestID, app)
	if canAddAccessor != true {
		return ReturnDeliverTxLog(code.RequestIsNotSpecial, "Request is not special", "")
	}
	setCanAddAccessorToFalse(funcParam.RequestID, app)

	var accessor = Accessor{
		funcParam.AccessorType,
		funcParam.AccessorPublicKey,
		funcParam.AccessorGroupID,
		true,
		nodeID,
	}

	accessorJSON, err := json.Marshal(accessor)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(accessorKey), []byte(accessorJSON))
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func registerMsqDestination(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("RegisterMsqDestination, Parameter: %s", param)
	var funcParam RegisterMsqDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	maxIalAalKey := "MaxIalAalNode" + "|" + nodeID
	_, maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
	if maxIalAalValue != nil {
		var maxIalAal MaxIalAal
		err := json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// Validate user's ial is <= node's max_ial
		for _, user := range funcParam.Users {
			if user.Ial > maxIalAal.MaxIal {
				return ReturnDeliverTxLog(code.IALError, "IAL must be less than or equals to registered node's MAX IAL", "")
			}
		}
	}

	// If validate passed then add Msq Destination
	for _, user := range funcParam.Users {
		key := "MsqDestination" + "|" + user.HashID
		_, chkExists := app.state.db.Get(prefixKey([]byte(key)))

		if chkExists != nil {
			var nodes []Node
			err = json.Unmarshal([]byte(chkExists), &nodes)
			if err != nil {
				return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
			}

			newNode := Node{
				user.Ial,
				nodeID,
				true,
				user.First,
			}

			// Check duplicate before add
			chkDup := false
			for _, node := range nodes {
				if newNode == node {
					chkDup = true
					break
				}
			}

			// Check frist
			if user.First {
				return ReturnDeliverTxLog(code.NotFirstIdP, "This node is not first IdP", "")
			}

			if chkDup == false {
				nodes = append(nodes, newNode)
				value, err := json.Marshal(nodes)
				if err != nil {
					return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
				}
				app.SetStateDB([]byte(key), []byte(value))
			}

		} else {
			var nodes []Node
			newNode := Node{
				user.Ial,
				nodeID,
				true,
				user.First,
			}
			nodes = append(nodes, newNode)
			value, err := json.Marshal(nodes)
			if err != nil {
				return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
			}
			app.SetStateDB([]byte(key), []byte(value))
		}
	}

	return ReturnDeliverTxLog(code.OK, "success", "")
}

func createIdpResponse(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateIdpResponse, Parameter: %s", param)
	var funcParam CreateIdpResponseParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	key := "Request" + "|" + funcParam.RequestID
	var response Response
	response.Ial = funcParam.Ial
	response.Aal = funcParam.Aal
	response.Status = funcParam.Status
	response.Signature = funcParam.Signature
	response.IdpID = nodeID
	response.IdentityProof = funcParam.IdentityProof
	response.PrivateProofHash = funcParam.PrivateProofHash
	_, value := app.state.db.Get(prefixKey([]byte(key)))

	if value == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request Request
	err = json.Unmarshal([]byte(value), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check duplicate before add
	chk := false
	for _, oldResponse := range request.Responses {
		if response == oldResponse {
			chk = true
			break
		}
	}

	// Check AAL
	if request.MinAal > response.Aal {
		return ReturnDeliverTxLog(code.AALError, "Response's AAL is less than min AAL", "")
	}

	// Check IAL
	if request.MinIal > response.Ial {
		return ReturnDeliverTxLog(code.IALError, "Response's IAL is less than min IAL", "")
	}

	// Check AAL, IAL with MaxIalAal
	maxIalAalKey := "MaxIalAalNode" + "|" + nodeID
	_, maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
	if maxIalAalValue != nil {
		var maxIalAal MaxIalAal
		err = json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		if response.Aal > maxIalAal.MaxAal {
			return ReturnDeliverTxLog(code.AALError, "Response's AAL is greater than max AAL", "")
		}
		if response.Ial > maxIalAal.MaxIal {
			return ReturnDeliverTxLog(code.IALError, "Response's IAL is greater than max IAL", "")
		}
	}

	// Check min_idp
	if len(request.Responses) >= request.MinIdp {
		return ReturnDeliverTxLog(code.RequestIsCompleted, "Can't response a request that's complete response", "")
	}

	// Check IsClosed
	if request.IsClosed {
		return ReturnDeliverTxLog(code.RequestIsClosed, "Can't response a request that's closed", "")
	}

	// Check IsTimedOut
	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.RequestIsTimedOut, "Can't response a request that's timed out", "")
	}

	// Check identity proof if mode == 3
	if request.Mode == 3 {
		identityProofKey := "IdentityProof" + "|" + funcParam.RequestID + "|" + nodeID
		_, identityProofValue := app.state.db.Get(prefixKey([]byte(identityProofKey)))
		proofPassed := false
		if identityProofValue != nil {
			if funcParam.IdentityProof == string(identityProofValue) {
				proofPassed = true
			}
		}
		if proofPassed == false {
			return ReturnDeliverTxLog(code.WrongIdentityProof, "Identity proof is wrong", "")
		}
	}

	if chk == false {
		request.Responses = append(request.Responses, response)

		// NO data request. If accept >= min_idp, then auto close request
		// if len(request.DataRequestList) == 0 {
		// 	app.logger.Info("Auto close")
		// 	accept := 0
		// 	reject := 0
		// 	for _, response := range request.Responses {
		// 		if response.Status == "accept" {
		// 			accept++
		// 		} else {
		// 			reject++
		// 		}
		// 	}
		// 	if accept >= request.MinIdp && reject == 0 {
		// 		request.IsClosed = true
		// 	}
		// } else {
		// 	app.logger.Info("No auto close")
		// }

		value, err := json.Marshal(request)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
		return ReturnDeliverTxLog(code.OK, "success", funcParam.RequestID)
	}
	return ReturnDeliverTxLog(code.DuplicateResponse, "Duplicate Response", "")
}

func updateIdentity(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateIdentity, Parameter: %s", param)
	var funcParam UpdateIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check IAL must less than Max IAL
	maxIalAalKey := "MaxIalAalNode" + "|" + nodeID
	_, maxIalAalValue := app.state.db.Get(prefixKey([]byte(maxIalAalKey)))
	if maxIalAalValue != nil {
		var maxIalAal MaxIalAal
		err := json.Unmarshal([]byte(maxIalAalValue), &maxIalAal)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		if funcParam.Ial > maxIalAal.MaxIal {
			return ReturnDeliverTxLog(code.IALError, "New IAL is greater than max IAL", "")
		}
	}

	msqDesKey := "MsqDestination" + "|" + funcParam.HashID
	_, msqDesValue := app.state.db.Get(prefixKey([]byte(msqDesKey)))
	if msqDesValue != nil {
		var msqDes []Node
		err := json.Unmarshal([]byte(msqDesValue), &msqDes)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		// Selective update
		if funcParam.Ial > 0 {
			for index := range msqDes {
				if msqDes[index].NodeID == nodeID {
					msqDes[index].Ial = funcParam.Ial
					break
				}
			}
		}
		msqDesJSON, err := json.Marshal(msqDes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(msqDesKey), []byte(msqDesJSON))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.HashIDNotFound, "Hash ID not found", "")
}

func declareIdentityProof(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DeclareIdentityProof, Parameter: %s", param)
	var funcParam DeclareIdentityProofParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check the request
	requestKey := "Request" + "|" + funcParam.RequestID
	_, requestValue := app.state.db.Get(prefixKey([]byte(requestKey)))

	if requestValue == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request Request
	err = json.Unmarshal([]byte(requestValue), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// check number of responses
	if len(request.Responses) >= request.MinIdp {
		return ReturnDeliverTxLog(code.RequestIsCompleted, "Can't declare identity proof for the request that's completed response", "")
	}

	// Check IsClosed
	if request.IsClosed {
		return ReturnDeliverTxLog(code.RequestIsClosed, "Can't declare identity proof for the request that's closed", "")
	}

	// Check IsTimedOut
	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.RequestIsTimedOut, "Can't declare identity proof for the request that's timed out", "")
	}

	identityProofKey := "IdentityProof" + "|" + funcParam.RequestID + "|" + nodeID
	_, identityProofValue := app.state.db.Get(prefixKey([]byte(identityProofKey)))
	if identityProofValue == nil {
		identityProofValue := funcParam.IdentityProof
		app.SetStateDB([]byte(identityProofKey), []byte(identityProofValue))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.DuplicateIdentityProof, "Duplicate Identity Proof", "")
}

func disableMsqDestination(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableMsqDestination, Parameter: %s", param)
	var funcParam DisableMsqDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	msqDesKey := "MsqDestination" + "|" + funcParam.HashID
	_, msqDesValue := app.state.db.Get(prefixKey([]byte(msqDesKey)))

	if msqDesValue != nil {
		var nodes []Node
		err = json.Unmarshal([]byte(msqDesValue), &nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index := range nodes {
			if nodes[index].NodeID == nodeID {
				nodes[index].Active = false
				break
			}
		}

		msqDesJSON, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(msqDesKey), []byte(msqDesJSON))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.HashIDNotFound, "Hash ID not found", "")
}

func disableAccessorMethod(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableAccessorMethod, Parameter: %s", param)
	var funcParam DisableAccessorMethodParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	_, accessorValue := app.state.db.Get(prefixKey([]byte(accessorKey)))

	if accessorValue != nil {
		var accessor Accessor
		err = json.Unmarshal([]byte(accessorValue), &accessor)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// check owner of accessor
		if accessor.Owner != nodeID {
			return ReturnDeliverTxLog(code.NotOwnerOfAccessor, "This node is not owner of this accessor", "")
		}

		accessor.Active = false
		accessorJSON, err := json.Marshal(accessor)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		app.SetStateDB([]byte(accessorKey), []byte(accessorJSON))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}

	return ReturnDeliverTxLog(code.AccessorIDNotFound, "Accessor ID not found", "")
}

func enableMsqDestination(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableMsqDestination, Parameter: %s", param)
	var funcParam DisableMsqDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	msqDesKey := "MsqDestination" + "|" + funcParam.HashID
	_, msqDesValue := app.state.db.Get(prefixKey([]byte(msqDesKey)))

	if msqDesValue != nil {
		var nodes []Node
		err = json.Unmarshal([]byte(msqDesValue), &nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		for index := range nodes {
			if nodes[index].NodeID == nodeID {
				nodes[index].Active = true
				break
			}
		}

		msqDesJSON, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(msqDesKey), []byte(msqDesJSON))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}
	return ReturnDeliverTxLog(code.HashIDNotFound, "Hash ID not found", "")
}

func enableAccessorMethod(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableAccessorMethod, Parameter: %s", param)
	var funcParam DisableAccessorMethodParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	accessorKey := "Accessor" + "|" + funcParam.AccessorID
	_, accessorValue := app.state.db.Get(prefixKey([]byte(accessorKey)))

	if accessorValue != nil {
		var accessor Accessor
		err = json.Unmarshal([]byte(accessorValue), &accessor)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// check owner of accessor
		if accessor.Owner != nodeID {
			return ReturnDeliverTxLog(code.NotOwnerOfAccessor, "This node is not owner of this accessor", "")
		}

		accessor.Active = true
		accessorJSON, err := json.Marshal(accessor)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}

		app.SetStateDB([]byte(accessorKey), []byte(accessorJSON))
		return ReturnDeliverTxLog(code.OK, "success", "")
	}

	return ReturnDeliverTxLog(code.AccessorIDNotFound, "Accessor ID not found", "")
}
