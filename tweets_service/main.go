package main

import (
	"context"
	"fmt"
	"io"
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
	pool        *pgxpool.Pool
	usersClient pb.UsersClient
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
	err = tx.QueryRow(ctx,
		"insert into tweets (parent_id, creator, content) select $1, $2, $3 returning id",
		in.GetParentId(),
		in.GetCreator(),
		in.GetContent(),
	).Scan(&id)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	if _, err = usersClient.OnTweetCreated(ctx, &pb.OnTweetCreatedRequest{
		Username: in.GetCreator(),
		TweetId:  int64(id),
	}); err != nil {
		return nil, err
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

	if rows.Err() != nil {
		return status.Errorf(codes.Internal, "%s", rows.Err())
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

	if _, err := pool.Exec(ctx, "update tweets set content = $1 where id = $2",
		in.GetContent(), in.GetId()); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.Empty{}, nil
}

func (*TweetsServerImpl) DeleteTweets(ctx context.Context, in *pb.DeleteTweetsRequest) (*pb.Empty, error) {
	if in.GetId() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "'id' field can't be omitted")
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, fmt.Sprintf("delete from tweets where id in (%s)",
		strings.Trim(strings.Replace(fmt.Sprint(in.GetId()), " ", ", ", -1), "[]"))); err != nil {
		return nil, err
	}

	if in.GetContext() == pb.DeleteTweetsRequest_Client {
		// Add request to users service
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	return &pb.Empty{}, nil
}

func (*TweetsServerImpl) GetOrderedTimeline(stream pb.Tweets_GetOrderedTimelineServer) error {
	ctx := context.Background()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Println(err)
		} else {
			if rows, err := pool.Query(ctx, fmt.Sprintf("select id from tweets where id in (%s) group by id order by creation_timestamp",
				strings.Trim(strings.Replace(fmt.Sprint(in.GetTimeline()), " ", ", ", -1), "[]"))); err != nil {
				log.Println(err)
			} else {
				timeline := pb.Timeline{}
				var tweetId int64

				for rows.Next() {
					if err := rows.Scan(&tweetId); err != nil {
						log.Println(err)
					} else {
						timeline.Timeline = append(timeline.Timeline, tweetId)
					}
				}

				if err = stream.Send(&timeline); err != nil {
					log.Println(err)
				}

				if rows.Err() != nil {
					log.Println(err)
				}
			}
		}
	}
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

	conn, err := grpc.Dial("localhost:8001", grpc.WithInsecure())
	if err != nil {
	}
	defer conn.Close()
	usersClient = pb.NewUsersClient(conn)

	s := grpc.NewServer()
	pb.RegisterTweetsServer(s, &TweetsServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
