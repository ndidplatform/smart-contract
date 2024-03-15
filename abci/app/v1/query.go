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

package app

import (
	"github.com/tendermint/tendermint/abci/types"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
)

// ReturnQuery return types.ResponseQuery
func (app *ABCIApplication) ReturnQuery(value []byte, log string, height int64) types.ResponseQuery {
	app.logger.Infof("Query result: %s", string(value))
	var res types.ResponseQuery
	res.Value = value
	res.Log = log
	res.Height = height
	return res
}

// QueryRouter is Pointer to function
func (app *ABCIApplication) QueryRouter(method string, param []byte, height int64) types.ResponseQuery {
	result := app.callQuery(method, param, height)
	return result
}

func (app *ABCIApplication) callQuery(name string, param []byte, height int64) types.ResponseQuery {
	switch name {
	case "GetNodeSigningPublicKey":
		return app.getNodeSigningPublicKey(param)
	case "GetNodeEncryptionPublicKey":
		return app.getNodeEncryptionPublicKey(param)
	case "GetIdpNodes":
		return app.getIdpNodes(param)
	case "GetRequest":
		return app.getRequest(param, height)
	case "GetRequestDetail":
		return app.getRequestDetail(param, height, true)
	case "GetAsNodesByServiceId":
		return app.getAsNodesByServiceId(param)
	case "GetMqAddresses":
		return app.getMqAddresses(param)
	case "GetNodeToken":
		return app.getNodeToken(param, true)
	case "GetPriceFunc":
		return app.getPriceFunc(param, true)
	case "GetServiceDetail":
		return app.getServiceDetail(param)
	case "GetNamespaceList":
		return app.getNamespaceList(param)
	case "CheckExistingIdentity":
		return app.checkExistingIdentity(param)
	case "GetAccessorKey":
		return app.getAccessorKey(param)
	case "GetServiceList":
		return app.getServiceList(param)
	case "GetNodeSigningMasterPublicKey":
		return app.getNodeSigningMasterPublicKey(param)
	case "GetNodeInfo":
		return app.getNodeInfo(param)
	case "GetNodePublicKeyList":
		return app.getNodePublicKeyList(param)
	case "CheckExistingAccessorID":
		return app.checkExistingAccessorID(param)
	case "GetIdentityInfo":
		return app.getIdentityInfo(param)
	case "GetDataSignature":
		return app.getDataSignature(param)
	case "GetServicesByAsID":
		return app.getServicesByAsID(param)
	case "GetIdpNodesInfo":
		return app.getIdpNodesInfo(param)
	case "GetAsNodesInfoByServiceId":
		return app.getAsNodesInfoByServiceId(param)
	case "GetNodesBehindProxyNode":
		return app.getNodesBehindProxyNode(param)
	case "GetNodeIDList":
		return app.getNodeIDList(param)
	case "GetAccessorOwner":
		return app.getAccessorOwner(param)
	case "GetErrorCodeList":
		return app.getErrorCodeList(param)
	case "IsInitEnded":
		return app.isInitEnded(param)
	case "GetChainHistory":
		return app.getChainHistory(param)
	case "GetReferenceGroupCode":
		return app.GetReferenceGroupCode(param)
	case "GetReferenceGroupCodeByAccessorID":
		return app.GetReferenceGroupCodeByAccessorID(param)
	case "GetAllowedIALList":
		return app.GetAllowedIALList(param, true)
	case "GetAllowedAALList":
		return app.GetAllowedAALList(param, true)
	case "GetAllowedModeList":
		return app.GetAllowedModeList(param)
	case "GetAllowedMinIalForRegisterIdentityAtFirstIdp":
		return app.GetAllowedMinIalForRegisterIdentityAtFirstIdp(param)
	case "GetServicePriceList":
		return app.getServicePriceList(param)
	case "GetServicePriceCeiling":
		return app.getServicePriceCeiling(param)
	case "GetServicePriceMinEffectiveDatetimeDelay":
		return app.getServicePriceMinEffectiveDatetimeDelay(param)
	case "GetMessage":
		return app.getMessage(param, height)
	case "GetMessageDetail":
		return app.getMessageDetail(param, height, true)
	case "GetRequestTypeList":
		return app.getRequestTypeList(param, height)
	case "GetSuppressedIdentityModificationNotificationNodeList":
		return app.getSuppressedIdentityModificationNotificationNodeList(param, height)
	case "IsSuppressedIdentityModificationNotificationNode":
		return app.isSuppressedIdentityModificationNotificationNode(param, height)
	case "GetAllowedNodeSupportedFeatureList":
		return app.getAllowedNodeSupportedFeatureList(param, height)
	default:
		return types.ResponseQuery{Code: code.UnknownMethod, Log: "Unknown method name"}
	}
}
