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

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
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

func (app *ABCIApplication) setServicePrice(param []byte, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetServicePrice, Parameter: %s", param)
	var funcParam SetServicePriceParam
	err := json.Unmarshal(param, &funcParam)
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

				if priceToSet.MinPrice > priceToSet.MaxPrice {
					return app.ReturnDeliverTxLog(code.ServicePriceMinCannotBeGreaterThanMax, "Service minimum price cannot be greater than maximum price", "")
				}

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
		EffectiveDatetime:   funcParam.EffectiveDatetime.UnixMilli(), // in milliseconds
		MoreInfoUrl:         funcParam.MoreInfoURL,
		Detail:              funcParam.Detail,
		CreationBlockHeight: app.state.CurrentBlockHeight,
		CreationChainId:     app.CurrentChain,
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

func (app *ABCIApplication) getServicePriceList(param []byte) types.ResponseQuery {
	app.logger.Infof("GetServicePriceList, Parameter: %s", param)
	var funcParam GetServicePriceListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if serviceValue == nil {
		return app.ReturnQuery(nil, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
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
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		retVal.ServicePriceListByNode = append(retVal.ServicePriceListByNode, *servicePriceListByNode)
	} else {
		if !service.Active {
			retValJSON, err := json.Marshal(retVal)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}

			return app.ReturnQuery(retValJSON, "service is not active", app.state.Height)
		}

		// Get all AS nodes providing service
		serviceDestinationListKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
		serviceDestinationListValue, err := app.state.Get([]byte(serviceDestinationListKey), true)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if serviceDestinationListValue == nil {
			retValJSON, err := json.Marshal(retVal)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}

			return app.ReturnQuery(retValJSON, "success", app.state.Height)
		}

		var serviceDestinationList data.ServiceDesList
		err = proto.Unmarshal([]byte(serviceDestinationListValue), &serviceDestinationList)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
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
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			retVal.ServicePriceListByNode = append(retVal.ServicePriceListByNode, *servicePriceListByNode)
		}
	}

	retValJSON, err := json.Marshal(retVal)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	return app.ReturnQuery(retValJSON, "success", app.state.Height)
}
