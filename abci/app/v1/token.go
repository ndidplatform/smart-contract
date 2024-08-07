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
	"errors"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

func (app *ABCIApplication) getTokenPriceByFunc(fnName string, committedState bool) float64 {
	key := tokenPriceFuncKeyPrefix + keySeparator + fnName
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		panic(err)
	}
	if value == nil {
		// if not set price of Function --> return price=1
		return 1.0
	}
	var tokenPrice data.TokenPrice
	err = proto.Unmarshal(value, &tokenPrice)
	if err != nil {
		return 1.0
	}
	return tokenPrice.Price
}

func (app *ABCIApplication) setTokenPriceByFunc(fnName string, price float64) error {
	key := tokenPriceFuncKeyPrefix + keySeparator + fnName
	var tokenPrice data.TokenPrice
	tokenPrice.Price = price
	value, err := utils.ProtoDeterministicMarshal(&tokenPrice)
	if err != nil {
		return &ApplicationError{
			Code:    code.MarshalError,
			Message: err.Error(),
		}
	}
	app.state.Set([]byte(key), []byte(value))

	return nil
}

func (app *ABCIApplication) createTokenAccount(nodeID string) error {
	key := tokenKeyPrefix + keySeparator + nodeID
	var token data.Token
	token.Amount = 0
	value, err := utils.ProtoDeterministicMarshal(&token)
	if err != nil {
		return err
	}
	app.state.Set([]byte(key), []byte(value))

	return nil
}

func (app *ABCIApplication) setToken(nodeID string, amount float64) error {
	key := tokenKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return &ApplicationError{
			Code:    code.TokenAccountNotFound,
			Message: "token account not found",
		}
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	token.Amount = amount
	value, err = utils.ProtoDeterministicMarshal(&token)
	if err != nil {
		return &ApplicationError{
			Code:    code.MarshalError,
			Message: err.Error(),
		}
	}
	app.state.Set([]byte(key), []byte(value))

	return nil
}

type SetPriceFuncParam struct {
	Func  string  `json:"func"`
	Price float64 `json:"price"`
}

func (app *ABCIApplication) validateSetPriceFunc(funcParam SetPriceFuncParam, callerNodeID string, committedState bool, checktx bool) error {
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

	return nil
}

func (app *ABCIApplication) setPriceFuncCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam SetPriceFuncParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetPriceFunc(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setPriceFunc(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetPriceFunc, Parameter: %s", param)
	var funcParam SetPriceFuncParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetPriceFunc(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	err = app.setTokenPriceByFunc(funcParam.Func, funcParam.Price)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	return app.NewExecTxResult(code.OK, "success", "")
}

type GetPriceFuncParam struct {
	Func string `json:"func"`
}

type GetPriceFuncResult struct {
	Price float64 `json:"price"`
}

func (app *ABCIApplication) getPriceFunc(param []byte, committedState bool) *abcitypes.ResponseQuery {
	app.logger.Infof("GetPriceFunc, Parameter: %s", param)
	var funcParam GetPriceFuncParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	price := app.getTokenPriceByFunc(funcParam.Func, committedState)
	var res = GetPriceFuncResult{
		price,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(value, "success", app.state.Height)
}

func (app *ABCIApplication) addToken(nodeID string, amount float64) error {
	key := tokenKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return &ApplicationError{
			Code:    code.TokenAccountNotFound,
			Message: "token account not found",
		}
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	token.Amount = token.Amount + amount
	value, err = utils.ProtoDeterministicMarshal(&token)
	if err != nil {
		return &ApplicationError{
			Code:    code.MarshalError,
			Message: err.Error(),
		}
	}
	app.state.Set([]byte(key), []byte(value))

	return nil
}

func (app *ABCIApplication) checkTokenAccount(nodeID string, committedState bool) (bool, error) {
	key := tokenKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return false, nil
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return false, &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}

	return true, nil
}

func (app *ABCIApplication) reduceToken(nodeID string, amount float64) error {
	key := tokenKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return &ApplicationError{
			Code:    code.AppStateError,
			Message: err.Error(),
		}
	}
	if value == nil {
		return &ApplicationError{
			Code:    code.TokenAccountNotFound,
			Message: "token account not found",
		}
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return &ApplicationError{
			Code:    code.UnmarshalError,
			Message: err.Error(),
		}
	}
	if amount > token.Amount {
		return &ApplicationError{
			Code:    code.TokenNotEnough,
			Message: "token not enough",
		}
	}
	token.Amount = token.Amount - amount
	value, err = utils.ProtoDeterministicMarshal(&token)
	if err != nil {
		return &ApplicationError{
			Code:    code.MarshalError,
			Message: err.Error(),
		}
	}
	app.state.Set([]byte(key), []byte(value))

	return nil
}

func (app *ABCIApplication) getToken(nodeID string, committedState bool) (float64, error) {
	key := tokenKeyPrefix + keySeparator + nodeID
	value, err := app.state.Get([]byte(key), committedState)
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, errors.New("token account not found")
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return 0, errors.New("token account not found")
	}
	return token.Amount, nil
}

type SetNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

func (app *ABCIApplication) validateSetNodeToken(funcParam SetNodeTokenParam, callerNodeID string, committedState bool, checktx bool) error {
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

	// stateless

	// Validate parameters
	if funcParam.NodeID == "" {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID cannot be empty",
		}
	}

	if funcParam.Amount < 0 {
		return &ApplicationError{
			Code:    code.AmountMustBeGreaterOrEqualToZero,
			Message: "Amount must be greater than or equal to zero",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	// Check token account
	tokenAccountFound, err := app.checkTokenAccount(funcParam.NodeID, committedState)
	if err != nil {
		return err
	}
	if !tokenAccountFound {
		return &ApplicationError{
			Code:    code.TokenAccountNotFound,
			Message: "token account not found",
		}
	}

	return nil
}

func (app *ABCIApplication) setNodeTokenCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam SetNodeTokenParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetNodeToken(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setNodeToken(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetNodeToken, Parameter: %s", param)
	var funcParam SetNodeTokenParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetNodeToken(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	err = app.setToken(funcParam.NodeID, funcParam.Amount)
	if err != nil {
		return app.NewExecTxResult(code.TokenAccountNotFound, err.Error(), "")
	}

	return app.NewExecTxResult(code.OK, "success", "")
}

type AddNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

func (app *ABCIApplication) validateAddNodeToken(funcParam AddNodeTokenParam, callerNodeID string, committedState bool, checktx bool) error {
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

	// stateless

	// Validate parameters
	if funcParam.NodeID == "" {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID cannot be empty",
		}
	}

	if funcParam.Amount < 0 {
		return &ApplicationError{
			Code:    code.AmountMustBeGreaterOrEqualToZero,
			Message: "Amount must be greater than or equal to zero",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	// Check token account
	tokenAccountFound, err := app.checkTokenAccount(funcParam.NodeID, committedState)
	if err != nil {
		return err
	}
	if !tokenAccountFound {
		return &ApplicationError{
			Code:    code.TokenAccountNotFound,
			Message: "token account not found",
		}
	}

	return nil
}

func (app *ABCIApplication) addNodeTokenCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam AddNodeTokenParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateAddNodeToken(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) addNodeToken(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("AddNodeToken, Parameter: %s", param)
	var funcParam AddNodeTokenParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateAddNodeToken(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	err = app.addToken(funcParam.NodeID, funcParam.Amount)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	return app.NewExecTxResult(code.OK, "success", "")
}

type ReduceNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

func (app *ABCIApplication) validateReduceNodeToken(funcParam ReduceNodeTokenParam, callerNodeID string, committedState bool, checktx bool) error {
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

	// stateless

	// Validate parameters
	if funcParam.NodeID == "" {
		return &ApplicationError{
			Code:    code.NodeIDNotFound,
			Message: "Node ID cannot be empty",
		}
	}

	if funcParam.Amount < 0 {
		return &ApplicationError{
			Code:    code.AmountMustBeGreaterOrEqualToZero,
			Message: "Amount must be greater than or equal to zero",
		}
	}

	if checktx {
		return nil
	}

	// stateful

	// Check token account
	tokenAccountFound, err := app.checkTokenAccount(funcParam.NodeID, committedState)
	if err != nil {
		return err
	}
	if !tokenAccountFound {
		return &ApplicationError{
			Code:    code.TokenAccountNotFound,
			Message: "token account not found",
		}
	}

	return nil
}

func (app *ABCIApplication) reduceNodeTokenCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam ReduceNodeTokenParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateReduceNodeToken(funcParam, callerNodeID, true, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) reduceNodeToken(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("ReduceNodeToken, Parameter: %s", param)
	var funcParam ReduceNodeTokenParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateReduceNodeToken(funcParam, callerNodeID, false, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	err = app.reduceToken(funcParam.NodeID, funcParam.Amount)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	return app.NewExecTxResult(code.OK, "success", "")
}

type GetNodeTokenParam struct {
	NodeID string `json:"node_id"`
}

type GetNodeTokenResult struct {
	Amount float64 `json:"amount"`
}

func (app *ABCIApplication) getNodeToken(param []byte, committedState bool) *abcitypes.ResponseQuery {
	app.logger.Infof("GetNodeToken, Parameter: %s", param)
	var funcParam GetNodeTokenParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewResponseQuery([]byte("{}"), err.Error(), app.state.Height)
	}
	tokenAmount, err := app.getToken(funcParam.NodeID, committedState)
	if err != nil {
		return app.NewResponseQuery([]byte("{}"), "not found", app.state.Height)
	}
	var res = GetNodeTokenResult{
		tokenAmount,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(value, "success", app.state.Height)
}
