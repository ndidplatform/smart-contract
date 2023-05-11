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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/ndidplatform/smart-contract/v8/abci/code"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}

func (app *ABCIApplication) Validators() (validators []types.Validator) {
	app.logger.Infof("Validators")
	itr, err := app.state.db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		validator := new(types.Validator)
		err := types.ReadMessage(bytes.NewBuffer(key), validator)
		if err != nil {
			panic(err)
		}
		validators = append(validators, *validator)
	}
	return
}

// add, update, or remove a validator
func (app *ABCIApplication) updateValidator(v types.ValidatorUpdate) types.ResponseDeliverTx {
	pubKeyBase64 := base64.StdEncoding.EncodeToString(v.PubKey.GetEd25519())
	key := []byte(validatorKeyPrefix + keySeparator + pubKeyBase64)

	if v.Power == 0 {
		// remove validator
		hasKey, err := app.state.Has(key, false)
		if err != nil {
			panic(err)
		}
		if !hasKey {
			return app.ReturnDeliverTxLog(code.Unauthorized, fmt.Sprintf("Cannot remove non-existent validator %X", key), "")
		}
		app.state.Delete(key)
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return app.ReturnDeliverTxLog(code.EncodingError, fmt.Sprintf("Error encoding validator: %v", err), "")
		}
		app.state.Set(key, value.Bytes())
	}

	app.valUpdates[pubKeyBase64] = v
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

type SetValidatorParam struct {
	PublicKey string `json:"public_key"`
	Power     int64  `json:"power"`
}

func (app *ABCIApplication) validateSetValidator(funcParam SetValidatorParam, callerNodeID string, committedState bool) error {
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

func (app *ABCIApplication) setValidatorCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam SetValidatorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetValidator(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) setValidator(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetValidator, Parameter: %s", param)
	var funcParam SetValidatorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetValidator(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	pubKey, err := base64.StdEncoding.DecodeString(string(funcParam.PublicKey))
	if err != nil {
		return app.ReturnDeliverTxLog(code.DecodingError, err.Error(), "")
	}

	return app.updateValidator(types.UpdateValidator(pubKey, funcParam.Power, ed25519.KeyType))
}
