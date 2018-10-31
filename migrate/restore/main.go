package main

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
	"github.com/ndidplatform/smart-contract/migrate/utils"
	"github.com/tendermint/tendermint/libs/common"
)

func main() {
	ndidKeyFile, err := os.Open("migrate/key/ndid")
	if err != nil {
		log.Fatal(err)
	}
	defer ndidKeyFile.Close()
	data, err := ioutil.ReadAll(ndidKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	ndidPrivKey := utils.GetPrivateKeyFromString(string(data))
	initNDID(ndidPrivKey)
	// TODO read path backup file from env var
	file, err := os.Open("migrate/data/data.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	maximum := 100
	count := 0
	var param did.SetInitDataParam
	param.KVList = make([]did.KeyValue, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonStr := scanner.Text()
		var kv did.KeyValue
		err := json.Unmarshal([]byte(jsonStr), &kv)
		if err != nil {
			panic(err)
		}
		param.KVList = append(param.KVList, kv)
		count++
		if count == maximum {
			setInitData(param, ndidPrivKey)
			count = 0
			param.KVList = make([]did.KeyValue, 0)
		}
	}
	if count > 0 {
		setInitData(param, ndidPrivKey)
	}
	endInit(ndidPrivKey)
}

func initNDID(ndidKey *rsa.PrivateKey) {
	ndidpublicKeyBytes, err := utils.GeneratePublicKey(&ndidKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var initNDIDparam did.InitNDIDParam
	initNDIDparam.NodeID = "NDID"
	initNDIDparam.PublicKey = string(ndidpublicKeyBytes)
	initNDIDparam.MasterPublicKey = string(ndidpublicKeyBytes)
	paramJSON, err := json.Marshal(initNDIDparam)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "InitNDID"
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(initNDIDparam.NodeID))
	resultObj, _ := result.(utils.ResponseTx)
	fmt.Println(resultObj.Result.DeliverTx.Log)
}

func setInitData(param did.SetInitDataParam, ndidKey *rsa.PrivateKey) {
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "SetInitData"
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte("NDID"))
	resultObj, _ := result.(utils.ResponseTx)
	fmt.Println(resultObj.Result.DeliverTx.Log)
}

func endInit(ndidKey *rsa.PrivateKey) {
	var param did.EndInitParam
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "EndInit"
	nodeID := "NDID"
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(utils.ResponseTx)
	fmt.Println(resultObj.Result.DeliverTx.Log)
}
