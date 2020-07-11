package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

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

func (s *UsersServerImpl) CreateUser(ctx context.Context, in *pb.CreateRequest) (*pb.EmptyReply, error) {
	if in.GetUsername() == "" {
		return &pb.EmptyReply{}, fmt.Errorf("username field can't be empty")
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
		return &pb.EmptyReply{}, err
	}

	notExist, err := rdb.SetNX(ctx, in.Username, userProto, 0).Result()
	if err != nil {
		return &pb.EmptyReply{}, err
	}
	if !notExist {
		return &pb.EmptyReply{}, fmt.Errorf("user with username: '%s' already exists", in.Username)
	}
	return &pb.EmptyReply{}, nil
}

func (s *UsersServerImpl) DeleteUser(ctx context.Context, in *pb.DeleteRequest) (*pb.EmptyReply, error) {
	if in.GetUsername() == "" {
		return &pb.EmptyReply{}, fmt.Errorf("username field can't be empty")
	}
	if err := rdb.Del(ctx, in.Username).Err(); err != nil {
		return &pb.EmptyReply{}, err
	}
	return &pb.EmptyReply{}, nil
}

func (s *UsersServerImpl) GetUserInfoSummary(ctx context.Context, in *pb.GetSummaryRequest) (*pb.GetSummaryReply, error) {
	if in.GetUsername() == "" {
		return &pb.GetSummaryReply{}, fmt.Errorf("username field can't be empty")
	}

	userProto, err := rdb.Get(ctx, in.Username).Result()
	if err != nil {
		if err == redis.Nil {
			return &pb.GetSummaryReply{}, fmt.Errorf("user with username: '%s' doesn't exist", in.Username)
		}
		return &pb.GetSummaryReply{}, err
	}

	var user pb.User
	if err := proto.Unmarshal([]byte(userProto), &user); err != nil {
		return &pb.GetSummaryReply{}, err
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
