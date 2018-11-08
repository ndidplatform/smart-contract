package main

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
	"github.com/ndidplatform/smart-contract/migrate/utils"
	"github.com/tendermint/tendermint/abci/types"
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
	// // TODO read path backup file from env var
	file, err := os.Open("migrate/data/validators.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonStr := scanner.Text()
		var kv did.KeyValue
		err := json.Unmarshal([]byte(jsonStr), &kv)
		if err != nil {
			panic(err)
		}
		validator := new(types.Validator)
		err = types.ReadMessage(bytes.NewBuffer(kv.Value), validator)
		if err != nil {
			panic(err)
		}
		publicKey := after(string(kv.Key), `val:`)
		var param did.SetValidatorParam
		param.PublicKey = publicKey
		param.Power = validator.Power
		SetValidator(param, ndidPrivKey)
	}
}

func SetValidator(param did.SetValidatorParam, ndidKey *rsa.PrivateKey) {
	ndidNodeID := "NDID"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "SetValidator"
	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	tempPSSmessage := append([]byte(fnName), paramJSON...)
	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(utils.ResponseTx)
	fmt.Println(resultObj.Result.DeliverTx.Log)
}

func after(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}

// func initNDID(ndidKey *rsa.PrivateKey) {
// 	ndidpublicKeyBytes, err := utils.GeneratePublicKey(&ndidKey.PublicKey)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	var initNDIDparam did.InitNDIDParam
// 	initNDIDparam.NodeID = "NDID"
// 	initNDIDparam.PublicKey = string(ndidpublicKeyBytes)
// 	initNDIDparam.MasterPublicKey = string(ndidpublicKeyBytes)
// 	paramJSON, err := json.Marshal(initNDIDparam)
// 	if err != nil {
// 		fmt.Println("error:", err)
// 	}
// 	fnName := "InitNDID"
// 	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
// 	tempPSSmessage := append([]byte(fnName), paramJSON...)
// 	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
// 	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
// 	newhash := crypto.SHA256
// 	pssh := newhash.New()
// 	pssh.Write(PSSmessage)
// 	hashed := pssh.Sum(nil)
// 	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
// 	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(initNDIDparam.NodeID))
// 	resultObj, _ := result.(utils.ResponseTx)
// 	fmt.Println(resultObj.Result.DeliverTx.Log)
// }

// func setInitData(param did.SetInitDataParam, ndidKey *rsa.PrivateKey) {
// 	paramJSON, err := json.Marshal(param)
// 	if err != nil {
// 		fmt.Println("error:", err)
// 	}
// 	fnName := "SetInitData"
// 	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
// 	tempPSSmessage := append([]byte(fnName), paramJSON...)
// 	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
// 	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
// 	newhash := crypto.SHA256
// 	pssh := newhash.New()
// 	pssh.Write(PSSmessage)
// 	hashed := pssh.Sum(nil)
// 	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
// 	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte("NDID"))
// 	resultObj, _ := result.(utils.ResponseTx)
// 	fmt.Println(resultObj.Result.DeliverTx.Log)
// }

// func endInit(ndidKey *rsa.PrivateKey) {
// 	var param did.EndInitParam
// 	paramJSON, err := json.Marshal(param)
// 	if err != nil {
// 		fmt.Println("error:", err)
// 	}
// 	fnName := "EndInit"
// 	nodeID := "NDID"
// 	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
// 	tempPSSmessage := append([]byte(fnName), paramJSON...)
// 	tempPSSmessage = append(tempPSSmessage, []byte(nonce)...)
// 	PSSmessage := []byte(base64.StdEncoding.EncodeToString(tempPSSmessage))
// 	newhash := crypto.SHA256
// 	pssh := newhash.New()
// 	pssh.Write(PSSmessage)
// 	hashed := pssh.Sum(nil)
// 	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
// 	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(nodeID))
// 	resultObj, _ := result.(utils.ResponseTx)
// 	fmt.Println(resultObj.Result.DeliverTx.Log)
// }
