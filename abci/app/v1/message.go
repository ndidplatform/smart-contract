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

	"github.com/tendermint/tendermint/abci/types"
	"google.golang.org/protobuf/proto"

	"github.com/ndidplatform/smart-contract/v9/abci/code"
	"github.com/ndidplatform/smart-contract/v9/abci/utils"
	data "github.com/ndidplatform/smart-contract/v9/protos/data"
)

type CreateMessageParam struct {
	MessageID string `json:"message_id"`
	Message   string `json:"message"`
	Purpose   string `json:"purpose"`
}

func (app *ABCIApplication) validateCreateMessage(funcParam CreateMessageParam, callerNodeID string, committedState bool) error {
	ok, err := app.isRPNodeByNodeID(callerNodeID, committedState)
	if err != nil {
		return err
	}
	if !ok {
		return &ApplicationError{
			Code:    code.NoPermissionForCallRPMethod,
			Message: "This node does not have permission to call RP method",
		}
	}

	return nil
}

func (app *ABCIApplication) createMessageCheckTx(param []byte, callerNodeID string) types.ResponseCheckTx {
	var funcParam CreateMessageParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return ReturnCheckTx(code.UnmarshalError, err.Error())
	}

	err = app.validateCreateMessage(funcParam, callerNodeID, true)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return ReturnCheckTx(appErr.Code, appErr.Message)
		}
		return ReturnCheckTx(code.UnknownError, err.Error())
	}

	return ReturnCheckTx(code.OK, "")
}

func (app *ABCIApplication) createMessage(param []byte, callerNodeID string) types.ResponseDeliverTx {
	app.logger.Infof("CreateMessage, Parameter: %s", param)
	var funcParam CreateMessageParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	err = app.validateCreateMessage(funcParam, callerNodeID, false)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			return app.ReturnDeliverTxLog(appErr.Code, appErr.Message, "")
		}
		return app.ReturnDeliverTxLog(code.UnknownError, err.Error(), "")
	}

	// log chain ID
	app.logger.Infof("CreateMessage, Chain ID: %s", app.CurrentChain)
	var message data.Message
	// set request data
	message.MessageId = funcParam.MessageID

	key := messageKeyPrefix + keySeparator + message.MessageId
	messageIDExist, err := app.state.Has([]byte(key), false)
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	if messageIDExist {
		return app.ReturnDeliverTxLog(code.DuplicateMessageID, "Duplicate message ID", "")
	}

	message.Message = funcParam.Message
	message.Purpose = funcParam.Purpose

	// set Owner
	message.Owner = callerNodeID
	// set creation_block_height
	message.CreationBlockHeight = app.state.CurrentBlockHeight
	// set chain_id
	message.ChainId = app.CurrentChain

	value, err := utils.ProtoDeterministicMarshal(&message)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), []byte(value))
	if err != nil {
		return app.ReturnDeliverTxLog(code.AppStateError, err.Error(), "")
	}
	return app.ReturnDeliverTxLog(code.OK, "success", message.MessageId)
}

type GetMessageParam struct {
	MessageID string `json:"message_id"`
}

type GetMessageResult struct {
	Message string `json:"message"`
}

func (app *ABCIApplication) getMessage(param []byte, height int64) types.ResponseQuery {
	app.logger.Infof("GetMessage, Parameter: %s", param)
	var funcParam GetMessageParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := messageKeyPrefix + keySeparator + funcParam.MessageID
	value, err := app.state.Get([]byte(key), true)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	if value == nil {
		valueJSON := []byte("{}")
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	var message data.Message
	err = proto.Unmarshal([]byte(value), &message)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var res GetMessageResult
	res.Message = message.Message

	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)
}

type GetMessageDetailResult struct {
	MessageID           string `json:"message_id"`
	Message             string `json:"message"`
	Purpose             string `json:"purpose"`
	RequesterNodeID     string `json:"requester_node_id"`
	CreationBlockHeight int64  `json:"creation_block_height"`
	CreationChainID     string `json:"creation_chain_id"`
}

func (app *ABCIApplication) getMessageDetail(param []byte, height int64, committedState bool) types.ResponseQuery {
	app.logger.Infof("GetMessageDetail, Parameter: %s", param)
	var funcParam GetMessageParam
	err := json.Unmarshal(param, &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	key := messageKeyPrefix + keySeparator + funcParam.MessageID
	var value []byte
	value, err = app.state.Get([]byte(key), committedState)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	if value == nil {
		valueJSON := []byte("{}")
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}

	var result GetMessageDetailResult
	var message data.Message
	err = proto.Unmarshal([]byte(value), &message)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	result.MessageID = message.MessageId
	result.Message = message.Message

	// Set purpose
	result.Purpose = message.Purpose

	// Set requester_node_id
	result.RequesterNodeID = message.Owner

	// Set creation_block_height
	result.CreationBlockHeight = message.CreationBlockHeight

	// Set creation_chain_id
	result.CreationChainID = message.ChainId

	resultJSON, err := json.Marshal(result)
	if err != nil {
		value = []byte("")
		return app.ReturnQuery(value, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}
