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
)

type IsInitEndedResult struct {
	InitEnded bool `json:"init_ended"`
}

func (app *ABCIApplication) isInitEnded(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("IsInitEnded, Parameter: %s", param)
	var result IsInitEndedResult
	result.InitEnded = false
	value, err := app.state.Get(initStateKeyBytes, true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	if string(value) == "false" {
		result.InitEnded = true
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getChainHistory(param []byte) *abcitypes.ResponseQuery {
	app.logger.Infof("GetChainHistory, Parameter: %s", param)
	chainHistoryInfoKey := "ChainHistoryInfo"
	value, err := app.state.Get([]byte(chainHistoryInfoKey), true)
	if err != nil {
		return app.NewResponseQuery(nil, err.Error(), app.state.Height)
	}
	return app.NewResponseQuery(value, "success", app.state.Height)
}
