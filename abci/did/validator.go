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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/tendermint/abci/types"
	crypto "github.com/tendermint/tendermint/crypto"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}

func (app *DIDApplication) Validators() (validators []types.Validator) {
	app.logger.Infof("Validators")
	// itr := app.state.db.Iterate(nil, nil)
	// for ; itr.Valid(); itr.Next() {
	// 	if isValidatorTx(itr.Key()) {
	// 		validator := new(types.Validator)
	// 		err := types.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		validators = append(validators, *validator)
	// 	}
	// }

	// viewed := []string{}
	app.state.db.Iterate(func(key []byte, value []byte) bool {
		// viewed = append(viewed, string(key))

		validator := new(types.Validator)
		err := types.ReadMessage(bytes.NewBuffer(key), validator)
		if err != nil {
			panic(err)
		}
		validators = append(validators, *validator)

		return false
	})

	return
}

// add, update, or remove a validator
func (app *DIDApplication) updateValidator(v types.Validator) types.ResponseDeliverTx {
	key := []byte("val:" + base64.StdEncoding.EncodeToString(v.PubKey.GetData()))

	if v.Power == 0 {
		// remove validator
		if !app.state.db.Has(key) {
			return types.ResponseDeliverTx{
				Code: code.Unauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %X", key)}
		}
		app.state.db.Remove(key)
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return types.ResponseDeliverTx{
				Code: code.EncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.state.db.Set(key, value.Bytes())
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)

	return ReturnDeliverTxLog(code.OK, "success", "")
}

func setValidator(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetValidator, Parameter: %s", param)
	var funcParam SetValidatorParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	pubKey, err := base64.StdEncoding.DecodeString(string(funcParam.PublicKey))
	if err != nil {
		return ReturnDeliverTxLog(code.DecodingError, err.Error(), "")
	}
	var pubKeyEd crypto.PubKeyEd25519
	copy(pubKeyEd[:], pubKey)

	var pubKeyObj types.PubKey
	pubKeyObj.Type = "ed25519"
	pubKeyObj.Data = pubKey

	return app.updateValidator(types.Validator{pubKeyEd.Address(), pubKeyObj, funcParam.Power})
}
