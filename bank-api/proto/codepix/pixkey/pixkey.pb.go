// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.1
// source: proto/codepix/pixkey/pixkey.proto

package pixkey

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

type Type int32

const (
	Type__     Type = 0
	Type_CPF   Type = 1
	Type_Phone Type = 2
	Type_Email Type = 3
)

// Enum value maps for Type.
var (
	Type_name = map[int32]string{
		0: "_",
		1: "CPF",
		2: "Phone",
		3: "Email",
	}
	Type_value = map[string]int32{
		"_":     0,
		"CPF":   1,
		"Phone": 2,
		"Email": 3,
	}
)

func (x Type) Enum() *Type {
	p := new(Type)
	*p = x
	return p
}

func (x Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Type) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_codepix_pixkey_pixkey_proto_enumTypes[0].Descriptor()
}

func (Type) Type() protoreflect.EnumType {
	return &file_proto_codepix_pixkey_pixkey_proto_enumTypes[0]
}

func (x Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Type.Descriptor instead.
func (Type) EnumDescriptor() ([]byte, []int) {
	return file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP(), []int{0}
}

type RegisterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type      Type   `protobuf:"varint,1,opt,name=type,proto3,enum=codepix.pixkey.Type" json:"type,omitempty" validate:"required,oneof=1 2 3"`  // @gotags: validate:"required,oneof=1 2 3"
	Key       string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty" validate:"required,max=100" mod:"trim"`                              // @gotags: validate:"required,max=100" mod:"trim"
	AccountId []byte `protobuf:"bytes,3,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty" validate:"required,len=16"` // @gotags: validate:"required,len=16"
}

func (x *RegisterRequest) Reset() {
	*x = RegisterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterRequest) ProtoMessage() {}

func (x *RegisterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterRequest.ProtoReflect.Descriptor instead.
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterRequest) GetType() Type {
	if x != nil {
		return x.Type
	}
	return Type__
}

func (x *RegisterRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *RegisterRequest) GetAccountId() []byte {
	if x != nil {
		return x.AccountId
	}
	return nil
}

type RegisterReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *RegisterReply) Reset() {
	*x = RegisterReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterReply) ProtoMessage() {}

func (x *RegisterReply) ProtoReflect() protoreflect.Message {
	mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterReply.ProtoReflect.Descriptor instead.
func (*RegisterReply) Descriptor() ([]byte, []int) {
	return file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterReply) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

type FindRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty" validate:"required"` // @gotags: validate:"required"
}

func (x *FindRequest) Reset() {
	*x = FindRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindRequest) ProtoMessage() {}

func (x *FindRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindRequest.ProtoReflect.Descriptor instead.
func (*FindRequest) Descriptor() ([]byte, []int) {
	return file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP(), []int{2}
}

func (x *FindRequest) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

type FindReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Type      Type   `protobuf:"varint,2,opt,name=type,proto3,enum=codepix.pixkey.Type" json:"type,omitempty"`
	Key       string `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	AccountId []byte `protobuf:"bytes,4,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
}

func (x *FindReply) Reset() {
	*x = FindReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindReply) ProtoMessage() {}

func (x *FindReply) ProtoReflect() protoreflect.Message {
	mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindReply.ProtoReflect.Descriptor instead.
func (*FindReply) Descriptor() ([]byte, []int) {
	return file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP(), []int{3}
}

func (x *FindReply) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *FindReply) GetType() Type {
	if x != nil {
		return x.Type
	}
	return Type__
}

func (x *FindReply) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *FindReply) GetAccountId() []byte {
	if x != nil {
		return x.AccountId
	}
	return nil
}

type ListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AccountId []byte `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty" validate:"required"` // @gotags: validate:"required"
}

func (x *ListRequest) Reset() {
	*x = ListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRequest) ProtoMessage() {}

func (x *ListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRequest.ProtoReflect.Descriptor instead.
func (*ListRequest) Descriptor() ([]byte, []int) {
	return file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP(), []int{4}
}

func (x *ListRequest) GetAccountId() []byte {
	if x != nil {
		return x.AccountId
	}
	return nil
}

type ListItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Type Type   `protobuf:"varint,2,opt,name=type,proto3,enum=codepix.pixkey.Type" json:"type,omitempty"`
	Key  string `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *ListItem) Reset() {
	*x = ListItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListItem) ProtoMessage() {}

func (x *ListItem) ProtoReflect() protoreflect.Message {
	mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListItem.ProtoReflect.Descriptor instead.
func (*ListItem) Descriptor() ([]byte, []int) {
	return file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP(), []int{5}
}

func (x *ListItem) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *ListItem) GetType() Type {
	if x != nil {
		return x.Type
	}
	return Type__
}

func (x *ListItem) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

type ListReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*ListItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ListReply) Reset() {
	*x = ListReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListReply) ProtoMessage() {}

func (x *ListReply) ProtoReflect() protoreflect.Message {
	mi := &file_proto_codepix_pixkey_pixkey_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListReply.ProtoReflect.Descriptor instead.
func (*ListReply) Descriptor() ([]byte, []int) {
	return file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP(), []int{6}
}

func (x *ListReply) GetItems() []*ListItem {
	if x != nil {
		return x.Items
	}
	return nil
}

var File_proto_codepix_pixkey_pixkey_proto protoreflect.FileDescriptor

var file_proto_codepix_pixkey_pixkey_proto_rawDesc = []byte{
	0x0a, 0x21, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2f,
	0x70, 0x69, 0x78, 0x6b, 0x65, 0x79, 0x2f, 0x70, 0x69, 0x78, 0x6b, 0x65, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x63, 0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78,
	0x6b, 0x65, 0x79, 0x22, 0x6c, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x28, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70,
	0x69, 0x78, 0x6b, 0x65, 0x79, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49,
	0x64, 0x22, 0x1f, 0x0a, 0x0d, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02,
	0x69, 0x64, 0x22, 0x1d, 0x0a, 0x0b, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69,
	0x64, 0x22, 0x76, 0x0a, 0x09, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x28,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x63,
	0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78, 0x6b, 0x65, 0x79, 0x2e, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x22, 0x2c, 0x0a, 0x0b, 0x4c, 0x69, 0x73,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x22, 0x56, 0x0a, 0x08, 0x4c, 0x69, 0x73, 0x74, 0x49,
	0x74, 0x65, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x28, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x14, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78, 0x6b,
	0x65, 0x79, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22,
	0x3b, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x2e, 0x0a, 0x05,
	0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x63, 0x6f,
	0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78, 0x6b, 0x65, 0x79, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x2a, 0x2c, 0x0a, 0x04,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x05, 0x0a, 0x01, 0x5f, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x43,
	0x50, 0x46, 0x10, 0x01, 0x12, 0x09, 0x0a, 0x05, 0x50, 0x68, 0x6f, 0x6e, 0x65, 0x10, 0x02, 0x12,
	0x09, 0x0a, 0x05, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x10, 0x03, 0x32, 0xdb, 0x01, 0x0a, 0x07, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4c, 0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x12, 0x1f, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78,
	0x6b, 0x65, 0x79, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69,
	0x78, 0x6b, 0x65, 0x79, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x22, 0x00, 0x12, 0x40, 0x0a, 0x04, 0x46, 0x69, 0x6e, 0x64, 0x12, 0x1b, 0x2e, 0x63,
	0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78, 0x6b, 0x65, 0x79, 0x2e, 0x46, 0x69,
	0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x63, 0x6f, 0x64, 0x65,
	0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78, 0x6b, 0x65, 0x79, 0x2e, 0x46, 0x69, 0x6e, 0x64, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x40, 0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1b,
	0x2e, 0x63, 0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78, 0x6b, 0x65, 0x79, 0x2e,
	0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x63, 0x6f,
	0x64, 0x65, 0x70, 0x69, 0x78, 0x2e, 0x70, 0x69, 0x78, 0x6b, 0x65, 0x79, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x27, 0x5a, 0x25, 0x63, 0x6f, 0x64, 0x65,
	0x70, 0x69, 0x78, 0x2f, 0x62, 0x61, 0x6e, 0x6b, 0x2d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x70, 0x69, 0x78, 0x2f, 0x70, 0x69, 0x78, 0x6b, 0x65,
	0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_codepix_pixkey_pixkey_proto_rawDescOnce sync.Once
	file_proto_codepix_pixkey_pixkey_proto_rawDescData = file_proto_codepix_pixkey_pixkey_proto_rawDesc
)

func file_proto_codepix_pixkey_pixkey_proto_rawDescGZIP() []byte {
	file_proto_codepix_pixkey_pixkey_proto_rawDescOnce.Do(func() {
		file_proto_codepix_pixkey_pixkey_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_codepix_pixkey_pixkey_proto_rawDescData)
	})
	return file_proto_codepix_pixkey_pixkey_proto_rawDescData
}

var file_proto_codepix_pixkey_pixkey_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_codepix_pixkey_pixkey_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_proto_codepix_pixkey_pixkey_proto_goTypes = []interface{}{
	(Type)(0),               // 0: codepix.pixkey.Type
	(*RegisterRequest)(nil), // 1: codepix.pixkey.RegisterRequest
	(*RegisterReply)(nil),   // 2: codepix.pixkey.RegisterReply
	(*FindRequest)(nil),     // 3: codepix.pixkey.FindRequest
	(*FindReply)(nil),       // 4: codepix.pixkey.FindReply
	(*ListRequest)(nil),     // 5: codepix.pixkey.ListRequest
	(*ListItem)(nil),        // 6: codepix.pixkey.ListItem
	(*ListReply)(nil),       // 7: codepix.pixkey.ListReply
}
var file_proto_codepix_pixkey_pixkey_proto_depIdxs = []int32{
	0, // 0: codepix.pixkey.RegisterRequest.type:type_name -> codepix.pixkey.Type
	0, // 1: codepix.pixkey.FindReply.type:type_name -> codepix.pixkey.Type
	0, // 2: codepix.pixkey.ListItem.type:type_name -> codepix.pixkey.Type
	6, // 3: codepix.pixkey.ListReply.items:type_name -> codepix.pixkey.ListItem
	1, // 4: codepix.pixkey.Service.Register:input_type -> codepix.pixkey.RegisterRequest
	3, // 5: codepix.pixkey.Service.Find:input_type -> codepix.pixkey.FindRequest
	5, // 6: codepix.pixkey.Service.List:input_type -> codepix.pixkey.ListRequest
	2, // 7: codepix.pixkey.Service.Register:output_type -> codepix.pixkey.RegisterReply
	4, // 8: codepix.pixkey.Service.Find:output_type -> codepix.pixkey.FindReply
	7, // 9: codepix.pixkey.Service.List:output_type -> codepix.pixkey.ListReply
	7, // [7:10] is the sub-list for method output_type
	4, // [4:7] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_codepix_pixkey_pixkey_proto_init() }
func file_proto_codepix_pixkey_pixkey_proto_init() {
	if File_proto_codepix_pixkey_pixkey_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_codepix_pixkey_pixkey_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterRequest); i {
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
		file_proto_codepix_pixkey_pixkey_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterReply); i {
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
		file_proto_codepix_pixkey_pixkey_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindRequest); i {
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
		file_proto_codepix_pixkey_pixkey_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindReply); i {
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
		file_proto_codepix_pixkey_pixkey_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRequest); i {
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
		file_proto_codepix_pixkey_pixkey_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListItem); i {
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
		file_proto_codepix_pixkey_pixkey_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListReply); i {
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
			RawDescriptor: file_proto_codepix_pixkey_pixkey_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_codepix_pixkey_pixkey_proto_goTypes,
		DependencyIndexes: file_proto_codepix_pixkey_pixkey_proto_depIdxs,
		EnumInfos:         file_proto_codepix_pixkey_pixkey_proto_enumTypes,
		MessageInfos:      file_proto_codepix_pixkey_pixkey_proto_msgTypes,
	}.Build()
	File_proto_codepix_pixkey_pixkey_proto = out.File
	file_proto_codepix_pixkey_pixkey_proto_rawDesc = nil
	file_proto_codepix_pixkey_pixkey_proto_goTypes = nil
	file_proto_codepix_pixkey_pixkey_proto_depIdxs = nil
}
