package did

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/sirupsen/logrus"
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
	logger     *logrus.Entry
	Version    string
}

func NewDIDApplication() *DIDApplication {
	logger := logrus.WithFields(logrus.Fields{"module": "abci-app"})
	logger.Infoln("NewDIDApplication")
	var dbDir = getEnv("DB_NAME", "DID")
	name := "didDB"
	db := dbm.NewDB(name, "leveldb", dbDir)
	state := loadState(db)
	return &DIDApplication{state: state,
		logger:  logger,
		Version: "0.0.1", // Hard code set version
	}
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
	res.Version = app.Version
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	return res
}

// Save the validators in the merkle tree
func (app *DIDApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}

// Track the block hash and header information
func (app *DIDApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.logger.Infof("BeginBlock: %d", req.Header.Height)
	// reset valset changes
	app.ValUpdates = make([]types.Validator, 0)
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *DIDApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.logger.Infof("EndBlock: %d", req.Height)
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

func (app *DIDApplication) DeliverTx(tx []byte) (res types.ResponseDeliverTx) {
	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Error("Recovered in f", r)
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
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]
	nonce := parts[2]
	signature := parts[3]
	nodeID := parts[4]

	app.logger.Infof("DeliverTx: %s, NodeID: %s", method, nodeID)

	if method != "" {
		return DeliverTxRouter(method, param, nonce, signature, nodeID, app)
	}
	return ReturnDeliverTxLog(code.MethodCanNotBeEmpty, "method can not be empty", "")
}

func (app *DIDApplication) CheckTx(tx []byte) (res types.ResponseCheckTx) {
	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Error("Recovered in f", r)
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
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]
	nonce := parts[2]
	signature := parts[3]
	nodeID := parts[4]

	app.logger.Infof("CheckTx: %s, NodeID: %s", method, nodeID)

	if method != "" && param != "" && nonce != "" && signature != "" && nodeID != "" {
		// If can decode and field != "" always return true
		return ReturnCheckTx(true)
	} else {
		return ReturnCheckTx(false)
	}
}

func (app *DIDApplication) Commit() types.ResponseCommit {
	app.logger.Infof("Commit")
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

	// Recover when panic
	defer func() {
		if r := recover(); r != nil {
			app.logger.Error("Recovered in f", r)
			res = ReturnQuery(nil, "wrong query format", app.state.Height, app)
		}
	}()

	txString, err := base64.StdEncoding.DecodeString(string(reqQuery.Data))
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height, app)
	}
	parts := strings.Split(string(txString), "|")

	method := parts[0]
	param := parts[1]

	app.logger.Infof("Query: %s", method)

	if method != "" {
		return QueryRouter(method, param, app)
	}
	return ReturnQuery(nil, "method can't empty", app.state.Height, app)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
