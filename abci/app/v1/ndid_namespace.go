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

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type AddNamespaceParam struct {
	Namespace                                    string `json:"namespace"`
	Description                                  string `json:"description"`
	Active                                       bool   `json:"active"`
	AllowedIdentifierCountInReferenceGroup       int32  `json:"allowed_identifier_count_in_reference_group"`
	AllowedActiveIdentifierCountInReferenceGroup int32  `json:"allowed_active_identifier_count_in_reference_group"`
}

func (app *ABCIApplication) validateAddNamespace(funcParam AddNamespaceParam, callerNodeID string, committedState bool) error {
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

	allNamespacesValue, err := app.state.Get(allNamespaceKeyBytes, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}

	if allNamespacesValue != nil {
		var namespaces data.NamespaceList
		err = proto.Unmarshal([]byte(allNamespacesValue), &namespaces)
		if err != nil {
			return &ApplicationError{
				Code:    code.UnmarshalError,
				Message: err.Error(),
			}
		}

		// Check duplicate namespace
		for _, namespace := range namespaces.Namespaces {
			if namespace.Namespace == funcParam.Namespace {
				return &ApplicationError{
					Code:    code.DuplicateNamespace,
					Message: "Duplicate namespace",
				}
			}
		}
	}

	return nil
}

func (app *ABCIApplication) addNamespaceCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam AddNamespaceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddNamespace(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) addNamespace(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNamespace, Parameter: %s", param)
	var funcParam AddNamespaceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddNamespace(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	allNamespacesValue, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var namespaces data.NamespaceList
	if allNamespacesValue != nil {
		err = proto.Unmarshal([]byte(allNamespacesValue), &namespaces)
		if err != nil {
			return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
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

func (app *ABCIApplication) validateEnableNamespace(funcParam EnableNamespaceParam, callerNodeID string, committedState bool) error {
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

	allNamespacesValue, err := app.state.Get(allNamespaceKeyBytes, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if allNamespacesValue == nil {
		return &ApplicationError{
			Code:    code.NamespaceNotFound,
			Message: "List of namespaces not found",
		}
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespacesValue), &namespaces)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	found := false
	for _, namespace := range namespaces.Namespaces {
		if namespace.Namespace == funcParam.Namespace {
			found = true
			break
		}
	}

	if !found {
		return &ApplicationError{
			Code:    code.NamespaceNotFound,
			Message: "Namespace not found",
		}
	}

	return nil
}

func (app *ABCIApplication) enableNamespaceCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam EnableNamespaceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableNamespace(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableNamespace(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("EnableNamespace, Parameter: %s", param)
	var funcParam EnableNamespaceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableNamespace(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	allNamespacesValue, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespacesValue), &namespaces)
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

func (app *ABCIApplication) validateDisableNamespace(funcParam DisableNamespaceParam, callerNodeID string, committedState bool) error {
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

	allNamespacesValue, err := app.state.Get(allNamespaceKeyBytes, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if allNamespacesValue == nil {
		return &ApplicationError{
			Code:    code.NamespaceNotFound,
			Message: "List of namespaces not found",
		}
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespacesValue), &namespaces)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	found := false
	for _, namespace := range namespaces.Namespaces {
		if namespace.Namespace == funcParam.Namespace {
			found = true
			break
		}
	}

	if !found {
		return &ApplicationError{
			Code:    code.NamespaceNotFound,
			Message: "Namespace not found",
		}
	}

	return nil
}

func (app *ABCIApplication) disableNamespaceCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam DisableNamespaceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableNamespace(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableNamespace(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("DisableNamespace, Parameter: %s", param)
	var funcParam DisableNamespaceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableNamespace(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	allNamespacesValue, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespacesValue), &namespaces)
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

func (app *ABCIApplication) validateUpdateNamespace(funcParam UpdateNamespaceParam, callerNodeID string, committedState bool) error {
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

	allNamespacesValue, err := app.state.Get(allNamespaceKeyBytes, committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if allNamespacesValue == nil {
		return &ApplicationError{
			Code:    code.NamespaceNotFound,
			Message: "List of namespaces not found",
		}
	}
	var namespaces data.NamespaceList
	err = proto.Unmarshal([]byte(allNamespacesValue), &namespaces)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	found := false
	for _, namespace := range namespaces.Namespaces {
		if namespace.Namespace == funcParam.Namespace {
			found = true
			break
		}
	}

	if !found {
		return &ApplicationError{
			Code:    code.NamespaceNotFound,
			Message: "Namespace not found",
		}
	}

	return nil
}

func (app *ABCIApplication) updateNamespaceCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam UpdateNamespaceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateUpdateNamespace(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) updateNamespace(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNamespace, Parameter: %s", param)
	var funcParam UpdateNamespaceParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateUpdateNamespace(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	allNamespaceValue, err := app.state.Get(allNamespaceKeyBytes, false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
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
