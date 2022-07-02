// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.1
// source: em.proto

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

type EmSendRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From    string            `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To      []string          `protobuf:"bytes,2,rep,name=to,proto3" json:"to,omitempty"`
	Headers map[string]string `protobuf:"bytes,3,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Parts   []*EmSendPart     `protobuf:"bytes,4,rep,name=parts,proto3" json:"parts,omitempty"`
}

func (x *EmSendRequest) Reset() {
	*x = EmSendRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_em_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmSendRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmSendRequest) ProtoMessage() {}

func (x *EmSendRequest) ProtoReflect() protoreflect.Message {
	mi := &file_em_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmSendRequest.ProtoReflect.Descriptor instead.
func (*EmSendRequest) Descriptor() ([]byte, []int) {
	return file_em_proto_rawDescGZIP(), []int{0}
}

func (x *EmSendRequest) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *EmSendRequest) GetTo() []string {
	if x != nil {
		return x.To
	}
	return nil
}

func (x *EmSendRequest) GetHeaders() map[string]string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *EmSendRequest) GetParts() []*EmSendPart {
	if x != nil {
		return x.Parts
	}
	return nil
}

type EmSendPart struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mime string `protobuf:"bytes,1,opt,name=mime,proto3" json:"mime,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *EmSendPart) Reset() {
	*x = EmSendPart{}
	if protoimpl.UnsafeEnabled {
		mi := &file_em_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmSendPart) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmSendPart) ProtoMessage() {}

func (x *EmSendPart) ProtoReflect() protoreflect.Message {
	mi := &file_em_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmSendPart.ProtoReflect.Descriptor instead.
func (*EmSendPart) Descriptor() ([]byte, []int) {
	return file_em_proto_rawDescGZIP(), []int{1}
}

func (x *EmSendPart) GetMime() string {
	if x != nil {
		return x.Mime
	}
	return ""
}

func (x *EmSendPart) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type EmSendResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Errors []string `protobuf:"bytes,1,rep,name=errors,proto3" json:"errors,omitempty"`
}

func (x *EmSendResponse) Reset() {
	*x = EmSendResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_em_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EmSendResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmSendResponse) ProtoMessage() {}

func (x *EmSendResponse) ProtoReflect() protoreflect.Message {
	mi := &file_em_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmSendResponse.ProtoReflect.Descriptor instead.
func (*EmSendResponse) Descriptor() ([]byte, []int) {
	return file_em_proto_rawDescGZIP(), []int{2}
}

func (x *EmSendResponse) GetErrors() []string {
	if x != nil {
		return x.Errors
	}
	return nil
}

var File_em_proto protoreflect.FileDescriptor

var file_em_proto_rawDesc = []byte{
	0x0a, 0x08, 0x65, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x65, 0x6d, 0x22, 0xcf,
	0x01, 0x0a, 0x0d, 0x45, 0x6d, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x02, 0x74, 0x6f, 0x12, 0x38, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x65, 0x6d, 0x2e, 0x45, 0x6d, 0x53, 0x65, 0x6e,
	0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x24,
	0x0a, 0x05, 0x70, 0x61, 0x72, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e,
	0x65, 0x6d, 0x2e, 0x45, 0x6d, 0x53, 0x65, 0x6e, 0x64, 0x50, 0x61, 0x72, 0x74, 0x52, 0x05, 0x70,
	0x61, 0x72, 0x74, 0x73, 0x1a, 0x3a, 0x0a, 0x0c, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x22, 0x34, 0x0a, 0x0a, 0x45, 0x6d, 0x53, 0x65, 0x6e, 0x64, 0x50, 0x61, 0x72, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x6d, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6d, 0x69,
	0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x28, 0x0a, 0x0e, 0x45, 0x6d, 0x53, 0x65, 0x6e, 0x64,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73,
	0x32, 0x3a, 0x0a, 0x09, 0x45, 0x6d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x2d, 0x0a,
	0x04, 0x53, 0x65, 0x6e, 0x64, 0x12, 0x11, 0x2e, 0x65, 0x6d, 0x2e, 0x45, 0x6d, 0x53, 0x65, 0x6e,
	0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x65, 0x6d, 0x2e, 0x45, 0x6d,
	0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x1c, 0x5a, 0x1a,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x63, 0x66, 0x77, 0x2f,
	0x64, 0x69, 0x64, 0x65, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_em_proto_rawDescOnce sync.Once
	file_em_proto_rawDescData = file_em_proto_rawDesc
)

func file_em_proto_rawDescGZIP() []byte {
	file_em_proto_rawDescOnce.Do(func() {
		file_em_proto_rawDescData = protoimpl.X.CompressGZIP(file_em_proto_rawDescData)
	})
	return file_em_proto_rawDescData
}

var file_em_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_em_proto_goTypes = []interface{}{
	(*EmSendRequest)(nil),  // 0: em.EmSendRequest
	(*EmSendPart)(nil),     // 1: em.EmSendPart
	(*EmSendResponse)(nil), // 2: em.EmSendResponse
	nil,                    // 3: em.EmSendRequest.HeadersEntry
}
var file_em_proto_depIdxs = []int32{
	3, // 0: em.EmSendRequest.headers:type_name -> em.EmSendRequest.HeadersEntry
	1, // 1: em.EmSendRequest.parts:type_name -> em.EmSendPart
	0, // 2: em.EmService.Send:input_type -> em.EmSendRequest
	2, // 3: em.EmService.Send:output_type -> em.EmSendResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_em_proto_init() }
func file_em_proto_init() {
	if File_em_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_em_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmSendRequest); i {
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
		file_em_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmSendPart); i {
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
		file_em_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EmSendResponse); i {
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
			RawDescriptor: file_em_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_em_proto_goTypes,
		DependencyIndexes: file_em_proto_depIdxs,
		MessageInfos:      file_em_proto_msgTypes,
	}.Build()
	File_em_proto = out.File
	file_em_proto_rawDesc = nil
	file_em_proto_goTypes = nil
	file_em_proto_depIdxs = nil
}