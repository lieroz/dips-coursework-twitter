package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	pb "github.com/lieroz/dips-coursework-twitter/users_service/protos"
)

var (
	rdb *redis.ClusterClient
)

const (
	port = ":8001"
)

type UsersServerImpl struct {
	pb.UnimplementedUsersServer
}

func (*UsersServerImpl) CreateUser(ctx context.Context, in *pb.CreateRequest) (*pb.EmptyReply, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username field can't be empty")
	}

	user := pb.User{
		Username:              in.Username,
		Firstname:             in.Firstname,
		Lastname:              in.Lastname,
		Description:           in.Description,
		RegistrationTimestamp: time.Now().Unix(),
	}

	userProto, err := proto.Marshal(&user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	notExist, err := rdb.SetNX(ctx, in.Username, userProto, 0).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	if !notExist {
		return nil, status.Errorf(codes.AlreadyExists, "user with username: '%s' already exists", in.Username)
	}
	return &pb.EmptyReply{}, nil
}

func (*UsersServerImpl) DeleteUser(ctx context.Context, in *pb.DeleteRequest) (*pb.EmptyReply, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username field can't be empty")
	}
	if err := rdb.Del(ctx, in.Username).Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	return &pb.EmptyReply{}, nil
}

func (*UsersServerImpl) GetUserInfoSummary(ctx context.Context, in *pb.GetSummaryRequest) (*pb.GetSummaryReply, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username field can't be empty")
	}

	userProto, err := rdb.Get(ctx, in.Username).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, status.Errorf(codes.NotFound, "user with username: '%s' doesn't exist", in.Username)
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	var user pb.User
	if err := proto.Unmarshal([]byte(userProto), &user); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.GetSummaryReply{
		Username:              user.Username,
		Firstname:             user.Firstname,
		Lastname:              user.Lastname,
		Description:           user.Description,
		RegistrationTimestamp: user.RegistrationTimestamp,
		FollowersCount:        int64(len(user.Followers)),
		FollowingCount:        int64(len(user.Following)),
		TweetsCount:           int64(len(user.Tweets)),
	}, nil
}

func (*UsersServerImpl) GetUsers(in *pb.GetUsersRequest, stream pb.Users_GetUsersServer) error {
	return status.Errorf(codes.Unimplemented, "method GetUsers not implemented")
}

func (*UsersServerImpl) UpdateFollowers(ctx context.Context, in *pb.UpdateFollowersRequest) (*pb.GenericUpdateReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateFollowers not implemented")
}

func (*UsersServerImpl) UpdateFollowing(ctx context.Context, in *pb.UpdateFollowedRequest) (*pb.GenericUpdateReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateFollowing not implemented")
}

func (*UsersServerImpl) UpdateTweets(ctx context.Context, in *pb.UpdateTweetsRequest) (*pb.GenericUpdateReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTweets not implemented")
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Start listening on: %s", port)
	rdb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{":7000", ":7001", ":7002", ":7003", ":7004", ":7005"},
	})

	s := grpc.NewServer()
	pb.RegisterUsersServer(s, &UsersServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
