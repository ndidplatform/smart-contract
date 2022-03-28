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
	"encoding/json"
	"strings"

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

var modeFunctionMap = map[string]bool{
	"RegisterIdentity":          true,
	"AddIdentity":               true,
	"AddAccessor":               true,
	"RevokeAccessor":            true,
	"RevokeIdentityAssociation": true,
	"UpdateIdentityModeList":    true,
	"RevokeAndAddAccessor":      true,
}

var (
	masterNDIDKeyBytes                            = []byte("MasterNDID")
	initStateKeyBytes                             = []byte("InitState")
	lastBlockKeyBytes                             = []byte("lastBlock")
	idpListKeyBytes                               = []byte("IdPList")
	allNamespaceKeyBytes                          = []byte("AllNamespace")
	servicePriceMinEffectiveDatetimeDelayKeyBytes = []byte("ServicePriceMinEffectiveDatetimeDelay")
)

const (
	keySeparator                                         = "|"
	nodeIDKeyPrefix                                      = "NodeID"
	behindProxyNodeKeyPrefix                             = "BehindProxyNode"
	tokenKeyPrefix                                       = "Token"
	tokenPriceFuncKeyPrefix                              = "TokenPriceFunc"
	serviceKeyPrefix                                     = "Service"
	serviceDestinationKeyPrefix                          = "ServiceDestination"
	approvedServiceKeyPrefix                             = "ApproveKey"
	providedServicesKeyPrefix                            = "ProvideService"
	refGroupCodeKeyPrefix                                = "RefGroupCode"
	identityToRefCodeKeyPrefix                           = "identityToRefCodeKey"
	accessorToRefCodeKeyPrefix                           = "accessorToRefCodeKey"
	allowedModeListKeyPrefix                             = "AllowedModeList"
	requestKeyPrefix                                     = "Request"
	messageKeyPrefix                                     = "Message"
	dataSignatureKeyPrefix                               = "SignData"
	errorCodeKeyPrefix                                   = "ErrorCode"
	errorCodeListKeyPrefix                               = "ErrorCodeList"
	servicePriceCeilingKeyPrefix                         = "ServicePriceCeiling"
	servicePriceMinEffectiveDatetimeDelayKeyPrefix       = "ServicePriceMinEffectiveDatetimeDelay"
	servicePriceListKeyPrefix                            = "ServicePriceListKey"
	requestTypeKeyPrefix                                 = "RequestType"
	suppressedIdentityModificationNotificationNodePrefix = "SuppressedIdentityModificationNotificationNode"
)

func (app *ABCIApplication) getServiceDetail(param string) types.ResponseQuery {
	app.logger.Infof("GetServiceDetail, Parameter: %s", param)
	var funcParam GetServiceDetailParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		value = []byte("{}")
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal(value, &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	returnValue, err := json.Marshal(&service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) checkExistingIdentity(param string) types.ResponseQuery {
	app.logger.Infof("CheckExistingIdentity, Parameter: %s", param)
	var funcParam CheckExistingIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result CheckExistingIdentityResult
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "Found reference group code and identity detail in parameter", app.state.Height)
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if refGroupCodeFromDB == nil {
			returnValue, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			return app.ReturnQuery(returnValue, "success", app.state.Height)
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "success", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "success", app.state.Height)
	}
	result.Exist = true
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getAccessorKey(param string) types.ResponseQuery {
	app.logger.Infof("GetAccessorKey, Parameter: %s", param)
	var funcParam GetAccessorKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAccessorKeyResult
	result.AccessorPublicKey = ""
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.AccessorPublicKey = accessor.AccessorPublicKey
				result.Active = accessor.Active
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServiceList(param string) types.ResponseQuery {
	app.logger.Infof("GetServiceList, Parameter: %s", param)
	key := "AllService"
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		result := make([]ServiceDetail, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	result := make([]*data.ServiceDetail, 0)
	// filter flag==true
	var services data.ServiceDetailList
	err = proto.Unmarshal([]byte(value), &services)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for _, service := range services.Services {
		if service.Active {
			result = append(result, service)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServiceNameByServiceID(serviceID string) string {
	key := serviceKeyPrefix + keySeparator + serviceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		panic(err)
	}
	if value == nil {
		return ""
	}
	var result ServiceDetail
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		return ""
	}
	return result.ServiceName
}

func (app *ABCIApplication) checkExistingAccessorID(param string) types.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorID, Parameter: %s", param)
	var funcParam CheckExistingAccessorIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result CheckExistingResult
	result.Exist = false
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.Exist = true
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getIdentityInfo(param string) types.ResponseQuery {
	app.logger.Infof("GetIdentityInfo, Parameter: %s", param)
	var funcParam GetIdentityInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetIdentityInfoResult
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "Found reference group code and identity detail in parameter", app.state.Height)
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if refGroupCodeFromDB == nil {
			returnValue, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			return app.ReturnQuery(returnValue, "Reference group not found", app.state.Height)
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "Reference group not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "Reference group not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		if funcParam.NodeID == idp.NodeId && idp.Active {
			result.Ial = idp.Ial
			if idp.Lial != nil {
				result.Lial = &idp.Lial.Value
			}
			if idp.Laal != nil {
				result.Laal = &idp.Laal.Value
			}
			result.ModeList = idp.Mode
			break
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if result.Ial <= 0.0 {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getDataSignature(param string) types.ResponseQuery {
	app.logger.Infof("GetDataSignature, Parameter: %s", param)
	var funcParam GetDataSignatureParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	signDataKey := dataSignatureKeyPrefix + keySeparator + funcParam.NodeID + keySeparator + funcParam.ServiceID + keySeparator + funcParam.RequestID
	signDataValue, err := app.state.Get([]byte(signDataKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if signDataValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var result GetDataSignatureResult
	result.Signature = string(signDataValue)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServicesByAsID(param string) types.ResponseQuery {
	app.logger.Infof("GetServicesByAsID, Parameter: %s", param)
	var funcParam GetServicesByAsIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetServicesByAsIDResult
	result.Services = make([]Service, 0)
	provideServiceKey := providedServicesKeyPrefix + keySeparator + funcParam.AsID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if provideServiceValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	var services data.ServiceList
	err = proto.Unmarshal([]byte(provideServiceValue), &services)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.AsID
	nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if nodeDetailValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for index, provideService := range services.Services {
		serviceKey := serviceKeyPrefix + keySeparator + provideService.ServiceId
		serviceValue, err := app.state.Get([]byte(serviceKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if serviceValue == nil {
			continue
		}
		var service data.ServiceDetail
		err = proto.Unmarshal([]byte(serviceValue), &service)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if nodeDetail.Active && service.Active {
			// Set suspended from NDID
			approveServiceKey := approvedServiceKeyPrefix + keySeparator + provideService.ServiceId + keySeparator + funcParam.AsID
			approveServiceJSON, err := app.state.Get([]byte(approveServiceKey), true)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			if approveServiceJSON == nil {
				continue
			}
			var approveService data.ApproveService
			err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
			if err == nil {
				services.Services[index].Suspended = !approveService.Active
			}
			var newRow Service
			newRow.Active = services.Services[index].Active
			newRow.MinAal = services.Services[index].MinAal
			newRow.MinIal = services.Services[index].MinIal
			newRow.ServiceID = services.Services[index].ServiceId
			newRow.Suspended = services.Services[index].Suspended
			newRow.SupportedNamespaceList = services.Services[index].SupportedNamespaceList
			result.Services = append(result.Services, newRow)
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.Services) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getAccessorOwner(param string) types.ResponseQuery {
	app.logger.Infof("GetAccessorOwner, Parameter: %s", param)
	var funcParam GetAccessorOwnerParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAccessorOwnerResult
	result.NodeID = ""
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, err := app.state.Get([]byte(refGroupKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.NodeID = idp.NodeId
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetReferenceGroupCode(param string) types.ResponseQuery {
	app.logger.Infof("GetReferenceGroupCode, Parameter: %s", param)
	var funcParam GetReferenceGroupCodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
	refGroupCodeFromDB, err := app.state.Get([]byte(identityToRefCodeKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		refGroupCodeFromDB = []byte("")
	}
	var result GetReferenceGroupCodeResult
	result.ReferenceGroupCode = string(refGroupCodeFromDB)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if string(refGroupCodeFromDB) == "" {
		return app.ReturnQuery(returnValue, "not found", app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetReferenceGroupCodeByAccessorID(param string) types.ResponseQuery {
	app.logger.Infof("GetReferenceGroupCodeByAccessorID, Parameter: %s", param)
	var funcParam GetReferenceGroupCodeByAccessorIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, err := app.state.Get([]byte(accessorToRefCodeKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if refGroupCodeFromDB == nil {
		refGroupCodeFromDB = []byte("")
	}
	var result GetReferenceGroupCodeResult
	result.ReferenceGroupCode = string(refGroupCodeFromDB)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetAllowedModeList(param string) types.ResponseQuery {
	app.logger.Infof("GetAllowedModeList, Parameter: %s", param)
	var funcParam GetAllowedModeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAllowedModeListResult
	result.AllowedModeList = app.GetAllowedModeFromStateDB(funcParam.Purpose, true)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetAllowedModeFromStateDB(purpose string, committedState bool) (result []int32) {
	allowedModeKey := "AllowedModeList" + keySeparator + purpose
	var allowedModeList data.AllowedModeList
	allowedModeValue, err := app.state.Get([]byte(allowedModeKey), committedState)
	if err != nil {
		return nil
	}
	if allowedModeValue == nil {
		// return default value
		if !modeFunctionMap[purpose] {
			result = append(result, 1)
		}
		result = append(result, 2)
		result = append(result, 3)
		return result
	}
	err = proto.Unmarshal(allowedModeValue, &allowedModeList)
	if err != nil {
		return result
	}
	result = allowedModeList.Mode
	return result
}

func (app *ABCIApplication) GetAllowedMinIalForRegisterIdentityAtFirstIdp(param string) types.ResponseQuery {
	app.logger.Infof("GetAllowedMinIalForRegisterIdentityAtFirstIdp, Parameter: %s", param)
	var result GetAllowedMinIalForRegisterIdentityAtFirstIdpResult
	result.MinIal = app.GetAllowedMinIalForRegisterIdentityAtFirstIdpFromStateDB(true)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetAllowedMinIalForRegisterIdentityAtFirstIdpFromStateDB(committedState bool) float64 {
	allowedMinIalKey := "AllowedMinIalForRegisterIdentityAtFirstIdp"
	var allowedMinIal data.AllowedMinIalForRegisterIdentityAtFirstIdp
	allowedMinIalValue, err := app.state.Get([]byte(allowedMinIalKey), committedState)
	if err != nil {
		return 0
	}
	if allowedMinIalValue == nil {
		return 0
	}
	err = proto.Unmarshal(allowedMinIalValue, &allowedMinIal)
	if err != nil {
		return 0
	}
	return allowedMinIal.MinIal
}

func (app *ABCIApplication) getErrorCodeList(param string) types.ResponseQuery {
	var funcParam GetErrorCodeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	// convert funcParam to lowercase and fetch the code list
	funcParam.Type = strings.ToLower(funcParam.Type)
	errorCodeListKey := errorCodeListKeyPrefix + keySeparator + funcParam.Type
	errorCodeListBytes, err := app.state.Get([]byte(errorCodeListKey), false)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var errorCodeList data.ErrorCodeList
	err = proto.Unmarshal(errorCodeListBytes, &errorCodeList)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	// parse result into response format
	result := make([]*GetErrorCodeListResult, 0, len(errorCodeList.ErrorCode))
	for _, errorCode := range errorCodeList.ErrorCode {
		result = append(result, &GetErrorCodeListResult{
			ErrorCode:   errorCode.ErrorCode,
			Description: errorCode.Description,
		})
	}

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}
