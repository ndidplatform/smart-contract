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

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
)

func ProtoDeterministicMarshal(m proto.Message) ([]byte, error) {
	var b proto.Buffer
	b.SetDeterministic(true)
	if err := b.Marshal(m); err != nil {
		return nil, err
	}
	retBytes := b.Bytes()
	if retBytes == nil {
		retBytes = make([]byte, 0)
	}
	return retBytes, nil
}

func WriteEventLogTx(filename string, time time.Time, name string, function string, nonce string) {
	createDirIfNotExist("event_log")
	f, err := os.OpenFile("event_log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var eventLog EventLogTx
	eventLog.Datetime = time.UnixNano() / 1000000
	eventLog.Name = name
	eventLog.Function = function
	eventLog.Nonce = nonce
	eventLogJSON, err := json.Marshal(eventLog)
	if err != nil {
		fmt.Println("error:", err)
	}
	_, err = f.WriteString(string(eventLogJSON) + "\r\n")
	if err != nil {
		panic(err)
	}
}

type EventLogTx struct {
	Datetime int64  `json:"datetime"`
	Name     string `json:"name"`
	Function string `json:"function"`
	Nonce    string `json:"nonce"`
}

func WriteEventLogBeginBlock(filename string, time time.Time, name string, height int64, numTxs int64) {
	createDirIfNotExist("event_log")
	f, err := os.OpenFile("event_log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var eventLog EventLogBeginBlock
	eventLog.Datetime = time.UnixNano() / 1000000
	eventLog.Name = name
	eventLog.Height = height
	eventLog.NumTxs = numTxs
	eventLogJSON, err := json.Marshal(eventLog)
	if err != nil {
		fmt.Println("error:", err)
	}
	_, err = f.WriteString(string(eventLogJSON) + "\r\n")
	if err != nil {
		panic(err)
	}
}

type EventLogBeginBlock struct {
	Datetime int64  `json:"datetime"`
	Name     string `json:"name"`
	Height   int64  `json:"height"`
	NumTxs   int64  `json:"numTxs"`
}

func WriteEventLog(filename string, time time.Time, name string) {
	createDirIfNotExist("event_log")
	f, err := os.OpenFile("event_log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var eventLog EventLog
	eventLog.Datetime = time.UnixNano() / 1000000
	eventLog.Name = name
	eventLogJSON, err := json.Marshal(eventLog)
	if err != nil {
		fmt.Println("error:", err)
	}
	_, err = f.WriteString(string(eventLogJSON) + "\r\n")
	if err != nil {
		panic(err)
	}
}

type EventLog struct {
	Datetime int64  `json:"datetime"`
	Name     string `json:"name"`
}

func WriteEventLogQuery(filename string, time time.Time, name string, function string) {
	createDirIfNotExist("event_log")
	f, err := os.OpenFile("event_log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var eventLog EventLogQuery
	eventLog.Datetime = time.UnixNano() / 1000000
	eventLog.Name = name
	eventLog.Function = function
	eventLogJSON, err := json.Marshal(eventLog)
	if err != nil {
		fmt.Println("error:", err)
	}
	_, err = f.WriteString(string(eventLogJSON) + "\r\n")
	if err != nil {
		panic(err)
	}
}

type EventLogQuery struct {
	Datetime int64  `json:"datetime"`
	Name     string `json:"name"`
	Function string `json:"function"`
}

func WriteDurationLog(filename string, durationTime int64, function string, optionalParams ...string) {
	createDirIfNotExist("event_log")
	f, err := os.OpenFile("event_log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var methodName string
	var nonceBase64 string

	if len(optionalParams) > 0 {
		if function == "CheckTx" {
			nonceBase64 = optionalParams[1]
		}
		methodName = optionalParams[0]
	}

	var eventLog DurationLog
	eventLog.Duration = durationTime
	eventLog.Function = function
	eventLog.Method = methodName
	eventLog.Nonce = nonceBase64
	eventLogJSON, err := json.Marshal(eventLog)
	if err != nil {
		fmt.Println("error:", err)
	}
	_, err = f.WriteString(string(eventLogJSON) + "\r\n")
	if err != nil {
		panic(err)
	}
}

type DurationLog struct {
	Duration int64  `json:"duration"`
	Function string `json:"function"`
	Method   string `json:"method"`
	Nonce    string `json:"nonce"`
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}
