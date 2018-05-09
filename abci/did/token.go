package did

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/tendermint/abci/types"
)

func getTokenPriceByFunc(fnName string, app *DIDApplication) float64 {
	key := "TokenPriceFunc" + "|" + fnName
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, _ := strconv.ParseFloat(string(value), 64)
		return s
	}
	return 1
}

func createTokenAccount(nodeID string, app *DIDApplication) {
	key := "Token" + "|" + nodeID
	value := strconv.FormatFloat(0, 'f', -1, 64)
	app.state.Size++
	app.state.db.Set(prefixKey([]byte(key)), []byte(value))
}

func setToken(nodeID string, amount float64, app *DIDApplication) error {
	key := "Token" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		value := strconv.FormatFloat(amount, 'f', -1, 64)
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
		return nil
	}
	return errors.New("not found token account")
}

func addToken(nodeID string, amount float64, app *DIDApplication) error {
	key := "Token" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return err
		}
		s = s + amount
		value := strconv.FormatFloat(s, 'f', -1, 64)
		app.state.Size++
		app.state.db.Set(prefixKey([]byte(key)), []byte(value))
		return nil
	}
	return errors.New("not found token account")
}

func reduceToken(nodeID string, amount float64, app *DIDApplication) error {
	key := "Token" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return err
		}
		if s-amount >= 0 {
			s = s - amount
			value := strconv.FormatFloat(s, 'f', -1, 64)
			app.state.Size++
			app.state.db.Set(prefixKey([]byte(key)), []byte(value))
			return nil
		}
		return errors.New("token not enough")
	}
	return errors.New("not found token account")
}

func getToken(nodeID string, app *DIDApplication) (float64, error) {
	key := "Token" + "|" + nodeID
	value := app.state.db.Get(prefixKey([]byte(key)))
	if value != nil {
		s, _ := strconv.ParseFloat(string(value), 64)
		return s, nil
	}
	return 0, errors.New("not found token account")
}

func setNodeToken(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("SetNodeToken")
	var funcParam SetNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	err = setToken(funcParam.NodeID, funcParam.Amount, app)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	return ReturnDeliverTxLog("success")
}

func addNodeToken(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("AddNodeToken")
	var funcParam AddNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	err = addToken(funcParam.NodeID, funcParam.Amount, app)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	return ReturnDeliverTxLog("success")
}

func reduceNodeToken(param string, app *DIDApplication) types.ResponseDeliverTx {
	fmt.Println("ReduceNodeToken")
	var funcParam ReduceNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	err = reduceToken(funcParam.NodeID, funcParam.Amount, app)
	if err != nil {
		return ReturnDeliverTxLog(err.Error())
	}
	return ReturnDeliverTxLog("success")
}

func getNodeToken(param string, app *DIDApplication) types.ResponseQuery {
	fmt.Println("GetNodeToken")
	var funcParam GetNodeTokenParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	tokenAmount, err := getToken(funcParam.NodeID, app)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var res = GetNodeTokenResult{
		tokenAmount,
	}
	value, err := json.Marshal(res)
	if err != nil {
		return ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return ReturnQuery(value, "success", app.state.Height)
}
