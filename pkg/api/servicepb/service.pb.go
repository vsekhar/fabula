// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.3
// source: service.proto

package servicepb

import (
	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/protoc-gen-go/descriptor"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	refpb "github.com/vsekhar/fabula/pkg/api/refpb"
	regionpb "github.com/vsekhar/fabula/pkg/api/regionpb"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// For a given entry, a Sequencing represents that entry's location in the top
// level sequence and down through the tree of packs.
type Sequencing struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SeqNo         uint64               `protobuf:"varint,1,opt,name=seq_no,json=seqNo,proto3" json:"seq_no,omitempty"`
	SequenceEntry *refpb.SequenceRef   `protobuf:"bytes,2,opt,name=sequence_entry,json=sequenceEntry,proto3" json:"sequence_entry,omitempty"`
	Timestamp     *timestamp.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// For a given entry, batches are listed in top-down order linking the
	// notarization entry to its sequence entry.
	//
	// For priors, batches is empty.
	Batches []*refpb.BatchRef `protobuf:"bytes,4,rep,name=batches,proto3" json:"batches,omitempty"`
}

func (x *Sequencing) Reset() {
	*x = Sequencing{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Sequencing) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Sequencing) ProtoMessage() {}

func (x *Sequencing) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Sequencing.ProtoReflect.Descriptor instead.
func (*Sequencing) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{0}
}

func (x *Sequencing) GetSeqNo() uint64 {
	if x != nil {
		return x.SeqNo
	}
	return 0
}

func (x *Sequencing) GetSequenceEntry() *refpb.SequenceRef {
	if x != nil {
		return x.SequenceEntry
	}
	return nil
}

func (x *Sequencing) GetTimestamp() *timestamp.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *Sequencing) GetBatches() []*refpb.BatchRef {
	if x != nil {
		return x.Batches
	}
	return nil
}

type NotarizeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Region   regionpb.Region `protobuf:"varint,1,opt,name=region,proto3,enum=fabula.Region" json:"region,omitempty"`
	Document []byte          `protobuf:"bytes,2,opt,name=document,proto3" json:"document,omitempty"` // must be 64 bytes
}

func (x *NotarizeRequest) Reset() {
	*x = NotarizeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotarizeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotarizeRequest) ProtoMessage() {}

func (x *NotarizeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotarizeRequest.ProtoReflect.Descriptor instead.
func (*NotarizeRequest) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{1}
}

func (x *NotarizeRequest) GetRegion() regionpb.Region {
	if x != nil {
		return x.Region
	}
	return regionpb.Region_REGION_UNSPECIFIED
}

func (x *NotarizeRequest) GetDocument() []byte {
	if x != nil {
		return x.Document
	}
	return nil
}

type NotarizeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prior      *Sequencing  `protobuf:"bytes,1,opt,name=prior,proto3" json:"prior,omitempty"`           // contains prior.timestamp
	Entry      *refpb.Entry `protobuf:"bytes,2,opt,name=entry,proto3" json:"entry,omitempty"`           // contains notarization_sha3512
	Sequencing *Sequencing  `protobuf:"bytes,3,opt,name=sequencing,proto3" json:"sequencing,omitempty"` // contains current entry timestamp
}

func (x *NotarizeResponse) Reset() {
	*x = NotarizeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotarizeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotarizeResponse) ProtoMessage() {}

func (x *NotarizeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotarizeResponse.ProtoReflect.Descriptor instead.
func (*NotarizeResponse) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{2}
}

func (x *NotarizeResponse) GetPrior() *Sequencing {
	if x != nil {
		return x.Prior
	}
	return nil
}

func (x *NotarizeResponse) GetEntry() *refpb.Entry {
	if x != nil {
		return x.Entry
	}
	return nil
}

func (x *NotarizeResponse) GetSequencing() *Sequencing {
	if x != nil {
		return x.Sequencing
	}
	return nil
}

type BatchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Batch *refpb.BatchRef `protobuf:"bytes,1,opt,name=batch,proto3" json:"batch,omitempty"`
}

func (x *BatchRequest) Reset() {
	*x = BatchRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchRequest) ProtoMessage() {}

func (x *BatchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchRequest.ProtoReflect.Descriptor instead.
func (*BatchRequest) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{3}
}

func (x *BatchRequest) GetBatch() *refpb.BatchRef {
	if x != nil {
		return x.Batch
	}
	return nil
}

type BatchResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Batches []*refpb.BatchRef `protobuf:"bytes,1,rep,name=batches,proto3" json:"batches,omitempty"`
	Entries []*refpb.Entry    `protobuf:"bytes,2,rep,name=entries,proto3" json:"entries,omitempty"`
}

func (x *BatchResponse) Reset() {
	*x = BatchResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchResponse) ProtoMessage() {}

func (x *BatchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchResponse.ProtoReflect.Descriptor instead.
func (*BatchResponse) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{4}
}

func (x *BatchResponse) GetBatches() []*refpb.BatchRef {
	if x != nil {
		return x.Batches
	}
	return nil
}

func (x *BatchResponse) GetEntries() []*refpb.Entry {
	if x != nil {
		return x.Entries
	}
	return nil
}

type SequenceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SeqNo uint64 `protobuf:"varint,1,opt,name=seq_no,json=seqNo,proto3" json:"seq_no,omitempty"`
}

func (x *SequenceRequest) Reset() {
	*x = SequenceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SequenceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SequenceRequest) ProtoMessage() {}

func (x *SequenceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SequenceRequest.ProtoReflect.Descriptor instead.
func (*SequenceRequest) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{5}
}

func (x *SequenceRequest) GetSeqNo() uint64 {
	if x != nil {
		return x.SeqNo
	}
	return 0
}

type SequenceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Root      *refpb.BatchRef      `protobuf:"bytes,1,opt,name=root,proto3" json:"root,omitempty"`
	Timestamp *timestamp.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Prior     *refpb.SequenceRef   `protobuf:"bytes,3,opt,name=prior,proto3" json:"prior,omitempty"`
}

func (x *SequenceResponse) Reset() {
	*x = SequenceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SequenceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SequenceResponse) ProtoMessage() {}

func (x *SequenceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SequenceResponse.ProtoReflect.Descriptor instead.
func (*SequenceResponse) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{6}
}

func (x *SequenceResponse) GetRoot() *refpb.BatchRef {
	if x != nil {
		return x.Root
	}
	return nil
}

func (x *SequenceResponse) GetTimestamp() *timestamp.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *SequenceResponse) GetPrior() *refpb.SequenceRef {
	if x != nil {
		return x.Prior
	}
	return nil
}

type RootRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RootRequest) Reset() {
	*x = RootRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RootRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RootRequest) ProtoMessage() {}

func (x *RootRequest) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RootRequest.ProtoReflect.Descriptor instead.
func (*RootRequest) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{7}
}

type RootResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *RootResponse) Reset() {
	*x = RootResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_service_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RootResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RootResponse) ProtoMessage() {}

func (x *RootResponse) ProtoReflect() protoreflect.Message {
	mi := &file_service_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RootResponse.ProtoReflect.Descriptor instead.
func (*RootResponse) Descriptor() ([]byte, []int) {
	return file_service_proto_rawDescGZIP(), []int{8}
}

func (x *RootResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_service_proto protoreflect.FileDescriptor

var file_service_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x06, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x09, 0x72, 0x65, 0x66, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xc5, 0x01, 0x0a, 0x0a, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x69, 0x6e, 0x67,
	0x12, 0x15, 0x0a, 0x06, 0x73, 0x65, 0x71, 0x5f, 0x6e, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x05, 0x73, 0x65, 0x71, 0x4e, 0x6f, 0x12, 0x3a, 0x0a, 0x0e, 0x73, 0x65, 0x71, 0x75, 0x65,
	0x6e, 0x63, 0x65, 0x5f, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63,
	0x65, 0x52, 0x65, 0x66, 0x52, 0x0d, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x2a, 0x0a,
	0x07, 0x62, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10,
	0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x66,
	0x52, 0x07, 0x62, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x22, 0x55, 0x0a, 0x0f, 0x4e, 0x6f, 0x74,
	0x61, 0x72, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x06,
	0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x66,
	0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x52, 0x06, 0x72, 0x65,
	0x67, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74,
	0x22, 0x95, 0x01, 0x0a, 0x10, 0x4e, 0x6f, 0x74, 0x61, 0x72, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x53, 0x65,
	0x71, 0x75, 0x65, 0x6e, 0x63, 0x69, 0x6e, 0x67, 0x52, 0x05, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x12,
	0x23, 0x0a, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d,
	0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x65,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x32, 0x0a, 0x0a, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x69,
	0x6e, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c,
	0x61, 0x2e, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x69, 0x6e, 0x67, 0x52, 0x0a, 0x73, 0x65,
	0x71, 0x75, 0x65, 0x6e, 0x63, 0x69, 0x6e, 0x67, 0x22, 0x36, 0x0a, 0x0c, 0x42, 0x61, 0x74, 0x63,
	0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x05, 0x62, 0x61, 0x74, 0x63,
	0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61,
	0x2e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x66, 0x52, 0x05, 0x62, 0x61, 0x74, 0x63, 0x68,
	0x22, 0x64, 0x0a, 0x0d, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x2a, 0x0a, 0x07, 0x62, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x42, 0x61, 0x74, 0x63,
	0x68, 0x52, 0x65, 0x66, 0x52, 0x07, 0x62, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x12, 0x27, 0x0a,
	0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d,
	0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x65,
	0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x22, 0x28, 0x0a, 0x0f, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e,
	0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x73, 0x65, 0x71,
	0x5f, 0x6e, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x73, 0x65, 0x71, 0x4e, 0x6f,
	0x22, 0x9d, 0x01, 0x0a, 0x10, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x04, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x52, 0x65, 0x66, 0x52, 0x04, 0x72, 0x6f, 0x6f, 0x74, 0x12, 0x38, 0x0a, 0x09, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x29, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x53, 0x65,
	0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x66, 0x52, 0x05, 0x70, 0x72, 0x69, 0x6f, 0x72,
	0x22, 0x0d, 0x0a, 0x0b, 0x52, 0x6f, 0x6f, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22,
	0x28, 0x0a, 0x0c, 0x52, 0x6f, 0x6f, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x88, 0x03, 0x0a, 0x06, 0x46, 0x61,
	0x62, 0x75, 0x6c, 0x61, 0x12, 0x40, 0x0a, 0x04, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x13, 0x2e, 0x66,
	0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x52, 0x6f, 0x6f, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x14, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x52, 0x6f, 0x6f, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x0d, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x07, 0x12,
	0x05, 0x2f, 0x72, 0x6f, 0x6f, 0x74, 0x12, 0x79, 0x0a, 0x08, 0x4e, 0x6f, 0x74, 0x61, 0x72, 0x69,
	0x7a, 0x65, 0x12, 0x17, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x4e, 0x6f, 0x74, 0x61,
	0x72, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x66, 0x61,
	0x62, 0x75, 0x6c, 0x61, 0x2e, 0x4e, 0x6f, 0x74, 0x61, 0x72, 0x69, 0x7a, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x3a, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x34, 0x22, 0x2f, 0x2f,
	0x76, 0x31, 0x2f, 0x7b, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x3d, 0x72, 0x65, 0x67, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x2a, 0x7d, 0x2f, 0x6e, 0x6f, 0x74, 0x61, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2f, 0x7b, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x7d, 0x3a, 0x01,
	0x2a, 0x12, 0x61, 0x0a, 0x05, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x14, 0x2e, 0x66, 0x61, 0x62,
	0x75, 0x6c, 0x61, 0x2e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x15, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x25, 0x12,
	0x23, 0x2f, 0x76, 0x31, 0x2f, 0x7b, 0x62, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x62, 0x61, 0x74, 0x63,
	0x68, 0x5f, 0x73, 0x68, 0x61, 0x33, 0x35, 0x31, 0x32, 0x3d, 0x62, 0x61, 0x74, 0x63, 0x68, 0x65,
	0x73, 0x2f, 0x2a, 0x7d, 0x12, 0x5e, 0x0a, 0x08, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65,
	0x12, 0x17, 0x2e, 0x66, 0x61, 0x62, 0x75, 0x6c, 0x61, 0x2e, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e,
	0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x66, 0x61, 0x62, 0x75,
	0x6c, 0x61, 0x2e, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x1f, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19, 0x12, 0x17, 0x2f, 0x76, 0x31,
	0x2f, 0x7b, 0x73, 0x65, 0x71, 0x5f, 0x6e, 0x6f, 0x3d, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63,
	0x65, 0x2f, 0x2a, 0x7d, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x76, 0x73, 0x65, 0x6b, 0x68, 0x61, 0x72, 0x2f, 0x66, 0x61, 0x62, 0x75, 0x6c,
	0x61, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_service_proto_rawDescOnce sync.Once
	file_service_proto_rawDescData = file_service_proto_rawDesc
)

func file_service_proto_rawDescGZIP() []byte {
	file_service_proto_rawDescOnce.Do(func() {
		file_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_service_proto_rawDescData)
	})
	return file_service_proto_rawDescData
}

var file_service_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_service_proto_goTypes = []interface{}{
	(*Sequencing)(nil),          // 0: fabula.Sequencing
	(*NotarizeRequest)(nil),     // 1: fabula.NotarizeRequest
	(*NotarizeResponse)(nil),    // 2: fabula.NotarizeResponse
	(*BatchRequest)(nil),        // 3: fabula.BatchRequest
	(*BatchResponse)(nil),       // 4: fabula.BatchResponse
	(*SequenceRequest)(nil),     // 5: fabula.SequenceRequest
	(*SequenceResponse)(nil),    // 6: fabula.SequenceResponse
	(*RootRequest)(nil),         // 7: fabula.RootRequest
	(*RootResponse)(nil),        // 8: fabula.RootResponse
	(*refpb.SequenceRef)(nil),   // 9: fabula.SequenceRef
	(*timestamp.Timestamp)(nil), // 10: google.protobuf.Timestamp
	(*refpb.BatchRef)(nil),      // 11: fabula.BatchRef
	(regionpb.Region)(0),        // 12: fabula.Region
	(*refpb.Entry)(nil),         // 13: fabula.Entry
}
var file_service_proto_depIdxs = []int32{
	9,  // 0: fabula.Sequencing.sequence_entry:type_name -> fabula.SequenceRef
	10, // 1: fabula.Sequencing.timestamp:type_name -> google.protobuf.Timestamp
	11, // 2: fabula.Sequencing.batches:type_name -> fabula.BatchRef
	12, // 3: fabula.NotarizeRequest.region:type_name -> fabula.Region
	0,  // 4: fabula.NotarizeResponse.prior:type_name -> fabula.Sequencing
	13, // 5: fabula.NotarizeResponse.entry:type_name -> fabula.Entry
	0,  // 6: fabula.NotarizeResponse.sequencing:type_name -> fabula.Sequencing
	11, // 7: fabula.BatchRequest.batch:type_name -> fabula.BatchRef
	11, // 8: fabula.BatchResponse.batches:type_name -> fabula.BatchRef
	13, // 9: fabula.BatchResponse.entries:type_name -> fabula.Entry
	11, // 10: fabula.SequenceResponse.root:type_name -> fabula.BatchRef
	10, // 11: fabula.SequenceResponse.timestamp:type_name -> google.protobuf.Timestamp
	9,  // 12: fabula.SequenceResponse.prior:type_name -> fabula.SequenceRef
	7,  // 13: fabula.Fabula.Root:input_type -> fabula.RootRequest
	1,  // 14: fabula.Fabula.Notarize:input_type -> fabula.NotarizeRequest
	3,  // 15: fabula.Fabula.Batch:input_type -> fabula.BatchRequest
	5,  // 16: fabula.Fabula.Sequence:input_type -> fabula.SequenceRequest
	8,  // 17: fabula.Fabula.Root:output_type -> fabula.RootResponse
	2,  // 18: fabula.Fabula.Notarize:output_type -> fabula.NotarizeResponse
	4,  // 19: fabula.Fabula.Batch:output_type -> fabula.BatchResponse
	6,  // 20: fabula.Fabula.Sequence:output_type -> fabula.SequenceResponse
	17, // [17:21] is the sub-list for method output_type
	13, // [13:17] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_service_proto_init() }
func file_service_proto_init() {
	if File_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Sequencing); i {
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
		file_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotarizeRequest); i {
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
		file_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotarizeResponse); i {
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
		file_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchRequest); i {
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
		file_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchResponse); i {
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
		file_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SequenceRequest); i {
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
		file_service_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SequenceResponse); i {
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
		file_service_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RootRequest); i {
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
		file_service_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RootResponse); i {
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
			RawDescriptor: file_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_service_proto_goTypes,
		DependencyIndexes: file_service_proto_depIdxs,
		MessageInfos:      file_service_proto_msgTypes,
	}.Build()
	File_service_proto = out.File
	file_service_proto_rawDesc = nil
	file_service_proto_goTypes = nil
	file_service_proto_depIdxs = nil
}
