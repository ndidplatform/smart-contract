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

	"github.com/ndidplatform/smart-contract/v6/abci/code"
	"github.com/ndidplatform/smart-contract/v6/abci/utils"
	data "github.com/ndidplatform/smart-contract/v6/protos/data"
)

func (app *ABCIApplication) setServicePriceCeiling(param string) types.ResponseDeliverTx {
	app.logger.Infof("SetServicePriceCeiling, Parameter: %s", param)
	var funcParam SetServicePriceCeilingParam
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

	// set/overwrite service's price ceiling list
	var priceCeilingByCurrencyList []*data.ServicePriceCeilingByCurency = make([]*data.ServicePriceCeilingByCurency, 0)

	for _, inputPriceCeiling := range funcParam.PriceCeilingByCurrencyList {
		var servicePriceCeilingByCurrency data.ServicePriceCeilingByCurency
		servicePriceCeilingByCurrency.Currency = inputPriceCeiling.Currency
		servicePriceCeilingByCurrency.Price = inputPriceCeiling.Price
		priceCeilingByCurrencyList = append(priceCeilingByCurrencyList, &servicePriceCeilingByCurrency)
	}

	var servicePriceCeilingList data.ServicePriceCeilingList
	servicePriceCeilingList = data.ServicePriceCeilingList{
		PriceCeilingByCurrencyList: priceCeilingByCurrencyList,
	}

	servicePriceCeilingListBytes, err := utils.ProtoDeterministicMarshal(&servicePriceCeilingList)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	servicePriceCeilingKey := servicePriceCeilingKeyPrefix + keySeparator + funcParam.ServiceID
	app.state.Set([]byte(servicePriceCeilingKey), []byte(servicePriceCeilingListBytes))

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) getServicePriceCeiling(param string) types.ResponseQuery {
	app.logger.Infof("GetServicePriceCeiling, Parameter: %s", param)
	var funcParam GetServicePriceCeilingParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceExists, err := app.state.Has([]byte(serviceKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if !serviceExists {
		return app.ReturnQuery(nil, "not found", app.state.Height)
	}

	servicePriceCeilingKey := servicePriceCeilingKeyPrefix + keySeparator + funcParam.ServiceID
	servicePriceCeilingBytes, err := app.state.Get([]byte(servicePriceCeilingKey), false)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	if servicePriceCeilingBytes == nil {
		var retVal GetServicePriceCeilingResult
		retVal.PriceCeilingByCurrencyList = make([]PriceCeilingByCurrency, 0)
		retValJSON, err := json.Marshal(retVal)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		return app.ReturnQuery(retValJSON, "success", app.state.Height)
	}

	var servicePriceCelingList data.ServicePriceCeilingList
	err = proto.Unmarshal([]byte(servicePriceCeilingBytes), &servicePriceCelingList)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var retVal GetServicePriceCeilingResult
	retVal.PriceCeilingByCurrencyList = make([]PriceCeilingByCurrency, 0)
	for _, priceByCurrency := range servicePriceCelingList.PriceCeilingByCurrencyList {
		var retValPriceCeilingByCurrency PriceCeilingByCurrency = PriceCeilingByCurrency{
			Currency: priceByCurrency.Currency,
			Price:    priceByCurrency.Price,
		}

		retVal.PriceCeilingByCurrencyList = append(retVal.PriceCeilingByCurrencyList, retValPriceCeilingByCurrency)
	}
	retValJSON, err := json.Marshal(retVal)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(retValJSON, "success", app.state.Height)
}

func (app *ABCIApplication) setServicePriceMinEffectiveDatetimeDelay(param string) types.ResponseDeliverTx {
	app.logger.Infof("SetServicePriceMinEffectiveDatetimeDelay, Parameter: %s", param)
	var funcParam SetServicePriceMinEffectiveDatetimeDelayParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	var servicePriceMinEffectiveDatetimeDelay data.ServicePriceMinEffectiveDatetimeDelay
	servicePriceMinEffectiveDatetimeDelay = data.ServicePriceMinEffectiveDatetimeDelay{
		DurationSecond: funcParam.DurationSecond,
	}
	servicePriceMinEffectiveDatetimeDelayBytes, err := utils.ProtoDeterministicMarshal(&servicePriceMinEffectiveDatetimeDelay)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.state.Set(servicePriceMinEffectiveDatetimeDelayKeyBytes, servicePriceMinEffectiveDatetimeDelayBytes)

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) getServicePriceMinEffectiveDatetimeDelay(param string) types.ResponseQuery {
	app.logger.Infof("GetServicePriceMinEffectiveDatetimeDelay, Parameter: %s", param)

	servicePriceMinEffectiveDatetimeDelayBytes, err := app.state.Get(servicePriceMinEffectiveDatetimeDelayKeyBytes, false)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	if servicePriceMinEffectiveDatetimeDelayBytes == nil {
		var retVal GetServicePriceMinEffectiveDatetimeDelayResult
		retVal.DurationSecond = uint32(12 * time.Hour / time.Second) // return default (12 hrs)
		retValJSON, err := json.Marshal(retVal)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		return app.ReturnQuery(retValJSON, "success", app.state.Height)
	}

	var servicePriceMinEffectiveDatetimeDelay data.ServicePriceMinEffectiveDatetimeDelay
	err = proto.Unmarshal([]byte(servicePriceMinEffectiveDatetimeDelayBytes), &servicePriceMinEffectiveDatetimeDelay)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var retVal GetServicePriceMinEffectiveDatetimeDelayResult
	retVal.DurationSecond = servicePriceMinEffectiveDatetimeDelay.DurationSecond
	retValJSON, err := json.Marshal(retVal)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(retValJSON, "success", app.state.Height)
}
