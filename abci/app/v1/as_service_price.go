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

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type SetServicePriceParam struct {
	ServiceID           string                   `json:"service_id"`
	PriceByCurrencyList []ServicePriceByCurrency `json:"price_by_currency_list"`
	EffectiveDatetime   time.Time                `json:"effective_datetime"`
	MoreInfoURL         string                   `json:"more_info_url"`
	Detail              string                   `json:"detail"`
}

type ServicePriceByCurrency struct {
	Currency string  `json:"currency"`
	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
}

func (app *ABCIApplication) validateSetServicePrice(funcParam SetServicePriceParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
	ok, err := app.isASNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallASMethod,
			Message: "This node does not have permission to call AS method",
		}
	}

	// stateless

	for _, priceToSet := range funcParam.PriceByCurrencyList {
		if priceToSet.MinPrice > priceToSet.MaxPrice {
			return &ApplicationError{
				Code:    code.ServicePriceMinCannotBeGreaterThanMax,
				Message: "Service minimum price cannot be greater than maximum price",
			}
		}
	}

	if checktx {
		return nil
	}

	// stateful

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

	servicePriceCeilingKey := servicePriceCeilingKeyPrefix + keySeparator + funcParam.ServiceID
	servicePriceCeilingListBytes, err := app.state.Get([]byte(servicePriceCeilingKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if servicePriceCeilingListBytes == nil {
		return &ApplicationError{
			Code:    code.ServicePriceCeilingListNotFound,
			Message: "Service price ceiling list not found",
		}
	}
	var servicePriceCeilingList data.ServicePriceCeilingList
	err = proto.Unmarshal([]byte(servicePriceCeilingListBytes), &servicePriceCeilingList)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

priceToSetLoop:
	for _, priceToSet := range funcParam.PriceByCurrencyList {
		for _, priceCeiling := range servicePriceCeilingList.PriceCeilingByCurrencyList {
			if priceCeiling.Currency == priceToSet.Currency {
				if priceToSet.MaxPrice >= 0 {
					if priceToSet.MaxPrice > priceCeiling.Price {
						return &ApplicationError{
							Code:    code.ServiceMaxPriceGreaterThanPriceCeiling,
							Message: "Service max price is greater than price ceiling",
						}
					}
				}
				// less then 0 price ceiling means no ceiling/limit

				if priceToSet.MinPrice > priceToSet.MaxPrice {
					return &ApplicationError{
						Code:    code.ServicePriceMinCannotBeGreaterThanMax,
						Message: "Service minimum price cannot be greater than maximum price",
					}
				}

				continue priceToSetLoop
			}
		}
		return &ApplicationError{
			Code:    code.ServicePriceCeilingNotFound,
			Message: "Service price ceiling not found",
		}
	}

	return nil
}

func (app *ABCIApplication) setServicePriceCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam SetServicePriceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetServicePrice(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setServicePrice(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetServicePrice, Parameter: %s", param)
	var funcParam SetServicePriceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetServicePrice(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Get service's price ceiling
	servicePriceCeilingKey := servicePriceCeilingKeyPrefix + keySeparator + funcParam.ServiceID
	servicePriceCeilingListBytes, err := app.state.Get([]byte(servicePriceCeilingKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var servicePriceCeilingList data.ServicePriceCeilingList
	err = proto.Unmarshal([]byte(servicePriceCeilingListBytes), &servicePriceCeilingList)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	var servicePriceByCurrencyListToSet []*data.ServicePriceByCurrency = make([]*data.ServicePriceByCurrency, 0)

priceToSetLoop:
	for _, priceToSet := range funcParam.PriceByCurrencyList {
		for _, priceCeiling := range servicePriceCeilingList.PriceCeilingByCurrencyList {
			if priceCeiling.Currency == priceToSet.Currency {
				servicePriceByCurrencyListToSet = append(servicePriceByCurrencyListToSet, &data.ServicePriceByCurrency{
					Currency: priceToSet.Currency,
					MinPrice: priceToSet.MinPrice,
					MaxPrice: priceToSet.MaxPrice,
				})

				continue priceToSetLoop
			}
		}
	}

	var newServicePrice *data.ServicePrice = &data.ServicePrice{
		PriceByCurrencyList: servicePriceByCurrencyListToSet,
		EffectiveDatetime:   funcParam.EffectiveDatetime.UnixMilli(), // in milliseconds
		MoreInfoUrl:         funcParam.MoreInfoURL,
		Detail:              funcParam.Detail,
		CreationBlockHeight: app.state.CurrentBlockHeight,
		CreationChainId:     app.CurrentChain,
	}

	servicePriceListKey := servicePriceListKeyPrefix + keySeparator + callerNodeID + keySeparator + funcParam.ServiceID
	currentServicePriceListBytes, err := app.state.Get([]byte(servicePriceListKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	if currentServicePriceListBytes != nil {
		var currentServicePriceList data.ServicePriceList
		err = proto.Unmarshal([]byte(currentServicePriceListBytes), &currentServicePriceList)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}

		// prepend new service price to list
		currentServicePriceList.ServicePriceList = append([]*data.ServicePrice{newServicePrice}, currentServicePriceList.ServicePriceList...)

		newCurrentServicePriceListBytes, err := utils.ProtoDeterministicMarshal(&currentServicePriceList)
		if err != nil {
			return app.NewExecTxResult(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(servicePriceListKey), []byte(newCurrentServicePriceListBytes))
	} else {

		var newServicePriceList data.ServicePriceList = data.ServicePriceList{
			ServicePriceList: []*data.ServicePrice{newServicePrice},
		}

		newCurrentServicePriceListBytes, err := utils.ProtoDeterministicMarshal(&newServicePriceList)
		if err != nil {
			return app.NewExecTxResult(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(servicePriceListKey), []byte(newCurrentServicePriceListBytes))
	}

	return app.NewExecTxResult(code.OK, "success", "")
}

type GetServicePriceListParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
}

type GetServicePriceListResult struct {
	ServicePriceListByNode []ServicePriceListByNode `json:"price_list_by_node"`
}

type ServicePriceListByNode struct {
	NodeID           string         `json:"node_id"`
	ServicePriceList []ServicePrice `json:"price_list"`
}

type ServicePrice struct {
	PriceByCurrencyList []ServicePriceByCurrency `json:"price_by_currency_list"`
	EffectiveDatetime   time.Time                `json:"effective_datetime"`
	MoreInfoURL         string                   `json:"more_info_url"`
	Detail              string                   `json:"detail"`
	CreationBlockHeight int64                    `json:"creation_block_height"`
	CreationChainID     string                   `json:"creation_chain_id"`
}

func (app *ABCIApplication) getServicePriceList(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetServicePriceList, Parameter: %s", param)
	var funcParam GetServicePriceListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if serviceValue == nil {
		return app.NewResponseQuery(nil, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	getServicePriceListByNode := func(nodeID string) (*ServicePriceListByNode, error) {
		servicePriceListKey := servicePriceListKeyPrefix + keySeparator + nodeID + keySeparator + funcParam.ServiceID
		servicePriceListBytes, err := app.state.Get([]byte(servicePriceListKey), false)
		if err != nil {
			return nil, err
		}

		if servicePriceListBytes == nil {
			var servicePriceByNode *ServicePriceListByNode = &ServicePriceListByNode{
				NodeID:           nodeID,
				ServicePriceList: make([]ServicePrice, 0),
			}

			return servicePriceByNode, nil
		}

		var servicePriceList data.ServicePriceList
		err = proto.Unmarshal([]byte(servicePriceListBytes), &servicePriceList)
		if err != nil {
			return nil, err
		}

		var servicePriceListByNode *ServicePriceListByNode = &ServicePriceListByNode{
			NodeID:           nodeID,
			ServicePriceList: make([]ServicePrice, 0),
		}
		for _, servicePrice := range servicePriceList.ServicePriceList {
			var retValServicePrice ServicePrice = ServicePrice{
				PriceByCurrencyList: make([]ServicePriceByCurrency, 0),
				EffectiveDatetime:   time.UnixMilli(servicePrice.EffectiveDatetime), // in milliseconds
				MoreInfoURL:         servicePrice.MoreInfoUrl,
				Detail:              servicePrice.Detail,
				CreationBlockHeight: servicePrice.CreationBlockHeight,
				CreationChainID:     servicePrice.CreationChainId,
			}

			for _, priceByCurrency := range servicePrice.PriceByCurrencyList {
				var retValServicePriceByCurrency ServicePriceByCurrency = ServicePriceByCurrency{
					Currency: priceByCurrency.Currency,
					MinPrice: priceByCurrency.MinPrice,
					MaxPrice: priceByCurrency.MaxPrice,
				}
				retValServicePrice.PriceByCurrencyList = append(retValServicePrice.PriceByCurrencyList, retValServicePriceByCurrency)
			}

			servicePriceListByNode.ServicePriceList = append(servicePriceListByNode.ServicePriceList, retValServicePrice)
		}

		return servicePriceListByNode, nil
	}

	var retVal GetServicePriceListResult
	retVal.ServicePriceListByNode = make([]ServicePriceListByNode, 0)
	if funcParam.NodeID != "" {
		servicePriceListByNode, err := getServicePriceListByNode(funcParam.NodeID)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		retVal.ServicePriceListByNode = append(retVal.ServicePriceListByNode, *servicePriceListByNode)
	} else {
		if !service.Active {
			retValJSON, err := json.Marshal(retVal)
			if err != nil {
				return app.NewResponseQuery(nil, err.Error(), app.state.Height)
			}

			return app.NewResponseQuery(retValJSON, "service is not active", app.state.Height)
		}

		// Get all AS nodes providing service
		serviceDestinationListKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
		serviceDestinationListValue, err := app.state.Get([]byte(serviceDestinationListKey), true)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}
		if serviceDestinationListValue == nil {
			retValJSON, err := json.Marshal(retVal)
			if err != nil {
				return app.NewResponseQuery(nil, err.Error(), app.state.Height)
			}

			return app.NewResponseQuery(retValJSON, "success", app.state.Height)
		}

		var serviceDestinationList data.ServiceDesList
		err = proto.Unmarshal([]byte(serviceDestinationListValue), &serviceDestinationList)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}

		for _, serviceDestination := range serviceDestinationList.Node {
			// filter out inactive service destination
			if !serviceDestination.Active {
				continue
			}

			// filter out not approved by NDID
			approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + serviceDestination.NodeId
			approveServiceJSON, err := app.state.Get([]byte(approveServiceKey), true)
			if err != nil {
				continue
			}
			if approveServiceJSON == nil {
				continue
			}
			var approveService data.ApproveService
			err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
			if err != nil {
				continue
			}
			if !approveService.Active {
				continue
			}

			nodeDetailKey := nodeIDKeyPrefix + keySeparator + serviceDestination.NodeId
			nodeDetailValue, err := app.state.Get([]byte(nodeDetailKey), true)
			if err != nil {
				continue
			}
			if nodeDetailValue == nil {
				continue
			}
			var nodeDetail data.NodeDetail
			err = proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				continue
			}

			// filter out inactive node
			if !nodeDetail.Active {
				continue
			}

			servicePriceListByNode, err := getServicePriceListByNode(serviceDestination.NodeId)
			if err != nil {
				return app.NewResponseQuery(nil, err.Error(), app.state.Height)
			}
			retVal.ServicePriceListByNode = append(retVal.ServicePriceListByNode, *servicePriceListByNode)
		}
	}

	retValJSON, err := json.Marshal(retVal)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(retValJSON, "success", app.state.Height)
}
