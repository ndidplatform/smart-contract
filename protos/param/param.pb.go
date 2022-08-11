// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.8
// source: param.proto

package ndid_abci_param_v8

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type KeyValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *KeyValue) Reset() {
	*x = KeyValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_param_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyValue) ProtoMessage() {}

func (x *KeyValue) ProtoReflect() protoreflect.Message {
	mi := &file_param_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyValue.ProtoReflect.Descriptor instead.
func (*KeyValue) Descriptor() ([]byte, []int) {
	return file_param_proto_rawDescGZIP(), []int{0}
}

func (x *KeyValue) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *KeyValue) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

type SetInitDataParam struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KvList []*KeyValue `protobuf:"bytes,1,rep,name=kv_list,json=kvList,proto3" json:"kv_list,omitempty"`
}

func (x *SetInitDataParam) Reset() {
	*x = SetInitDataParam{}
	if protoimpl.UnsafeEnabled {
		mi := &file_param_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetInitDataParam) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetInitDataParam) ProtoMessage() {}

func (x *SetInitDataParam) ProtoReflect() protoreflect.Message {
	mi := &file_param_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetInitDataParam.ProtoReflect.Descriptor instead.
func (*SetInitDataParam) Descriptor() ([]byte, []int) {
	return file_param_proto_rawDescGZIP(), []int{1}
}

func (x *SetInitDataParam) GetKvList() []*KeyValue {
	if x != nil {
		return x.KvList
	}
	return nil
}

var File_param_proto protoreflect.FileDescriptor

var file_param_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6e,
	0x64, 0x69, 0x64, 0x5f, 0x61, 0x62, 0x63, 0x69, 0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x5f, 0x76,
	0x38, 0x22, 0x32, 0x0a, 0x08, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x49, 0x0a, 0x10, 0x53, 0x65, 0x74, 0x49, 0x6e, 0x69, 0x74,
	0x44, 0x61, 0x74, 0x61, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x12, 0x35, 0x0a, 0x07, 0x6b, 0x76, 0x5f,
	0x6c, 0x69, 0x73, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6e, 0x64, 0x69,
	0x64, 0x5f, 0x61, 0x62, 0x63, 0x69, 0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x5f, 0x76, 0x38, 0x2e,
	0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x06, 0x6b, 0x76, 0x4c, 0x69, 0x73, 0x74,
	0x42, 0x17, 0x5a, 0x15, 0x2e, 0x2f, 0x3b, 0x6e, 0x64, 0x69, 0x64, 0x5f, 0x61, 0x62, 0x63, 0x69,
	0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x5f, 0x76, 0x38, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_param_proto_rawDescOnce sync.Once
	file_param_proto_rawDescData = file_param_proto_rawDesc
)

func file_param_proto_rawDescGZIP() []byte {
	file_param_proto_rawDescOnce.Do(func() {
		file_param_proto_rawDescData = protoimpl.X.CompressGZIP(file_param_proto_rawDescData)
	})
	return file_param_proto_rawDescData
}

var file_param_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_param_proto_goTypes = []interface{}{
	(*KeyValue)(nil),         // 0: ndid_abci_param_v8.KeyValue
	(*SetInitDataParam)(nil), // 1: ndid_abci_param_v8.SetInitDataParam
}
var file_param_proto_depIdxs = []int32{
	0, // 0: ndid_abci_param_v8.SetInitDataParam.kv_list:type_name -> ndid_abci_param_v8.KeyValue
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_param_proto_init() }
func file_param_proto_init() {
	if File_param_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_param_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeyValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_param_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetInitDataParam); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_param_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_param_proto_goTypes,
		DependencyIndexes: file_param_proto_depIdxs,
		MessageInfos:      file_param_proto_msgTypes,
	}.Build()
	File_param_proto = out.File
	file_param_proto_rawDesc = nil
	file_param_proto_goTypes = nil
	file_param_proto_depIdxs = nil
}
