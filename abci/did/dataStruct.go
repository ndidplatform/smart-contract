package did

type NodePublicKey struct {
	NodeID    string `json:"node_id"`
	PublicKey string `json:"public_key"`
}

type GetNodePublicKeyParam struct {
	NodeID string `json:"node_id"`
}

type GetNodePublicKeyResult struct {
	PublicKey string `json:"public_key"`
}

type User struct {
	HashID string  `json:"hash_id"`
	Ial    float64 `json:"ial"`
}

type RegisterMsqDestinationParam struct {
	Users  []User `json:"users"`
	NodeID string `json:"node_id"`
}

type Node struct {
	Ial    float64 `json:"ial"`
	NodeID string  `json:"node_id"`
}

type MsqDestination struct {
	Nodes []Node `json:"nodes"`
}

type GetMsqDestinationParam struct {
	HashID string  `json:"hash_id"`
	MinIal float64 `json:"min_ial"`
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
	MinAal          float64       `json:"min_aal"`
	MinIal          float64       `json:"min_ial"`
	Timeout         int           `json:"timeout"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"message_hash"`
	Responses       []Response    `json:"responses"`
	IsClosed        bool          `json:"is_closed"`
	IsTimedOut      bool          `json:"is_timed_out"`
	SignDataCount   int           `json:"sign_data_count"`
}

type Response struct {
	RequestID     string  `json:"request_id"`
	Aal           float64 `json:"aal"`
	Ial           float64 `json:"ial"`
	Status        string  `json:"status"`
	Signature     string  `json:"signature"`
	AccessorID    string  `json:"accessor_id"`
	IdentityProof string  `json:"identity_proof"`
}

type GetRequestParam struct {
	RequestID string `json:"requestId"`
}

type GetRequestResult struct {
	Status      string `json:"status"`
	IsClosed    bool   `json:"is_closed"`
	IsTimedOut  bool   `json:"is_timed_out"`
	MessageHash string `json:"messageHash"`
}

type Callback struct {
	RequestID string `json:"requestId"`
	Height    int64  `json:"height"`
}

type SignDataParam struct {
	NodeID    string `json:"node_id"`
	RequestID string `json:"request_id"`
	Signature string `json:"signature"`
}

type RegisterServiceDestinationParam struct {
	AsServiceID string `json:"service_id"`
	NodeID      string `json:"node_id"`
}

type GetServiceDestinationParam struct {
	AsServiceID string `json:"service_id"`
}

type GetServiceDestinationResult struct {
	NodeID []string `json:"node_id"`
}

type InitNDIDParam struct {
	NodeID    string `json:"node_id"`
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

type RegisterMsqAddressParam struct {
	NodeID string `json:"node_id"`
	IP     string `json:"ip"`
	Port   int64  `json:"port"`
}

type GetMsqAddressParam struct {
	NodeID string `json:"node_id"`
}

type MsqAddress struct {
	IP   string `json:"ip"`
	Port int64  `json:"port"`
}

type SetNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

type AddNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

type ReduceNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

type GetNodeTokenParam struct {
	NodeID string `json:"node_id"`
}

type GetNodeTokenResult struct {
	Amount float64 `json:"amount"`
}

type SetPriceFuncParam struct {
	Func  string  `json:"func"`
	Price float64 `json:"price"`
}

type GetPriceFuncParam struct {
	Func string `json:"func"`
}

type GetPriceFuncResult struct {
	Price float64 `json:"price"`
}

type Report struct {
	Method string  `json:"method"`
	Price  float64 `json:"price"`
	Data   string  `json:"data"`
}

type GetUsedTokenReportParam struct {
	NodeID string `json:"node_id"`
}

type RequestIDParam struct {
	RequestID string `json:"requestId"`
}
