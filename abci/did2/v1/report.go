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

	"github.com/tendermint/tendermint/abci/types"
)

func writeBurnTokenReport(nodeID string, method string, price float64, data string, app *DIDApplication) error {
	key := "SpendGas" + "|" + nodeID
	_, chkExists := app.state.db.Get(prefixKey([]byte(key)))
	newReport := Report{
		method,
		price,
		data,
	}
	if chkExists != nil {
		var reports []Report
		err := json.Unmarshal([]byte(chkExists), &reports)
		if err != nil {
			return err
		}
		reports = append(reports, newReport)
		value, err := json.Marshal(reports)
		if err != nil {
			return err
		}
		app.SetStateDB([]byte(key), []byte(value))
	} else {
		var reports []Report
		reports = append(reports, newReport)
		value, err := json.Marshal(reports)
		if err != nil {
			return err
		}
		app.SetStateDB([]byte(key), []byte(value))
	}
	return nil
}

func getUsedTokenReport(param string, app *DIDApplication, height int64) types.ResponseQuery {
	app.logger.Infof("GetUsedTokenReport, Parameter: %s", param)
	var funcParam GetUsedTokenReportParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.db.Version64(), app)
	}
	key := "SpendGas" + "|" + funcParam.NodeID
	_, value := app.state.db.GetVersioned(prefixKey([]byte(key)), height)
	if value == nil {
		value = []byte("")
		return ReturnQuery(value, "not found", app.state.db.Version64(), app)
	}
	return ReturnQuery(value, "success", app.state.db.Version64(), app)
}
