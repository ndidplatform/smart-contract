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
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/tendermint/tendermint/abci/types"

	"github.com/ndidplatform/smart-contract/v5/abci/code"
	"github.com/ndidplatform/smart-contract/v5/abci/utils"
	data "github.com/ndidplatform/smart-contract/v5/protos/data"
)

func (app *ABCIApplication) setServicePrice(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetServicePrice, Parameter: %s", param)
	var funcParam SetServicePriceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// check if service ID exists
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	exists, err := app.state.Has([]byte(serviceKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if !exists {
		return app.ReturnDeliverTxLog(code.ServiceIDNotFound, "Service ID not found", "")
	}

	// check effective date/time if it's before latest block time + configured duration by NDID
	servicePriceMinEffectiveDatetimeDelayBytes, err := app.state.Get(servicePriceMinEffectiveDatetimeDelayKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var effectiveDatetimeMinDelayDuration time.Duration
	if servicePriceMinEffectiveDatetimeDelayBytes == nil {
		effectiveDatetimeMinDelayDuration = time.Duration(12 * time.Hour) // default 12 hrs
	} else {
		var servicePriceMinEffectiveDatetimeDelay data.ServicePriceMinEffectiveDatetimeDelay
		err = proto.Unmarshal([]byte(servicePriceMinEffectiveDatetimeDelayBytes), &servicePriceMinEffectiveDatetimeDelay)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		effectiveDatetimeMinDelayDuration = time.Duration(servicePriceMinEffectiveDatetimeDelay.DurationSecond) * time.Second
	}

	startingAllowedTime := app.lastBlockTime.Add(effectiveDatetimeMinDelayDuration)
	if funcParam.EffectiveDatetime.Before(startingAllowedTime) {
		return app.ReturnDeliverTxLog(code.ServicePriceEffectiveDatetimeBeforeAllowed, "Service price effective datetime is before allowed datetime", "")
	}

	// Get service's price ceiling
	servicePriceCeilingKey := servicePriceCeilingKeyPrefix + keySeparator + funcParam.ServiceID
	servicePriceCeilingListBytes, err := app.state.Get([]byte(servicePriceCeilingKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if servicePriceCeilingListBytes == nil {
		return app.ReturnDeliverTxLog(code.ServicePriceCeilingListNotFound, "Service price ceiling list not found", "")
	}
	var servicePriceCeilingList data.ServicePriceCeilingList
	err = proto.Unmarshal([]byte(servicePriceCeilingListBytes), &servicePriceCeilingList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	var servicePriceByCurrencyListToSet []*data.ServicePriceByCurrency = make([]*data.ServicePriceByCurrency, 0)

priceToSetLoop:
	for _, priceToSet := range funcParam.PriceByCurrencyList {
		for _, priceCeiling := range servicePriceCeilingList.PriceCeilingByCurrencyList {
			if priceCeiling.Currency == priceToSet.Currency {
				if priceToSet.MaxPrice >= 0 {
					if priceToSet.MaxPrice > priceCeiling.Price {
						return app.ReturnDeliverTxLog(code.ServiceMaxPriceGreaterThanPriceCeiling, "Service max price is greater than price ceiling", "")
					}
				}
				// less then 0 price ceiling means no ceiling/limit

				servicePriceByCurrencyListToSet = append(servicePriceByCurrencyListToSet, &data.ServicePriceByCurrency{
					Currency: priceToSet.Currency,
					MinPrice: priceToSet.MinPrice,
					MaxPrice: priceToSet.MaxPrice,
				})

				continue priceToSetLoop
			}
		}
		return app.ReturnDeliverTxLog(code.ServicePriceCeilingNotFound, "Service price ceiling not found", "")
	}

	var newServicePrice *data.ServicePrice = &data.ServicePrice{
		PriceByCurrencyList: servicePriceByCurrencyListToSet,
		EffectiveDatetime:   funcParam.EffectiveDatetime.Unix(),
		MoreInfoUrl:         funcParam.MoreInfoURL,
		Detail:              funcParam.Detail,
	}

	servicePriceListKey := servicePriceListKeyPrefix + keySeparator + nodeID + keySeparator + funcParam.ServiceID
	currentServicePriceListBytes, err := app.state.Get([]byte(servicePriceListKey), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}

	if currentServicePriceListBytes != nil {
		var currentServicePriceList data.ServicePriceList
		err = proto.Unmarshal([]byte(currentServicePriceListBytes), &currentServicePriceList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// prepend new service price to list
		currentServicePriceList.ServicePriceList = append([]*data.ServicePrice{newServicePrice}, currentServicePriceList.ServicePriceList...)

		newCurrentServicePriceListBytes, err := utils.ProtoDeterministicMarshal(&currentServicePriceList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(servicePriceListKey), []byte(newCurrentServicePriceListBytes))
	} else {

		var newServicePriceList data.ServicePriceList = data.ServicePriceList{
			ServicePriceList: []*data.ServicePrice{newServicePrice},
		}

		newCurrentServicePriceListBytes, err := utils.ProtoDeterministicMarshal(&newServicePriceList)
		if err != nil {
			return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(servicePriceListKey), []byte(newCurrentServicePriceListBytes))
	}

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) getServicePriceList(param string) types.ResponseQuery {
	app.logger.Infof("GetServicePriceList, Parameter: %s", param)
	var funcParam GetServicePriceListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	servicePriceListKey := servicePriceListKeyPrefix + keySeparator + funcParam.NodeID + keySeparator + funcParam.ServiceID
	servicePriceListBytes, err := app.state.Get([]byte(servicePriceListKey), false)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	if servicePriceListBytes == nil {
		var retVal GetServicePriceListResult
		retVal.ServicePriceList = make([]ServicePrice, 0)
		retValJSON, err := json.Marshal(retVal)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		return app.ReturnQuery(retValJSON, "success", app.state.Height)
	}

	var servicePriceList data.ServicePriceList
	err = proto.Unmarshal([]byte(servicePriceListBytes), &servicePriceList)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var retVal GetServicePriceListResult
	retVal.ServicePriceList = make([]ServicePrice, 0)
	for _, servicePrice := range servicePriceList.ServicePriceList {
		var retValServicePrice ServicePrice = ServicePrice{
			PriceByCurrencyList: make([]ServicePriceByCurrency, 0),
			EffectiveDatetime:   time.Unix(servicePrice.EffectiveDatetime, 0),
			MoreInfoURL:         servicePrice.MoreInfoUrl,
			Detail:              servicePrice.Detail,
		}

		for _, priceByCurrency := range servicePrice.PriceByCurrencyList {
			var retValServicePriceByCurrency ServicePriceByCurrency = ServicePriceByCurrency{
				Currency: priceByCurrency.Currency,
				MinPrice: priceByCurrency.MinPrice,
				MaxPrice: priceByCurrency.MaxPrice,
			}
			retValServicePrice.PriceByCurrencyList = append(retValServicePrice.PriceByCurrencyList, retValServicePriceByCurrency)
		}

		retVal.ServicePriceList = append(retVal.ServicePriceList, retValServicePrice)
	}
	retValJSON, err := json.Marshal(retVal)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(retValJSON, "success", app.state.Height)
}
