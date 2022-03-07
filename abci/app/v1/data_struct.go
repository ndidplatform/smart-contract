/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package app

import "time"

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

type Identity struct {
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
}

type RegisterIdentityParam struct {
	ReferenceGroupCode string     `json:"reference_group_code"`
	NewIdentityList    []Identity `json:"new_identity_list"`
	Ial                float64    `json:"ial"`
	Lial               *bool      `json:"lial"`
	Laal               *bool      `json:"laal"`
	ModeList           []int32    `json:"mode_list"`
	AccessorID         string     `json:"accessor_id"`
	AccessorPublicKey  string     `json:"accessor_public_key"`
	AccessorType       string     `json:"accessor_type"`
	RequestID          string     `json:"request_id"`
}

type AddIdentityParam struct {
	ReferenceGroupCode string     `json:"reference_group_code"`
	NewIdentityList    []Identity `json:"new_identity_list"`
	RequestID          string     `json:"request_id"`
}

type Node struct {
	Ial          float64 `json:"ial"`
	NodeID       string  `json:"node_id"`
	Active       bool    `json:"active"`
	First        bool    `json:"first"`
	TimeoutBlock int64   `json:"timeout_block"`
}

type MsqDestination struct {
	Nodes []Node `json:"nodes"`
}

type GetIdpNodesParam struct {
	ReferenceGroupCode                     string   `json:"reference_group_code"`
	IdentityNamespace                      string   `json:"identity_namespace"`
	IdentityIdentifierHash                 string   `json:"identity_identifier_hash"`
	FilterForNodeID                        *string  `json:"filter_for_node_id"`
	IsIdpAgent                             *bool    `json:"agent"`
	MinAal                                 float64  `json:"min_aal"`
	MinIal                                 float64  `json:"min_ial"`
	OnTheFlySupport                        *bool    `json:"on_the_fly_support"`
	NodeIDList                             []string `json:"node_id_list"`
	SupportedRequestMessageDataUrlTypeList []string `json:"supported_request_message_data_url_type_list"`
	ModeList                               []int32  `json:"mode_list"`
}

type MsqDestinationNode struct {
	ID                                     string   `json:"node_id"`
	Name                                   string   `json:"node_name"`
	MaxIal                                 float64  `json:"max_ial"`
	MaxAal                                 float64  `json:"max_aal"`
	OnTheFlySupport                        bool     `json:"on_the_fly_support"`
	Ial                                    *float64 `json:"ial,omitempty"`
	Lial                                   *bool    `json:"lial"`
	Laal                                   *bool    `json:"laal"`
	ModeList                               *[]int32 `json:"mode_list,omitempty"`
	SupportedRequestMessageDataUrlTypeList []string `json:"supported_request_message_data_url_type_list"`
	IsIdpAgent                             bool     `json:"agent"`
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

type ASResponse struct {
	AsID         string `json:"as_id"`
	Signed       *bool  `json:"signed,omitempty"`
	ReceivedData *bool  `json:"received_data,omitempty"`
	ErrorCode    *int32 `json:"error_code,omitempty"`
}

type DataRequest struct {
	ServiceID         string       `json:"service_id"`
	As                []string     `json:"as_id_list"`
	Count             int          `json:"min_as"`
	RequestParamsHash string       `json:"request_params_hash"`
	ResponseList      []ASResponse `json:"response_list"`
}

type CreateRequestParam struct {
	RequestID       string        `json:"request_id"`
	MinIdp          int           `json:"min_idp"`
	MinAal          float64       `json:"min_aal"`
	MinIal          float64       `json:"min_ial"`
	Timeout         int           `json:"request_timeout"`
	IdPIDList       []string      `json:"idp_id_list"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"request_message_hash"`
	Purpose         string        `json:"purpose"`
	Mode            int32         `json:"mode"`
}

type Response struct {
	Ial            *float64 `json:"ial,omitempty"`
	Aal            *float64 `json:"aal,omitempty"`
	Status         *string  `json:"status,omitempty"`
	Signature      *string  `json:"signature,omitempty"`
	IdpID          string   `json:"idp_id"`
	ValidIal       *bool    `json:"valid_ial"`
	ValidSignature *bool    `json:"valid_signature"`
	ErrorCode      *int32   `json:"error_code,omitempty"`
}

type CreateIdpResponseParam struct {
	Aal       float64 `json:"aal"`
	Ial       float64 `json:"ial"`
	RequestID string  `json:"request_id"`
	Signature string  `json:"signature"`
	Status    string  `json:"status"`
	ErrorCode *int32  `json:"error_code"`
}

type CreateMessageParam struct {
	MessageID string `json:"message_id"`
	Message   string `json:"message"`
	Purpose   string `json:"purpose"`
}

type GetRequestParam struct {
	RequestID string `json:"request_id"`
}

type GetRequestResult struct {
	IsClosed    bool   `json:"closed"`
	IsTimedOut  bool   `json:"timed_out"`
	MessageHash string `json:"request_message_hash"`
	Mode        int32  `json:"mode"`
}

type GetRequestDetailResult struct {
	RequestID           string        `json:"request_id"`
	MinIdp              int           `json:"min_idp"`
	MinAal              float64       `json:"min_aal"`
	MinIal              float64       `json:"min_ial"`
	Timeout             int           `json:"request_timeout"`
	IdPIDList           []string      `json:"idp_id_list"`
	DataRequestList     []DataRequest `json:"data_request_list"`
	MessageHash         string        `json:"request_message_hash"`
	Responses           []Response    `json:"response_list"`
	IsClosed            bool          `json:"closed"`
	IsTimedOut          bool          `json:"timed_out"`
	Purpose             string        `json:"purpose"`
	Mode                int32         `json:"mode"`
	RequesterNodeID     string        `json:"requester_node_id"`
	CreationBlockHeight int64         `json:"creation_block_height"`
	CreationChainID     string        `json:"creation_chain_id"`
}

type GetMessageParam struct {
	MessageID string `json:"message_id"`
}

type GetMessageResult struct {
	Message string `json:"message"`
}

type GetMessageDetailResult struct {
	MessageID           string `json:"message_id"`
	Message             string `json:"message"`
	Purpose             string `json:"purpose"`
	RequesterNodeID     string `json:"requester_node_id"`
	CreationBlockHeight int64  `json:"creation_block_height"`
	CreationChainID     string `json:"creation_chain_id"`
}

type CreateAsResponseParam struct {
	ServiceID string `json:"service_id"`
	RequestID string `json:"request_id"`
	Signature string `json:"signature"`
	ErrorCode *int32 `json:"error_code"`
}

type AddServiceParam struct {
	ServiceID         string `json:"service_id"`
	ServiceName       string `json:"service_name"`
	DataSchema        string `json:"data_schema"`
	DataSchemaVersion string `json:"data_schema_version"`
}

type DisableServiceParam struct {
	ServiceID string `json:"service_id"`
}

type RegisterServiceDestinationParam struct {
	MinAal                 float64  `json:"min_aal"`
	MinIal                 float64  `json:"min_ial"`
	ServiceID              string   `json:"service_id"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

type GetServiceDetailParam struct {
	ServiceID string `json:"service_id"`
}

type GetAsNodesByServiceIdParam struct {
	ServiceID  string   `json:"service_id"`
	NodeIDList []string `json:"node_id_list"`
}

type ASNode struct {
	ID        string  `json:"node_id"`
	Name      string  `json:"node_name"`
	MinIal    float64 `json:"min_ial"`
	MinAal    float64 `json:"min_aal"`
	ServiceID string  `json:"service_id"`
	Active    bool    `json:"active"`
}

type GetAsNodesByServiceIdResult struct {
	Node []ASNode `json:"node"`
}

type ASNodeResult struct {
	ID                     string   `json:"node_id"`
	Name                   string   `json:"node_name"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

type GetAsNodesByServiceIdWithNameResult struct {
	Node []ASNodeResult `json:"node"`
}

type InitNDIDParam struct {
	NodeID           string `json:"node_id"`
	PublicKey        string `json:"public_key"`
	MasterPublicKey  string `json:"master_public_key"`
	ChainHistoryInfo string `json:"chain_history_info"`
}

type TransferNDIDParam struct {
	PublicKey string `json:"public_key"`
}

type RegisterNode struct {
	NodeID          string   `json:"node_id"`
	PublicKey       string   `json:"public_key"`
	MasterPublicKey string   `json:"master_public_key"`
	NodeName        string   `json:"node_name"`
	Role            string   `json:"role"`
	MaxIal          float64  `json:"max_ial"`            // IdP only attribute
	MaxAal          float64  `json:"max_aal"`            // IdP only attribute
	OnTheFlySupport *bool    `json:"on_the_fly_support"` // IdP only attribute
	IsIdPAgent      *bool    `json:"agent"`              // IdP only attribute
	UseWhitelist    *bool    `json:"node_id_whitelist_active"`
	Whitelist       []string `json:"node_id_whitelist"`
}

type NodeDetail struct {
	PublicKey       string `json:"public_key"`
	MasterPublicKey string `json:"master_public_key"`
	NodeName        string `json:"node_name"`
	Role            string `json:"role"`
	Active          bool   `json:"active"`
}

type MaxIalAal struct {
	MaxIal float64 `json:"max_ial"`
	MaxAal float64 `json:"max_aal"`
}

type SetMqAddressesParam struct {
	Addresses []MsqAddress `json:"addresses"`
}

type GetMqAddressesParam struct {
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

type RequestIDParam struct {
	RequestID string `json:"request_id"`
}

type Namespace struct {
	Namespace                                    string `json:"namespace"`
	Description                                  string `json:"description"`
	Active                                       bool   `json:"active"`
	AllowedIdentifierCountInReferenceGroup       int32  `json:"allowed_identifier_count_in_reference_group"`
	AllowedActiveIdentifierCountInReferenceGroup int32  `json:"allowed_active_identifier_count_in_reference_group"`
}

type DisableNamespaceParam struct {
	Namespace string `json:"namespace"`
}

type UpdateNodeParam struct {
	PublicKey                              string   `json:"public_key"`
	MasterPublicKey                        string   `json:"master_public_key"`
	SupportedRequestMessageDataUrlTypeList []string `json:"supported_request_message_data_url_type_list"`
}

type RegisterAccessorParam struct {
	AccessorID        string `json:"accessor_id"`
	AccessorType      string `json:"accessor_type"`
	AccessorPublicKey string `json:"accessor_public_key"`
	AccessorGroupID   string `json:"accessor_group_id"`
}

type Accessor struct {
	AccessorType      string `json:"accessor_type"`
	AccessorPublicKey string `json:"accessor_public_key"`
	AccessorGroupID   string `json:"accessor_group_id"`
	Active            bool   `json:"active"`
	Owner             string `json:"owner"`
}

type AddAccessorParam struct {
	ReferenceGroupCode     string `json:"reference_group_code"`
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
	AccessorID             string `json:"accessor_id"`
	AccessorPublicKey      string `json:"accessor_public_key"`
	AccessorType           string `json:"accessor_type"`
	RequestID              string `json:"request_id"`
}

type CheckExistingIdentityParam struct {
	ReferenceGroupCode     string `json:"reference_group_code"`
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
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
	Active            bool   `json:"active"`
}

type SetValidatorParam struct {
	PublicKey string `json:"public_key"`
	Power     int64  `json:"power"`
}

type SetDataReceivedParam struct {
	RequestID string `json:"request_id"`
	ServiceID string `json:"service_id"`
	AsID      string `json:"as_id"`
}

type ServiceDetail struct {
	ServiceID         string `json:"service_id"`
	ServiceName       string `json:"service_name"`
	DataSchema        string `json:"data_schema"`
	DataSchemaVersion string `json:"data_schema_version"`
	Active            bool   `json:"active"`
}

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

type ProxyNodeInfo struct {
	NodeID          string       `json:"node_id"`
	NodeName        string       `json:"node_name"`
	PublicKey       string       `json:"public_key"`
	MasterPublicKey string       `json:"master_public_key"`
	Mq              []MsqAddress `json:"mq"`
	Config          string       `json:"config"`
}

type GetNodeInfoResult struct {
	PublicKey       string `json:"public_key"`
	MasterPublicKey string `json:"master_public_key"`
	NodeName        string `json:"node_name"`
	Role            string `json:"role"`
	// for IdP
	MaxIal                                 *float64  `json:"max_ial,omitempty"`
	MaxAal                                 *float64  `json:"max_aal,omitempty"`
	OnTheFlySupport                        *bool     `json:"on_the_fly_support,omitempty"`
	SupportedRequestMessageDataUrlTypeList *[]string `json:"supported_request_message_data_url_type_list,omitempty"`
	IsIdpAgent                             *bool     `json:"agent,omitempty"`
	// for IdP and RP
	UseWhitelist *bool     `json:"node_id_whitelist_active,omitempty"`
	Whitelist    *[]string `json:"node_id_whitelist,omitempty"`
	// for node behind proxy
	Proxy *ProxyNodeInfo `json:"proxy,omitempty"`
	// for all
	Mq     []MsqAddress `json:"mq"`
	Active bool         `json:"active"`
}

type GetIdentityInfoParam struct {
	ReferenceGroupCode     string `json:"reference_group_code"`
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
	NodeID                 string `json:"node_id"`
}

type GetIdentityInfoResult struct {
	Ial      float64 `json:"ial"`
	Lial     *bool   `json:"lial"`
	Laal     *bool   `json:"laal"`
	ModeList []int32 `json:"mode_list"`
}

type UpdateNodeByNDIDParam struct {
	NodeID          string   `json:"node_id"`
	MaxIal          float64  `json:"max_ial"`
	MaxAal          float64  `json:"max_aal"`
	OnTheFlySupport *bool    `json:"on_the_fly_support"`
	NodeName        string   `json:"node_name"`
	IsIdPAgent      *bool    `json:"agent"`
	UseWhitelist    *bool    `json:"node_id_whitelist_active"`
	Whitelist       []string `json:"node_id_whitelist"`
}

type UpdateIdentityParam struct {
	ReferenceGroupCode     string   `json:"reference_group_code"`
	IdentityNamespace      string   `json:"identity_namespace"`
	IdentityIdentifierHash string   `json:"identity_identifier_hash"`
	Ial                    *float64 `json:"ial"`
	Lial                   *bool    `json:"lial"`
	Laal                   *bool    `json:"laal"`
}

type CloseRequestParam struct {
	RequestID         string          `json:"request_id"`
	ResponseValidList []ResponseValid `json:"response_valid_list"`
}

type TimeOutRequestParam struct {
	RequestID         string          `json:"request_id"`
	ResponseValidList []ResponseValid `json:"response_valid_list"`
}

type ResponseValid struct {
	IdpID          string `json:"idp_id"`
	ValidIal       *bool  `json:"valid_ial"`
	ValidSignature *bool  `json:"valid_signature"`
}

type GetDataSignatureParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
	RequestID string `json:"request_id"`
}

type GetDataSignatureResult struct {
	Signature string `json:"signature"`
}

type UpdateServiceDestinationParam struct {
	ServiceID              string   `json:"service_id"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

type UpdateServiceParam struct {
	ServiceID         string `json:"service_id"`
	ServiceName       string `json:"service_name"`
	DataSchema        string `json:"data_schema"`
	DataSchemaVersion string `json:"data_schema_version"`
}

type DisableMsqDestinationParam struct {
	HashID string `json:"hash_id"`
}

type DisableAccessorMethodParam struct {
	AccessorID string `json:"accessor_id"`
}

type RegisterServiceDestinationByNDIDParam struct {
	ServiceID string `json:"service_id"`
	NodeID    string `json:"node_id"`
}

type DisableNodeParam struct {
	NodeID string `json:"node_id"`
}

type Service struct {
	ServiceID              string   `json:"service_id"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	Active                 bool     `json:"active"`
	Suspended              bool     `json:"suspended"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
}

type GetServicesByAsIDParam struct {
	AsID string `json:"as_id"`
}

type GetServicesByAsIDResult struct {
	Services []Service `json:"services"`
}

type DisableServiceDestinationByNDIDParam struct {
	ServiceID string `json:"service_id"`
	NodeID    string `json:"node_id"`
}

type ApproveService struct {
	Active bool `json:"active"`
}

type DisableServiceDestinationParam struct {
	ServiceID string `json:"service_id"`
}

type ClearRegisterIdentityTimeoutParam struct {
	HashID string `json:"hash_id"`
}

type TimeOutBlockRegisterIdentity struct {
	TimeOutBlock int64 `json:"time_out_block"`
}

type GetIdpNodesInfoResult struct {
	Node []IdpNode `json:"node"`
}

type IdpNodeProxy struct {
	NodeID    string       `json:"node_id"`
	PublicKey string       `json:"public_key"`
	Mq        []MsqAddress `json:"mq"`
	Config    string       `json:"config"`
}

type IdpNode struct {
	NodeID                                 string        `json:"node_id"`
	Name                                   string        `json:"name"`
	MaxIal                                 float64       `json:"max_ial"`
	MaxAal                                 float64       `json:"max_aal"`
	OnTheFlySupport                        bool          `json:"on_the_fly_support"`
	PublicKey                              string        `json:"public_key"`
	Mq                                     []MsqAddress  `json:"mq"`
	IsIdpAgent                             bool          `json:"agent"`
	UseWhitelist                           *bool         `json:"node_id_whitelist_active,omitempty"`
	Whitelist                              *[]string     `json:"node_id_whitelist,omitempty"`
	Ial                                    *float64      `json:"ial,omitempty"`
	ModeList                               *[]int32      `json:"mode_list,omitempty"`
	SupportedRequestMessageDataUrlTypeList []string      `json:"supported_request_message_data_url_type_list"`
	Proxy                                  *IdpNodeProxy `json:"proxy,omitempty"`
}

type ASWithMqNode struct {
	ID                     string       `json:"node_id"`
	Name                   string       `json:"name"`
	MinIal                 float64      `json:"min_ial"`
	MinAal                 float64      `json:"min_aal"`
	PublicKey              string       `json:"public_key"`
	Mq                     []MsqAddress `json:"mq"`
	SupportedNamespaceList []string     `json:"supported_namespace_list"`
}

type GetAsNodesInfoByServiceIdResult struct {
	Node []interface{} `json:"node"`
}

type AddNodeToProxyNodeParam struct {
	NodeID      string `json:"node_id"`
	ProxyNodeID string `json:"proxy_node_id"`
	Config      string `json:"config"`
}

type UpdateNodeProxyNodeParam struct {
	NodeID      string `json:"node_id"`
	ProxyNodeID string `json:"proxy_node_id"`
	Config      string `json:"config"`
}

type RemoveNodeFromProxyNode struct {
	NodeID string `json:"node_id"`
}

type ASWithMqNodeBehindProxy struct {
	NodeID                 string   `json:"node_id"`
	Name                   string   `json:"name"`
	MinIal                 float64  `json:"min_ial"`
	MinAal                 float64  `json:"min_aal"`
	PublicKey              string   `json:"public_key"`
	SupportedNamespaceList []string `json:"supported_namespace_list"`
	Proxy                  struct {
		NodeID    string       `json:"node_id"`
		PublicKey string       `json:"public_key"`
		Mq        []MsqAddress `json:"mq"`
		Config    string       `json:"config"`
	} `json:"proxy"`
}

type GetNodesBehindProxyNodeParam struct {
	ProxyNodeID string `json:"proxy_node_id"`
}

type GetNodesBehindProxyNodeResult struct {
	Nodes []interface{} `json:"nodes"`
}

type IdPBehindProxy struct {
	NodeID                                 string   `json:"node_id"`
	NodeName                               string   `json:"node_name"`
	Role                                   string   `json:"role"`
	PublicKey                              string   `json:"public_key"`
	MasterPublicKey                        string   `json:"master_public_key"`
	MaxIal                                 float64  `json:"max_ial"`
	MaxAal                                 float64  `json:"max_aal"`
	OnTheFlySupport                        bool     `json:"on_the_fly_support"`
	IsIdpAgent                             bool     `json:"agent"`
	Config                                 string   `json:"config"`
	SupportedRequestMessageDataUrlTypeList []string `json:"supported_request_message_data_url_type_list"`
}

type ASorRPBehindProxy struct {
	NodeID          string `json:"node_id"`
	NodeName        string `json:"node_name"`
	Role            string `json:"role"`
	PublicKey       string `json:"public_key"`
	MasterPublicKey string `json:"master_public_key"`
	Config          string `json:"config"`
}

type Proxy struct {
	ProxyNodeID string `json:"proxy_node_id"`
	Config      string `json:"config"`
}

type GetNodeIDListParam struct {
	Role string `json:"role"`
}

type GetNodeIDListResult struct {
	NodeIDList []string `json:"node_id_list"`
}

type GetMqAddressesResult []MsqAddress

type GetAccessorsInAccessorGroupParam struct {
	AccessorGroupID string `json:"accessor_group_id"`
	IdpID           string `json:"idp_id"`
}

type GetAccessorsInAccessorGroupResult struct {
	AccessorList []string `json:"accessor_list"`
}

type RevokeAccessorParam struct {
	AccessorIDList []string `json:"accessor_id_list"`
	RequestID      string   `json:"request_id"`
}

type GetAccessorOwnerParam struct {
	AccessorID string `json:"accessor_id"`
}

type GetAccessorOwnerResult struct {
	NodeID string `json:"node_id"`
}

type KeyValue struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type SetInitDataParam struct {
	KVList []KeyValue `json:"kv_list"`
}

type EndInitParam struct{}

type SetLastBlockParam struct {
	BlockHeight int64 `json:"block_height"`
}

type IsInitEndedParam struct{}

type IsInitEndedResult struct {
	InitEnded bool `json:"init_ended"`
}

type GetReferenceGroupCodeParam struct {
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
}

type GetReferenceGroupCodeResult struct {
	ReferenceGroupCode string `json:"reference_group_code"`
}

type GetReferenceGroupCodeByAccessorIDParam struct {
	AccessorID string `json:"accessor_id"`
}

type RevokeIdentityAssociationParam struct {
	ReferenceGroupCode     string `json:"reference_group_code"`
	IdentityNamespace      string `json:"identity_namespace"`
	IdentityIdentifierHash string `json:"identity_identifier_hash"`
	RequestID              string `json:"request_id"`
}

type GetAllowedModeListResult struct {
	AllowedModeList []int32 `json:"allowed_mode_list"`
}

type UpdateIdentityModeListParam struct {
	ReferenceGroupCode     string  `json:"reference_group_code"`
	IdentityNamespace      string  `json:"identity_namespace"`
	IdentityIdentifierHash string  `json:"identity_identifier_hash"`
	ModeList               []int32 `json:"mode_list"`
	RequestID              string  `json:"request_id"`
}

type SetAllowedModeListParam struct {
	Purpose         string  `json:"purpose"`
	AllowedModeList []int32 `json:"allowed_mode_list"`
}

type GetAllowedModeListParam struct {
	Purpose string `json:"purpose"`
}

type SetAllowedMinIalForRegisterIdentityAtFirstIdpParam struct {
	MinIal float64 `json:"min_ial"`
}

type GetAllowedMinIalForRegisterIdentityAtFirstIdpResult struct {
	MinIal float64 `json:"min_ial"`
}

type UpdateNamespaceParam struct {
	Namespace                                    string `json:"namespace"`
	Description                                  string `json:"description"`
	AllowedIdentifierCountInReferenceGroup       int32  `json:"allowed_identifier_count_in_reference_group"`
	AllowedActiveIdentifierCountInReferenceGroup int32  `json:"allowed_active_identifier_count_in_reference_group"`
}

type RevokeAndAddAccessorParam struct {
	RevokingAccessorID string `json:"revoking_accessor_id"`
	AccessorID         string `json:"accessor_id"`
	AccessorPublicKey  string `json:"accessor_public_key"`
	AccessorType       string `json:"accessor_type"`
	RequestID          string `json:"request_id"`
}

type AddErrorCodeParam struct {
	ErrorCode   int32  `json:"error_code"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type RemoveErrorCodeParam struct {
	ErrorCode int32  `json:"error_code"`
	Type      string `json:"type"`
}

type GetErrorCodeListParam struct {
	Type string `json:"type"`
}

type GetErrorCodeListResult struct {
	ErrorCode   int32  `json:"error_code"`
	Description string `json:"description"`
}

type SetServicePriceCeilingParam struct {
	ServiceID                  string                   `json:"service_id"`
	PriceCeilingByCurrencyList []PriceCeilingByCurrency `json:"price_ceiling_by_currency_list"`
}

type PriceCeilingByCurrency struct {
	Currency string  `json:"currency"`
	Price    float64 `json:"price"`
}

type GetServicePriceCeilingParam struct {
	ServiceID string `json:"service_id"`
}

type GetServicePriceCeilingResult struct {
	PriceCeilingByCurrencyList []PriceCeilingByCurrency `json:"price_ceiling_by_currency_list"`
}

type SetServicePriceMinEffectiveDatetimeDelayParam struct {
	ServiceID      string `json:"service_id"`
	DurationSecond uint32 `json:"duration_second"`
}

type GetServicePriceMinEffectiveDatetimeDelayParam struct {
	ServiceID string `json:"service_id"`
}

type GetServicePriceMinEffectiveDatetimeDelayResult struct {
	DurationSecond uint32 `json:"duration_second"`
}

type SetServicePriceParam struct {
	ServiceID           string                   `json:"service_id"`
	PriceByCurrencyList []ServicePriceByCurrency `json:"price_by_currency_list"`
	EffectiveDatetime   time.Time                `json:"effective_datetime"`
	MoreInfoURL         string                   `json:"more_info_url"`
	Detail              string                   `json:"detail"`
}

type ServicePriceByCurrency struct {
	Currency string  `json:"currency"`
	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
}

type GetServicePriceListParam struct {
	NodeID    string `json:"node_id"`
	ServiceID string `json:"service_id"`
}

type GetServicePriceListResult struct {
	ServicePriceListByNode []ServicePriceListByNode `json:"price_list_by_node"`
}

type ServicePriceListByNode struct {
	NodeID           string         `json:"node_id"`
	ServicePriceList []ServicePrice `json:"price_list"`
}

type ServicePrice struct {
	PriceByCurrencyList []ServicePriceByCurrency `json:"price_by_currency_list"`
	EffectiveDatetime   time.Time                `json:"effective_datetime"`
	MoreInfoURL         string                   `json:"more_info_url"`
	Detail              string                   `json:"detail"`
	CreationBlockHeight int64                    `json:"creation_block_height"`
	CreationChainID     string                   `json:"creation_chain_id"`
}
