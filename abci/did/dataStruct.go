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
