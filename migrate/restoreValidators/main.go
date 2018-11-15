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
	// Variable
	ndidID := getEnv("NDID_NODE_ID", "NDID")
	backupValidatorFileName := getEnv("BACKUP_VALIDATORS_FILE", "validators")
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
	file, err := os.Open("migrate/data/" + backupValidatorFileName + ".txt")
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
		SetValidator(param, ndidPrivKey, ndidID)
	}
}

func SetValidator(param did.SetValidatorParam, ndidKey *rsa.PrivateKey, ndidID string) {
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
	result, _ := utils.CallTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidID))
	resultObj, _ := result.(utils.ResponseTx)
	fmt.Println(resultObj.Result.DeliverTx.Log)
}

func after(value string, a string) string {
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

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
