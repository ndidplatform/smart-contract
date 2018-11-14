package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
	"github.com/ndidplatform/smart-contract/migrate/utils"
	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	kvPairPrefixKey = []byte("kvPairKey:")
)

func main() {
	// Variable
	dbDir := getEnv("DB_DIR", "DB1")
	dbName := getEnv("DB_NAME", "didDB")
	backupDBDir := getEnv("BACKUP_DB_DIR", "Backup_DB")
	backupDataFileName := getEnv("BACKUP_DATA_FILE", "data")
	backupValidatorFileName := getEnv("BACKUP_VALIDATORS_FILE", "validators")
	chainHistoryFileName := getEnv("CHAIN_HISTORY_FILE", "chain_history")
	backupBlockNumberStr := getEnv("BLOCK_NUMBER", "")

	// Delete backup file
	deleteFile("migrate/data/" + backupDataFileName + ".txt")
	deleteFile("migrate/data/" + backupValidatorFileName + ".txt")
	os.Remove(backupDBDir)
	deleteFile("migrate/data/" + chainHistoryFileName + ".txt")

	// Save previous chain info
	resStatus := utils.GetTendermintStatus()
	if backupBlockNumberStr == "" {
		backupBlockNumberStr = resStatus.Result.SyncInfo.LatestBlockHeight
	}
	backupBlockNumber, err := strconv.ParseInt(backupBlockNumberStr, 10, 64)
	if err != nil {
		panic(err)
	}
	blockStatus := utils.GetBlockStatus(backupBlockNumber)
	chainID := blockStatus.Result.Block.Header.ChainID
	latestBlockHeight := blockStatus.Result.Block.Header.Height
	latestBlockHash := blockStatus.Result.BlockMeta.BlockID.Hash
	latestAppHash := blockStatus.Result.Block.Header.AppHash
	fmt.Printf("--- Chain info at block: %s ---\n", backupBlockNumberStr)
	fmt.Println("Chain ID: " + chainID)
	fmt.Println("Latest Block Height: " + latestBlockHeight)
	fmt.Println("Latest Block Hash: " + latestBlockHash)
	fmt.Println("Latest App Hash: " + latestAppHash)

	// Copy stateDB dir
	copyDir(dbDir, backupDBDir)

	// Save kv from backup DB
	db := dbm.NewDB(dbName, "leveldb", backupDBDir)
	oldTree := iavl.NewMutableTree(db, 0)
	oldTree.Load()
	tree, _ := oldTree.GetImmutable(backupBlockNumber)
	_, ndidNodeID := tree.Get(prefixKey([]byte("MasterNDID")))
	tree.Iterate(func(key []byte, value []byte) (stop bool) {
		// Validator
		if strings.Contains(string(key), "val:") {
			var kv did.KeyValue
			kv.Key = key
			kv.Value = value
			jsonStr, err := json.Marshal(kv)
			if err != nil {
				panic(err)
			}
			fWriteLn(backupValidatorFileName, jsonStr)
			return false
		}
		// Chain history info
		if strings.Contains(string(key), "ChainHistoryInfo") {
			var chainHistory ChainHistory
			if string(value) != "" {
				err := json.Unmarshal([]byte(value), &chainHistory)
				if err != nil {
					panic(err)
				}
			}
			var prevChain ChainHistoryDetail
			prevChain.ChainID = chainID
			prevChain.LatestBlockHeight = latestBlockHeight
			prevChain.LatestBlockHash = latestBlockHash
			prevChain.LatestAppHash = latestAppHash
			chainHistory.Chains = append(chainHistory.Chains, prevChain)
			chainHistoryStr, err := json.Marshal(chainHistory)
			if err != nil {
				panic(err)
			}
			fWriteLn(chainHistoryFileName, chainHistoryStr)
			return false
		}
		if strings.Contains(string(key), string(ndidNodeID)) {
			return false
		}
		if strings.Contains(string(key), "MasterNDID") {
			return false
		}
		if strings.Contains(string(key), "InitState") {
			return false
		}
		var kv did.KeyValue
		kv.Key = key
		kv.Value = value
		jsonStr, err := json.Marshal(kv)
		if err != nil {
			panic(err)
		}
		fWriteLn(backupDataFileName, jsonStr)
		return false
	})
}

func copyDir(source string, dest string) (err error) {
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}
	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()
		if obj.IsDir() {
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()
	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()
	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}
	return
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

func fWriteLn(filename string, data []byte) {
	createDirIfNotExist("migrate/data")
	f, err := os.OpenFile("migrate/data/"+filename+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString("\r\n")
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

func deleteFile(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		return
	}
	err = os.Remove(dir)
	if err != nil {
		panic(err)
	}
}

type ChainHistoryDetail struct {
	ChainID           string `json:"chain_id"`
	LatestBlockHash   string `json:"latest_block_hash"`
	LatestAppHash     string `json:"latest_app_hash"`
	LatestBlockHeight string `json:"latest_block_height"`
}

type ChainHistory struct {
	Chains []ChainHistoryDetail `json:"chains"`
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
