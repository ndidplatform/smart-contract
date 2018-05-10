package did

import (
	"bytes"
	"encoding/hex"
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
	fmt.Println("Validators")
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
	pubKeyAndPower := strings.Split(string(tx), "/")
	if len(pubKeyAndPower) < 1 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey'. Got %v", pubKeyAndPower),
		}
	}
	pubkeyS, powerS := pubKeyAndPower[0], "10"
	if len(pubKeyAndPower) > 1 {
		powerS = "0"
	}

	//get the pubkey and power
	// pubKeyAndPower := strings.Split(string(tx), "/")
	// if len(pubKeyAndPower) != 2 {
	// 	return types.ResponseDeliverTx{
	// 		Code: code.CodeTypeEncodingError,
	// 		Log:  fmt.Sprintf("Expected 'pubkey/power'. Got %v", pubKeyAndPower)}
	// }
	// pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]

	// decode the pubkey, ensuring its go-crypto encoded
	pubkey, err := hex.DecodeString(pubkeyS)
	var pubKeyEd crypto.PubKeyEd25519
	copy(pubKeyEd[:], pubkey)

	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%s) is invalid hex", pubkeyS)}
	}
	/*_, err = crypto.PubKeyFromBytes(pubkey)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%X) is invalid go-crypto encoded", pubkey)}
	}*/

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	// update
	return app.updateValidator(types.Validator{pubKeyEd.Bytes(), power})
}

// add, update, or remove a validator
func (app *DIDApplication) updateValidator(v types.Validator) types.ResponseDeliverTx {
	key := []byte("val:" + string(v.PubKey))
	if v.Power == 0 {
		// remove validator
		if !app.state.db.Has(key) {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeUnauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %X", key)}
		}
		app.state.db.Delete(key)
		app.state.Size--
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.state.db.Set(key, value.Bytes())
		app.state.Size++
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)

	return types.ResponseDeliverTx{Code: code.CodeTypeOK}
}
