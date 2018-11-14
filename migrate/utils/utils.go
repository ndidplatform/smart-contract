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
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	protoTm "github.com/ndidplatform/smart-contract/protos/tendermint"
	"github.com/tendermint/tendermint/libs/common"
)

var tendermintAddr = GetEnv("TENDERMINT_ADDRESS", "http://localhost:45000")

func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func GetPrivateKeyFromString(privK string) *rsa.PrivateKey {
	privK = strings.Replace(privK, "\t", "", -1)
	block, _ := pem.Decode([]byte(privK))
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
	}
	return privateKey
}

func GeneratePublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
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

func CallTendermint(fnName []byte, param []byte, nonce []byte, signature []byte, nodeID []byte) (interface{}, error) {

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

type ResponseTx struct {
	Result struct {
		Height  int `json:"height"`
		CheckTx struct {
			Code int      `json:"code"`
			Log  string   `json:"log"`
			Fee  struct{} `json:"fee"`
		} `json:"check_tx"`
		DeliverTx struct {
			Log  string   `json:"log"`
			Fee  struct{} `json:"fee"`
			Tags []common.KVPair
		} `json:"deliver_tx"`
		Hash string `json:"hash"`
	} `json:"result"`
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
}

type ResponseQuery struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		Response struct {
			Log    string `json:"log"`
			Value  string `json:"value"`
			Height string `json:"height"`
		} `json:"response"`
	} `json:"result"`
}

type ResponseStatus struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		NodeInfo struct {
			ID         string   `json:"id"`
			ListenAddr string   `json:"listen_addr"`
			Network    string   `json:"network"`
			Version    string   `json:"version"`
			Channels   string   `json:"channels"`
			Moniker    string   `json:"moniker"`
			Other      []string `json:"other"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash   string    `json:"latest_block_hash"`
			LatestAppHash     string    `json:"latest_app_hash"`
			LatestBlockHeight string    `json:"latest_block_height"`
			LatestBlockTime   time.Time `json:"latest_block_time"`
			CatchingUp        bool      `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address string `json:"address"`
			PubKey  struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	} `json:"result"`
}

func GetTendermintStatus() ResponseStatus {
	var URL *url.URL
	URL, err := url.Parse(tendermintAddr)
	if err != nil {
		panic(err)
	}
	URL.Path += "/status"
	parameters := url.Values{}
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest("GET", encodedURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var body ResponseStatus
	json.NewDecoder(resp.Body).Decode(&body)
	return body
}

type BlockResult struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		BlockMeta struct {
			BlockID struct {
				Hash  string `json:"hash"`
				Parts struct {
					Total string `json:"total"`
					Hash  string `json:"hash"`
				} `json:"parts"`
			} `json:"block_id"`
			Header struct {
				ChainID     string    `json:"chain_id"`
				Height      string    `json:"height"`
				Time        time.Time `json:"time"`
				NumTxs      string    `json:"num_txs"`
				LastBlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"last_block_id"`
				TotalTxs        string `json:"total_txs"`
				LastCommitHash  string `json:"last_commit_hash"`
				DataHash        string `json:"data_hash"`
				ValidatorsHash  string `json:"validators_hash"`
				ConsensusHash   string `json:"consensus_hash"`
				AppHash         string `json:"app_hash"`
				LastResultsHash string `json:"last_results_hash"`
				EvidenceHash    string `json:"evidence_hash"`
			} `json:"header"`
		} `json:"block_meta"`
		Block struct {
			Header struct {
				ChainID     string    `json:"chain_id"`
				Height      string    `json:"height"`
				Time        time.Time `json:"time"`
				NumTxs      string    `json:"num_txs"`
				LastBlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"last_block_id"`
				TotalTxs        string `json:"total_txs"`
				LastCommitHash  string `json:"last_commit_hash"`
				DataHash        string `json:"data_hash"`
				ValidatorsHash  string `json:"validators_hash"`
				ConsensusHash   string `json:"consensus_hash"`
				AppHash         string `json:"app_hash"`
				LastResultsHash string `json:"last_results_hash"`
				EvidenceHash    string `json:"evidence_hash"`
			} `json:"header"`
			Data struct {
				Txs interface{} `json:"txs"`
			} `json:"data"`
			Evidence struct {
				Evidence interface{} `json:"evidence"`
			} `json:"evidence"`
			LastCommit struct {
				BlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total string `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"block_id"`
				Precommits []struct {
					ValidatorAddress string    `json:"validator_address"`
					ValidatorIndex   string    `json:"validator_index"`
					Height           string    `json:"height"`
					Round            string    `json:"round"`
					Timestamp        time.Time `json:"timestamp"`
					Type             int       `json:"type"`
					BlockID          struct {
						Hash  string `json:"hash"`
						Parts struct {
							Total string `json:"total"`
							Hash  string `json:"hash"`
						} `json:"parts"`
					} `json:"block_id"`
					Signature struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"signature"`
				} `json:"precommits"`
			} `json:"last_commit"`
		} `json:"block"`
	} `json:"result"`
}

func GetBlockStatus(height int64) BlockResult {
	var URL *url.URL
	URL, err := url.Parse(tendermintAddr)
	if err != nil {
		panic(err)
	}
	URL.Path += "/block"
	parameters := url.Values{}
	parameters.Add("height", strconv.FormatInt(height, 10))
	URL.RawQuery = parameters.Encode()
	encodedURL := URL.String()
	req, err := http.NewRequest("GET", encodedURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var body BlockResult
	json.NewDecoder(resp.Body).Decode(&body)
	return body
}
