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

	proto, err := proto.Marshal(&user)
	if err != nil {
		return &pb.EmptyReply{}, err
	}

	notExist, err := rdb.SetNX(ctx, in.Username, proto, 0).Result()
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
	return &pb.GetSummaryReply{}, nil
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
