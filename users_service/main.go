package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	pb "github.com/lieroz/dips-coursework-twitter/protos"
)

var (
	pool *pgxpool.Pool
)

const (
	port = ":8001"
)

func difference(a, b []int) (diff []int) {
	m := make(map[int]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

type UsersServerImpl struct {
	pb.UnimplementedUsersServer
}

func (*UsersServerImpl) CreateUser(ctx context.Context, in *pb.CreateRequest) (*pb.Empty, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
	}

	cmdTag, err := pool.Exec(ctx,
		"insert into users (username, firstname, lastname, description) values ($1, $2, $3, $4) on conflict do nothing",
		in.GetUsername(),
		in.GetFirstname(),
		in.GetLastname(),
		in.GetDescription(),
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return nil, status.Errorf(codes.AlreadyExists, "user with username: '%s' already exists", in.Username)
	}

	return &pb.Empty{}, nil
}

// TODO: add tweets deletion from tweets table and followers timelines
func (*UsersServerImpl) DeleteUser(ctx context.Context, in *pb.DeleteRequest) (*pb.Empty, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	defer tx.Rollback(ctx)

	var followers, following []string
	var tweets []int

	if err = tx.QueryRow(ctx,
		"delete from users where username = $1 returning followers, following, tweets",
		in.GetUsername()).Scan(
		&followers,
		&following,
		&tweets,
	); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	query := func(field, users string) string {
		return fmt.Sprintf("update users set %[1]v = array_remove(%[1]v , $1) where username in ('%[2]v')",
			field, users)
	}

	if _, err = tx.Exec(ctx, query("following", strings.Join(followers[:], "', '")),
		in.GetUsername()); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if _, err = tx.Exec(ctx, query("followers", strings.Join(following[:], "', '")),
		in.GetUsername()); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) GetUserInfoSummary(ctx context.Context, in *pb.GetSummaryRequest) (*pb.GetSummaryReply, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
	}

	var timestamp time.Time

	summary := &pb.GetSummaryReply{}
	err := pool.QueryRow(ctx, `select username, firstname, lastname, description,
		registration_timestamp, cardinality(followers), cardinality(following), cardinality(tweets)
		from users where username = $1`, in.GetUsername()).Scan(
		&summary.Username,
		&summary.Firstname,
		&summary.Lastname,
		&summary.Description,
		&timestamp,
		&summary.FollowersCount,
		&summary.FollowingCount,
		&summary.TweetsCount,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound,
				"user with username: '%s' doesn't exist", in.GetUsername())
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	summary.RegistrationTimestamp = timestamp.Unix()
	return summary, nil
}

func (*UsersServerImpl) GetUsers(in *pb.GetUsersRequest, stream pb.Users_GetUsersServer) error {
	if in.GetUsername() == "" {
		return status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
	}

	query := "select %s from users where username = $1"
	switch in.GetRole() {
	case pb.Role_Follower:
		query = fmt.Sprintf(query, "followers")
	case pb.Role_Followed:
		query = fmt.Sprintf(query, "following")
	}

	var users []string
	if err := pool.QueryRow(context.Background(), query,
		in.GetUsername()).Scan(&users); err != nil {
		return status.Errorf(codes.Internal, "%s", err)
	}

	query = fmt.Sprintf(`select username, firstname, lastname, description, 
		registration_timestamp, followers, following, tweets 
		from users where username in ('%s')`, strings.Join(users[:], "', '"))
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return status.Errorf(codes.Internal, "%s", err)
	}

	user := pb.User{}
	var timestamp time.Time
	var reply pb.GetUsersReply

	for rows.Next() {
		if err = rows.Scan(
			&user.Username,
			&user.Firstname,
			&user.Lastname,
			&user.Description,
			&timestamp,
			&user.Followers,
			&user.Following,
			&user.Tweets); err != nil {
			reply.Reply = &pb.GetUsersReply_Error{Error: true}
			log.Println(err)
		} else {
			user.RegistrationTimestamp = timestamp.Unix()
			reply.Reply = &pb.GetUsersReply_User{User: &user}
			if err = stream.Send(&reply); err != nil {
				return status.Errorf(codes.Internal, "%s", err)
			}
		}
	}

	if rows.Err() != nil {
		return status.Errorf(codes.Internal, "%s", rows.Err())
	}

	return nil
}

func (*UsersServerImpl) Follow(ctx context.Context, in *pb.FollowRequest) (*pb.Empty, error) {
	if in.GetFollower() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'follower' field can't be empty")
	}
	if in.GetFollowed() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'followed' field can't be empty")
	}
	if in.GetFollower() == in.GetFollowed() {
		return nil, status.Errorf(codes.InvalidArgument, "user can't follow himself")
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	defer tx.Rollback(ctx)

	query := func(v string) string {
		return fmt.Sprintf(`update users set %[1]v = array_append(%[1]v, $1::varchar) where username = $2 
			and exists (select 1 from users where username = $2 and not ($1 = any(%[1]v)))`, v)
	}

	var timeline []int
	if err = tx.QueryRow(ctx, fmt.Sprintf(query("following"))+" returning timeline",
		in.GetFollowed(), in.GetFollower()).Scan(&timeline); err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound,
				"user '%s' is already following '%s' or doesn't exist",
				in.GetFollower(), in.GetFollowed())
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	var tweets []int
	if err = tx.QueryRow(ctx, fmt.Sprintf(query("followers"))+" returning tweets",
		in.GetFollower(), in.GetFollowed()).Scan(&tweets); err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound,
				"user '%s' is already followed by '%s' or doesn't exist",
				in.GetFollowed(), in.GetFollower())
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	timeline = append(timeline, tweets...)

	// FIXME: remove select from tweets, cause it depends on foreign data
	if len(timeline) != 0 {
		if _, err = tx.Exec(ctx, fmt.Sprintf(`update users set timeline = array(select id from tweets
		where id in (%s) group by id order by creation_timestamp) where username = $1`,
			strings.Trim(strings.Replace(fmt.Sprint(timeline), " ", ", ", -1), "[]")),
			in.GetFollower()); err != nil {
			return nil, status.Errorf(codes.Internal, "%s", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) Unfollow(ctx context.Context, in *pb.FollowRequest) (*pb.Empty, error) {
	if in.GetFollower() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'follower' field can't be empty")
	}
	if in.GetFollowed() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'followed' field can't be empty")
	}
	if in.GetFollower() == in.GetFollowed() {
		return nil, status.Errorf(codes.InvalidArgument, "user can't unfollow himself")
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	defer tx.Rollback(ctx)

	query := func(v string) string {
		return fmt.Sprintf(`update users set %[1]v = array_remove(%[1]v, $1::varchar) where username = $2 
			and exists (select 1 from users where username = $2 and $1 = any(%[1]v))`, v)
	}

	var timeline []int
	if err = tx.QueryRow(ctx, fmt.Sprintf(query("following"))+" returning timeline",
		in.GetFollowed(), in.GetFollower()).Scan(&timeline); err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound,
				"user '%s' is not following '%s' or doesn't exist",
				in.GetFollower(), in.GetFollowed())
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	var tweets []int
	if err = tx.QueryRow(ctx, fmt.Sprintf(query("followers"))+" returning tweets",
		in.GetFollower(), in.GetFollowed()).Scan(&tweets); err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound,
				"user '%s' is not followed by '%s' or doesn't exist",
				in.GetFollowed(), in.GetFollower())
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	newTimeline := difference(timeline, tweets)
	if _, err = tx.Exec(ctx, fmt.Sprintf(`update users set timeline = array[%s]::integer[] where username = $1`,
		strings.Trim(strings.Replace(fmt.Sprint(newTimeline), " ", ", ", -1), "[]")),
		in.GetFollower()); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) OnTweetCreated(ctx context.Context, in *pb.OnTweetCreatedRequest) (*pb.Empty, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
	}
	if in.GetTweetId() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "'tweet_id' field can't be empty")
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	defer tx.Rollback(ctx)

	var followers []string
	if err = tx.QueryRow(ctx, "update users set tweets = array_append(tweets, $1) where username = $2 returning followers",
		in.GetTweetId(), in.GetUsername()).Scan(&followers); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if _, err = tx.Exec(ctx, fmt.Sprintf("update users set timeline = array_append(timeline, $1) where username in ('%s')",
		strings.Join(followers[:], "', '")), in.GetTweetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Start listening on: %s", port)

	pool, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUsersServer(s, &UsersServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
