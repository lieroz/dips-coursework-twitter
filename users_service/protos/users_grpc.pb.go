// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package protos

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// UsersClient is the client API for Users service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UsersClient interface {
	CreateUser(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (*EmptyReply, error)
	DeleteUser(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*EmptyReply, error)
	GetUserInfoSummary(ctx context.Context, in *GetSummaryRequest, opts ...grpc.CallOption) (*GetSummaryReply, error)
	GetUsers(ctx context.Context, in *GetUsersRequest, opts ...grpc.CallOption) (Users_GetUsersClient, error)
	Follow(ctx context.Context, in *FollowRequest, opts ...grpc.CallOption) (*EmptyReply, error)
	UpdateTweets(ctx context.Context, in *UpdateTweetsRequest, opts ...grpc.CallOption) (*EmptyReply, error)
}

type usersClient struct {
	cc grpc.ClientConnInterface
}

func NewUsersClient(cc grpc.ClientConnInterface) UsersClient {
	return &usersClient{cc}
}

func (c *usersClient) CreateUser(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (*EmptyReply, error) {
	out := new(EmptyReply)
	err := c.cc.Invoke(ctx, "/users.Users/CreateUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) DeleteUser(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*EmptyReply, error) {
	out := new(EmptyReply)
	err := c.cc.Invoke(ctx, "/users.Users/DeleteUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) GetUserInfoSummary(ctx context.Context, in *GetSummaryRequest, opts ...grpc.CallOption) (*GetSummaryReply, error) {
	out := new(GetSummaryReply)
	err := c.cc.Invoke(ctx, "/users.Users/GetUserInfoSummary", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) GetUsers(ctx context.Context, in *GetUsersRequest, opts ...grpc.CallOption) (Users_GetUsersClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Users_serviceDesc.Streams[0], "/users.Users/GetUsers", opts...)
	if err != nil {
		return nil, err
	}
	x := &usersGetUsersClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Users_GetUsersClient interface {
	Recv() (*GetUsersReply, error)
	grpc.ClientStream
}

type usersGetUsersClient struct {
	grpc.ClientStream
}

func (x *usersGetUsersClient) Recv() (*GetUsersReply, error) {
	m := new(GetUsersReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *usersClient) Follow(ctx context.Context, in *FollowRequest, opts ...grpc.CallOption) (*EmptyReply, error) {
	out := new(EmptyReply)
	err := c.cc.Invoke(ctx, "/users.Users/Follow", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersClient) UpdateTweets(ctx context.Context, in *UpdateTweetsRequest, opts ...grpc.CallOption) (*EmptyReply, error) {
	out := new(EmptyReply)
	err := c.cc.Invoke(ctx, "/users.Users/UpdateTweets", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UsersServer is the server API for Users service.
// All implementations must embed UnimplementedUsersServer
// for forward compatibility
type UsersServer interface {
	CreateUser(context.Context, *CreateRequest) (*EmptyReply, error)
	DeleteUser(context.Context, *DeleteRequest) (*EmptyReply, error)
	GetUserInfoSummary(context.Context, *GetSummaryRequest) (*GetSummaryReply, error)
	GetUsers(*GetUsersRequest, Users_GetUsersServer) error
	Follow(context.Context, *FollowRequest) (*EmptyReply, error)
	UpdateTweets(context.Context, *UpdateTweetsRequest) (*EmptyReply, error)
	mustEmbedUnimplementedUsersServer()
}

// UnimplementedUsersServer must be embedded to have forward compatible implementations.
type UnimplementedUsersServer struct {
}

func (*UnimplementedUsersServer) CreateUser(context.Context, *CreateRequest) (*EmptyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}
func (*UnimplementedUsersServer) DeleteUser(context.Context, *DeleteRequest) (*EmptyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (*UnimplementedUsersServer) GetUserInfoSummary(context.Context, *GetSummaryRequest) (*GetSummaryReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserInfoSummary not implemented")
}
func (*UnimplementedUsersServer) GetUsers(*GetUsersRequest, Users_GetUsersServer) error {
	return status.Errorf(codes.Unimplemented, "method GetUsers not implemented")
}
func (*UnimplementedUsersServer) Follow(context.Context, *FollowRequest) (*EmptyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Follow not implemented")
}
func (*UnimplementedUsersServer) UpdateTweets(context.Context, *UpdateTweetsRequest) (*EmptyReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTweets not implemented")
}
func (*UnimplementedUsersServer) mustEmbedUnimplementedUsersServer() {}

func RegisterUsersServer(s *grpc.Server, srv UsersServer) {
	s.RegisterService(&_Users_serviceDesc, srv)
}

func _Users_CreateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).CreateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/users.Users/CreateUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).CreateUser(ctx, req.(*CreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/users.Users/DeleteUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).DeleteUser(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_GetUserInfoSummary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSummaryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).GetUserInfoSummary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/users.Users/GetUserInfoSummary",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).GetUserInfoSummary(ctx, req.(*GetSummaryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_GetUsers_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetUsersRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UsersServer).GetUsers(m, &usersGetUsersServer{stream})
}

type Users_GetUsersServer interface {
	Send(*GetUsersReply) error
	grpc.ServerStream
}

type usersGetUsersServer struct {
	grpc.ServerStream
}

func (x *usersGetUsersServer) Send(m *GetUsersReply) error {
	return x.ServerStream.SendMsg(m)
}

func _Users_Follow_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FollowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).Follow(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/users.Users/Follow",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).Follow(ctx, req.(*FollowRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Users_UpdateTweets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateTweetsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersServer).UpdateTweets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/users.Users/UpdateTweets",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersServer).UpdateTweets(ctx, req.(*UpdateTweetsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Users_serviceDesc = grpc.ServiceDesc{
	ServiceName: "users.Users",
	HandlerType: (*UsersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateUser",
			Handler:    _Users_CreateUser_Handler,
		},
		{
			MethodName: "DeleteUser",
			Handler:    _Users_DeleteUser_Handler,
		},
		{
			MethodName: "GetUserInfoSummary",
			Handler:    _Users_GetUserInfoSummary_Handler,
		},
		{
			MethodName: "Follow",
			Handler:    _Users_Follow_Handler,
		},
		{
			MethodName: "UpdateTweets",
			Handler:    _Users_UpdateTweets_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetUsers",
			Handler:       _Users_GetUsers_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "users.proto",
}
