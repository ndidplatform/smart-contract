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

	"google.golang.org/protobuf/proto"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
	goleveldbutil "github.com/syndtr/goleveldb/leveldb/util"
)

type YourDataServiceDetail struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
	Active      bool   `json:"active"`
}

type GetYourDataServiceDetailParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) getYourDataServiceDetail(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetYourDataServiceDetail, Parameter: %s", param)
	var funcParam GetYourDataServiceDetailParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	key := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		value = []byte("{}")
		return app.NewResponseQuery(value, "not found", app.state.Height)
	}
	var service data.YourDataServiceDetail
	err = proto.Unmarshal(value, &service)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	returnValue, err := json.Marshal(&service)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getYourDataServiceList(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetYourDataServiceList, Parameter: %s", param)

	result := make([]*data.YourDataServiceDetail, 0)

	yourDataServiceKeyIteratorPrefix := yourDataServiceKeyPrefix + keySeparator
	r := goleveldbutil.BytesPrefix([]byte(yourDataServiceKeyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		runes := []rune(string(key))
		serviceID := string(runes[len(yourDataServiceKeyIteratorPrefix):])

		var service *data.YourDataServiceDetail
		err = proto.Unmarshal([]byte(value), service)
		if err != nil {
			app.logger.Errorf("failed to unmarshal YourData service data: %+v", err)
			continue
		}

		if service == nil {
			app.logger.Errorf("unexpected no YourData service data for service ID: %s", serviceID)
			continue
		}

		if service.Active {
			result = append(result, service)
		}
	}
	iter.Close()

	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getYourDataServiceNameByServiceID(serviceID string) string {
	key := serviceKeyPrefix + keySeparator + serviceID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		panic(err)
	}
	if value == nil {
		return ""
	}
	var result YourDataServiceDetail
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		return ""
	}

	return result.ServiceName
}
