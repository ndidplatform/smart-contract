// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protos/data/data.proto

package data

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type NodeDetail struct {
	PublicKey            string   `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	MasterPublicKey      string   `protobuf:"bytes,2,opt,name=master_public_key,json=masterPublicKey,proto3" json:"master_public_key,omitempty"`
	NodeName             string   `protobuf:"bytes,3,opt,name=node_name,json=nodeName,proto3" json:"node_name,omitempty"`
	Role                 string   `protobuf:"bytes,4,opt,name=role,proto3" json:"role,omitempty"`
	MaxIal               float64  `protobuf:"fixed64,5,opt,name=max_ial,json=maxIal,proto3" json:"max_ial,omitempty"`
	MaxAal               float64  `protobuf:"fixed64,6,opt,name=max_aal,json=maxAal,proto3" json:"max_aal,omitempty"`
	Mq                   *MQ      `protobuf:"bytes,7,opt,name=mq,proto3" json:"mq,omitempty"`
	Active               bool     `protobuf:"varint,8,opt,name=active,proto3" json:"active,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeDetail) Reset()         { *m = NodeDetail{} }
func (m *NodeDetail) String() string { return proto.CompactTextString(m) }
func (*NodeDetail) ProtoMessage()    {}
func (*NodeDetail) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_d70173269bd0f6a7, []int{0}
}
func (m *NodeDetail) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeDetail.Unmarshal(m, b)
}
func (m *NodeDetail) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeDetail.Marshal(b, m, deterministic)
}
func (dst *NodeDetail) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeDetail.Merge(dst, src)
}
func (m *NodeDetail) XXX_Size() int {
	return xxx_messageInfo_NodeDetail.Size(m)
}
func (m *NodeDetail) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeDetail.DiscardUnknown(m)
}

var xxx_messageInfo_NodeDetail proto.InternalMessageInfo

func (m *NodeDetail) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func (m *NodeDetail) GetMasterPublicKey() string {
	if m != nil {
		return m.MasterPublicKey
	}
	return ""
}

func (m *NodeDetail) GetNodeName() string {
	if m != nil {
		return m.NodeName
	}
	return ""
}

func (m *NodeDetail) GetRole() string {
	if m != nil {
		return m.Role
	}
	return ""
}

func (m *NodeDetail) GetMaxIal() float64 {
	if m != nil {
		return m.MaxIal
	}
	return 0
}

func (m *NodeDetail) GetMaxAal() float64 {
	if m != nil {
		return m.MaxAal
	}
	return 0
}

func (m *NodeDetail) GetMq() *MQ {
	if m != nil {
		return m.Mq
	}
	return nil
}

func (m *NodeDetail) GetActive() bool {
	if m != nil {
		return m.Active
	}
	return false
}

type MQ struct {
	Ip                   string   `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Port                 int64    `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MQ) Reset()         { *m = MQ{} }
func (m *MQ) String() string { return proto.CompactTextString(m) }
func (*MQ) ProtoMessage()    {}
func (*MQ) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_d70173269bd0f6a7, []int{1}
}
func (m *MQ) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MQ.Unmarshal(m, b)
}
func (m *MQ) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MQ.Marshal(b, m, deterministic)
}
func (dst *MQ) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MQ.Merge(dst, src)
}
func (m *MQ) XXX_Size() int {
	return xxx_messageInfo_MQ.Size(m)
}
func (m *MQ) XXX_DiscardUnknown() {
	xxx_messageInfo_MQ.DiscardUnknown(m)
}

var xxx_messageInfo_MQ proto.InternalMessageInfo

func (m *MQ) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

func (m *MQ) GetPort() int64 {
	if m != nil {
		return m.Port
	}
	return 0
}

type IdPList struct {
	NodeId               []string `protobuf:"bytes,1,rep,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IdPList) Reset()         { *m = IdPList{} }
func (m *IdPList) String() string { return proto.CompactTextString(m) }
func (*IdPList) ProtoMessage()    {}
func (*IdPList) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_d70173269bd0f6a7, []int{2}
}
func (m *IdPList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IdPList.Unmarshal(m, b)
}
func (m *IdPList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IdPList.Marshal(b, m, deterministic)
}
func (dst *IdPList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IdPList.Merge(dst, src)
}
func (m *IdPList) XXX_Size() int {
	return xxx_messageInfo_IdPList.Size(m)
}
func (m *IdPList) XXX_DiscardUnknown() {
	xxx_messageInfo_IdPList.DiscardUnknown(m)
}

var xxx_messageInfo_IdPList proto.InternalMessageInfo

func (m *IdPList) GetNodeId() []string {
	if m != nil {
		return m.NodeId
	}
	return nil
}

type NamespaceList struct {
	Namespaces           []*Namespace `protobuf:"bytes,1,rep,name=namespaces,proto3" json:"namespaces,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *NamespaceList) Reset()         { *m = NamespaceList{} }
func (m *NamespaceList) String() string { return proto.CompactTextString(m) }
func (*NamespaceList) ProtoMessage()    {}
func (*NamespaceList) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_d70173269bd0f6a7, []int{3}
}
func (m *NamespaceList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NamespaceList.Unmarshal(m, b)
}
func (m *NamespaceList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NamespaceList.Marshal(b, m, deterministic)
}
func (dst *NamespaceList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NamespaceList.Merge(dst, src)
}
func (m *NamespaceList) XXX_Size() int {
	return xxx_messageInfo_NamespaceList.Size(m)
}
func (m *NamespaceList) XXX_DiscardUnknown() {
	xxx_messageInfo_NamespaceList.DiscardUnknown(m)
}

var xxx_messageInfo_NamespaceList proto.InternalMessageInfo

func (m *NamespaceList) GetNamespaces() []*Namespace {
	if m != nil {
		return m.Namespaces
	}
	return nil
}

type Namespace struct {
	Namespace            string   `protobuf:"bytes,1,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Active               bool     `protobuf:"varint,3,opt,name=active,proto3" json:"active,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Namespace) Reset()         { *m = Namespace{} }
func (m *Namespace) String() string { return proto.CompactTextString(m) }
func (*Namespace) ProtoMessage()    {}
func (*Namespace) Descriptor() ([]byte, []int) {
	return fileDescriptor_data_d70173269bd0f6a7, []int{4}
}
func (m *Namespace) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Namespace.Unmarshal(m, b)
}
func (m *Namespace) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Namespace.Marshal(b, m, deterministic)
}
func (dst *Namespace) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Namespace.Merge(dst, src)
}
func (m *Namespace) XXX_Size() int {
	return xxx_messageInfo_Namespace.Size(m)
}
func (m *Namespace) XXX_DiscardUnknown() {
	xxx_messageInfo_Namespace.DiscardUnknown(m)
}

var xxx_messageInfo_Namespace proto.InternalMessageInfo

func (m *Namespace) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *Namespace) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Namespace) GetActive() bool {
	if m != nil {
		return m.Active
	}
	return false
}

func init() {
	proto.RegisterType((*NodeDetail)(nil), "NodeDetail")
	proto.RegisterType((*MQ)(nil), "MQ")
	proto.RegisterType((*IdPList)(nil), "IdPList")
	proto.RegisterType((*NamespaceList)(nil), "NamespaceList")
	proto.RegisterType((*Namespace)(nil), "Namespace")
}

func init() { proto.RegisterFile("protos/data/data.proto", fileDescriptor_data_d70173269bd0f6a7) }

var fileDescriptor_data_d70173269bd0f6a7 = []byte{
	// 325 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x91, 0xc1, 0x6a, 0xe3, 0x30,
	0x10, 0x86, 0x91, 0x9d, 0x75, 0xe2, 0x09, 0xbb, 0x4b, 0x55, 0x48, 0x05, 0x6d, 0xc1, 0xf8, 0x64,
	0x72, 0x48, 0x21, 0x3d, 0xf6, 0x54, 0xe8, 0x25, 0xb4, 0x09, 0x89, 0x5e, 0xc0, 0x4c, 0x2c, 0x1d,
	0x44, 0x25, 0x4b, 0xb1, 0xd5, 0x92, 0x3c, 0x71, 0x5f, 0xa3, 0x58, 0x49, 0x8c, 0x2f, 0x66, 0xe6,
	0xfb, 0x3f, 0xb0, 0xff, 0x31, 0xcc, 0x5c, 0x63, 0xbd, 0x6d, 0x9f, 0x04, 0x7a, 0x0c, 0x8f, 0x45,
	0x00, 0xf9, 0x0f, 0x01, 0xd8, 0x58, 0x21, 0xdf, 0xa4, 0x47, 0xa5, 0xe9, 0x23, 0x80, 0xfb, 0xda,
	0x6b, 0x55, 0x95, 0x9f, 0xf2, 0xc4, 0x48, 0x46, 0x8a, 0x94, 0xa7, 0x67, 0xf2, 0x2e, 0x4f, 0x74,
	0x0e, 0x37, 0x06, 0x5b, 0x2f, 0x9b, 0x72, 0x60, 0x45, 0xc1, 0xfa, 0x7f, 0x0e, 0xb6, 0xbd, 0x7b,
	0x0f, 0x69, 0x6d, 0x85, 0x2c, 0x6b, 0x34, 0x92, 0xc5, 0xc1, 0x99, 0x74, 0x60, 0x83, 0x46, 0x52,
	0x0a, 0xa3, 0xc6, 0x6a, 0xc9, 0x46, 0x81, 0x87, 0x99, 0xde, 0xc1, 0xd8, 0xe0, 0xb1, 0x54, 0xa8,
	0xd9, 0x9f, 0x8c, 0x14, 0x84, 0x27, 0x06, 0x8f, 0x2b, 0xd4, 0xd7, 0x00, 0x51, 0xb3, 0xa4, 0x0f,
	0x5e, 0x51, 0xd3, 0x5b, 0x88, 0xcc, 0x81, 0x8d, 0x33, 0x52, 0x4c, 0x97, 0xf1, 0x62, 0xbd, 0xe3,
	0x91, 0x39, 0xd0, 0x19, 0x24, 0x58, 0x79, 0xf5, 0x2d, 0xd9, 0x24, 0x23, 0xc5, 0x84, 0x5f, 0xb6,
	0xbc, 0x80, 0x68, 0xbd, 0xa3, 0xff, 0x20, 0x52, 0xee, 0x52, 0x2c, 0x52, 0xae, 0xfb, 0x10, 0x67,
	0x1b, 0x1f, 0x4a, 0xc4, 0x3c, 0xcc, 0x79, 0x0e, 0xe3, 0x95, 0xd8, 0x7e, 0xa8, 0xd6, 0x77, 0xaf,
	0x0e, 0x25, 0x94, 0x60, 0x24, 0x8b, 0x8b, 0x94, 0x27, 0xdd, 0xba, 0x12, 0xf9, 0x0b, 0xfc, 0xed,
	0x8a, 0xb4, 0x0e, 0x2b, 0x19, 0xcc, 0x39, 0x40, 0x7d, 0x05, 0x6d, 0x90, 0xa7, 0x4b, 0x58, 0xf4,
	0x0e, 0x1f, 0xa4, 0x79, 0x05, 0x69, 0x1f, 0xd0, 0x07, 0x48, 0xfb, 0xe8, 0x7a, 0xf1, 0x1e, 0xd0,
	0x0c, 0xa6, 0x42, 0xb6, 0x55, 0xa3, 0x9c, 0x57, 0xb6, 0xbe, 0xdc, 0x7a, 0x88, 0x06, 0x7d, 0xe3,
	0x61, 0xdf, 0x7d, 0x12, 0x7e, 0xf0, 0xf3, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x98, 0xd2, 0x02,
	0x3f, 0xfa, 0x01, 0x00, 0x00,
}