// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protos/users.proto

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
	return fileDescriptor_3e973bd211a9df00, []int{0}
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
	return fileDescriptor_3e973bd211a9df00, []int{0}
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

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_3e973bd211a9df00, []int{1}
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
	return fileDescriptor_3e973bd211a9df00, []int{2}
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
	return fileDescriptor_3e973bd211a9df00, []int{3}
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
	return fileDescriptor_3e973bd211a9df00, []int{4}
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
	return fileDescriptor_3e973bd211a9df00, []int{5}
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
	Role                 Role     `protobuf:"varint,2,opt,name=role,proto3,enum=users.Role" json:"role,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetUsersRequest) Reset()         { *m = GetUsersRequest{} }
func (m *GetUsersRequest) String() string { return proto.CompactTextString(m) }
func (*GetUsersRequest) ProtoMessage()    {}
func (*GetUsersRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3e973bd211a9df00, []int{6}
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
	return fileDescriptor_3e973bd211a9df00, []int{7}
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
	return fileDescriptor_3e973bd211a9df00, []int{8}
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

func init() {
	proto.RegisterEnum("users.Role", Role_name, Role_value)
	proto.RegisterType((*User)(nil), "users.User")
	proto.RegisterType((*Empty)(nil), "users.Empty")
	proto.RegisterType((*CreateRequest)(nil), "users.CreateRequest")
	proto.RegisterType((*DeleteRequest)(nil), "users.DeleteRequest")
	proto.RegisterType((*GetSummaryRequest)(nil), "users.GetSummaryRequest")
	proto.RegisterType((*GetSummaryReply)(nil), "users.GetSummaryReply")
	proto.RegisterType((*GetUsersRequest)(nil), "users.GetUsersRequest")
	proto.RegisterType((*GetUsersReply)(nil), "users.GetUsersReply")
	proto.RegisterType((*FollowRequest)(nil), "users.FollowRequest")
}

func init() {
	proto.RegisterFile("protos/users.proto", fileDescriptor_3e973bd211a9df00)
}

var fileDescriptor_3e973bd211a9df00 = []byte{
	// 572 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xd4, 0x54, 0xcf, 0x6e, 0xd3, 0x4c,
	0x10, 0x8f, 0x93, 0x38, 0x71, 0x27, 0x49, 0xdb, 0x6f, 0xd5, 0x2f, 0xb2, 0x22, 0x24, 0x52, 0x5f,
	0x88, 0x80, 0x26, 0x55, 0x11, 0x17, 0x90, 0x10, 0x6a, 0x21, 0x2d, 0x17, 0x0e, 0x2e, 0xbd, 0x70,
	0x89, 0xdc, 0x64, 0x62, 0x56, 0xd8, 0x5e, 0xb3, 0xbb, 0x26, 0x0a, 0x2f, 0x80, 0x78, 0x20, 0x1e,
	0x83, 0x77, 0x42, 0xbb, 0x6b, 0x3b, 0x4e, 0xcb, 0x21, 0x37, 0xc4, 0xc9, 0xfa, 0xfd, 0x99, 0xd5,
	0xcf, 0xb3, 0x33, 0x0b, 0x24, 0xe5, 0x4c, 0x32, 0x31, 0xc9, 0x04, 0x72, 0x31, 0xd6, 0x80, 0xd8,
	0x1a, 0x78, 0x3f, 0xea, 0xd0, 0xbc, 0x11, 0xc8, 0xc9, 0x00, 0x1c, 0xc5, 0x24, 0x41, 0x8c, 0xae,
	0x35, 0xb4, 0x46, 0x7b, 0x7e, 0x89, 0xc9, 0x03, 0xd8, 0x5b, 0x52, 0x2e, 0xa4, 0x16, 0xeb, 0x5a,
	0xdc, 0x10, 0xaa, 0x32, 0x0a, 0x72, 0xb1, 0x61, 0x2a, 0x0b, 0x4c, 0x86, 0xd0, 0x59, 0xa0, 0x98,
	0x73, 0x9a, 0x4a, 0xca, 0x12, 0xb7, 0xa9, 0xe5, 0x2a, 0x45, 0x9e, 0x43, 0x9f, 0x63, 0x48, 0x85,
	0xe4, 0x81, 0xc2, 0x33, 0x49, 0x63, 0x14, 0x32, 0x88, 0x53, 0xd7, 0x1e, 0x5a, 0xa3, 0x86, 0xff,
	0x7f, 0x55, 0xfd, 0x50, 0x88, 0x3a, 0x12, 0x8b, 0x22, 0xb6, 0x42, 0x2e, 0xdc, 0xd6, 0xb0, 0xa1,
	0x23, 0x15, 0xc4, 0x46, 0xa5, 0x49, 0xe8, 0xb6, 0xab, 0x2a, 0x4d, 0x42, 0xd2, 0x87, 0x96, 0x5c,
	0x21, 0x4a, 0xe1, 0x3a, 0xc3, 0xc6, 0xa8, 0xe1, 0xe7, 0xc8, 0x6b, 0x83, 0xfd, 0x36, 0x4e, 0xe5,
	0xda, 0xfb, 0x6e, 0x41, 0xef, 0x82, 0x63, 0x20, 0xd1, 0xc7, 0x2f, 0x19, 0x0a, 0xf9, 0xb7, 0xba,
	0xe3, 0x3d, 0x81, 0xde, 0x1b, 0x8c, 0x70, 0xa7, 0x20, 0xde, 0x04, 0xfe, 0xbb, 0x44, 0x79, 0x9d,
	0xc5, 0x71, 0xc0, 0xd7, 0xbb, 0x14, 0xfc, 0xac, 0xc3, 0x41, 0xb5, 0x22, 0x8d, 0xd6, 0xff, 0xda,
	0x1c, 0x3c, 0x82, 0x83, 0xf2, 0xda, 0x67, 0x73, 0x96, 0x25, 0xd2, 0x6d, 0x69, 0xff, 0x7e, 0x49,
	0x5f, 0x28, 0x76, 0x63, 0xa4, 0x49, 0x98, 0x1b, 0xdb, 0x55, 0x23, 0x4d, 0x42, 0x63, 0x3c, 0x86,
	0xae, 0x99, 0x87, 0xdc, 0xe5, 0x68, 0x57, 0xc7, 0x70, 0xda, 0xe2, 0xbd, 0xd7, 0x6d, 0x53, 0x6b,
	0x23, 0x76, 0x19, 0x90, 0x87, 0xd0, 0xe4, 0x2c, 0x32, 0x1d, 0xdb, 0x3f, 0xeb, 0x8c, 0xcd, 0x1a,
	0xfa, 0x2c, 0x42, 0x5f, 0x0b, 0xde, 0x35, 0xf4, 0x36, 0xe7, 0xa9, 0x4b, 0x38, 0x86, 0xa6, 0x32,
	0xe9, 0x93, 0x3a, 0x65, 0x85, 0x32, 0x5c, 0xd5, 0x7c, 0x2d, 0x91, 0x3e, 0xd8, 0xc8, 0x39, 0xe3,
	0xfa, 0x54, 0xe7, 0xaa, 0xe6, 0x1b, 0x78, 0xde, 0x06, 0x9b, 0xab, 0x33, 0xbc, 0x4b, 0xe8, 0x4d,
	0xf5, 0x9f, 0x55, 0x22, 0x16, 0x3d, 0x29, 0x22, 0x16, 0xb8, 0xa2, 0x2d, 0xf2, 0x8b, 0x2d, 0xf1,
	0x63, 0x0f, 0x9a, 0x2a, 0x2b, 0xe9, 0x82, 0x33, 0xcd, 0xfd, 0x87, 0xb5, 0x0a, 0x5a, 0x1c, 0x5a,
	0x67, 0xbf, 0xea, 0x60, 0xeb, 0xfc, 0xe4, 0x14, 0xc0, 0xac, 0x8e, 0x7e, 0x55, 0x8e, 0xf2, 0xe8,
	0x5b, 0xdb, 0x34, 0xe8, 0xe6, 0xac, 0xde, 0x36, 0x55, 0x61, 0x66, 0x7c, 0xab, 0x62, 0x6b, 0xec,
	0xef, 0x54, 0x4c, 0x81, 0xe4, 0xfd, 0x7a, 0x97, 0x2c, 0x59, 0x3e, 0xbe, 0xc4, 0xcd, 0x3d, 0xf7,
	0x76, 0x60, 0xd0, 0xff, 0x83, 0xa2, 0xda, 0xfc, 0x02, 0x9c, 0xa2, 0xef, 0xa4, 0xe2, 0xa9, 0x5e,
	0xec, 0xe0, 0xe8, 0x1e, 0x9f, 0x46, 0xeb, 0x53, 0x8b, 0x3c, 0x85, 0x96, 0xf9, 0xff, 0x32, 0xf1,
	0x56, 0xb7, 0xef, 0x24, 0x1e, 0x83, 0x73, 0x93, 0x2c, 0x77, 0xf6, 0x9f, 0xbf, 0xfe, 0xf8, 0x2a,
	0xa4, 0xf2, 0x53, 0x76, 0x3b, 0x9e, 0xb3, 0x78, 0x12, 0x51, 0xe4, 0xec, 0xdb, 0x64, 0x41, 0x53,
	0x71, 0x32, 0x67, 0x19, 0x17, 0xb8, 0x62, 0xfc, 0xf3, 0x89, 0x5c, 0x51, 0x29, 0x91, 0x9b, 0x67,
	0x7d, 0x26, 0x90, 0x7f, 0xa5, 0x73, 0x7c, 0x69, 0xde, 0xfa, 0xdb, 0x96, 0xfe, 0x3e, 0xfb, 0x1d,
	0x00, 0x00, 0xff, 0xff, 0x2d, 0x5c, 0x7d, 0x87, 0xfc, 0x05, 0x00, 0x00,
}
