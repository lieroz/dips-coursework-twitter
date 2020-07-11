package main

import (
	"context"
	"log"
	"net"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"

	pb "github.com/lieroz/dips-coursework-twitter/users_service/protos"
)

var (
	ctx = context.Background()
	rdb *redis.ClusterClient
)

const (
	port = ":8001"
)

type UsersServerImpl struct {
	pb.UnimplementedUsersServer
}

func (s *UsersServerImpl) CreateUser(ctx context.Context, in *pb.CreateRequest) (*pb.EmptyReply, error) {
	return &pb.EmptyReply{}, nil
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
