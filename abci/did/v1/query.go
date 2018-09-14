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
	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/tendermint/abci/types"
)

// ReturnQuery return types.ResponseQuery
func ReturnQuery(value []byte, log string, height int64, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("Query result: %s", string(value))
	var res types.ResponseQuery
	res.Value = value
	res.Log = log
	res.Height = height
	return res
}

// QueryRouter is Pointer to function
func QueryRouter(method string, param string, app *DIDApplication, height int64) types.ResponseQuery {
	result := callQuery(method, param, app, height)
	return result
}

func callQuery(name string, param string, app *DIDApplication, height int64) types.ResponseQuery {
	switch name {
	case "GetNodePublicKey":
		return getNodePublicKey(param, app, height)
	case "GetIdpNodes":
		return getIdpNodes(param, app, height)
	case "GetRequest":
		return getRequest(param, app, height)
	case "GetRequestDetail":
		return getRequestDetail(param, app, height)
	case "GetAsNodesByServiceId":
		return getAsNodesByServiceId(param, app, height)
	case "GetMsqAddress":
		return getMsqAddress(param, app, height)
	case "GetNodeToken":
		return getNodeToken(param, app, height)
	case "GetPriceFunc":
		return getPriceFunc(param, app, height)
	case "GetUsedTokenReport":
		return getUsedTokenReport(param, app, height)
	case "GetServiceDetail":
		return getServiceDetail(param, app, height)
	case "GetNamespaceList":
		return getNamespaceList(param, app, height)
	case "CheckExistingIdentity":
		return checkExistingIdentity(param, app, height)
	case "GetAccessorGroupID":
		return getAccessorGroupID(param, app, height)
	case "GetAccessorKey":
		return getAccessorKey(param, app, height)
	case "GetServiceList":
		return getServiceList(param, app, height)
	case "GetNodeMasterPublicKey":
		return getNodeMasterPublicKey(param, app, height)
	case "GetNodeInfo":
		return getNodeInfo(param, app, height)
	case "CheckExistingAccessorID":
		return checkExistingAccessorID(param, app, height)
	case "CheckExistingAccessorGroupID":
		return checkExistingAccessorGroupID(param, app, height)
	case "GetIdentityInfo":
		return getIdentityInfo(param, app, height)
	case "GetDataSignature":
		return getDataSignature(param, app, height)
	case "GetIdentityProof":
		return getIdentityProof(param, app, height)
	case "GetServicesByAsID":
		return getServicesByAsID(param, app, height)
	case "GetIdpNodesInfo":
		return getIdpNodesInfo(param, app, height)
	case "GetAsNodesInfoByServiceId":
		return getAsNodesInfoByServiceId(param, app, height)
	case "GetNodesBehindProxyNode":
		return getNodesBehindProxyNode(param, app, height)
	default:
		return types.ResponseQuery{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}
