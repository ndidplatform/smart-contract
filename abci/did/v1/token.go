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

package did

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/tendermint/abci/types"
)

func (app *DIDApplication) getTokenPriceByFunc(fnName string, height int64) float64 {
	key := "TokenPriceFunc" + "|" + fnName
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	if value != nil {
		s, _ := strconv.ParseFloat(string(value), 64)
		return s
	}
	// if not set price of Function --> return price=1
	return 1.0
}

func (app *DIDApplication) setTokenPriceByFunc(fnName string, price float64) {
	key := "TokenPriceFunc" + "|" + fnName
	value := strconv.FormatFloat(price, 'f', -1, 64)
	app.SetStateDB([]byte(key), []byte(value))
}

func (app *DIDApplication) createTokenAccount(nodeID string) {
	key := "Token" + "|" + nodeID
	value := strconv.FormatFloat(0, 'f', -1, 64)
	app.SetStateDB([]byte(key), []byte(value))
}

func (app *DIDApplication) setToken(nodeID string, amount float64) error {
	key := "Token" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		value := strconv.FormatFloat(amount, 'f', -1, 64)
		app.SetStateDB([]byte(key), []byte(value))
		return nil
	}
	return errors.New("token account not found")
}

func (app *DIDApplication) setPriceFunc(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetPriceFunc, Parameter: %s", param)
	var funcParam SetPriceFuncParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	app.setTokenPriceByFunc(funcParam.Func, funcParam.Price)
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) getPriceFunc(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetPriceFunc, Parameter: %s", param)
	var funcParam GetPriceFuncParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	price := app.getTokenPriceByFunc(funcParam.Func, height)
	var res = GetPriceFuncResult{
		price,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(value, "success", app.state.db.Version64())
}

func (app *DIDApplication) addToken(nodeID string, amount float64) error {
	key := "Token" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return err
		}
		s = s + amount
		value := strconv.FormatFloat(s, 'f', -1, 64)
		app.SetStateDB([]byte(key), []byte(value))
		return nil
	}
	return errors.New("token account not found")
}

func (app *DIDApplication) checkTokenAccount(nodeID string) bool {
	key := "Token" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value == nil {
		return false
	}
	_, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return false
	}
	return true
}

func (app *DIDApplication) reduceToken(nodeID string, amount float64) error {
	key := "Token" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return err
		}
		if s-amount >= 0 {
			s = s - amount
			value := strconv.FormatFloat(s, 'f', -1, 64)
			app.SetStateDB([]byte(key), []byte(value))
			return nil
		}
		return errors.New("token not enough")
	}
	return errors.New("token account not found")
}

func (app *DIDApplication) getToken(nodeID string) (float64, error) {
	key := "Token" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, _ := strconv.ParseFloat(string(value), 64)
		return s, nil
	}
	return 0, errors.New("token account not found")
}

func (app *DIDApplication) setNodeToken(param string, nodeID string) types.ResponseDeliverTx {
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

func (app *DIDApplication) addNodeToken(param string, nodeID string) types.ResponseDeliverTx {
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

func (app *DIDApplication) reduceNodeToken(param string, nodeID string) types.ResponseDeliverTx {
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
	err = app.reduceToken(funcParam.NodeID, funcParam.Amount)
	if err != nil {
		return app.ReturnDeliverTxLog(code.TokenNotEnough, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *DIDApplication) getNodeToken(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodeToken, Parameter: %s", param)
	var funcParam GetNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), err.Error(), app.state.db.Version64())
	}
	tokenAmount, err := app.getToken(funcParam.NodeID)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.db.Version64())
	}
	var res = GetNodeTokenResult{
		tokenAmount,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.db.Version64())
	}
	return app.ReturnQuery(value, "success", app.state.db.Version64())
}
