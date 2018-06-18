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

type GetNodeMasterPublicKeyParam struct {
	NodeID string `json:"node_id"`
}

type GetNodeMasterPublicKeyResult struct {
	MasterPublicKey string `json:"master_public_key"`
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

type GetIdpNodesParam struct {
	HashID string  `json:"hash_id"`
	MinIal float64 `json:"min_ial"`
	MinAal float64 `json:"min_aal"`
}

type MsqDestinationNode struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	MaxIal float64 `json:"max_ial"`
	MaxAal float64 `json:"max_aal"`
}

type GetIdpNodesResult struct {
	Node []MsqDestinationNode `json:"node"`
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
	ServiceID            string   `json:"service_id"`
	As                   []string `json:"as_id_list"`
	Count                int      `json:"min_as"`
	RequestParamsHash    string   `json:"request_params_hash"`
	AnsweredAsIdList     []string `json:"answered_as_id_list"`
	ReceivedDataFromList []string `json:"received_data_from_list"`
}

type Request struct {
	RequestID       string        `json:"request_id"`
	MinIdp          int           `json:"min_idp"`
	MinAal          float64       `json:"min_aal"`
	MinIal          float64       `json:"min_ial"`
	Timeout         int           `json:"request_timeout"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"request_message_hash"`
	Responses       []Response    `json:"responses"`
	IsClosed        bool          `json:"closed"`
	IsTimedOut      bool          `json:"timed_out"`
	CanAddAccessor  bool          `json:"can_add_accessor"`
	Owner           string        `json:"owner"`
	Mode            int           `json:"mode"`
}

type Response struct {
	RequestID        string  `json:"request_id"`
	Aal              float64 `json:"aal"`
	Ial              float64 `json:"ial"`
	Status           string  `json:"status"`
	Signature        string  `json:"signature"`
	IdentityProof    string  `json:"identity_proof"`
	PrivateProofHash string  `json:"private_proof_hash"`
	IdpID            string  `json:"idp_id"`
}

type GetRequestParam struct {
	RequestID string `json:"requestId"`
}

type GetRequestResult struct {
	IsClosed    bool   `json:"closed"`
	IsTimedOut  bool   `json:"timed_out"`
	MessageHash string `json:"request_message_hash"`
	Mode        int    `json:"mode"`
}

type GetRequestDetailResult struct {
	RequestID       string        `json:"request_id"`
	MinIdp          int           `json:"min_idp"`
	MinAal          float64       `json:"min_aal"`
	MinIal          float64       `json:"min_ial"`
	Timeout         int           `json:"request_timeout"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"request_message_hash"`
	Responses       []Response    `json:"responses"`
	IsClosed        bool          `json:"closed"`
	IsTimedOut      bool          `json:"timed_out"`
	Special         bool          `json:"special"`
	Mode            int           `json:"mode"`
}

type Callback struct {
	RequestID string `json:"requestId"`
	Height    int64  `json:"height"`
}

type SignDataParam struct {
	ServiceID string `json:"service_id"`
	RequestID string `json:"request_id"`
	Signature string `json:"signature"`
}

type AddServiceParam struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type DeleteServiceParam struct {
	ServiceID string `json:"service_id"`
}

type RegisterServiceDestinationParam struct {
	ServiceID string  `json:"service_id"`
	NodeID    string  `json:"node_id"`
	MinIal    float64 `json:"min_ial"`
	MinAal    float64 `json:"min_aal"`
}

type Service struct {
	ServiceName string `json:"service_name"`
}

type GetServiceDetailParam struct {
	ServiceID string `json:"service_id"`
}

type GetAsNodesByServiceIdParam struct {
	ServiceID string `json:"service_id"`
}

type ASNode struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	MinIal    float64 `json:"min_ial"`
	MinAal    float64 `json:"min_aal"`
	ServiceID string  `json:"service_id"`
}

type GetAsNodesByServiceIdResult struct {
	Node []ASNode `json:"node"`
}

type ASNodeResult struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	MinIal      float64 `json:"min_ial"`
	MinAal      float64 `json:"min_aal"`
	ServiceName string  `json:"service_name"`
}

type GetAsNodesByServiceIdWithNameResult struct {
	Node []ASNodeResult `json:"node"`
}

type InitNDIDParam struct {
	NodeID    string `json:"node_id"`
	PublicKey string `json:"public_key"`
}

type TransferNDIDParam struct {
	PublicKey string `json:"public_key"`
}

type RegisterNode struct {
	NodeID          string  `json:"node_id"`
	PublicKey       string  `json:"public_key"`
	MasterPublicKey string  `json:"master_public_key"`
	NodeName        string  `json:"node_name"`
	Role            string  `json:"role"`
	MaxIal          float64 `json:"max_ial"`
	MaxAal          float64 `json:"max_aal"`
}

type NodeDetail struct {
	PublicKey       string `json:"public_key"`
	MasterPublicKey string `json:"master_public_key"`
	NodeName        string `json:"node_name"`
}

type MaxIalAal struct {
	MaxIal float64 `json:"max_ial"`
	MaxAal float64 `json:"max_aal"`
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

type Namespace struct {
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

type DeleteNamespaceParam struct {
	Namespace string `json:"namespace"`
}

type UpdateNodeParam struct {
	PublicKey       string `json:"public_key"`
	MasterPublicKey string `json:"master_public_key"`
}

type CreateIdentityParam struct {
	AccessorID        string `json:"accessor_id"`
	AccessorType      string `json:"accessor_type"`
	AccessorPublicKey string `json:"accessor_public_key"`
	AccessorGroupID   string `json:"accessor_group_id"`
}

type Accessor struct {
	AccessorType      string `json:"accessor_type"`
	AccessorPublicKey string `json:"accessor_public_key"`
	AccessorGroupID   string `json:"accessor_group_id"`
}

type AccessorMethod struct {
	AccessorID        string `json:"accessor_id"`
	AccessorType      string `json:"accessor_type"`
	AccessorPublicKey string `json:"accessor_public_key"`
	AccessorGroupID   string `json:"accessor_group_id"`
	RequestID         string `json:"request_id"`
}

type CheckExistingIdentityParam struct {
	HashID string `json:"hash_id"`
}

type CheckExistingIdentityResult struct {
	Exist bool `json:"exist"`
}

type GetAccessorGroupIDParam struct {
	AccessorID string `json:"accessor_id"`
}

type GetAccessorGroupIDResult struct {
	AccessorGroupID string `json:"accessor_group_id"`
}

type GetAccessorKeyParam struct {
	AccessorID string `json:"accessor_id"`
}

type GetAccessorKeyResult struct {
	AccessorPublicKey string `json:"accessor_public_key"`
}

type SetValidatorParam struct {
	PublicKey string `json:"public_key"`
	Power     int64  `json:"power"`
}

type SetDataReceivedParam struct {
	RequestID string `json:"requestId"`
	ServiceID string `json:"service_id"`
	AsID      string `json:"as_id"`
}

type ServiceDetail struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// type GetNodeInfoResult struct {
// 	Version string `json:"version"`
// }

type CheckExistingAccessorIDParam struct {
	AccessorID string `json:"accessor_id"`
}

type CheckExistingAccessorGroupIDParam struct {
	AccessorGroupID string `json:"accessor_group_id"`
}

type CheckExistingResult struct {
	Exist bool `json:"exist"`
}

type GetNodeInfoParam struct {
	NodeID string `json:"node_id"`
}

type GetNodeInfoResult struct {
	PublicKey       string  `json:"public_key"`
	MasterPublicKey string  `json:"master_public_key"`
	NodeName        string  `json:"node_name"`
	Role            string  `json:"role"`
	MaxIal          float64 `json:"max_ial"`
	MaxAal          float64 `json:"max_aal"`
}

type GetIdentityInfoParam struct {
	HashID string `json:"hash_id"`
	NodeID string `json:"node_id"`
}

type GetIdentityInfoResult struct {
	Ial float64 `json:"ial"`
}
