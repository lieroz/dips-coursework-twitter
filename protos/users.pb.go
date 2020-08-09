// Code generated by protoc-gen-go. DO NOT EDIT.
// source: users.proto

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

type Role int32

const (
	Role_Follower Role = 0
	Role_Followed Role = 1
)

var Role_name = map[int32]string{
	0: "Follower",
	1: "Followed",
}

var Role_value = map[string]int32{
	"Follower": 0,
	"Followed": 1,
}

func (x Role) String() string {
	return proto.EnumName(Role_name, int32(x))
}

func (Role) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{0}
}

type User struct {
	Username              string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Firstname             string   `protobuf:"bytes,2,opt,name=firstname,proto3" json:"firstname,omitempty"`
	Lastname              string   `protobuf:"bytes,3,opt,name=lastname,proto3" json:"lastname,omitempty"`
	Description           string   `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	RegistrationTimestamp int64    `protobuf:"varint,5,opt,name=registration_timestamp,json=registrationTimestamp,proto3" json:"registration_timestamp,omitempty"`
	Followers             []string `protobuf:"bytes,6,rep,name=followers,proto3" json:"followers,omitempty"`
	Following             []string `protobuf:"bytes,7,rep,name=following,proto3" json:"following,omitempty"`
	Tweets                []int64  `protobuf:"varint,8,rep,packed,name=tweets,proto3" json:"tweets,omitempty"`
	XXX_NoUnkeyedLiteral  struct{} `json:"-"`
	XXX_unrecognized      []byte   `json:"-"`
	XXX_sizecache         int32    `json:"-"`
}

func (m *User) Reset()         { *m = User{} }
func (m *User) String() string { return proto.CompactTextString(m) }
func (*User) ProtoMessage()    {}
func (*User) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{0}
}

func (m *User) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_User.Unmarshal(m, b)
}
func (m *User) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_User.Marshal(b, m, deterministic)
}
func (m *User) XXX_Merge(src proto.Message) {
	xxx_messageInfo_User.Merge(m, src)
}
func (m *User) XXX_Size() int {
	return xxx_messageInfo_User.Size(m)
}
func (m *User) XXX_DiscardUnknown() {
	xxx_messageInfo_User.DiscardUnknown(m)
}

var xxx_messageInfo_User proto.InternalMessageInfo

func (m *User) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *User) GetFirstname() string {
	if m != nil {
		return m.Firstname
	}
	return ""
}

func (m *User) GetLastname() string {
	if m != nil {
		return m.Lastname
	}
	return ""
}

func (m *User) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *User) GetRegistrationTimestamp() int64 {
	if m != nil {
		return m.RegistrationTimestamp
	}
	return 0
}

func (m *User) GetFollowers() []string {
	if m != nil {
		return m.Followers
	}
	return nil
}

func (m *User) GetFollowing() []string {
	if m != nil {
		return m.Following
	}
	return nil
}

func (m *User) GetTweets() []int64 {
	if m != nil {
		return m.Tweets
	}
	return nil
}

type CreateRequest struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Firstname            string   `protobuf:"bytes,2,opt,name=firstname,proto3" json:"firstname,omitempty"`
	Lastname             string   `protobuf:"bytes,3,opt,name=lastname,proto3" json:"lastname,omitempty"`
	Description          string   `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateRequest) Reset()         { *m = CreateRequest{} }
func (m *CreateRequest) String() string { return proto.CompactTextString(m) }
func (*CreateRequest) ProtoMessage()    {}
func (*CreateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{1}
}

func (m *CreateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateRequest.Unmarshal(m, b)
}
func (m *CreateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateRequest.Marshal(b, m, deterministic)
}
func (m *CreateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateRequest.Merge(m, src)
}
func (m *CreateRequest) XXX_Size() int {
	return xxx_messageInfo_CreateRequest.Size(m)
}
func (m *CreateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateRequest proto.InternalMessageInfo

func (m *CreateRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *CreateRequest) GetFirstname() string {
	if m != nil {
		return m.Firstname
	}
	return ""
}

func (m *CreateRequest) GetLastname() string {
	if m != nil {
		return m.Lastname
	}
	return ""
}

func (m *CreateRequest) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

type DeleteRequest struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteRequest) Reset()         { *m = DeleteRequest{} }
func (m *DeleteRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteRequest) ProtoMessage()    {}
func (*DeleteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{2}
}

func (m *DeleteRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteRequest.Unmarshal(m, b)
}
func (m *DeleteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteRequest.Marshal(b, m, deterministic)
}
func (m *DeleteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteRequest.Merge(m, src)
}
func (m *DeleteRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteRequest.Size(m)
}
func (m *DeleteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteRequest proto.InternalMessageInfo

func (m *DeleteRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

type GetSummaryRequest struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetSummaryRequest) Reset()         { *m = GetSummaryRequest{} }
func (m *GetSummaryRequest) String() string { return proto.CompactTextString(m) }
func (*GetSummaryRequest) ProtoMessage()    {}
func (*GetSummaryRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{3}
}

func (m *GetSummaryRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetSummaryRequest.Unmarshal(m, b)
}
func (m *GetSummaryRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetSummaryRequest.Marshal(b, m, deterministic)
}
func (m *GetSummaryRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetSummaryRequest.Merge(m, src)
}
func (m *GetSummaryRequest) XXX_Size() int {
	return xxx_messageInfo_GetSummaryRequest.Size(m)
}
func (m *GetSummaryRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetSummaryRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetSummaryRequest proto.InternalMessageInfo

func (m *GetSummaryRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

type GetSummaryReply struct {
	Username              string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Firstname             string   `protobuf:"bytes,2,opt,name=firstname,proto3" json:"firstname,omitempty"`
	Lastname              string   `protobuf:"bytes,3,opt,name=lastname,proto3" json:"lastname,omitempty"`
	Description           string   `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	RegistrationTimestamp int64    `protobuf:"varint,5,opt,name=registration_timestamp,json=registrationTimestamp,proto3" json:"registration_timestamp,omitempty"`
	FollowersCount        int64    `protobuf:"varint,6,opt,name=followers_count,json=followersCount,proto3" json:"followers_count,omitempty"`
	FollowingCount        int64    `protobuf:"varint,7,opt,name=following_count,json=followingCount,proto3" json:"following_count,omitempty"`
	TweetsCount           int64    `protobuf:"varint,8,opt,name=tweets_count,json=tweetsCount,proto3" json:"tweets_count,omitempty"`
	XXX_NoUnkeyedLiteral  struct{} `json:"-"`
	XXX_unrecognized      []byte   `json:"-"`
	XXX_sizecache         int32    `json:"-"`
}

func (m *GetSummaryReply) Reset()         { *m = GetSummaryReply{} }
func (m *GetSummaryReply) String() string { return proto.CompactTextString(m) }
func (*GetSummaryReply) ProtoMessage()    {}
func (*GetSummaryReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{4}
}

func (m *GetSummaryReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetSummaryReply.Unmarshal(m, b)
}
func (m *GetSummaryReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetSummaryReply.Marshal(b, m, deterministic)
}
func (m *GetSummaryReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetSummaryReply.Merge(m, src)
}
func (m *GetSummaryReply) XXX_Size() int {
	return xxx_messageInfo_GetSummaryReply.Size(m)
}
func (m *GetSummaryReply) XXX_DiscardUnknown() {
	xxx_messageInfo_GetSummaryReply.DiscardUnknown(m)
}

var xxx_messageInfo_GetSummaryReply proto.InternalMessageInfo

func (m *GetSummaryReply) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *GetSummaryReply) GetFirstname() string {
	if m != nil {
		return m.Firstname
	}
	return ""
}

func (m *GetSummaryReply) GetLastname() string {
	if m != nil {
		return m.Lastname
	}
	return ""
}

func (m *GetSummaryReply) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *GetSummaryReply) GetRegistrationTimestamp() int64 {
	if m != nil {
		return m.RegistrationTimestamp
	}
	return 0
}

func (m *GetSummaryReply) GetFollowersCount() int64 {
	if m != nil {
		return m.FollowersCount
	}
	return 0
}

func (m *GetSummaryReply) GetFollowingCount() int64 {
	if m != nil {
		return m.FollowingCount
	}
	return 0
}

func (m *GetSummaryReply) GetTweetsCount() int64 {
	if m != nil {
		return m.TweetsCount
	}
	return 0
}

type GetUsersRequest struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Role                 Role     `protobuf:"varint,2,opt,name=role,proto3,enum=Role" json:"role,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetUsersRequest) Reset()         { *m = GetUsersRequest{} }
func (m *GetUsersRequest) String() string { return proto.CompactTextString(m) }
func (*GetUsersRequest) ProtoMessage()    {}
func (*GetUsersRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{5}
}

func (m *GetUsersRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetUsersRequest.Unmarshal(m, b)
}
func (m *GetUsersRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetUsersRequest.Marshal(b, m, deterministic)
}
func (m *GetUsersRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetUsersRequest.Merge(m, src)
}
func (m *GetUsersRequest) XXX_Size() int {
	return xxx_messageInfo_GetUsersRequest.Size(m)
}
func (m *GetUsersRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetUsersRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetUsersRequest proto.InternalMessageInfo

func (m *GetUsersRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *GetUsersRequest) GetRole() Role {
	if m != nil {
		return m.Role
	}
	return Role_Follower
}

type GetUsersReply struct {
	// Types that are valid to be assigned to Reply:
	//	*GetUsersReply_User
	//	*GetUsersReply_Error
	Reply                isGetUsersReply_Reply `protobuf_oneof:"reply"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *GetUsersReply) Reset()         { *m = GetUsersReply{} }
func (m *GetUsersReply) String() string { return proto.CompactTextString(m) }
func (*GetUsersReply) ProtoMessage()    {}
func (*GetUsersReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{6}
}

func (m *GetUsersReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetUsersReply.Unmarshal(m, b)
}
func (m *GetUsersReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetUsersReply.Marshal(b, m, deterministic)
}
func (m *GetUsersReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetUsersReply.Merge(m, src)
}
func (m *GetUsersReply) XXX_Size() int {
	return xxx_messageInfo_GetUsersReply.Size(m)
}
func (m *GetUsersReply) XXX_DiscardUnknown() {
	xxx_messageInfo_GetUsersReply.DiscardUnknown(m)
}

var xxx_messageInfo_GetUsersReply proto.InternalMessageInfo

type isGetUsersReply_Reply interface {
	isGetUsersReply_Reply()
}

type GetUsersReply_User struct {
	User *User `protobuf:"bytes,1,opt,name=user,proto3,oneof"`
}

type GetUsersReply_Error struct {
	Error bool `protobuf:"varint,2,opt,name=error,proto3,oneof"`
}

func (*GetUsersReply_User) isGetUsersReply_Reply() {}

func (*GetUsersReply_Error) isGetUsersReply_Reply() {}

func (m *GetUsersReply) GetReply() isGetUsersReply_Reply {
	if m != nil {
		return m.Reply
	}
	return nil
}

func (m *GetUsersReply) GetUser() *User {
	if x, ok := m.GetReply().(*GetUsersReply_User); ok {
		return x.User
	}
	return nil
}

func (m *GetUsersReply) GetError() bool {
	if x, ok := m.GetReply().(*GetUsersReply_Error); ok {
		return x.Error
	}
	return false
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*GetUsersReply) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*GetUsersReply_User)(nil),
		(*GetUsersReply_Error)(nil),
	}
}

type FollowRequest struct {
	Follower             string   `protobuf:"bytes,1,opt,name=follower,proto3" json:"follower,omitempty"`
	Followed             string   `protobuf:"bytes,2,opt,name=followed,proto3" json:"followed,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FollowRequest) Reset()         { *m = FollowRequest{} }
func (m *FollowRequest) String() string { return proto.CompactTextString(m) }
func (*FollowRequest) ProtoMessage()    {}
func (*FollowRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{7}
}

func (m *FollowRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FollowRequest.Unmarshal(m, b)
}
func (m *FollowRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FollowRequest.Marshal(b, m, deterministic)
}
func (m *FollowRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FollowRequest.Merge(m, src)
}
func (m *FollowRequest) XXX_Size() int {
	return xxx_messageInfo_FollowRequest.Size(m)
}
func (m *FollowRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FollowRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FollowRequest proto.InternalMessageInfo

func (m *FollowRequest) GetFollower() string {
	if m != nil {
		return m.Follower
	}
	return ""
}

func (m *FollowRequest) GetFollowed() string {
	if m != nil {
		return m.Followed
	}
	return ""
}

type OnTweetCreatedRequest struct {
	Username             string   `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	TweetId              int64    `protobuf:"varint,2,opt,name=tweet_id,json=tweetId,proto3" json:"tweet_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OnTweetCreatedRequest) Reset()         { *m = OnTweetCreatedRequest{} }
func (m *OnTweetCreatedRequest) String() string { return proto.CompactTextString(m) }
func (*OnTweetCreatedRequest) ProtoMessage()    {}
func (*OnTweetCreatedRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_030765f334c86cea, []int{8}
}

func (m *OnTweetCreatedRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OnTweetCreatedRequest.Unmarshal(m, b)
}
func (m *OnTweetCreatedRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OnTweetCreatedRequest.Marshal(b, m, deterministic)
}
func (m *OnTweetCreatedRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OnTweetCreatedRequest.Merge(m, src)
}
func (m *OnTweetCreatedRequest) XXX_Size() int {
	return xxx_messageInfo_OnTweetCreatedRequest.Size(m)
}
func (m *OnTweetCreatedRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_OnTweetCreatedRequest.DiscardUnknown(m)
}

var xxx_messageInfo_OnTweetCreatedRequest proto.InternalMessageInfo

func (m *OnTweetCreatedRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *OnTweetCreatedRequest) GetTweetId() int64 {
	if m != nil {
		return m.TweetId
	}
	return 0
}

func init() {
	proto.RegisterEnum("Role", Role_name, Role_value)
	proto.RegisterType((*User)(nil), "User")
	proto.RegisterType((*CreateRequest)(nil), "CreateRequest")
	proto.RegisterType((*DeleteRequest)(nil), "DeleteRequest")
	proto.RegisterType((*GetSummaryRequest)(nil), "GetSummaryRequest")
	proto.RegisterType((*GetSummaryReply)(nil), "GetSummaryReply")
	proto.RegisterType((*GetUsersRequest)(nil), "GetUsersRequest")
	proto.RegisterType((*GetUsersReply)(nil), "GetUsersReply")
	proto.RegisterType((*FollowRequest)(nil), "FollowRequest")
	proto.RegisterType((*OnTweetCreatedRequest)(nil), "OnTweetCreatedRequest")
}

func init() {
	proto.RegisterFile("users.proto", fileDescriptor_030765f334c86cea)
}

var fileDescriptor_030765f334c86cea = []byte{
	// 555 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xd4, 0x54, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0xad, 0xe3, 0xd8, 0x71, 0x27, 0x8d, 0x1b, 0x56, 0x6a, 0x64, 0x02, 0x07, 0x63, 0x21, 0x11,
	0x81, 0xb4, 0x54, 0x41, 0x5c, 0xe0, 0xd6, 0x42, 0x9b, 0x1e, 0x00, 0xc9, 0xb4, 0x17, 0x2e, 0x91,
	0x69, 0x36, 0xd1, 0x4a, 0xfe, 0x08, 0xbb, 0x1b, 0x55, 0xf9, 0x05, 0x88, 0x1f, 0x84, 0xf8, 0x7b,
	0x68, 0x67, 0x6d, 0xc7, 0x06, 0x84, 0x72, 0x43, 0x9c, 0xac, 0x37, 0xef, 0xcd, 0x7a, 0xe6, 0xcd,
	0xce, 0x42, 0x7f, 0x23, 0x99, 0x90, 0x74, 0x2d, 0x0a, 0x55, 0x8c, 0xfb, 0x1b, 0xc5, 0xd3, 0x12,
	0x44, 0xdf, 0x3a, 0xd0, 0xbd, 0x91, 0x4c, 0x90, 0x31, 0x78, 0x5a, 0x94, 0x27, 0x19, 0x0b, 0xac,
	0xd0, 0x9a, 0x1c, 0xc6, 0x35, 0x26, 0x0f, 0xe1, 0x70, 0xc9, 0x85, 0x54, 0x48, 0x76, 0x90, 0xdc,
	0x05, 0x74, 0x66, 0x9a, 0x94, 0xa4, 0x6d, 0x32, 0x2b, 0x4c, 0x42, 0xe8, 0x2f, 0x98, 0xbc, 0x15,
	0x7c, 0xad, 0x78, 0x91, 0x07, 0x5d, 0xa4, 0x9b, 0x21, 0xf2, 0x12, 0x46, 0x82, 0xad, 0xb8, 0x54,
	0x22, 0xd1, 0x78, 0xae, 0x78, 0xc6, 0xa4, 0x4a, 0xb2, 0x75, 0xe0, 0x84, 0xd6, 0xc4, 0x8e, 0x4f,
	0x9a, 0xec, 0x75, 0x45, 0x62, 0x49, 0x45, 0x9a, 0x16, 0x77, 0x4c, 0xc8, 0xc0, 0x0d, 0x6d, 0x2c,
	0xa9, 0x0a, 0xec, 0x58, 0x9e, 0xaf, 0x82, 0x5e, 0x93, 0xe5, 0xf9, 0x8a, 0x8c, 0xc0, 0x55, 0x77,
	0x8c, 0x29, 0x19, 0x78, 0xa1, 0x3d, 0xb1, 0xe3, 0x12, 0x45, 0x5f, 0x2d, 0x18, 0x9c, 0x0b, 0x96,
	0x28, 0x16, 0xb3, 0x2f, 0x1b, 0x26, 0xd5, 0xbf, 0x32, 0x25, 0x7a, 0x06, 0x83, 0x37, 0x2c, 0x65,
	0x7b, 0x15, 0x12, 0x3d, 0x87, 0x7b, 0x97, 0x4c, 0x7d, 0xdc, 0x64, 0x59, 0x22, 0xb6, 0xfb, 0x24,
	0x7c, 0xef, 0xc0, 0x71, 0x33, 0x63, 0x9d, 0x6e, 0xff, 0xb7, 0xf1, 0x3f, 0x81, 0xe3, 0x7a, 0xda,
	0xf3, 0xdb, 0x62, 0x93, 0xab, 0xc0, 0x45, 0xbd, 0x5f, 0x87, 0xcf, 0x75, 0x74, 0x27, 0xe4, 0xf9,
	0xaa, 0x14, 0xf6, 0x9a, 0x42, 0x9e, 0xaf, 0x8c, 0xf0, 0x11, 0x1c, 0x99, 0x6b, 0x50, 0xaa, 0x3c,
	0x54, 0xf5, 0x4d, 0x0c, 0x25, 0xd1, 0x0c, 0x6d, 0xd3, 0xdb, 0x22, 0xf7, 0xb9, 0x20, 0xf7, 0xa1,
	0x2b, 0x8a, 0xd4, 0x38, 0xe6, 0x4f, 0x1d, 0x1a, 0x17, 0x29, 0x8b, 0x31, 0x14, 0xbd, 0x83, 0xc1,
	0xee, 0x24, 0x6d, 0xff, 0x03, 0xe8, 0xea, 0x3c, 0x3c, 0xa3, 0x3f, 0x75, 0xa8, 0xa6, 0x66, 0x07,
	0x31, 0x06, 0xc9, 0x08, 0x1c, 0x26, 0x44, 0x21, 0xf0, 0x24, 0x6f, 0x76, 0x10, 0x1b, 0x78, 0xd6,
	0x03, 0x47, 0xe8, 0xec, 0xe8, 0x12, 0x06, 0x17, 0xd8, 0x4d, 0xa3, 0xac, 0xca, 0x87, 0xaa, 0xac,
	0x0a, 0x37, 0xb8, 0x45, 0x39, 0xcc, 0x1a, 0x47, 0xef, 0xe1, 0xe4, 0x43, 0x7e, 0xad, 0x5b, 0x36,
	0x7b, 0xb0, 0xd8, 0xaf, 0x4f, 0x0f, 0x5d, 0x9a, 0x73, 0x73, 0xa0, 0x1d, 0xf7, 0x10, 0x5f, 0x2d,
	0x9e, 0x46, 0xd0, 0xd5, 0x5d, 0x93, 0x23, 0xf0, 0x2e, 0xca, 0xff, 0x0f, 0x0f, 0x1a, 0x68, 0x31,
	0xb4, 0xa6, 0x3f, 0x3a, 0xe0, 0xa0, 0x13, 0xe4, 0x31, 0x80, 0xf9, 0x2d, 0x3e, 0x48, 0x3e, 0x6d,
	0xed, 0xe2, 0xd8, 0xa5, 0x6f, 0xb3, 0xb5, 0xda, 0x6a, 0x95, 0xd9, 0x8d, 0x52, 0xd5, 0x5a, 0x94,
	0x5a, 0xf5, 0x0a, 0x48, 0xe9, 0xf0, 0x55, 0xbe, 0x2c, 0xca, 0xab, 0x4e, 0x08, 0xfd, 0x6d, 0x53,
	0xc6, 0x43, 0xfa, 0xeb, 0x2e, 0x50, 0xf0, 0xaa, 0xe9, 0x10, 0x64, 0x9b, 0x23, 0x1f, 0xfb, 0xb4,
	0x35, 0xba, 0x53, 0x8b, 0x84, 0xe0, 0x9a, 0x7e, 0x88, 0x4f, 0x5b, 0x73, 0xa8, 0xab, 0x89, 0xc0,
	0xbb, 0xc9, 0x97, 0x7f, 0xd7, 0x9c, 0x82, 0xdf, 0xf6, 0x9e, 0x8c, 0xe8, 0x1f, 0x87, 0x51, 0x65,
	0x9c, 0xc1, 0x27, 0x8f, 0xbe, 0xc6, 0x67, 0x5c, 0x7e, 0x76, 0xf1, 0xfb, 0xe2, 0x67, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x4c, 0xd1, 0x63, 0xe1, 0xea, 0x05, 0x00, 0x00,
}