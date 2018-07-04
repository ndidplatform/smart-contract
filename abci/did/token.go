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

func getTokenPriceByFunc(fnName string, app *DIDApplication) float64 {
	key := "TokenPriceFunc" + "|" + fnName
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, _ := strconv.ParseFloat(string(value), 64)
		return s
	}
	// if not set price of Function --> return price=1
	return 1.0
}

func setTokenPriceByFunc(fnName string, price float64, app *DIDApplication) {
	key := "TokenPriceFunc" + "|" + fnName
	value := strconv.FormatFloat(price, 'f', -1, 64)
	app.SetStateDB([]byte(key), []byte(value))
}

func createTokenAccount(nodeID string, app *DIDApplication) {
	key := "Token" + "|" + nodeID
	value := strconv.FormatFloat(0, 'f', -1, 64)
	app.SetStateDB([]byte(key), []byte(value))
}

func setToken(nodeID string, amount float64, app *DIDApplication) error {
	key := "Token" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		value := strconv.FormatFloat(amount, 'f', -1, 64)
		app.SetStateDB([]byte(key), []byte(value))
		return nil
	}
	return errors.New("token account not found")
}

func setPriceFunc(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetPriceFunc, Parameter: %s", param)
	var funcParam SetPriceFuncParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	setTokenPriceByFunc(funcParam.Func, funcParam.Price, app)
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func getPriceFunc(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetPriceFunc, Parameter: %s", param)
	var funcParam GetPriceFuncParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	price := getTokenPriceByFunc(funcParam.Func, app)
	var res = GetPriceFuncResult{
		price,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}

func addToken(nodeID string, amount float64, app *DIDApplication) error {
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

func reduceToken(nodeID string, amount float64, app *DIDApplication) error {
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

func getToken(nodeID string, app *DIDApplication) (float64, error) {
	key := "Token" + "|" + nodeID
	_, value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, _ := strconv.ParseFloat(string(value), 64)
		return s, nil
	}
	return 0, errors.New("token account not found")
}

func setNodeToken(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetNodeToken, Parameter: %s", param)
	var funcParam SetNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	err = setToken(funcParam.NodeID, funcParam.Amount, app)
	if err != nil {
		return ReturnDeliverTxLog(code.TokenAccountNotFound, err.Error(), "")
	}
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func addNodeToken(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("AddNodeToken, Parameter: %s", param)
	var funcParam AddNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	err = addToken(funcParam.NodeID, funcParam.Amount, app)
	if err != nil {
		return ReturnDeliverTxLog(code.TokenAccountNotFound, err.Error(), "")
	}
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func reduceNodeToken(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("ReduceNodeToken, Parameter: %s", param)
	var funcParam ReduceNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	err = reduceToken(funcParam.NodeID, funcParam.Amount, app)
	if err != nil {
		return ReturnDeliverTxLog(code.TokenAccountNotFound, err.Error(), "")
	}
	return ReturnDeliverTxLog(code.OK, "success", "")
}

func getNodeToken(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetNodeToken, Parameter: %s", param)
	var funcParam GetNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	tokenAmount, err := getToken(funcParam.NodeID, app)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	var res = GetNodeTokenResult{
		tokenAmount,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}
