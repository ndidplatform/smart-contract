package types

type NodeRole string

const (
	NodeRoleNdid  NodeRole = "NDID"
	NodeRoleRp    NodeRole = "RP"
	NodeRoleIdp   NodeRole = "IdP"
	NodeRoleAs    NodeRole = "AS"
	NodeRoleProxy NodeRole = "Proxy"
)
