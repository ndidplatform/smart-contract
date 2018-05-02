package did

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
)

var (
	stateKey        = []byte("stateKey")
	kvPairPrefixKey = []byte("kvPairKey:")
)

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
	Owner   []byte
}

// TO DO save state as DB file
func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

var _ types.Application = (*DIDApplication)(nil)

type DIDApplication struct {
	types.BaseApplication
	state State
}

func NewDIDApplication() *DIDApplication {
	state := loadState(dbm.NewMemDB())
	return &DIDApplication{state: state}
}

func (app *DIDApplication) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: fmt.Sprintf("{\"size\":%v}", app.state.Size)}
}

func (app *DIDApplication) EndBlock(req types.RequestEndBlock) (resInfo types.ResponseEndBlock) {
	fmt.Println("EndBlock")
	return types.ResponseEndBlock{}
}

func (app *DIDApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	fmt.Println("DeliverTx")
	txString, err := base64.StdEncoding.DecodeString(strings.Replace(string(tx), " ", "+", -1))
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	fmt.Println(string(txString))
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]

	if method != "" {
		return DeliverTxRouter(method, param, app)
	} else {
		return ReturnDeliverTxLog("method can't empty")
	}
}

func (app *DIDApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	fmt.Println("CheckTx")
	txString, err := base64.StdEncoding.DecodeString(strings.Replace(string(tx), " ", "+", -1))
	if err != nil {
		// return ReturnDeliverTxLog(err.Error())
	}
	fmt.Println(string(txString))
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]
	nonce := parts[2]
	signature := parts[3]
	publicKey := parts[4]

	if method != "" {
		return CheckTxRouter(method, param, nonce, signature, publicKey, app)
	} else {
		return ReturnCheckTx(false)
	}
}

func (app *DIDApplication) Commit() types.ResponseCommit {
	fmt.Println("Commit")
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height += 1
	saveState(app.state)
	return types.ResponseCommit{Data: appHash}
}

func (app *DIDApplication) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	fmt.Println("Query")
	fmt.Println(string(reqQuery.Data))

	txString, err := base64.StdEncoding.DecodeString(strings.Replace(string(reqQuery.Data), " ", "+", -1))
	if err != nil {
		ReturnQuery(nil, err.Error())
	}
	fmt.Println(string(txString))
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]

	if method != "" {
		return QueryRouter(method, param, app)
	} else {
		return ReturnQuery(nil, "method can't empty")
	}
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
