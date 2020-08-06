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

	pb "github.com/lieroz/dips-coursework-twitter/tweets_service/protos"
)

var (
	pool *pgxpool.Pool
)

const (
	port = ":8002"
)

type TweetsServerImpl struct {
	pb.UnimplementedTweetsServer
}

func (*TweetsServerImpl) CreateTweet(ctx context.Context, in *pb.CreateTweetRequest) (*pb.Empty, error) {
	if in.GetCreator() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'creator' field can't be empty")
	}
	if in.GetContent() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'content' field can't be empty")
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	defer tx.Rollback(ctx)

	var id int
	// https://stackoverflow.com/questions/31733790/postgresql-parameter-issue-1
	err = pool.QueryRow(ctx,
		`insert into tweets (parent_id, creator, content) select $1, $2::varchar, $3 
			where exists (select 1 from users where username = $2) returning id`,
		in.GetParentId(),
		in.GetCreator(),
		in.GetContent(),
	).Scan(&id)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound,
				"user with username: '%s' doesn't exist and tweet can't be created", in.GetCreator())

		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	// maybe check for result? but this is under tx, so must be granular enough
	if _, err = pool.Exec(ctx, "update users set tweets = array_append(tweets, $1) where username = $2",
		id, in.GetCreator()); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.Empty{}, nil
}

func (*TweetsServerImpl) GetTweets(in *pb.GetTweetsRequest, stream pb.Tweets_GetTweetsServer) error {
	if in.GetTweets() == nil {
		return status.Errorf(codes.InvalidArgument, "'tweets' field can't be empty")
	}

	query := fmt.Sprintf("select * from tweets where id in (%s)",
		strings.Trim(strings.Replace(fmt.Sprint(in.GetTweets()), " ", ", ", -1), "[]"))
	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return status.Errorf(codes.Internal, "%s", err)
	}

	tweet := pb.Tweet{}
	var timestamp time.Time
	var reply pb.GetTweetsReply

	for rows.Next() {
		if err = rows.Scan(
			&tweet.Id,
			&tweet.ParentId,
			&tweet.Creator,
			&tweet.Content,
			&timestamp); err != nil {
			reply.Reply = &pb.GetTweetsReply_Error{Error: true}
			log.Println(err)
		} else {
			tweet.CreationTimestamp = timestamp.Unix()
			reply.Reply = &pb.GetTweetsReply_Tweet{Tweet: &tweet}
			if err = stream.Send(&reply); err != nil {
				return status.Errorf(codes.Internal, "%s", err)
			}
		}
	}

	return nil
}

func (*TweetsServerImpl) EditTweet(ctx context.Context, in *pb.EditTweetRequest) (*pb.Empty, error) {
	if in.GetId() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "'id' field can't be omitted")
	}
	if in.GetContent() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "'content' field can't be empty")
	}

	if cmdTag, err := pool.Exec(ctx, "update tweets set content = $1 where id = $2 and parent_id = 0",
		in.GetContent(), in.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	} else {
		if cmdTag.RowsAffected() == 0 {
			return nil, status.Errorf(codes.InvalidArgument,
				"retweeted tweet can't be updated")
		}
	}

	return &pb.Empty{}, nil
}

func (*TweetsServerImpl) DeleteTweet(ctx context.Context, in *pb.DeleteTweetsRequest) (*pb.Empty, error) {
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
	pb.RegisterTweetsServer(s, &TweetsServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
