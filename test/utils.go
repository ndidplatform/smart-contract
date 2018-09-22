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

package test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	protoTm "github.com/ndidplatform/smart-contract/protos/tendermint"
)

var tendermintAddr = getEnv("TENDERMINT_ADDRESS", "http://localhost:45000")

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func getPrivateKeyFromString(privK string) *rsa.PrivateKey {
	privK = strings.Replace(privK, "\t", "", -1)
	block, _ := pem.Decode([]byte(privK))
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
	}
	return privateKey
}

func generatePublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	privBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   pubKeyBytes,
	}
	publicPEM := pem.EncodeToMemory(&privBlock)
	return publicPEM, nil
}

func callTendermint(fnName []byte, param []byte, nonce []byte, signature []byte, nodeID []byte) (interface{}, error) {

	var tx protoTm.Tx
	tx.Method = string(fnName)
	tx.Params = string(param)
	tx.Nonce = nonce
	tx.Signature = signature
	tx.NodeId = string(nodeID)

	txByte, err := proto.Marshal(&tx)
	if err != nil {
		log.Printf("err: %s", err.Error())
	}

	txEncoded := hex.EncodeToString(txByte)

	var URL *url.URL
	URL, err = url.Parse(tendermintAddr)
	if err != nil {
		panic("boom")
	}
	URL.Path += "/broadcast_tx_commit"
	parameters := url.Values{}
	parameters.Add("tx", `0x`+txEncoded)
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest("GET", encodedURL, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body ResponseTx
	json.NewDecoder(resp.Body).Decode(&body)
	return body, nil
}

func queryTendermint(fnName []byte, param []byte) (interface{}, error) {
	// var path string
	// path += string(fnName)
	// path += "|"
	// path += base64.StdEncoding.EncodeToString(param)

	var data protoTm.Query
	data.Method = string(fnName)
	data.Params = string(param)

	dataByte, err := proto.Marshal(&data)
	if err != nil {
		log.Printf("err: %s", err.Error())
	}

	dataEncoded := hex.EncodeToString(dataByte)

	var URL *url.URL
	URL, err = url.Parse(tendermintAddr)
	if err != nil {
		panic("boom")
	}
	URL.Path += "/abci_query"
	parameters := url.Values{}
	parameters.Add("data", `0x`+dataEncoded)
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest("GET", encodedURL, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body ResponseQuery
	json.NewDecoder(resp.Body).Decode(&body)
	return body, nil
}

func getValidatorPubkey() string {
	var URL *url.URL
	URL, err := url.Parse(tendermintAddr)
	if err != nil {
		panic("boom")
	}
	URL.Path += "/status"
	parameters := url.Values{}
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest("GET", encodedURL, nil)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var body ResponseStatus
	json.NewDecoder(resp.Body).Decode(&body)
	return body.Result.ValidatorInfo.PubKey.Value
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func writeLog(filename string, time int64) {
	createDirIfNotExist("log")
	f, err := os.OpenFile("log/"+filename+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(strconv.FormatInt(time, 10) + "\r\n")
	if err != nil {
		panic(err)
	}
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}
