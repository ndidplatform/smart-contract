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

	"github.com/ndidplatform/smart-contract/v7/abci/code"
	"github.com/ndidplatform/smart-contract/v7/abci/utils"
	data "github.com/ndidplatform/smart-contract/v7/protos/data"
)

type AddNamespaceParam struct {
	Namespace                                    string `json:"namespace"`
	Description                                  string `json:"description"`
	Active                                       bool   `json:"active"`
	AllowedIdentifierCountInReferenceGroup       int32  `json:"allowed_identifier_count_in_reference_group"`
	AllowedActiveIdentifierCountInReferenceGroup int32  `json:"allowed_active_identifier_count_in_reference_group"`
}

func (app *ABCIApplication) addNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNamespace, Parameter: %s", param)
	var funcParam AddNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	chkExists, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var namespaces data.NamespaceList
	if chkExists != nil {
		err = proto.Unmarshal([]byte(chkExists), &namespaces)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}

		// Check duplicate namespace
		for _, namespace := range namespaces.Namespaces {
			if namespace.Namespace == funcParam.Namespace {
				return app.ReturnDeliverTxLog(code.DuplicateNamespace, "Duplicate namespace", "")
			}
		}
	}
	var newNamespace data.Namespace
	newNamespace.Namespace = funcParam.Namespace
	newNamespace.Description = funcParam.Description
	if funcParam.AllowedIdentifierCountInReferenceGroup != 0 {
		newNamespace.AllowedIdentifierCountInReferenceGroup = funcParam.AllowedIdentifierCountInReferenceGroup
	}
	if funcParam.AllowedActiveIdentifierCountInReferenceGroup != 0 {
		newNamespace.AllowedActiveIdentifierCountInReferenceGroup = funcParam.AllowedActiveIdentifierCountInReferenceGroup
	}
	// set active flag
	newNamespace.Active = true
	namespaces.Namespaces = append(namespaces.Namespaces, &newNamespace)
	value, err := utils.ProtoDeterministicMarshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(allNamespaceKeyBytes, []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type EnableNamespaceParam struct {
	Namespace string `json:"namespace"`
}

func (app *ABCIApplication) enableNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableNamespace, Parameter: %s", param)
	var funcParam EnableNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	chkExists, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var namespaces data.NamespaceList
	if chkExists == nil {
		return app.ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
	}
	err = proto.Unmarshal([]byte(chkExists), &namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for index, namespace := range namespaces.Namespaces {
		if namespace.Namespace == funcParam.Namespace {
			namespaces.Namespaces[index].Active = true
			break
		}
	}
	value, err := utils.ProtoDeterministicMarshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(allNamespaceKeyBytes, []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type DisableNamespaceParam struct {
	Namespace string `json:"namespace"`
}

func (app *ABCIApplication) disableNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableNamespace, Parameter: %s", param)
	var funcParam DisableNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	chkExists, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if chkExists == nil {
		return app.ReturnDeliverTxLog(code.NamespaceNotFound, "List of namespace not found", "")
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(chkExists), &namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for index, namespace := range namespaces.Namespaces {
		if namespace.Namespace == funcParam.Namespace {
			namespaces.Namespaces[index].Active = false
			break
		}
	}
	value, err := utils.ProtoDeterministicMarshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set(allNamespaceKeyBytes, []byte(value))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type UpdateNamespaceParam struct {
	Namespace                                    string `json:"namespace"`
	Description                                  string `json:"description"`
	AllowedIdentifierCountInReferenceGroup       int32  `json:"allowed_identifier_count_in_reference_group"`
	AllowedActiveIdentifierCountInReferenceGroup int32  `json:"allowed_active_identifier_count_in_reference_group"`
}

func (app *ABCIApplication) updateNamespace(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNamespace, Parameter: %s", param)
	var funcParam UpdateNamespaceParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	allNamespaceValue, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if allNamespaceValue == nil {
		return app.ReturnDeliverTxLog(code.NamespaceNotFound, "Namespace not found", "")
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespaceValue), &namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	for index, namespace := range namespaces.Namespaces {
		if namespace.Namespace == funcParam.Namespace {
			if funcParam.Description != "" {
				namespaces.Namespaces[index].Description = funcParam.Description
			}
			if funcParam.AllowedIdentifierCountInReferenceGroup != 0 {
				namespaces.Namespaces[index].AllowedIdentifierCountInReferenceGroup = funcParam.AllowedIdentifierCountInReferenceGroup
			}
			if funcParam.AllowedActiveIdentifierCountInReferenceGroup != 0 {
				namespaces.Namespaces[index].AllowedActiveIdentifierCountInReferenceGroup = funcParam.AllowedActiveIdentifierCountInReferenceGroup
			}
			break
		}
	}
	allNamespaceValue, err = utils.ProtoDeterministicMarshal(&namespaces)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.state.Set(allNamespaceKeyBytes, []byte(allNamespaceValue))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}
