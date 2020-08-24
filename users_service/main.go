package main

import (
	"bytes"
	"context"
	"fmt"
	// "io"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	nats "github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	pb "github.com/lieroz/dips-coursework-twitter/protos"
	"github.com/lieroz/dips-coursework-twitter/tools"
)

var (
	pool *pgxpool.Pool
	nc   *nats.Conn
	rdb  *redis.Client
)

const (
	port = ":8001"
)

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

func (*UsersServerImpl) DeleteUser(ctx context.Context, in *pb.DeleteRequest) (*pb.Empty, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'username' field can't be empty")
	}

	var followers, following []string
	var tweets []int64

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer psqlCtxCancel()

	if err := pool.QueryRow(psqlCtx,
		"select followers, following, tweets from users where username = $1",
		in.GetUsername()).Scan(
		&followers,
		&following,
		&tweets,
	); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	cmdProto := &pb.NatsDeleteUserMessage{Username: in.Username, Tweets: tweets}
	serializedCmd, err := proto.Marshal(cmdProto)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	msgProto := &pb.NatsMessage{Command: pb.NatsMessage_DeleteUser, Message: serializedCmd}
	serializedMsg, err := proto.Marshal(msgProto)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	rdbCtx, rdbCtxCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer rdbCtxCancel()

	key := fmt.Sprintf("%s:delete", in.Username)
	pipe := rdb.Pipeline()
	pipe.HSet(rdbCtx, key, "followers", followers, "following", following)
	pipe.Expire(rdbCtx, key, 10*time.Second)

	if _, err = pipe.Exec(rdbCtx); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if err := nc.Publish("tweets", serializedMsg); err != nil {
		// key in redis will be cleaned up after expiration time
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
	if err := pool.QueryRow(ctx, `select username, firstname, lastname, description,
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
	); err != nil {
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
		in.Username).Scan(&users); err != nil {
		return status.Errorf(codes.Internal, "%s", err)
	}

	query = fmt.Sprintf(`select username, firstname, lastname, description, 
		registration_timestamp, followers, following, tweets 
		from users where username in ('%s')`, strings.Join(users[:], "', '"))
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return status.Errorf(codes.Internal, "%s", err)
	}
	defer rows.Close()

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
	// if in.GetFollower() == "" {
	// 	return nil, status.Errorf(codes.InvalidArgument, "'follower' field can't be empty")
	// }
	// if in.GetFollowed() == "" {
	// 	return nil, status.Errorf(codes.InvalidArgument, "'followed' field can't be empty")
	// }
	// if in.GetFollower() == in.GetFollowed() {
	// 	return nil, status.Errorf(codes.InvalidArgument, "user can't follow himself")
	// }

	// tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	// if err != nil {
	// 	return nil, status.Errorf(codes.Internal, "%s", err)
	// }
	// defer tx.Rollback(ctx)

	// query := func(v string) string {
	// 	return fmt.Sprintf(`update users set %[1]v = array_append(%[1]v, $1::varchar) where username = $2
	// 		and exists (select 1 from users where username = $2 and not ($1 = any(%[1]v)))`, v)
	// }

	// var timeline []int64
	// if err = tx.QueryRow(ctx, fmt.Sprintf(query("following"))+" returning timeline",
	// 	in.GetFollowed(), in.GetFollower()).Scan(&timeline); err != nil {
	// 	if err == pgx.ErrNoRows {
	// 		return nil, status.Errorf(codes.NotFound,
	// 			"user '%s' is already following '%s' or doesn't exist",
	// 			in.GetFollower(), in.GetFollowed())
	// 	}
	// 	return nil, status.Errorf(codes.Internal, "%s", err)
	// }

	// var tweets []int64
	// if err = tx.QueryRow(ctx, fmt.Sprintf(query("followers"))+" returning tweets",
	// 	in.GetFollower(), in.GetFollowed()).Scan(&tweets); err != nil {
	// 	if err == pgx.ErrNoRows {
	// 		return nil, status.Errorf(codes.NotFound,
	// 			"user '%s' is already followed by '%s' or doesn't exist",
	// 			in.GetFollowed(), in.GetFollower())
	// 	}
	// 	return nil, status.Errorf(codes.Internal, "%s", err)
	// }

	// timeline = append(timeline, tweets...)

	// if len(timeline) != 0 {
	// 	// read below
	// 	stream, err := tweetsClient.GetOrderedTimeline(ctx)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	if err = stream.Send(&pb.Timeline{Timeline: timeline}); err != nil {
	// 		return nil, err
	// 	} else {
	// 		stream.CloseSend()
	// 		r, err := stream.Recv()
	// 		if err != nil && err != io.EOF {
	// 			return nil, err
	// 		}

	// 		if _, err = tx.Exec(ctx, fmt.Sprintf("update users set timeline = array[%s]::integer[] where username = $1",
	// 			tools.intArrayToString(r.GetTimeline())),
	// 			in.Follower); err != nil {
	// 			return nil, status.Errorf(codes.Internal, "%s", err)
	// 		}
	// 	}
	// }

	// // FIXME: here commit may fail only if smth is wrong with database itself
	// // or machine which runs it, so better send tx abort to tweets service to rollback
	// // but this is a prototype service for coursework, so just ignore it
	// if err = tx.Commit(ctx); err != nil {
	// 	return nil, status.Errorf(codes.Internal, "%s", err)
	// }

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

	var timeline []int64
	if err = tx.QueryRow(ctx, fmt.Sprintf(query("following"))+" returning timeline",
		in.GetFollowed(), in.GetFollower()).Scan(&timeline); err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound,
				"user '%s' is not following '%s' or doesn't exist",
				in.GetFollower(), in.GetFollowed())
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	var tweets []int64
	if err = tx.QueryRow(ctx, fmt.Sprintf(query("followers"))+" returning tweets",
		in.GetFollower(), in.GetFollowed()).Scan(&tweets); err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound,
				"user '%s' is not followed by '%s' or doesn't exist",
				in.GetFollowed(), in.GetFollower())
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	newTimeline := tools.Difference(timeline, tweets)
	if _, err = tx.Exec(ctx, fmt.Sprintf("update users set timeline = array[%s]::integer[] where username = $1",
		tools.IntArrayToString(newTimeline)),
		in.GetFollower()); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.Empty{}, nil
}

func handleDeleteUser(ctx context.Context, serializedMsg []byte) {
	msgProto := &pb.NatsDeleteUserMessage{}
	if err := proto.Unmarshal(serializedMsg, msgProto); err != nil {
		//TODO: add logging
		return
	}

	if msgProto.GetUsername() == "" {
		//TODO: add logging
		return
	}
	if msgProto.GetTweets() == nil {
		//TODO: add logging
		return
	}

	rdbCtx, rdbCtxCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer rdbCtxCancel()

	key := fmt.Sprintf("%s:delete", msgProto.Username)
	pipe := rdb.Pipeline()
	hmgetRes := pipe.HMGet(rdbCtx, key, "followers", "following")
	pipe.Del(rdbCtx, key)

	if _, err := pipe.Exec(rdbCtx); err != nil {
		//TODO: add logging
		return
	}

	vals := hmgetRes.Val()
	followers := tools.InterfaceToStringArray(vals[0])
	following := tools.InterfaceToStringArray(vals[1])

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer psqlCtxCancel()

	tx, err := pool.BeginTx(psqlCtx, pgx.TxOptions{})
	if err != nil {
		//TODO: add logging
		return
	}
	defer tx.Rollback(psqlCtx)

	query := func(field, users string) string {
		return fmt.Sprintf("update users set %[1]v = array_remove(%[1]v , $1) where username in ('%[2]v')",
			field, users)
	}

	rows, err := tx.Query(ctx, query("following", strings.Join(followers[:], "', '"))+" returning username, timeline",
		msgProto.Username)
	if err != nil {
		//TODO: add logging
		return
	}

	var username string
	var timeline []int64
	strQuery := bytes.NewBufferString("update users as u1 set timeline = u2.timeline from (values")

	for rows.Next() {
		if err = rows.Scan(&username, &timeline); err != nil {
			log.Println(err) //TODO: add logging
		} else {
			newTimeline := tools.Difference(timeline, msgProto.Tweets)[:]
			// sort by tweets id, cause greater id means that tweet was created later
			// id is stored in bigserial which is 2^63-1, so we don't have to worry about renewing it
			sort.Slice(newTimeline, func(i, j int) bool {
				return newTimeline[i] > newTimeline[j]
			})

			strQuery.WriteString(fmt.Sprintf("('%s', array[%s]::integer[]),",
				username, strings.Trim(strings.Replace(fmt.Sprint(newTimeline), " ", ", ", -1), "[]")))
		}
	}

	strQuery.Truncate(strQuery.Len() - 1)
	strQuery.WriteString(") as u2(username, timeline) where u2.username = u1.username")
	rows.Close()

	if rows.CommandTag().RowsAffected() > 0 && err == nil {
		if _, err = tx.Exec(ctx, strQuery.String()); err != nil {
			//TODO: add logging
			return
		}
	}

	if _, err = tx.Exec(ctx, query("followers", strings.Join(following[:], "', '")),
		msgProto.Username); err != nil {
		//TODO: add logging
		return
	}

	if err = tx.Commit(ctx); err != nil {
		//TODO: add logging
		return
	}
}

func handleCreateTweet(ctx context.Context, serializedMsg []byte) {
	msgProto := &pb.NatsCreateTweetMessage{}
	if err := proto.Unmarshal(serializedMsg, msgProto); err != nil {
		//TODO: add logging
		return
	}

	if msgProto.GetUsername() == "" {
		//TODO: add logging
		return
	}
	if msgProto.GetTweetId() == 0 {
		//TODO: add logging
		return
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		//TODO: add logging
		return
	}
	defer tx.Rollback(ctx)

	var followers []string
	if err = tx.QueryRow(ctx, "update users set tweets = array_append(tweets, $1) where username = $2 returning followers",
		msgProto.TweetId, msgProto.Username).Scan(&followers); err != nil {
		//TODO: add logging
		return
	}

	if _, err = tx.Exec(ctx, fmt.Sprintf("update users set timeline = array_append(timeline, $1) where username in ('%s')",
		strings.Join(followers[:], "', '")), msgProto.TweetId); err != nil {
		//TODO: add logging
		return
	}

	if err = tx.Commit(ctx); err != nil {
		//TODO: add logging
		return
	}
}

func handleDeleteTweets(ctx context.Context, serializedMsg []byte) {
	msgProto := &pb.NatsDeleteTweetsMessage{}
	if err := proto.Unmarshal(serializedMsg, msgProto); err != nil {
		//TODO: add logging
		return
	}

	if msgProto.GetTweets() == nil {
		//TODO: add logging
		return
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		//TODO: add logging
		return
	}
	defer tx.Rollback(ctx)

	strTweets := tools.IntArrayToString(msgProto.Tweets)
	rows, err := tx.Query(ctx, fmt.Sprintf("select username, tweets, timeline from users where timeline && array[%[1]v] or tweets && array[%[1]v]", strTweets))
	if err != nil {
		//TODO: add logging
		return
	}

	var username string
	var tweets []int64
	var timeline []int64
	strQuery := bytes.NewBufferString("update users as u1 set tweets = u2.tweets, timeline = u2.timeline from (values")

	for rows.Next() {
		if err = rows.Scan(&username, &tweets, &timeline); err != nil {
			//TODO: add logging
			return
		}

		newTimeline := tools.Difference(timeline, msgProto.Tweets)
		newTweets := tools.Difference(tweets, msgProto.Tweets)
		strQuery.WriteString(fmt.Sprintf("('%s', array[%s]::integer[], array[%s]::integer[]),",
			username, strings.Trim(strings.Replace(fmt.Sprint(newTweets), " ", ", ", -1), "[]"),
			strings.Trim(strings.Replace(fmt.Sprint(newTimeline), " ", ", ", -1), "[]")))
	}

	strQuery.Truncate(strQuery.Len() - 1)
	strQuery.WriteString(") as u2(username, tweets, timeline) where u2.username = u1.username")

	if _, err := tx.Exec(ctx, strQuery.String()); err != nil {
		//TODO: add logging
		return
	}

	if err = tx.Commit(ctx); err != nil {
		//TODO: add logging
		return
	}
}

func natsCallback(natsMsg *nats.Msg) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msg := &pb.NatsMessage{}
	proto.Unmarshal(natsMsg.Data, msg)

	switch msg.Command {
	case pb.NatsMessage_DeleteUser:
		handleDeleteUser(ctx, msg.Message)
	case pb.NatsMessage_CreateTweet:
		handleCreateTweet(ctx, msg.Message)
	case pb.NatsMessage_DeleteTweet:
		handleDeleteTweets(ctx, msg.Message)
	}
}

func main() {
	var err error
	if pool, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL")); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	//TODO: add connect timeout + ping/pong
	if nc, err = nats.Connect(nats.DefaultURL); err != nil {
		log.Fatal(err)
	}
	nc.QueueSubscribe("users", "tweets_queue", natsCallback)
	nc.Flush()
	defer nc.Close()

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		PoolSize: 2,
	})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Start listening on: %s", port)

	s := grpc.NewServer()
	pb.RegisterUsersServer(s, &UsersServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
