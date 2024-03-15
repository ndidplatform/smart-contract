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

var regulatorMethod = map[string]bool{
	"InitNDID":                         true,
	"RegisterNode":                     true,
	"AddNodeToken":                     true,
	"ReduceNodeToken":                  true,
	"SetNodeToken":                     true,
	"SetPriceFunc":                     true,
	"AddNamespace":                     true,
	"DisableNamespace":                 true,
	"SetValidator":                     true,
	"AddService":                       true,
	"DisableService":                   true,
	"UpdateNodeByNDID":                 true,
	"UpdateService":                    true,
	"RegisterServiceDestinationByNDID": true,
	"DisableNode":                      true,
	"DisableServiceDestinationByNDID":  true,
	"EnableNode":                       true,
	"EnableServiceDestinationByNDID":   true,
	"EnableNamespace":                  true,
	"EnableService":                    true,
	"SetTimeOutBlockRegisterIdentity":  true,
	"AddNodeToProxyNode":               true,
	"UpdateNodeProxyNode":              true,
	"RemoveNodeFromProxyNode":          true,
	"AddErrorCode":                     true,
	"RemoveErrorCode":                  true,
	"SetInitData":                      true,
	"EndInit":                          true,
	"SetLastBlock":                     true,
	"SetSupportedIALList":              true,
	"SetSupportedAALList":              true,
	"SetAllowedModeList":               true,
	"UpdateNamespace":                  true,
	"SetAllowedMinIalForRegisterIdentityAtFirstIdp":        true,
	"SetServicePriceCeiling":                               true,
	"SetServicePriceMinEffectiveDatetimeDelay":             true,
	"AddRequestType":                                       true,
	"RemoveRequestType":                                    true,
	"AddSuppressedIdentityModificationNotificationNode":    true,
	"RemoveSuppressedIdentityModificationNotificationNode": true,
	"AddAllowedNodeSupportedFeature":                       true,
	"RemoveAllowedNodeSupportedFeature":                    true,
}
