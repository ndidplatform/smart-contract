package did

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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
	db           dbm.DB
	Size         int64    `json:"size"`
	Height       int64    `json:"height"`
	AppHash      []byte   `json:"app_hash"`
	UncommitKeys []string `json:"uncommit_keys"`
	CommitStr    string   `json:"commit_str"`
}

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

	var dbDir = getEnv("DB_NAME", "DID")

	name := "didDB"
	db, err := dbm.NewGoLevelDB(name, dbDir)
	if err != nil {
		panic(err)
	}

	state := loadState(db)

	return &DIDApplication{state: state}
}

func (app *DIDApplication) SetStateDB(key, value []byte) {
	if string(key) != "stateKey" {
		app.state.UncommitKeys = append(app.state.UncommitKeys, string(key))
	}
	app.state.db.Set(prefixKey(key), value)
	app.state.Size++
}

func (app *DIDApplication) DeleteStateDB(key []byte) {
	app.state.db.Delete(prefixKey(key))
	app.state.Size--
}

func (app *DIDApplication) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	var res types.ResponseInfo
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	return res
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
	fmt.Print("BeginBlock: ")
	fmt.Println(req.Header.Height)
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
			res = ReturnDeliverTxLog(code.WrongTransactionFormat, "wrong transaction format", "")
		}
	}()

	// TODO change method add Validator
	// After scale test delete this
	if isValidatorTx(tx) {
		// update validators in the merkle tree
		// and in app.ValUpdates
		return app.execValidatorTx(tx)
	}
	// ---------------------

	txString, err := base64.StdEncoding.DecodeString(string(tx))
	if err != nil {
		return ReturnDeliverTxLog(code.DecodingError, err.Error(), "")
	}
	fmt.Println(string(txString))
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]
	nonce := parts[2]
	signature := parts[3]
	nodeID := parts[4]

	if method != "" {
		return DeliverTxRouter(method, param, nonce, signature, nodeID, app)
	}
	return ReturnDeliverTxLog(code.MethodCanNotBeEmpty, "method can not be empty", "")
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
	// After scale test delete this
	if isValidatorTx(tx) {
		return ReturnCheckTx(true)
	}
	// ---------------------

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

	if method != "" && param != "" && nonce != "" && signature != "" && nodeID != "" {
		// return CheckTxRouter(method, param, nonce, signature, nodeID, app)

		// If can decode and field != "" always return true
		return ReturnCheckTx(true)
	} else {
		return ReturnCheckTx(false)
	}
}

func (app *DIDApplication) Commit() types.ResponseCommit {
	fmt.Println("Commit")
	newAppHashString := ""
	for _, key := range app.state.UncommitKeys {
		value := app.state.db.Get(prefixKey([]byte(key)))
		if value != nil {
			newAppHashString += string(key) + string(value)
		}
	}
	h := sha256.New()
	if newAppHashString != "" {
		dbStat := app.state.db.Stats()
		newAppHashStr := app.state.CommitStr + newAppHashString + dbStat["database.size"]
		h.Write([]byte(newAppHashStr))
		newAppHash := h.Sum(nil)
		app.state.CommitStr = hex.EncodeToString(newAppHash)
	}
	app.state.AppHash = []byte(app.state.CommitStr)
	app.state.Height++
	saveState(app.state)
	app.state.UncommitKeys = nil
	return types.ResponseCommit{Data: app.state.AppHash}
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
