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

func getUser(ctx context.Context, username string) (*pb.User, error) {
	message, err := rdb.Get(ctx, username).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, status.Errorf(codes.NotFound, "user with username: '%s' doesn't exist", username)
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	var user pb.User
	if err = proto.Unmarshal([]byte(message), &user); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &user, nil
}

func setUserIfExists(ctx context.Context, user *pb.User) error {
	message, err := proto.Marshal(user)
	if err != nil {
		return status.Errorf(codes.Internal, "%s", err)
	}

	exists, err := rdb.SetXX(ctx, user.Username, message, 0).Result()
	if err != nil {
		return status.Errorf(codes.Internal, "%s", err)
	}
	if !exists {
		return status.Errorf(codes.AlreadyExists, "user with username: '%s' already exists", user.Username)
	}

	return nil
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func find(s []string, r string) bool {
	for _, v := range s {
		if v == r {
			return true
		}
	}
	return false
}

type UsersServerImpl struct {
	pb.UnimplementedUsersServer
}

func (*UsersServerImpl) CreateUser(ctx context.Context, in *pb.CreateRequest) (*pb.EmptyReply, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
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

	notExists, err := rdb.SetNX(ctx, in.Username, userProto, 0).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	if !notExists {
		return nil, status.Errorf(codes.AlreadyExists, "user with username: '%s' already exists", in.Username)
	}
	return &pb.EmptyReply{}, nil
}

// TODO: Add delete for tweets
func (*UsersServerImpl) DeleteUser(ctx context.Context, in *pb.DeleteRequest) (*pb.EmptyReply, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
	}

	user, err := getUser(ctx, in.Username)
	if err != nil {
		return nil, err
	}

	followers := user.GetFollowers()
	following := user.GetFollowing()

	// FIXME: here data may stay inconsistent and it is not very highload friendly
	for _, follower := range followers {
		if user, err = getUser(ctx, follower); err != nil {
			return nil, err
		}
		user.Following = remove(user.Following, in.Username)
		if err = setUserIfExists(ctx, user); err != nil {
			return nil, err
		}
	}

	for _, followed := range following {
		if user, err = getUser(ctx, followed); err != nil {
			return nil, err
		}
		user.Followers = remove(user.Followers, in.Username)
		if err = setUserIfExists(ctx, user); err != nil {
			return nil, err
		}
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

	user, err := getUser(ctx, in.Username)
	if err != nil {
		return nil, err
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
	if in.GetUsername() == "" {
		return status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
	}

	var userProto string
	var err error
	ctx := context.TODO()

	user, err := getUser(ctx, in.Username)
	if err != nil {
		return err
	}

	var usernames []string
	switch in.GetRole() {
	case pb.Role_Follower:
		usernames = user.GetFollowers()
	case pb.Role_Followed:
		usernames = user.GetFollowing()
	}

	var reply pb.GetUsersReply
	//FIXME: this is not very highload friendly
	for _, username := range usernames {
		userProto, err = rdb.Get(ctx, username).Result()
		if err != nil {
			if err == redis.Nil {
				log.Printf("user with username: '%s' doesn't exist\n", username)
			} else {
				log.Printf("%s\n", err)
			}
			reply.Reply = &pb.GetUsersReply_Error{Error: true}
		} else {
			if err = proto.Unmarshal([]byte(userProto), user); err != nil {
				reply.Reply = &pb.GetUsersReply_Error{Error: true}
				log.Printf("%s\n", err)
			} else {
				reply.Reply = &pb.GetUsersReply_User{User: user}
			}
		}

		if err = stream.Send(&reply); err != nil {
			return err
		}
	}

	return nil
}

func (*UsersServerImpl) Follow(ctx context.Context, in *pb.FollowRequest) (*pb.EmptyReply, error) {
	if in.GetFollower() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'follower' field can't be empty")
	}
	if in.GetFollowed() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'followed' field can't be empty")
	}

	follower, err := getUser(ctx, in.Follower)
	if err != nil {
		return nil, err
	}

	followed, err := getUser(ctx, in.Followed)
	if err != nil {
		return nil, err
	}

	switch in.GetAction() {
	case pb.FollowRequest_Follow:
		if !find(follower.Following, in.Followed) || !find(followed.Followers, in.Follower) {
			follower.Following = append(follower.Following, in.Followed)
			followed.Followers = append(followed.Followers, in.Follower)
		} else {
			return nil, status.Errorf(codes.AlreadyExists, "'%s' already follows '%s'", in.Follower, in.Followed)
		}
	case pb.FollowRequest_Unfollow:
		follower.Following = remove(follower.Following, in.Followed)
		followed.Followers = remove(followed.Followers, in.Follower)
	}

	if err := setUserIfExists(ctx, follower); err != nil {
		return nil, err
	}

	if err := setUserIfExists(ctx, followed); err != nil {
		return nil, err
	}

	return &pb.EmptyReply{}, nil
}

func (*UsersServerImpl) UpdateTweets(ctx context.Context, in *pb.UpdateTweetsRequest) (*pb.GenericUpdateReply, error) {
	return &pb.GenericUpdateReply{}, nil
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
