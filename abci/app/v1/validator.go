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

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
)

func (app *ABCIApplication) Validators() (validators []abcitypes.Validator) {
	app.logger.Infof("Validators")
	itr, err := app.state.db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		validator := new(abcitypes.Validator)
		err := abcitypes.ReadMessage(bytes.NewBuffer(key), validator)
		if err != nil {
			panic(err)
		}
		validators = append(validators, *validator)
	}
	return
}

// add, update, or remove a validator
func (app *ABCIApplication) updateValidator(v abcitypes.ValidatorUpdate) *abcitypes.ExecTxResult {
	pubKeyBase64 := base64.StdEncoding.EncodeToString(v.PubKey.GetEd25519())
	key := []byte(validatorKeyPrefix + keySeparator + pubKeyBase64)

	if v.Power == 0 {
		// remove validator
		hasKey, err := app.state.Has(key, false)
		if err != nil {
			panic(err)
		}
		if !hasKey {
			return app.NewExecTxResult(code.Unauthorized, fmt.Sprintf("Cannot remove non-existent validator %X", key), "")
		}
		app.state.Delete(key)
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := abcitypes.WriteMessage(&v, value); err != nil {
			return app.NewExecTxResult(code.EncodingError, fmt.Sprintf("Error encoding validator: %v", err), "")
		}
		app.state.Set(key, value.Bytes())
	}

	app.valUpdates[pubKeyBase64] = v
	return app.NewExecTxResult(code.OK, "success", "")
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

	// validate ed25519 public key
	pubKey, err := base64.StdEncoding.DecodeString(funcParam.PublicKey)
	if err != nil {
		return &ApplicationError{
			Code:    code.DecodingError,
			Message: err.Error(),
		}
	}
	if len(pubKey) != ed25519.PubKeySize {
		err = fmt.Errorf("invalid Ed25519 public key size. Got %d, expected %d", len(pubKey), ed25519.PubKeySize)
		return &ApplicationError{
			Code:    code.InvalidValidatorPublicKey,
			Message: err.Error(),
		}
	}

	if funcParam.Power < 0 {
		return &ApplicationError{
			Code:    code.InvalidValidatorVotingPower,
			Message: "voting power can't be negative",
		}
	}

	return nil
}

func (app *ABCIApplication) setValidatorCheckTx(param []byte, callerNodeID string) *abcitypes.ResponseCheckTx {
	var funcParam SetValidatorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return NewResponseCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateSetValidator(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return NewResponseCheckTx(appErr.Code, appErr.Message)
		}
		return NewResponseCheckTx(code.UnknownError, err.Error())
	}

	return NewResponseCheckTx(code.OK, "")
}

func (app *ABCIApplication) setValidator(param []byte, callerNodeID string) *abcitypes.ExecTxResult {
	app.logger.Infof("SetValidator, Parameter: %s", param)
	var funcParam SetValidatorParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.NewExecTxResult(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateSetValidator(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.NewExecTxResult(appErr.Code, appErr.Message, "")
		}
		return app.NewExecTxResult(code.UnknownError, err.Error(), "")
	}

	pubKey, err := base64.StdEncoding.DecodeString(funcParam.PublicKey)
	if err != nil {
		return app.NewExecTxResult(code.DecodingError, err.Error(), "")
	}

	return app.updateValidator(abcitypes.UpdateValidator(pubKey, funcParam.Power, ed25519.KeyType))
}
