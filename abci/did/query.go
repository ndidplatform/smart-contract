/*
Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED 

This file is part of NDID software.

NDID is the free software: you can redistribute it and/or modify  it under the terms of the Affero GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or any later version.

NDID is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the Affero GNU General Public License for more details.

You should have received a copy of the Affero GNU General Public License along with the NDID source code.  If not, see https://www.gnu.org/licenses/agpl.txt.

please contact info@ndid.co.th for any further questions
*/

package did

import (
	"reflect"

	"github.com/tendermint/abci/types"
)

// ReturnQuery return types.ResponseQuery
func ReturnQuery(value []byte, log string, height int64, app *DIDApplication) types.ResponseQuery {
	app.logger.Infof("Query reult: %s", string(value))
	var res types.ResponseQuery
	res.Value = value
	res.Log = log
	res.Height = height
	return res
}

// QueryRouter is Pointer to function
func QueryRouter(method string, param string, app *DIDApplication) types.ResponseQuery {
	funcs := map[string]interface{}{
		"GetNodePublicKey":             getNodePublicKey,
		"GetIdpNodes":                  getIdpNodes,
		"GetRequest":                   getRequest,
		"GetRequestDetail":             getRequestDetail,
		"GetAsNodesByServiceId":        getAsNodesByServiceId,
		"GetMsqAddress":                getMsqAddress,
		"GetNodeToken":                 getNodeToken,
		"GetPriceFunc":                 getPriceFunc,
		"GetUsedTokenReport":           getUsedTokenReport,
		"GetServiceDetail":             getServiceDetail,
		"GetNamespaceList":             getNamespaceList,
		"CheckExistingIdentity":        checkExistingIdentity,
		"GetAccessorGroupID":           getAccessorGroupID,
		"GetAccessorKey":               getAccessorKey,
		"GetServiceList":               getServiceList,
		"GetNodeMasterPublicKey":       getNodeMasterPublicKey,
		"GetNodeInfo":                  getNodeInfo,
		"CheckExistingAccessorID":      checkExistingAccessorID,
		"CheckExistingAccessorGroupID": checkExistingAccessorGroupID,
		"GetIdentityInfo":              getIdentityInfo,
		"GetDataSignature":             getDataSignature,
		"GetIdentityProof":             getIdentityProof,
	}
	value, _ := callQuery(funcs, method, param, app)
	return value[0].Interface().(types.ResponseQuery)
}

func callQuery(m map[string]interface{}, name string, param string, app *DIDApplication) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(param)
	in[1] = reflect.ValueOf(app)
	result = f.Call(in)
	return
}
