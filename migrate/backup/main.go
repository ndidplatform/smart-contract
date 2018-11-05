package main

import (
	"encoding/json"
	"os"
	"strings"

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	kvPairPrefixKey = []byte("kvPairKey:")
)

func main() {
	// Delete backup file
	fileName := "data"
	deleteFile("migrate/data/" + fileName + ".txt")

	var dbDir = "DB1"
	name := "didDB"
	db := dbm.NewDB(name, "leveldb", dbDir)
	oldTree := iavl.NewMutableTree(db, 0)
	oldTree.Load()
	tree, _ := oldTree.GetImmutable(oldTree.Version())
	_, ndidNodeID := tree.Get(prefixKey([]byte("MasterNDID")))
	tree.Iterate(func(key []byte, value []byte) (stop bool) {
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
		fWriteLn(fileName, jsonStr)
		return false
	})
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
