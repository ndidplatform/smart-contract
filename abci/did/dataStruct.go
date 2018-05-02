package did

type NodePublicKey struct {
	NodeID    string `json:"node_id"`
	PublicKey string `json:"public_key"`
}

type GetNodePublicKeyParam struct {
	NodeID string `json:"node_id"`
}

type GetNodePublicKeyPesult struct {
	PublicKey string `json:"public_key"`
}

type User struct {
	HashID string `json:"hash_id"`
	Ial    int    `json:"ial"`
}

type RegisterMsqDestinationParam struct {
	Users  []User `json:"users"`
	NodeID string `json:"node_id"`
}

type Node struct {
	Ial    int    `json:"ial"`
	NodeID string `json:"node_id"`
}

type MsqDestination struct {
	Nodes []Node `json:"nodes"`
}

type GetMsqDestinationParam struct {
	HashID string `json:"hash_id"`
	MinIal int    `json:"min_ial"`
}

type GetMsqDestinationResult struct {
	NodeID []string `json:"node_id"`
}

type AccessorMethod struct {
	AccessorID   string `json:"accessor_id"`
	AccessorType string `json:"accessor_type"`
	AccessorKey  string `json:"accessor_key"`
	Commitment   string `json:"commitment"`
}

type GetAccessorMethodParam struct {
	AccessorID string `json:"accessor_id"`
}

type GetAccessorMethodResult struct {
	AccessorType string `json:"accessor_type"`
	AccessorKey  string `json:"accessor_key"`
	Commitment   string `json:"commitment"`
}

type DataRequest struct {
	ServiceID         string   `json:"service_id"`
	As                []string `json:"as_id_list"`
	Count             int      `json:"count"`
	RequestParamsHash string   `json:"request_params_hash"`
}

type Request struct {
	RequestID       string        `json:"request_id"`
	MinIdp          int           `json:"min_idp"`
	MinAal          int           `json:"min_aal"`
	MinIal          int           `json:"min_ial"`
	Timeout         int           `json:"timeout"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"message_hash"`
	Responses       []Response    `json:"responses"`
}

type Response struct {
	RequestID     string `json:"request_id"`
	Aal           int    `json:"aal"`
	Ial           int    `json:"ial"`
	Status        string `json:"status"`
	Signature     string `json:"signature"`
	AccessorID    string `json:"accessor_id"`
	IdentityProof string `json:"identity_proof"`
}

type GetRequestParam struct {
	RequestID string `json:"requestId"`
}

type GetRequestResult struct {
	Status      string `json:"status"`
	MessageHash string `json:"messageHash"`
}

type Callback struct {
	RequestID string `json:"requestId"`
}

type SignDataParam struct {
	AsID      string `json:"as_id"`
	RequestID string `json:"request_id"`
	Signature string `json:"signature"`
}

type RegisterServiceDestinationParam struct {
	AsID        string `json:"as_id"`
	AsServiceID string `json:"service_id"`
	NodeID      string `json:"node_id"`
}

type GetServiceDestinationParam struct {
	AsID        string `json:"as_id"`
	AsServiceID string `json:"service_id"`
}

type GetServiceDestinationResult struct {
	NodeID string `json:"node_id"`
}

type InitNDIDParam struct {
	PublicKey string `json:"public_key"`
}

type TransferNDIDParam struct {
	PublicKey string `json:"public_key"`
}

type RegisterNode struct {
	NodeID    string `json:"node_id"`
	PublicKey string `json:"public_key"`
	Role      string `json:"role"`
}
