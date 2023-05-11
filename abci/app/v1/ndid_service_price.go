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

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v8/abci/code"
	"github.com/ndidplatform/smart-contract/v8/abci/utils"
	data "github.com/ndidplatform/smart-contract/v8/protos/data"
)

type SetServicePriceCeilingParam struct {
	ServiceID                  string                   `json:"service_id"`
	PriceCeilingByCurrencyList []PriceCeilingByCurrency `json:"price_ceiling_by_currency_list"`
}

type PriceCeilingByCurrency struct {
	Currency string  `json:"currency"`
	Price    float64 `json:"price"`
}

func (app *ABCIApplication) validateSetServicePriceCeiling(funcParam SetServicePriceCeilingParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	// check if service ID exists
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	exists, err := app.state.Has([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if !exists {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}

	return nil
}

func (app *ABCIApplication) setServicePriceCeilingCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetServicePriceCeilingParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetServicePriceCeiling(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setServicePriceCeiling(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetServicePriceCeiling, Parameter: %s", param)
	var funcParam SetServicePriceCeilingParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetServicePriceCeiling(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	// set/overwrite service's price ceiling list
	var priceCeilingByCurrencyList []*data.ServicePriceCeilingByCurency = make([]*data.ServicePriceCeilingByCurency, 0)

	for _, inputPriceCeiling := range funcParam.PriceCeilingByCurrencyList {
		var servicePriceCeilingByCurrency data.ServicePriceCeilingByCurency
		servicePriceCeilingByCurrency.Currency = inputPriceCeiling.Currency
		servicePriceCeilingByCurrency.Price = inputPriceCeiling.Price
		priceCeilingByCurrencyList = append(priceCeilingByCurrencyList, &servicePriceCeilingByCurrency)
	}

	servicePriceCeilingList := data.ServicePriceCeilingList{
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

type GetServicePriceCeilingParam struct {
	ServiceID string `json:"service_id"`
}

type GetServicePriceCeilingResult struct {
	PriceCeilingByCurrencyList []PriceCeilingByCurrency `json:"price_ceiling_by_currency_list"`
}

func (app *ABCIApplication) getServicePriceCeiling(param []byte) types.ResponseQuery {
	app.logger.Infof("GetServicePriceCeiling, Parameter: %s", param)
	var funcParam GetServicePriceCeilingParam
	err := json.Unmarshal(param, &funcParam)
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

type SetServicePriceMinEffectiveDatetimeDelayParam struct {
	ServiceID      string `json:"service_id"`
	DurationSecond uint32 `json:"duration_second"`
}

func (app *ABCIApplication) validateSetServicePriceMinEffectiveDatetimeDelay(funcParam SetServicePriceMinEffectiveDatetimeDelayParam, callerNodeID string, committedState bool) error {
	ok, err := app.isNDIDNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallNDIDMethod,
			Message: "This node does not have permission to call NDID method",
		}
	}

	if funcParam.ServiceID != "" {
		// check if service ID exists
		serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
		exists, err := app.state.Has([]byte(serviceKey), false)
		if err != nil {
			return &ApplicationError{
				Code:    code.AppStateError,
				Message: err.Error(),
			}
		}
		if !exists {
			return &ApplicationError{
				Code:    code.ServiceIDNotFound,
				Message: "Service ID not found",
			}
		}
	}

	return nil
}

func (app *ABCIApplication) setServicePriceMinEffectiveDatetimeDelayCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetServicePriceMinEffectiveDatetimeDelayParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetServicePriceMinEffectiveDatetimeDelay(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setServicePriceMinEffectiveDatetimeDelay(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetServicePriceMinEffectiveDatetimeDelay, Parameter: %s", param)
	var funcParam SetServicePriceMinEffectiveDatetimeDelayParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetServicePriceMinEffectiveDatetimeDelay(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	servicePriceMinEffectiveDatetimeDelay := data.ServicePriceMinEffectiveDatetimeDelay{
		DurationSecond: funcParam.DurationSecond,
	}
	servicePriceMinEffectiveDatetimeDelayBytes, err := utils.ProtoDeterministicMarshal(&servicePriceMinEffectiveDatetimeDelay)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	if funcParam.ServiceID != "" {
		key := servicePriceMinEffectiveDatetimeDelayKeyPrefix + keySeparator + funcParam.ServiceID
		app.state.Set([]byte(key), servicePriceMinEffectiveDatetimeDelayBytes)
	} else {
		// global / fallback from specific service ID
		app.state.Set(servicePriceMinEffectiveDatetimeDelayKeyBytes, servicePriceMinEffectiveDatetimeDelayBytes)
	}

	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type GetServicePriceMinEffectiveDatetimeDelayParam struct {
	ServiceID string `json:"service_id"`
}

type GetServicePriceMinEffectiveDatetimeDelayResult struct {
	DurationSecond uint32 `json:"duration_second"`
}

func (app *ABCIApplication) getServicePriceMinEffectiveDatetimeDelay(param []byte) types.ResponseQuery {
	app.logger.Infof("GetServicePriceMinEffectiveDatetimeDelay, Parameter: %s", param)
	var funcParam GetServicePriceMinEffectiveDatetimeDelayParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var servicePriceMinEffectiveDatetimeDelayBytes []byte
	if funcParam.ServiceID != "" {
		serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
		serviceExists, err := app.state.Has([]byte(serviceKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if !serviceExists {
			return app.ReturnQuery(nil, "not found", app.state.Height)
		}

		key := servicePriceMinEffectiveDatetimeDelayKeyPrefix + keySeparator + funcParam.ServiceID
		servicePriceMinEffectiveDatetimeDelayBytes, err = app.state.Get([]byte(key), false)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}

		// get global / fallback from specific service ID
		if servicePriceMinEffectiveDatetimeDelayBytes == nil {
			servicePriceMinEffectiveDatetimeDelayBytes, err = app.state.Get(servicePriceMinEffectiveDatetimeDelayKeyBytes, false)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
		}
	} else {
		servicePriceMinEffectiveDatetimeDelayBytes, err = app.state.Get(servicePriceMinEffectiveDatetimeDelayKeyBytes, false)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
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
