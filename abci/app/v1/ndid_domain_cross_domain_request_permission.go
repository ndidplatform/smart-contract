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

	"github.com/ndidplatform/smart-contract/v10/abci/code"
	"github.com/ndidplatform/smart-contract/v10/abci/utils"
	data "github.com/ndidplatform/smart-contract/v10/protos/data"
)

type EnableDomainCrossDomainRequestParam struct {
	Domain string `json:"domain"`
}

func (app *ABCIApplication) validateEnableDomainCrossDomainRequest(funcParam EnableDomainCrossDomainRequestParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
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

	if checktx {
		return nil
	}

	// stateful

	key := domainKeyPrefix + keySeparator + funcParam.Domain
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return &ApplicationError{
			Code:    code.DomainDoesNotExist,
			Message: "Domain does not exist",
		}
	}

	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if !domain.CrossDomainRequestDisabled {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already enabled",
		}
	}

	return nil
}

func (app *ABCIApplication) enableDomainCrossDomainRequestCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableDomainCrossDomainRequestParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableDomainCrossDomainRequest(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableDomainCrossDomainRequest(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableDomainCrossDomainRequest, Parameter: %s", param)
	var funcParam EnableDomainCrossDomainRequestParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableDomainCrossDomainRequest(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := domainKeyPrefix + keySeparator + funcParam.Domain

	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	domain.CrossDomainRequestDisabled = false

	value, err = utils.ProtoDeterministicMarshal(&domain)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableDomainCrossDomainRequestParam struct {
	Domain string `json:"domain"`
}

func (app *ABCIApplication) validateDisableDomainCrossDomainRequest(funcParam DisableDomainCrossDomainRequestParam, callerNodeID string, committedState bool, checktx bool) error {
	// permission
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

	if checktx {
		return nil
	}

	// stateful

	key := domainKeyPrefix + keySeparator + funcParam.Domain
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return &ApplicationError{
			Code:    code.DomainDoesNotExist,
			Message: "Domain does not exist",
		}
	}

	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if domain.CrossDomainRequestDisabled {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already disabled",
		}
	}

	return nil
}

func (app *ABCIApplication) disableDomainCrossDomainRequestCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableDomainCrossDomainRequestParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableDomainCrossDomainRequest(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableDomainCrossDomainRequest(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableDomainCrossDomainRequest, Parameter: %s", param)
	var funcParam DisableDomainCrossDomainRequestParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableDomainCrossDomainRequest(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := domainKeyPrefix + keySeparator + funcParam.Domain

	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return app.NewExecTxResult(code.AppStateError, err.Error(), "")
	}
	var domain data.Domain
	err = proto.Unmarshal(value, &domain)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	domain.CrossDomainRequestDisabled = true

	value, err = utils.ProtoDeterministicMarshal(&domain)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}
