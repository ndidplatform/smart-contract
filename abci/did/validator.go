package did

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}

func (app *DIDApplication) Validators() (validators []types.Validator) {
	app.logger.Infof("Validators")
	itr := app.state.db.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		if isValidatorTx(itr.Key()) {
			validator := new(types.Validator)
			err := types.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
	}
	return
}

// format is "val:pubkey"tx
func (app *DIDApplication) execValidatorTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ValidatorSetChangePrefix):]

	// TODO change get PubKey and Power when got ValidatorTx
	// Use "@" as separator since pubKey is base64 and may contain "/"
	pubKeyAndPower := strings.Split(string(tx), "@")
	if len(pubKeyAndPower) < 1 {
		return types.ResponseDeliverTx{
			Code: code.EncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey'. Got %v", pubKeyAndPower),
		}
	}
	pubkeyS, powerS := pubKeyAndPower[0], "10"
	if len(pubKeyAndPower) > 1 {
		powerS = "0"
	}

	// publicKey, _ := base64.StdEncoding.DecodeString(pubkeyS)
	publicKey := pubkeyS
	pubKey, _ := base64.StdEncoding.DecodeString(string(publicKey))
	var pubKeyEd crypto.PubKeyEd25519
	copy(pubKeyEd[:], pubKey)

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.EncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	// update
	return app.updateValidator(types.Validator{pubKeyEd.Bytes(), power})
}

// add, update, or remove a validator
func (app *DIDApplication) updateValidator(v types.Validator) types.ResponseDeliverTx {
	key := []byte("val:" + base64.StdEncoding.EncodeToString(v.PubKey))

	if v.Power == 0 {
		// remove validator
		if !app.state.db.Has(key) {
			return types.ResponseDeliverTx{
				Code: code.Unauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %X", key)}
		}
		app.state.db.Delete(key)
		app.state.Size--
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return types.ResponseDeliverTx{
				Code: code.EncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.state.db.Set(key, value.Bytes())
		app.state.Size++
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)

	return ReturnDeliverTxLog(code.OK, "success", "")
}

func setValidator(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetValidator, Parameter: %s", param)
	var funcParam SetValidatorParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	pubKey, err := base64.StdEncoding.DecodeString(string(funcParam.PublicKey))
	if err != nil {
		return ReturnDeliverTxLog(code.DecodingError, err.Error(), "")
	}
	var pubKeyEd crypto.PubKeyEd25519
	copy(pubKeyEd[:], pubKey)

	return app.updateValidator(types.Validator{pubKeyEd.Bytes(), funcParam.Power})
}
