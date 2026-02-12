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
	goleveldbutil "github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type AddDomainParam struct {
	Domain string `json:"domain"`
}

func (app *ABCIApplication) validateAddDomain(funcParam AddDomainParam, callerNodeID string, committedState bool, checktx bool) error {
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

	if funcParam.Domain == "" {
		return &ApplicationError{
			Code:    code.DomainCannotBeEmpty,
			Message: "Domain name cannot be empty",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	key := domainKeyPrefix + keySeparator + funcParam.Domain
	exists, err := app.state.Has([]byte(key), committedState)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if exists {
		return &ApplicationError{
			Code:    code.DomainAlreadyExists,
			Message: "domain already exists",
		}
	}

	return nil
}

func (app *ABCIApplication) addDomainCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddDomainParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddDomain(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

// regulator only
func (app *ABCIApplication) addDomain(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddDomain, Parameter: %s", param)
	var funcParam AddDomainParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddDomain(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	key := domainKeyPrefix + keySeparator + funcParam.Domain

	var domain data.Domain
	domain.Active = true
	domain.NodeWhitelistEnabled = false

	value, err := utils.ProtoDeterministicMarshal(&domain)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}

	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type EnableDomainParam struct {
	Domain string `json:"domain"`
}

func (app *ABCIApplication) validateEnableDomain(funcParam EnableDomainParam, callerNodeID string, committedState bool, checktx bool) error {
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
	if domain.Active {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already active",
		}
	}

	return nil
}

func (app *ABCIApplication) enableDomainCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam EnableDomainParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateEnableDomain(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) enableDomain(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("EnableDomain, Parameter: %s", param)
	var funcParam EnableDomainParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateEnableDomain(funcParam, callerNodeID, false, false)
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

	domain.Active = true

	value, err = utils.ProtoDeterministicMarshal(&domain)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type DisableDomainParam struct {
	Domain string `json:"domain"`
}

func (app *ABCIApplication) validateDisableDomain(funcParam DisableDomainParam, callerNodeID string, committedState bool, checktx bool) error {
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
	if !domain.Active {
		return &ApplicationError{
			Code:    code.InvalidStateChange,
			Message: "Already inactive",
		}
	}

	return nil
}

func (app *ABCIApplication) disableDomainCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam DisableDomainParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateDisableDomain(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) disableDomain(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("DisableDomain, Parameter: %s", param)
	var funcParam DisableDomainParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateDisableDomain(funcParam, callerNodeID, false, false)
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

	domain.Active = false

	value, err = utils.ProtoDeterministicMarshal(&domain)
	if err != nil {
		return app.NewExecTxResult(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), value)

	return app.NewExecTxResult(code.OK, "success", "")
}

type GetDomainListParam struct {
	Prefix string `json:"prefix"`
}

type Domain struct {
	Domain               string `json:"domain"`
	Acitve               bool   `json:"active"`
	NodeWhitelistEnabled bool   `json:"node_whitelist_enabled"`
}

type GetDomainListResult struct {
	DomainList []Domain `json:"domain_list"`
}

func (app *ABCIApplication) getDomainList(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetDomainList, Parameter: %s", param)
	var funcParam GetDomainListParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	domainList := make([]Domain, 0)

	domainKeyIteratorBasePrefix := domainKeyPrefix + keySeparator
	domainKeyIteratorPrefix := domainKeyIteratorBasePrefix + funcParam.Prefix
	r := goleveldbutil.BytesPrefix([]byte(domainKeyIteratorPrefix))
	iter, err := app.state.db.Iterator(r.Start, r.Limit)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		runes := []rune(string(key))
		domainName := string(runes[len(domainKeyIteratorBasePrefix):])

		var domain data.Domain
		err = proto.Unmarshal(value, &domain)
		if err != nil {
			return app.NewResponseQuery(nil, err.Error(), app.state.Height)
		}

		domainList = append(domainList, Domain{
			Domain:               domainName,
			Acitve:               domain.Active,
			NodeWhitelistEnabled: domain.NodeWhitelistEnabled,
		})
	}
	iter.Close()

	result := &GetDomainListResult{
		DomainList: domainList,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}

	return app.NewResponseQuery(resultJSON, "success", app.state.Height)
}
