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

// TweetsClient is the client API for Tweets service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TweetsClient interface {
	CreateTweet(ctx context.Context, in *CreateTweetRequest, opts ...grpc.CallOption) (*Empty, error)
	GetTweets(ctx context.Context, in *GetTweetsRequest, opts ...grpc.CallOption) (Tweets_GetTweetsClient, error)
	DeleteTweets(ctx context.Context, in *DeleteTweetsRequest, opts ...grpc.CallOption) (*Empty, error)
	// private methods
	GetOrderedTimeline(ctx context.Context, opts ...grpc.CallOption) (Tweets_GetOrderedTimelineClient, error)
}

type tweetsClient struct {
	cc grpc.ClientConnInterface
}

func NewTweetsClient(cc grpc.ClientConnInterface) TweetsClient {
	return &tweetsClient{cc}
}

func (c *tweetsClient) CreateTweet(ctx context.Context, in *CreateTweetRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/Tweets/CreateTweet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tweetsClient) GetTweets(ctx context.Context, in *GetTweetsRequest, opts ...grpc.CallOption) (Tweets_GetTweetsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Tweets_serviceDesc.Streams[0], "/Tweets/GetTweets", opts...)
	if err != nil {
		return nil, err
	}
	x := &tweetsGetTweetsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Tweets_GetTweetsClient interface {
	Recv() (*GetTweetsReply, error)
	grpc.ClientStream
}

type tweetsGetTweetsClient struct {
	grpc.ClientStream
}

func (x *tweetsGetTweetsClient) Recv() (*GetTweetsReply, error) {
	m := new(GetTweetsReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *tweetsClient) DeleteTweets(ctx context.Context, in *DeleteTweetsRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/Tweets/DeleteTweets", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tweetsClient) GetOrderedTimeline(ctx context.Context, opts ...grpc.CallOption) (Tweets_GetOrderedTimelineClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Tweets_serviceDesc.Streams[1], "/Tweets/GetOrderedTimeline", opts...)
	if err != nil {
		return nil, err
	}
	x := &tweetsGetOrderedTimelineClient{stream}
	return x, nil
}

type Tweets_GetOrderedTimelineClient interface {
	Send(*Timeline) error
	Recv() (*Timeline, error)
	grpc.ClientStream
}

type tweetsGetOrderedTimelineClient struct {
	grpc.ClientStream
}

func (x *tweetsGetOrderedTimelineClient) Send(m *Timeline) error {
	return x.ClientStream.SendMsg(m)
}

func (x *tweetsGetOrderedTimelineClient) Recv() (*Timeline, error) {
	m := new(Timeline)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TweetsServer is the server API for Tweets service.
// All implementations must embed UnimplementedTweetsServer
// for forward compatibility
type TweetsServer interface {
	CreateTweet(context.Context, *CreateTweetRequest) (*Empty, error)
	GetTweets(*GetTweetsRequest, Tweets_GetTweetsServer) error
	DeleteTweets(context.Context, *DeleteTweetsRequest) (*Empty, error)
	// private methods
	GetOrderedTimeline(Tweets_GetOrderedTimelineServer) error
	mustEmbedUnimplementedTweetsServer()
}

// UnimplementedTweetsServer must be embedded to have forward compatible implementations.
type UnimplementedTweetsServer struct {
}

func (*UnimplementedTweetsServer) CreateTweet(context.Context, *CreateTweetRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTweet not implemented")
}
func (*UnimplementedTweetsServer) GetTweets(*GetTweetsRequest, Tweets_GetTweetsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetTweets not implemented")
}
func (*UnimplementedTweetsServer) DeleteTweets(context.Context, *DeleteTweetsRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTweets not implemented")
}
func (*UnimplementedTweetsServer) GetOrderedTimeline(Tweets_GetOrderedTimelineServer) error {
	return status.Errorf(codes.Unimplemented, "method GetOrderedTimeline not implemented")
}
func (*UnimplementedTweetsServer) mustEmbedUnimplementedTweetsServer() {}

func RegisterTweetsServer(s *grpc.Server, srv TweetsServer) {
	s.RegisterService(&_Tweets_serviceDesc, srv)
}

func _Tweets_CreateTweet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateTweetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TweetsServer).CreateTweet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Tweets/CreateTweet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TweetsServer).CreateTweet(ctx, req.(*CreateTweetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tweets_GetTweets_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetTweetsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TweetsServer).GetTweets(m, &tweetsGetTweetsServer{stream})
}

type Tweets_GetTweetsServer interface {
	Send(*GetTweetsReply) error
	grpc.ServerStream
}

type tweetsGetTweetsServer struct {
	grpc.ServerStream
}

func (x *tweetsGetTweetsServer) Send(m *GetTweetsReply) error {
	return x.ServerStream.SendMsg(m)
}

func _Tweets_DeleteTweets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteTweetsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TweetsServer).DeleteTweets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Tweets/DeleteTweets",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TweetsServer).DeleteTweets(ctx, req.(*DeleteTweetsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tweets_GetOrderedTimeline_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TweetsServer).GetOrderedTimeline(&tweetsGetOrderedTimelineServer{stream})
}

type Tweets_GetOrderedTimelineServer interface {
	Send(*Timeline) error
	Recv() (*Timeline, error)
	grpc.ServerStream
}

type tweetsGetOrderedTimelineServer struct {
	grpc.ServerStream
}

func (x *tweetsGetOrderedTimelineServer) Send(m *Timeline) error {
	return x.ServerStream.SendMsg(m)
}

func (x *tweetsGetOrderedTimelineServer) Recv() (*Timeline, error) {
	m := new(Timeline)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Tweets_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Tweets",
	HandlerType: (*TweetsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateTweet",
			Handler:    _Tweets_CreateTweet_Handler,
		},
		{
			MethodName: "DeleteTweets",
			Handler:    _Tweets_DeleteTweets_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetTweets",
			Handler:       _Tweets_GetTweets_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetOrderedTimeline",
			Handler:       _Tweets_GetOrderedTimeline_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "tweets.proto",
}
