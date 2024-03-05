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
	nonceKeyPrefix                                       = "n"
	nodeIDKeyPrefix                                      = "NodeID"
	nodeKeyKeyPrefix                                     = "NodeKey"
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
	validatorKeyPrefix                                   = "Validator"
)
