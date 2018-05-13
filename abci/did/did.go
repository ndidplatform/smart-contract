package did

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/code"
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
	state      State
	ValUpdates []types.Validator
}

func NewDIDApplication() *DIDApplication {
	state := loadState(dbm.NewMemDB())
	return &DIDApplication{state: state}
}

func (app *DIDApplication) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: fmt.Sprintf("{\"size\":%v}", app.state.Size)}
}

// Save the validators in the merkle tree
func (app *DIDApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			fmt.Println("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}

// Track the block hash and header information
func (app *DIDApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	// reset valset changes
	app.ValUpdates = make([]types.Validator, 0)
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *DIDApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	fmt.Println("EndBlock")
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

func (app *DIDApplication) DeliverTx(tx []byte) (res types.ResponseDeliverTx) {
	fmt.Println("DeliverTx")

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			res = ReturnDeliverTxLog(code.CodeTypeError, "wrong create transaction format", "")
		}
	}()

	// TODO change method add Validator
	if isValidatorTx(tx) {
		// update validators in the merkle tree
		// and in app.ValUpdates
		return app.execValidatorTx(tx)
	}

	txString, err := base64.StdEncoding.DecodeString(string(tx))
	if err != nil {
		return ReturnDeliverTxLog(code.CodeTypeError, err.Error(), "")
	}
	fmt.Println(string(txString))
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]
	nodeID := parts[4]

	if method != "" {
		return DeliverTxRouter(method, param, nodeID, app)
	}
	return ReturnDeliverTxLog(code.CodeTypeError, "method can't empty", "")
}

func (app *DIDApplication) CheckTx(tx []byte) (res types.ResponseCheckTx) {
	fmt.Println("CheckTx")

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			res = ReturnCheckTx(false)
		}
	}()

	// TODO check permission before can add Validator
	if isValidatorTx(tx) {
		return ReturnCheckTx(true)
	}

	txString, err := base64.StdEncoding.DecodeString(strings.Replace(string(tx), " ", "+", -1))
	if err != nil {
		return ReturnCheckTx(false)
	}
	fmt.Println(string(txString))
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]
	nonce := parts[2]
	signature := parts[3]
	nodeID := parts[4]

	if method != "" {
		return CheckTxRouter(method, param, nonce, signature, nodeID, app)
	} else {
		return ReturnCheckTx(false)
	}
}

func (app *DIDApplication) Commit() types.ResponseCommit {
	fmt.Println("Commit")
	itr := app.state.db.Iterator(nil, nil)
	defer itr.Close()

	strAppHash := ""
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		v := itr.Value()
		if string(k) != "stateKey" {
			strAppHash += string(k) + string(v)
		}
		// fmt.Println(string(k) + "->" + string(v))
	}
	h := sha256.New()
	h.Write([]byte(strAppHash))
	appHash := h.Sum(nil)
	app.state.AppHash = appHash
	app.state.Height += 1
	saveState(app.state)
	return types.ResponseCommit{Data: appHash}
}

func (app *DIDApplication) Query(reqQuery types.RequestQuery) (res types.ResponseQuery) {
	fmt.Println("Query")

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			res = ReturnQuery(nil, "wrong query format", app.state.Height)
		}
	}()

	fmt.Println(string(reqQuery.Data))

	txString, err := base64.StdEncoding.DecodeString(string(reqQuery.Data))
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	fmt.Println(string(txString))
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]

	if method != "" {
		return QueryRouter(method, param, app)
	}
	return ReturnQuery(nil, "method can't empty", app.state.Height)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
