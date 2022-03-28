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

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

func (app *ABCIApplication) getNamespaceList(param string) types.ResponseQuery {
	app.logger.Infof("GetNamespaceList, Parameter: %s", param)
	value, err := app.state.Get(allNamespaceKeyBytes, true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		value = []byte("[]")
		return app.ReturnQuery(value, "not found", app.state.Height)
	}

	result := make([]*data.Namespace, 0)
	// filter flag==true
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(value), &namespaces)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for _, namespace := range namespaces.Namespaces {
		if namespace.Active {
			result = append(result, namespace)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetNamespaceMap(committedState bool) (result map[string]bool) {
	result = make(map[string]bool, 0)
	allNamespaceValue, err := app.state.Get(allNamespaceKeyBytes, committedState)
	if err != nil {
		return nil
	}
	if allNamespaceValue == nil {
		return result
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespaceValue), &namespaces)
	if err != nil {
		return result
	}
	for _, namespace := range namespaces.Namespaces {
		if namespace.Active {
			result[namespace.Namespace] = true
		}
	}
	return result
}

func (app *ABCIApplication) GetNamespaceAllowedIdentifierCountMap(committedState bool) (result map[string]int) {
	result = make(map[string]int, 0)
	allNamespaceValue, err := app.state.Get(allNamespaceKeyBytes, committedState)
	if err != nil {
		return nil
	}
	if allNamespaceValue == nil {
		return result
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespaceValue), &namespaces)
	if err != nil {
		return result
	}
	for _, namespace := range namespaces.Namespaces {
		if namespace.Active {
			if namespace.AllowedIdentifierCountInReferenceGroup == -1 {
				result[namespace.Namespace] = 0
			} else {
				result[namespace.Namespace] = int(namespace.AllowedIdentifierCountInReferenceGroup)
			}
		}
	}
	return result
}
