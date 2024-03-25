package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	mathRand "math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmRand "github.com/tendermint/tendermint/libs/rand"
	"google.golang.org/protobuf/proto"

	protoTm "github.com/ndidplatform/smart-contract/v9/protos/tendermint"
)

var tendermintAddr = GetEnv("TENDERMINT_ADDRESS", "http://localhost:45000")

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

func CreateSignatureAndNonce(fnName string, paramJSON []byte, privKey *rsa.PrivateKey) (nonce string, signature []byte) {
	status, err := Status()
	if err != nil {
		fmt.Println(err.Error())
		return "", nil
	}

	currentChainID := status.Result.NodeInfo.Network

	nonce = base64.StdEncoding.EncodeToString([]byte(tmRand.Str(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(currentChainID)...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := tempPSSmessage
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	signature, err = rsa.SignPKCS1v15(rand.Reader, privKey, newhash, hashed)
	if err != nil {
		fmt.Println(err.Error())
	}
	return nonce, signature
}

func Status() (*ResponseStatus, error) {
	var URL *url.URL
	URL, err := url.Parse(tendermintAddr)
	if err != nil {
		panic(err)
	}
	URL.Path += "/status"
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
	var body *ResponseStatus
	json.NewDecoder(resp.Body).Decode(&body)
	return body, nil
}

func CreateTxn(fnName []byte, param []byte, nonce []byte, signature []byte, nodeID []byte) (interface{}, error) {
	var tx protoTm.Tx
	tx.Method = string(fnName)
	tx.Params = param
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

func Query(fnName []byte, param []byte) (interface{}, error) {
	var data protoTm.Query
	data.Method = string(fnName)
	data.Params = param
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

func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	mathRand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[mathRand.Intn(len(letterRunes))]
	}
	return string(b)
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
			Tags []abcitypes.EventAttribute
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
	ID      int    `json:"id"`
	Result  struct {
		NodeInfo struct {
			ProtocolVersion struct {
				P2P   string `json:"p2p"`
				Block string `json:"block"`
				App   string `json:"app"`
			} `json:"protocol_version"`
			ID         string `json:"id"`
			ListenAddr string `json:"listen_addr"`
			Network    string `json:"network"`
			Version    string `json:"version"`
			Channels   string `json:"channels"`
			Moniker    string `json:"moniker"`
			Other      struct {
				TxIndex    string `json:"tx_index"`
				RPCAddress string `json:"rpc_address"`
			} `json:"other"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash     string    `json:"latest_block_hash"`
			LatestAppHash       string    `json:"latest_app_hash"`
			LatestBlockHeight   string    `json:"latest_block_height"`
			LatestBlockTime     time.Time `json:"latest_block_time"`
			EarliestBlockHash   string    `json:"earliest_block_hash"`
			EarliestAppHash     string    `json:"earliest_app_hash"`
			EarliestBlockHeight string    `json:"earliest_block_height"`
			EarliestBlockTime   time.Time `json:"earliest_block_time"`
			CatchingUp          bool      `json:"catching_up"`
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
