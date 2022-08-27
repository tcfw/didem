// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.4
// source: consensus.proto

package api

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

type ConsensusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Request:
	//	*ConsensusRequest_Tip
	Request isConsensusRequest_Request `protobuf_oneof:"Request"`
}

func (x *ConsensusRequest) Reset() {
	*x = ConsensusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_consensus_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConsensusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConsensusRequest) ProtoMessage() {}

func (x *ConsensusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_consensus_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConsensusRequest.ProtoReflect.Descriptor instead.
func (*ConsensusRequest) Descriptor() ([]byte, []int) {
	return file_consensus_proto_rawDescGZIP(), []int{0}
}

func (m *ConsensusRequest) GetRequest() isConsensusRequest_Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (x *ConsensusRequest) GetTip() *TipRequest {
	if x, ok := x.GetRequest().(*ConsensusRequest_Tip); ok {
		return x.Tip
	}
	return nil
}

type isConsensusRequest_Request interface {
	isConsensusRequest_Request()
}

type ConsensusRequest_Tip struct {
	Tip *TipRequest `protobuf:"bytes,1,opt,name=Tip,proto3,oneof"`
}

func (*ConsensusRequest_Tip) isConsensusRequest_Request() {}

type TipRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ChainId string `protobuf:"bytes,1,opt,name=chainId,proto3" json:"chainId,omitempty"`
}

func (x *TipRequest) Reset() {
	*x = TipRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_consensus_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TipRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TipRequest) ProtoMessage() {}

func (x *TipRequest) ProtoReflect() protoreflect.Message {
	mi := &file_consensus_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TipRequest.ProtoReflect.Descriptor instead.
func (*TipRequest) Descriptor() ([]byte, []int) {
	return file_consensus_proto_rawDescGZIP(), []int{1}
}

func (x *TipRequest) GetChainId() string {
	if x != nil {
		return x.ChainId
	}
	return ""
}

type TipResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tip     string `protobuf:"bytes,1,opt,name=tip,proto3" json:"tip,omitempty"`
	ChainId string `protobuf:"bytes,2,opt,name=chainId,proto3" json:"chainId,omitempty"`
}

func (x *TipResponse) Reset() {
	*x = TipResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_consensus_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TipResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TipResponse) ProtoMessage() {}

func (x *TipResponse) ProtoReflect() protoreflect.Message {
	mi := &file_consensus_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TipResponse.ProtoReflect.Descriptor instead.
func (*TipResponse) Descriptor() ([]byte, []int) {
	return file_consensus_proto_rawDescGZIP(), []int{2}
}

func (x *TipResponse) GetTip() string {
	if x != nil {
		return x.Tip
	}
	return ""
}

func (x *TipResponse) GetChainId() string {
	if x != nil {
		return x.ChainId
	}
	return ""
}

var File_consensus_proto protoreflect.FileDescriptor

var file_consensus_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x03, 0x70, 0x32, 0x70, 0x22, 0x42, 0x0a, 0x10, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e,
	0x73, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x03, 0x54, 0x69,
	0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x32, 0x70, 0x2e, 0x54, 0x69,
	0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x03, 0x54, 0x69, 0x70, 0x42,
	0x09, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x26, 0x0a, 0x0a, 0x54, 0x69,
	0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x49, 0x64, 0x22, 0x39, 0x0a, 0x0b, 0x54, 0x69, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x74, 0x69, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x42, 0x1c, 0x5a,
	0x1a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x63, 0x66, 0x77,
	0x2f, 0x64, 0x69, 0x64, 0x65, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_consensus_proto_rawDescOnce sync.Once
	file_consensus_proto_rawDescData = file_consensus_proto_rawDesc
)

func file_consensus_proto_rawDescGZIP() []byte {
	file_consensus_proto_rawDescOnce.Do(func() {
		file_consensus_proto_rawDescData = protoimpl.X.CompressGZIP(file_consensus_proto_rawDescData)
	})
	return file_consensus_proto_rawDescData
}

var file_consensus_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_consensus_proto_goTypes = []interface{}{
	(*ConsensusRequest)(nil), // 0: p2p.ConsensusRequest
	(*TipRequest)(nil),       // 1: p2p.TipRequest
	(*TipResponse)(nil),      // 2: p2p.TipResponse
}
var file_consensus_proto_depIdxs = []int32{
	1, // 0: p2p.ConsensusRequest.Tip:type_name -> p2p.TipRequest
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_consensus_proto_init() }
func file_consensus_proto_init() {
	if File_consensus_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_consensus_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConsensusRequest); i {
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
		file_consensus_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TipRequest); i {
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
		file_consensus_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TipResponse); i {
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
	file_consensus_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*ConsensusRequest_Tip)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_consensus_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_consensus_proto_goTypes,
		DependencyIndexes: file_consensus_proto_depIdxs,
		MessageInfos:      file_consensus_proto_msgTypes,
	}.Build()
	File_consensus_proto = out.File
	file_consensus_proto_rawDesc = nil
	file_consensus_proto_goTypes = nil
	file_consensus_proto_depIdxs = nil
}
