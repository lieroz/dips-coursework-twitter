// Code generated by protoc-gen-go. DO NOT EDIT.
// source: utils.proto

package protos

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type NatsMessage_Command int32

const (
	NatsMessage_DeleteUser   NatsMessage_Command = 0
	NatsMessage_CreateTweet  NatsMessage_Command = 1
	NatsMessage_DeleteTweets NatsMessage_Command = 2
)

var NatsMessage_Command_name = map[int32]string{
	0: "DeleteUser",
	1: "CreateTweet",
	2: "DeleteTweets",
}

var NatsMessage_Command_value = map[string]int32{
	"DeleteUser":   0,
	"CreateTweet":  1,
	"DeleteTweets": 2,
}

func (x NatsMessage_Command) String() string {
	return proto.EnumName(NatsMessage_Command_name, int32(x))
}

func (NatsMessage_Command) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_c91c651f4717a5f2, []int{4, 0}
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_c91c651f4717a5f2, []int{0}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type NatsDeleteUserMessage struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Tweets               []int64  `protobuf:"varint,2,rep,packed,name=tweets,proto3" json:"tweets,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NatsDeleteUserMessage) Reset()         { *m = NatsDeleteUserMessage{} }
func (m *NatsDeleteUserMessage) String() string { return proto.CompactTextString(m) }
func (*NatsDeleteUserMessage) ProtoMessage()    {}
func (*NatsDeleteUserMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_c91c651f4717a5f2, []int{1}
}

func (m *NatsDeleteUserMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NatsDeleteUserMessage.Unmarshal(m, b)
}
func (m *NatsDeleteUserMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NatsDeleteUserMessage.Marshal(b, m, deterministic)
}
func (m *NatsDeleteUserMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NatsDeleteUserMessage.Merge(m, src)
}
func (m *NatsDeleteUserMessage) XXX_Size() int {
	return xxx_messageInfo_NatsDeleteUserMessage.Size(m)
}
func (m *NatsDeleteUserMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_NatsDeleteUserMessage.DiscardUnknown(m)
}

var xxx_messageInfo_NatsDeleteUserMessage proto.InternalMessageInfo

func (m *NatsDeleteUserMessage) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *NatsDeleteUserMessage) GetTweets() []int64 {
	if m != nil {
		return m.Tweets
	}
	return nil
}

type NatsCreateTweetMessage struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	TweetId              int64    `protobuf:"varint,2,opt,name=tweet_id,json=tweetId,proto3" json:"tweet_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NatsCreateTweetMessage) Reset()         { *m = NatsCreateTweetMessage{} }
func (m *NatsCreateTweetMessage) String() string { return proto.CompactTextString(m) }
func (*NatsCreateTweetMessage) ProtoMessage()    {}
func (*NatsCreateTweetMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_c91c651f4717a5f2, []int{2}
}

func (m *NatsCreateTweetMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NatsCreateTweetMessage.Unmarshal(m, b)
}
func (m *NatsCreateTweetMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NatsCreateTweetMessage.Marshal(b, m, deterministic)
}
func (m *NatsCreateTweetMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NatsCreateTweetMessage.Merge(m, src)
}
func (m *NatsCreateTweetMessage) XXX_Size() int {
	return xxx_messageInfo_NatsCreateTweetMessage.Size(m)
}
func (m *NatsCreateTweetMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_NatsCreateTweetMessage.DiscardUnknown(m)
}

var xxx_messageInfo_NatsCreateTweetMessage proto.InternalMessageInfo

func (m *NatsCreateTweetMessage) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *NatsCreateTweetMessage) GetTweetId() int64 {
	if m != nil {
		return m.TweetId
	}
	return 0
}

type NatsDeleteTweetsMessage struct {
	Tweets               []int64  `protobuf:"varint,1,rep,packed,name=tweets,proto3" json:"tweets,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NatsDeleteTweetsMessage) Reset()         { *m = NatsDeleteTweetsMessage{} }
func (m *NatsDeleteTweetsMessage) String() string { return proto.CompactTextString(m) }
func (*NatsDeleteTweetsMessage) ProtoMessage()    {}
func (*NatsDeleteTweetsMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_c91c651f4717a5f2, []int{3}
}

func (m *NatsDeleteTweetsMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NatsDeleteTweetsMessage.Unmarshal(m, b)
}
func (m *NatsDeleteTweetsMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NatsDeleteTweetsMessage.Marshal(b, m, deterministic)
}
func (m *NatsDeleteTweetsMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NatsDeleteTweetsMessage.Merge(m, src)
}
func (m *NatsDeleteTweetsMessage) XXX_Size() int {
	return xxx_messageInfo_NatsDeleteTweetsMessage.Size(m)
}
func (m *NatsDeleteTweetsMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_NatsDeleteTweetsMessage.DiscardUnknown(m)
}

var xxx_messageInfo_NatsDeleteTweetsMessage proto.InternalMessageInfo

func (m *NatsDeleteTweetsMessage) GetTweets() []int64 {
	if m != nil {
		return m.Tweets
	}
	return nil
}

type NatsMessage struct {
	Command              NatsMessage_Command `protobuf:"varint,1,opt,name=command,proto3,enum=NatsMessage_Command" json:"command,omitempty"`
	Message              []byte              `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *NatsMessage) Reset()         { *m = NatsMessage{} }
func (m *NatsMessage) String() string { return proto.CompactTextString(m) }
func (*NatsMessage) ProtoMessage()    {}
func (*NatsMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_c91c651f4717a5f2, []int{4}
}

func (m *NatsMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NatsMessage.Unmarshal(m, b)
}
func (m *NatsMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NatsMessage.Marshal(b, m, deterministic)
}
func (m *NatsMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NatsMessage.Merge(m, src)
}
func (m *NatsMessage) XXX_Size() int {
	return xxx_messageInfo_NatsMessage.Size(m)
}
func (m *NatsMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_NatsMessage.DiscardUnknown(m)
}

var xxx_messageInfo_NatsMessage proto.InternalMessageInfo

func (m *NatsMessage) GetCommand() NatsMessage_Command {
	if m != nil {
		return m.Command
	}
	return NatsMessage_DeleteUser
}

func (m *NatsMessage) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

func init() {
	proto.RegisterEnum("NatsMessage_Command", NatsMessage_Command_name, NatsMessage_Command_value)
	proto.RegisterType((*Empty)(nil), "Empty")
	proto.RegisterType((*NatsDeleteUserMessage)(nil), "NatsDeleteUserMessage")
	proto.RegisterType((*NatsCreateTweetMessage)(nil), "NatsCreateTweetMessage")
	proto.RegisterType((*NatsDeleteTweetsMessage)(nil), "NatsDeleteTweetsMessage")
	proto.RegisterType((*NatsMessage)(nil), "NatsMessage")
}

func init() {
	proto.RegisterFile("utils.proto", fileDescriptor_c91c651f4717a5f2)
}

var fileDescriptor_c91c651f4717a5f2 = []byte{
	// 253 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0xc1, 0x4b, 0xc3, 0x30,
	0x18, 0xc5, 0x4d, 0x8b, 0x4b, 0x7d, 0x1d, 0xb3, 0x04, 0x9d, 0xd5, 0x53, 0xc9, 0xa9, 0xa7, 0x80,
	0x7a, 0xd4, 0x93, 0xd3, 0x83, 0x88, 0x0a, 0x41, 0x2f, 0x5e, 0x24, 0xda, 0x0f, 0x19, 0x2c, 0xeb,
	0x68, 0x32, 0xc4, 0x3f, 0xc4, 0xff, 0x57, 0x9a, 0xad, 0xb6, 0xde, 0x3c, 0x85, 0x97, 0xf7, 0xbe,
	0x97, 0x5f, 0x3e, 0xa4, 0x6b, 0x3f, 0x5f, 0x38, 0xb5, 0x6a, 0x6a, 0x5f, 0x4b, 0x8e, 0xdd, 0x1b,
	0xbb, 0xf2, 0x5f, 0xf2, 0x0e, 0x87, 0x0f, 0xc6, 0xbb, 0x6b, 0x5a, 0x90, 0xa7, 0x67, 0x47, 0xcd,
	0x3d, 0x39, 0x67, 0x3e, 0x48, 0x9c, 0x20, 0x59, 0x3b, 0x6a, 0x96, 0xc6, 0x52, 0xce, 0x0a, 0x56,
	0xee, 0xe9, 0x5f, 0x2d, 0xa6, 0x18, 0xf9, 0x4f, 0x22, 0xef, 0xf2, 0xa8, 0x88, 0xcb, 0x58, 0x6f,
	0x95, 0x7c, 0xc4, 0xb4, 0x2d, 0x9b, 0x35, 0x64, 0x3c, 0x3d, 0xb5, 0x77, 0xff, 0x69, 0x3b, 0x46,
	0x12, 0xe6, 0x5f, 0xe7, 0x55, 0x1e, 0x15, 0xac, 0x8c, 0x35, 0x0f, 0xfa, 0xb6, 0x92, 0xa7, 0x38,
	0xea, 0xe9, 0x42, 0xa1, 0xeb, 0x1a, 0x7b, 0x06, 0xf6, 0x87, 0xe1, 0x9b, 0x21, 0x6d, 0x67, 0xba,
	0x9c, 0x02, 0x7f, 0xaf, 0xad, 0x35, 0xcb, 0x2a, 0x3c, 0x3c, 0x39, 0x3b, 0x50, 0x03, 0x5b, 0xcd,
	0x36, 0x9e, 0xee, 0x42, 0x22, 0x07, 0xb7, 0x1b, 0x2f, 0xc0, 0x8c, 0x75, 0x27, 0xe5, 0x25, 0xf8,
	0x36, 0x2d, 0x26, 0x40, 0xbf, 0xb1, 0x6c, 0x47, 0xec, 0x23, 0x1d, 0x7c, 0x3a, 0x63, 0x22, 0xc3,
	0x78, 0x08, 0x9d, 0x45, 0x57, 0x78, 0x49, 0xd4, 0x45, 0x58, 0xbe, 0x7b, 0x1b, 0x85, 0xf3, 0xfc,
	0x27, 0x00, 0x00, 0xff, 0xff, 0x99, 0x93, 0x2c, 0xe2, 0x93, 0x01, 0x00, 0x00,
}
