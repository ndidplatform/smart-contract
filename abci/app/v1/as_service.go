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

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type RegisterServiceDestinationParam struct {
	MinAal                 float64  `json:"min_aal"`
	MinIal                 float64  `json:"min_ial"`
	ServiceID              string   `json:"service_id"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

func (app *ABCIApplication) validateRegisterServiceDestination(funcParam RegisterServiceDestinationParam, callerNodeID string, committedState bool) error {
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

	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceValue == nil {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	// Check service is active
	if !service.Active {
		return &ApplicationError{
			Code:    code.ServiceIsNotActive,
			Message: "Service is not active",
		}
	}

	provideServiceKey := providedServicesKeyPrefix + keySeparator + callerNodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return &ApplicationError{
				Code:    code.UnmarshalError,
				Message: err.Error(),
			}
		}
	}

	// Check duplicate service ID
	for _, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			return &ApplicationError{
				Code:    code.DuplicateServiceID,
				Message: "Duplicate service ID in provide service list",
			}
		}
	}

	// Check approve register service destination from NDID
	approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + callerNodeID
	approveServiceValue, err := app.state.Get([]byte(approveServiceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if approveServiceValue == nil {
		return &ApplicationError{
			Code:    code.NoPermissionForRegisterServiceDestination,
			Message: "This node does not have permission to register service destination",
		}
	}
	var approveService data.ApproveService
	err = proto.Unmarshal([]byte(approveServiceValue), &approveService)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if !approveService.Active {
		return &ApplicationError{
			Code:    code.NoPermissionForRegisterServiceDestination,
			Message: "This node does not have permission to register service destination",
		}
	}

	// Add ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}

	if serviceDestinationValue != nil {
		var nodes data.ServiceDesList
		err := proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
		if err != nil {
			return &ApplicationError{
				Code:    code.UnmarshalError,
				Message: err.Error(),
			}
		}

		// Check duplicate node ID before add
		for _, node := range nodes.Node {
			if node.NodeId == callerNodeID {
				return &ApplicationError{
					Code:    code.DuplicateNodeID,
					Message: "Duplicate node ID",
				}
			}
		}
	}

	return nil
}

func (app *ABCIApplication) registerServiceDestinationCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam RegisterServiceDestinationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateRegisterServiceDestination(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) registerServiceDestination(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("RegisterServiceDestination, Parameter: %s", param)
	var funcParam RegisterServiceDestinationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateRegisterServiceDestination(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	provideServiceKey := providedServicesKeyPrefix + keySeparator + callerNodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	}

	// Append to ProvideService list
	var newService data.Service
	newService.ServiceId = funcParam.ServiceID
	newService.MinAal = funcParam.MinAal
	newService.MinIal = funcParam.MinIal
	newService.Active = true
	newService.SupportedNamespaceList = funcParam.SupportedNamespaceList
	services.Services = append(services.Services, &newService)

	provideServiceValue, err = utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	// Add ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	if serviceDestinationValue != nil {
		var nodes data.ServiceDesList
		err := proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}

		var newNode data.ASNode
		newNode.NodeId = callerNodeID
		newNode.MinIal = funcParam.MinIal
		newNode.MinAal = funcParam.MinAal
		newNode.ServiceId = funcParam.ServiceID
		newNode.SupportedNamespaceList = funcParam.SupportedNamespaceList
		newNode.Active = true
		nodes.Node = append(nodes.Node, &newNode)
		value, err := utils.ProtoDeterministicMarshal(&nodes)
		if err != nil {
			return app.NewExecTxResult(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(serviceDestinationKey), []byte(value))
	} else {
		var nodes data.ServiceDesList
		var newNode data.ASNode
		newNode.NodeId = callerNodeID
		newNode.MinIal = funcParam.MinIal
		newNode.MinAal = funcParam.MinAal
		newNode.ServiceId = funcParam.ServiceID
		newNode.SupportedNamespaceList = funcParam.SupportedNamespaceList
		newNode.Active = true
		nodes.Node = append(nodes.Node, &newNode)
		value, err := utils.ProtoDeterministicMarshal(&nodes)
		if err != nil {
			return app.NewExecTxResult(code.MarshalError, err.Error(), "")
		}
		app.state.Set([]byte(serviceDestinationKey), []byte(value))
	}
	app.state.Set([]byte(provideServiceKey), []byte(provideServiceValue))

	return app.NewExecTxResult(code.OK, "success", "")
}

type UpdateServiceDestinationParam struct {
	ServiceID              string   `json:"service_id"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

func (app *ABCIApplication) validateUpdateServiceDestination(funcParam UpdateServiceDestinationParam, callerNodeID string, committedState bool) error {
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

	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceValue == nil {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}

	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceDestinationValue == nil {
		return &ApplicationError{
			Code:    code.ServiceDestinationNotFound,
			Message: "Service destination not found",
		}
	}

	return nil
}

func (app *ABCIApplication) updateServiceDestinationCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam UpdateServiceDestinationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateUpdateServiceDestination(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) updateServiceDestination(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("UpdateServiceDestination, Parameter: %s", param)
	var funcParam UpdateServiceDestinationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateUpdateServiceDestination(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	var nodes data.ServiceDesList
	err = proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	for index := range nodes.Node {
		if nodes.Node[index].NodeId == callerNodeID {
			// selective update
			if funcParam.MinAal > 0 {
				nodes.Node[index].MinAal = funcParam.MinAal
			}
			if funcParam.MinIal > 0 {
				nodes.Node[index].MinIal = funcParam.MinIal
			}
			if len(funcParam.SupportedNamespaceList) > 0 {
				nodes.Node[index].SupportedNamespaceList = funcParam.SupportedNamespaceList
			}
			break
		}
	}

	// Update ProvideService
	provideServiceKey := providedServicesKeyPrefix + keySeparator + callerNodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	}
	for index, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			if funcParam.MinAal > 0 {
				services.Services[index].MinAal = funcParam.MinAal
			}
			if funcParam.MinIal > 0 {
				services.Services[index].MinIal = funcParam.MinIal
			}
			if len(funcParam.SupportedNamespaceList) > 0 {
				services.Services[index].SupportedNamespaceList = funcParam.SupportedNamespaceList
			}
			break
		}
	}
	provideServiceValue, err = utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	serviceDestinationValue, err = utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(provideServiceKey), []byte(provideServiceValue))
	app.state.Set([]byte(serviceDestinationKey), []byte(serviceDestinationValue))

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableServiceDestinationParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateDisableServiceDestination(funcParam DisableServiceDestinationParam, callerNodeID string, committedState bool) error {
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

	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceValue == nil {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}

	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceDestinationValue == nil {
		return &ApplicationError{
			Code:    code.ServiceDestinationNotFound,
			Message: "Service destination not found",
		}
	}

	return nil
}

func (app *ABCIApplication) disableServiceDestinationCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableServiceDestinationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableServiceDestination(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableServiceDestination(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableServiceDestination, Parameter: %s", param)
	var funcParam DisableServiceDestinationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableServiceDestination(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	var nodes data.ServiceDesList
	err = proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	for index := range nodes.Node {
		if nodes.Node[index].NodeId == callerNodeID {
			nodes.Node[index].Active = false
			break
		}
	}

	// Update ProvideService
	provideServiceKey := providedServicesKeyPrefix + keySeparator + callerNodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	}
	for index, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			services.Services[index].Active = false
			break
		}
	}
	provideServiceValue, err = utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	serviceDestinationValue, err = utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(provideServiceKey), []byte(provideServiceValue))
	app.state.Set([]byte(serviceDestinationKey), []byte(serviceDestinationValue))

	return app.NewExecTxResult(code.OK, "success", "")
}

type EnableServiceDestinationParam struct {
	ServiceID string `json:"service_id"`
}

func (app *ABCIApplication) validateEnableServiceDestination(funcParam EnableServiceDestinationParam, callerNodeID string, committedState bool) error {
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

	// Check Service ID
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, err := app.state.Get([]byte(serviceKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceValue == nil {
		return &ApplicationError{
			Code:    code.ServiceIDNotFound,
			Message: "Service ID not found",
		}
	}

	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if serviceDestinationValue == nil {
		return &ApplicationError{
			Code:    code.ServiceDestinationNotFound,
			Message: "Service destination not found",
		}
	}

	return nil
}

func (app *ABCIApplication) enableServiceDestinationCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableServiceDestinationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableServiceDestination(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableServiceDestination(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableServiceDestination, Parameter: %s", param)
	var funcParam EnableServiceDestinationParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableServiceDestination(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	// Update ServiceDestination
	serviceDestinationKey := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	serviceDestinationValue, err := app.state.Get([]byte(serviceDestinationKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}

	var nodes data.ServiceDesList
	err = proto.Unmarshal([]byte(serviceDestinationValue), &nodes)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	for index := range nodes.Node {
		if nodes.Node[index].NodeId == callerNodeID {
			nodes.Node[index].Active = true
			break
		}
	}

	// Update ProvideService
	provideServiceKey := providedServicesKeyPrefix + keySeparator + callerNodeID
	provideServiceValue, err := app.state.Get([]byte(provideServiceKey), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var services data.ServiceList
	if provideServiceValue != nil {
		err := proto.Unmarshal([]byte(provideServiceValue), &services)
		if err != nil {
			return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
		}
	}
	for index, service := range services.Services {
		if service.ServiceId == funcParam.ServiceID {
			services.Services[index].Active = true
			break
		}
	}
	provideServiceJSON, err := utils.ProtoDeterministicMarshal(&services)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	serviceDestinationJSON, err := utils.ProtoDeterministicMarshal(&nodes)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(provideServiceKey), []byte(provideServiceJSON))
	app.state.Set([]byte(serviceDestinationKey), []byte(serviceDestinationJSON))

	return app.NewExecTxResult(code.OK, "success", "")
}
