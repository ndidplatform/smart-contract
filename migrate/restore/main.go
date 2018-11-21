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
	// Variable
	ndidID := getEnv("NDID_NODE_ID", "NDID")
	backupDataFileName := getEnv("BACKUP_DATA_FILE", "data")
	ndidKeyFile, err := os.Open("migrate/key/ndid")
	if err != nil {
		log.Fatal(err)
	}
	defer ndidKeyFile.Close()
	data, err := ioutil.ReadAll(ndidKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	ndidMasterKeyFile, err := os.Open("migrate/key/ndid_master")
	if err != nil {
		log.Fatal(err)
	}
	defer ndidMasterKeyFile.Close()
	dataMaster, err := ioutil.ReadAll(ndidMasterKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	ndidPrivKey := utils.GetPrivateKeyFromString(string(data))
	ndidMasterPrivKey := utils.GetPrivateKeyFromString(string(dataMaster))
	initNDID(ndidPrivKey, ndidMasterPrivKey, ndidID)
	file, err := os.Open("migrate/data/" + backupDataFileName + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	maximumBytes := 100000
	size := 0
	count := 0
	nTx := 0

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
		size += len(kv.Key) + len(kv.Value)
		nTx++
		if size > maximumBytes {
			setInitData(param, ndidPrivKey, ndidID)
			fmt.Print("Number of kv in param: ")
			fmt.Println(count)
			fmt.Print("Total number of kv: ")
			fmt.Println(nTx)
			count = 0
			size = 0
			param.KVList = make([]did.KeyValue, 0)
		}
	}
	if count > 0 {
		setInitData(param, ndidPrivKey, ndidID)
	}
	endInit(ndidPrivKey, ndidID)
}

func initNDID(ndidKey *rsa.PrivateKey, ndidMasterKey *rsa.PrivateKey, ndidID string) {
	// Variable
	chainHistoryFileName := getEnv("CHAIN_HISTORY_FILE", "chain_history")
	chainHistoryData, err := ioutil.ReadFile("migrate/data/" + chainHistoryFileName + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	ndidPublicKeyBytes, err := utils.GeneratePublicKey(&ndidKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	ndidMasterPublicKeyBytes, err := utils.GeneratePublicKey(&ndidKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var initNDIDparam did.InitNDIDParam
	initNDIDparam.NodeID = ndidID
	initNDIDparam.PublicKey = string(ndidPublicKeyBytes)
	initNDIDparam.MasterPublicKey = string(ndidMasterPublicKeyBytes)
	initNDIDparam.ChainHistoryInfo = string(chainHistoryData)
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

func setInitData(param did.SetInitDataParam, ndidKey *rsa.PrivateKey, ndidID string) {
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
	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidID))
	resultObj, _ := result.(utils.ResponseTx)
	// fmt.Println(resultObj)
	fmt.Println(resultObj.Result.DeliverTx.Log)
}

func endInit(ndidKey *rsa.PrivateKey, ndidID string) {
	var param did.EndInitParam
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	fnName := "EndInit"
	nodeID := ndidID
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

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
