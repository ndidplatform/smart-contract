package did

import (
	"encoding/json"
	"fmt"

	"github.com/ndidplatform/smart-contract/abci/code"
	"github.com/tendermint/abci/types"
)

func signData(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("SignData")
	var signData SignDataParam
	err := json.Unmarshal([]byte(param), &signData)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	requestKey := "Request" + "|" + signData.RequestID
	requestJSON := app.state.db.Get(prefixKey([]byte(requestKey)))
	if requestJSON == nil {
		return ReturnDeliverTxLog(code.RequestIDNotFound, "Request ID not found", "")
	}
	var request Request
	err = json.Unmarshal([]byte(requestJSON), &request)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Check IsClosed
	if request.IsClosed {
		return ReturnDeliverTxLog(code.RequestIsClosed, "Request is closed", "")
	}

	// Check IsTimedOut
	if request.IsTimedOut {
		return ReturnDeliverTxLog(code.RequestIsTimedOut, "Request is timed out", "")
	}

	// Check nodeID is exist in as_id_list
	exist := false
	for _, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == signData.ServiceID {
			for _, as := range dataRequest.As {
				if as == nodeID {
					exist = true
					break
				}
			}
		}
	}
	if exist == false {
		return ReturnDeliverTxLog(code.NodeIDIsNotExistInASList, "Node ID is not exist in AS list", "")
	}

	signDataKey := "SignData" + "|" + signData.Signature
	signDataJSON, err := json.Marshal(signData)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	// Update answered_as_id_list in request
	for index, dataRequest := range request.DataRequestList {
		if dataRequest.ServiceID == signData.ServiceID {
			request.DataRequestList[index].AnsweredAsIdList = append(dataRequest.AnsweredAsIdList, nodeID)
		}
	}

	requestJSON, err = json.Marshal(request)
	if err != nil {
		return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}

	app.SetStateDB([]byte(requestKey), []byte(requestJSON))
	app.SetStateDB([]byte(signDataKey), []byte(signDataJSON))
	return ReturnDeliverTxLog(code.OK, "success", signData.RequestID)
}

func registerServiceDestination(param string, app *DIDApplication, nodeID string) types.ResponseDeliverTx {
	fmt.Println("RegisterServiceDestination")
	var funcParam RegisterServiceDestinationParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}

	// Add Service Detail
	key := "Service" + "|" + funcParam.AsServiceID + "|" + funcParam.NodeID
	chkExists := app.state.db.Get(prefixKey([]byte(key)))
	if chkExists != nil {
		return ReturnDeliverTxLog(code.DuplicateServiceID, "Duplicate service ID", "")
	} else {
		var service = Service{
			funcParam.ServiceName,
			funcParam.MinIal,
			funcParam.MinAal,
		}
		value, err := json.Marshal(service)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
	}

	// Add ServiceDestination
	key = "ServiceDestination" + "|" + funcParam.AsServiceID
	chkExists = app.state.db.Get(prefixKey([]byte(key)))

	if chkExists != nil {
		var nodes GetAsNodesByServiceIdResult
		err := json.Unmarshal([]byte(chkExists), &nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
		}
		var newNode = ASNode{
			funcParam.NodeID,
			getNodeNameByNodeID(funcParam.NodeID, app),
			funcParam.MinIal,
			funcParam.MinAal,
			funcParam.ServiceName,
		}
		nodes.Node = append(nodes.Node, newNode)
		value, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
	} else {
		var nodes GetAsNodesByServiceIdResult
		var newNode = ASNode{
			funcParam.NodeID,
			getNodeNameByNodeID(funcParam.NodeID, app),
			funcParam.MinIal,
			funcParam.MinAal,
			funcParam.ServiceName,
		}
		nodes.Node = append(nodes.Node, newNode)
		value, err := json.Marshal(nodes)
		if err != nil {
			return ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
		}
		app.SetStateDB([]byte(key), []byte(value))
	}
	return ReturnDeliverTxLog(code.OK, "success", "")
}
