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

	"github.com/golang/protobuf/proto"
	"github.com/ndidplatform/smart-contract/v4/abci/code"
	"github.com/ndidplatform/smart-contract/v4/abci/utils"
	"github.com/tendermint/tendermint/abci/types"

	data "github.com/ndidplatform/smart-contract/v4/protos/data"
)

func (app *ABCIApplication) getTokenPriceByFunc(fnName string) float64 {
	key := "TokenPriceFunc" + "|" + fnName
	value, err := app.state.Get([]byte(key), true)
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
	key := "TokenPriceFunc" + "|" + fnName
	var tokenPrice data.TokenPrice
	tokenPrice.Price = price
	value, err := utils.ProtoDeterministicMarshal(&tokenPrice)
	if err != nil {
		return err
	}
	app.state.Set([]byte(key), []byte(value))
	return nil
}

func (app *ABCIApplication) createTokenAccount(nodeID string) {
	key := "Token" + "|" + nodeID
	var token data.Token
	token.Amount = 0
	value, _ := utils.ProtoDeterministicMarshal(&token)
	app.state.Set([]byte(key), []byte(value))
}

func (app *ABCIApplication) setToken(nodeID string, amount float64) error {
	key := "Token" + "|" + nodeID
	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return err
	}
	if value == nil {
		return errors.New("token account not found")
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return errors.New("token account not found")
	}
	token.Amount = amount
	value, err = utils.ProtoDeterministicMarshal(&token)
	if err != nil {
		return errors.New("token account not found")
	}
	app.state.Set([]byte(key), []byte(value))
	return nil
}

func (app *ABCIApplication) setPriceFunc(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetPriceFunc, Parameter: %s", param)
	var funcParam SetPriceFuncParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	err = app.setTokenPriceByFunc(funcParam.Func, funcParam.Price)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) getPriceFunc(param string) types.ResponseQuery {
	app.logger.Infof("GetPriceFunc, Parameter: %s", param)
	var funcParam GetPriceFuncParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	price := app.getTokenPriceByFunc(funcParam.Func)
	var res = GetPriceFuncResult{
		price,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(value, "success", app.state.Height)
}

func (app *ABCIApplication) addToken(nodeID string, amount float64) error {
	key := "Token" + "|" + nodeID
	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return err
	}
	if value == nil {
		return errors.New("token account not found")
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return errors.New("token account not found")
	}
	token.Amount = token.Amount + amount
	value, err = utils.ProtoDeterministicMarshal(&token)
	if err != nil {
		return errors.New("token account not found")
	}
	app.state.Set([]byte(key), []byte(value))
	return nil
}

func (app *ABCIApplication) checkTokenAccount(nodeID string) bool {
	key := "Token" + "|" + nodeID
	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		panic(err)
	}
	if value == nil {
		return false
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return false
	}
	return true
}

func (app *ABCIApplication) reduceToken(nodeID string, amount float64) (errorCode uint32, errorLog string) {
	key := "Token" + "|" + nodeID
	value, err := app.state.Get([]byte(key), false)
	if err != nil {
		return code.AppStateError, ""
	}
	if value == nil {
		return code.TokenAccountNotFound, "token account not found"
	}
	var token data.Token
	err = proto.Unmarshal(value, &token)
	if err != nil {
		return code.TokenAccountNotFound, "token account not found"
	}
	if amount > token.Amount {
		return code.TokenNotEnough, "token not enough"
	}
	token.Amount = token.Amount - amount
	value, err = utils.ProtoDeterministicMarshal(&token)
	if err != nil {
		return code.TokenAccountNotFound, "token account not found"
	}
	app.state.Set([]byte(key), []byte(value))
	return code.OK, ""
}

func (app *ABCIApplication) getToken(nodeID string) (float64, error) {
	key := "Token" + "|" + nodeID
	value, err := app.state.Get([]byte(key), false)
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

func (app *ABCIApplication) getTokenCommitted(nodeID string) (float64, error) {
	key := "Token" + "|" + nodeID
	value, err := app.state.Get([]byte(key), true)
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

func (app *ABCIApplication) setNodeToken(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetNodeToken, Parameter: %s", param)
	var funcParam SetNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Validate parameter
	if funcParam.Amount < 0 {
		return app.ReturnDeliverTxLog(code.AmountMustBeGreaterOrEqualToZero, "Amount must be greater than or equal to zero", "")
	}
	// Check token account
	if !app.checkTokenAccount(funcParam.NodeID) {
		return app.ReturnDeliverTxLog(code.TokenAccountNotFound, "token account not found", "")
	}
	err = app.setToken(funcParam.NodeID, funcParam.Amount)
	if err != nil {
		return app.ReturnDeliverTxLog(code.TokenAccountNotFound, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) addNodeToken(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNodeToken, Parameter: %s", param)
	var funcParam AddNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Validate parameter
	if funcParam.Amount < 0 {
		return app.ReturnDeliverTxLog(code.AmountMustBeGreaterOrEqualToZero, "Amount must be greater than or equal to zero", "")
	}
	// Check token account
	if !app.checkTokenAccount(funcParam.NodeID) {
		return app.ReturnDeliverTxLog(code.TokenAccountNotFound, "token account not found", "")
	}
	err = app.addToken(funcParam.NodeID, funcParam.Amount)
	if err != nil {
		return app.ReturnDeliverTxLog(code.TokenAccountNotFound, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) reduceNodeToken(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("ReduceNodeToken, Parameter: %s", param)
	var funcParam ReduceNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// Validate parameter
	if funcParam.Amount < 0 {
		return app.ReturnDeliverTxLog(code.AmountMustBeGreaterOrEqualToZero, "Amount must be greater than or equal to zero", "")
	}
	// Check token account
	if !app.checkTokenAccount(funcParam.NodeID) {
		return app.ReturnDeliverTxLog(code.TokenAccountNotFound, "token account not found", "")
	}
	errCode, errLog := app.reduceToken(funcParam.NodeID, funcParam.Amount)
	if errCode != code.OK {
		return app.ReturnDeliverTxLog(errCode, errLog, "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) getNodeToken(param string) types.ResponseQuery {
	app.logger.Infof("GetNodeToken, Parameter: %s", param)
	var funcParam GetNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), err.Error(), app.state.Height)
	}
	tokenAmount, err := app.getTokenCommitted(funcParam.NodeID)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var res = GetNodeTokenResult{
		tokenAmount,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(value, "success", app.state.Height)
}
