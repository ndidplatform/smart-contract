package did

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tendermint/abci/example/code"
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

//---------------------------------------------------

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

func (app *DIDApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	fmt.Println("DeliverTx")
	var key, value []byte
	parts := strings.Split(string(tx), ",")

	fmt.Println(string(tx))

	method := parts[1]
	namespace := parts[2]
	identifier := parts[3]

	if method == "CreateIdentity" {
		fmt.Println("CreateIdentity")
		key, uuid := namespace+"|"+identifier, namespace+identifier // TODO change UUID

		//check exist
		value := app.state.db.Get(prefixKey([]byte(key)))
		if value != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("identify already exists")}
		}

		app.state.db.Set(prefixKey([]byte(key)), []byte(uuid))
		app.state.Size += 1

		return types.ResponseDeliverTx{
			Code: code.CodeTypeOK,
			Log:  fmt.Sprintf("success")}

	} else if method == "CreateIDPResponse" {
		fmt.Println("CreateIDPResponse")
		// TODO add logic for store idp response
		return types.ResponseDeliverTx{
			Code: code.CodeTypeOK,
			Log:  fmt.Sprintf("success")}
	} else {
		fmt.Println("else")
		key, value = tx, tx
		app.state.db.Set(key, value)

		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("fail")}
	}
}

func (app *DIDApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	fmt.Println("CheckTx")
	return types.ResponseCheckTx{Code: code.CodeTypeOK}
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

func (app *DIDApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	fmt.Println("Query")
	fmt.Println(string(reqQuery.Data))
	parts := strings.Split(string(reqQuery.Data), ",")
	method := parts[0]
	namespace := parts[1]
	identifier := parts[2]

	if method == "GetIdentifier" {

		key := namespace + "|" + identifier
		value := app.state.db.Get(prefixKey([]byte(key)))

		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}

	} else if method == "CreateIDPResponse" {
		// TODO add query logic for idp response
		resQuery.Log = "success"
	} else {
		resQuery.Log = "wrong method name"
		return
	}

	return
}
