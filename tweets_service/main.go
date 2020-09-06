package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"github.com/jackc/pgx/v4/pgxpool"
	nats "github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"

	pb "github.com/lieroz/dips-coursework-twitter/protos"
	"github.com/lieroz/dips-coursework-twitter/tools"
)

var (
	pool *pgxpool.Pool
	nc   *nats.Conn
	rdb  *redis.Client
)

const (
	port           = ":8002"
	rdbCtxTimeout  = 100 * time.Millisecond
	psqlCtxTimeout = 100 * time.Millisecond
	ncWriteTimeout = 100 * time.Millisecond
)

type TweetsServerImpl struct {
	pb.UnimplementedTweetsServer
}

func (*TweetsServerImpl) CreateTweet(ctx context.Context, in *pb.CreateTweetRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetCreator() == "" {
		sublogger.Error().Msg("'creator' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}
	if in.GetContent() == "" {
		sublogger.Error().Msg("'content' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}

	query := "insert into tweets (parent_id, creator, content) select $1, $2, $3 returning id"
	var id int64

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	if err := pool.QueryRow(psqlCtx,
		query,
		in.ParentId,
		in.Creator,
		in.Content,
	).Scan(&id); err != nil {
		sublogger.Error().Err(err).
			Str("$1", strconv.FormatInt(in.ParentId, 10)).
			Str("$2", in.Creator).
			Str("$3", in.Content).
			Msg(query)
		return nil, tools.GrpcError(codes.Internal)
	}

	cmdProto := &pb.NatsCreateTweetMessage{Username: in.Creator, TweetId: id}
	serializedCmd, err := proto.Marshal(cmdProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsCreateTweetMessage").Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	msgProto := &pb.NatsMessage{Command: pb.NatsMessage_CreateTweet, Message: serializedCmd}
	serializedMsg, err := proto.Marshal(msgProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsMessage").Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	if err := nc.Publish("users", serializedMsg); err != nil {
		sublogger.Error().Err(err).Str("nats command",
			pb.NatsMessage_Command_name[int32(pb.NatsMessage_CreateTweet)]).Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	return &pb.Empty{}, nil
}

func (*TweetsServerImpl) GetTweets(in *pb.GetTweetsRequest, stream pb.Tweets_GetTweetsServer) error {
	if in.GetTweets() == nil {
		log.Error().Msg("'tweets' field can't be empty")
		return tools.GrpcError(codes.InvalidArgument)
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(context.Background(), psqlCtxTimeout)
	defer psqlCtxCancel()

	query := fmt.Sprintf("select * from tweets where id in (%s)", tools.IntArrayToString(in.Tweets))
	rows, err := pool.Query(psqlCtx, query)
	if err != nil {
		log.Error().Err(err).Msg(query)
		return tools.GrpcError(codes.Internal)
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
			log.Error().Err(err).Send()
		} else {
			tweet.CreationTimestamp = timestamp.Unix()
			reply.Reply = &pb.GetTweetsReply_Tweet{Tweet: &tweet}
			if err = stream.Send(&reply); err != nil {
				log.Error().Err(err).Send()
				return tools.GrpcError(codes.Internal)
			}
		}
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Send()
		return tools.GrpcError(codes.Internal)
	}

	return nil
}

func (*TweetsServerImpl) EditTweet(ctx context.Context, in *pb.EditTweetRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetId() == 0 {
		sublogger.Error().Msg("'id' field can't be omitted")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}
	if in.GetContent() == "" {
		sublogger.Error().Msg("'content' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}
	psqlCtx, psqlCtxCancel := context.WithTimeout(context.Background(), psqlCtxTimeout)
	defer psqlCtxCancel()

	query := "update tweets set content = $1 where id = $2"
	if _, err := pool.Exec(psqlCtx, query, in.Content, in.Id); err != nil {
		sublogger.Error().Err(err).
			Str("$1", in.Content).
			Str("$2", strconv.FormatInt(in.Id, 10)).
			Msg(query)
		return nil, tools.GrpcError(codes.Internal)
	}

	return &pb.Empty{}, nil
}

func (*TweetsServerImpl) DeleteTweets(ctx context.Context, in *pb.DeleteTweetsRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetId() == nil {
		sublogger.Error().Msg("'id' field can't be omitted")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(context.Background(), psqlCtxTimeout)
	defer psqlCtxCancel()

	query := fmt.Sprintf("delete from tweets where id in (%s)", tools.IntArrayToString(in.Id))
	if _, err := pool.Exec(psqlCtx, query); err != nil {
		sublogger.Error().Err(err).Msg(query)
		return nil, tools.GrpcError(codes.Internal)
	}

	cmdProto := &pb.NatsDeleteTweetsMessage{Tweets: in.Id}
	serializedCmd, err := proto.Marshal(cmdProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsDeleteTweetMessage").Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	msgProto := &pb.NatsMessage{Command: pb.NatsMessage_DeleteTweet, Message: serializedCmd}
	serializedMsg, err := proto.Marshal(msgProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsMessage").Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	if err := nc.Publish("users", serializedMsg); err != nil {
		sublogger.Error().Err(err).Str("nats command",
			pb.NatsMessage_Command_name[int32(pb.NatsMessage_DeleteTweet)]).Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	return &pb.Empty{}, nil
}

func natsCallback(natsMsg *nats.Msg) {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	msg := &pb.NatsMessage{}
	proto.Unmarshal(natsMsg.Data, msg)

	// switch msg.Command {
	// case pb.NatsMessage_DeleteUser:
	// 	handleDeleteUser(ctx, msg.Message)
	// case pb.NatsMessage_CreateTweet:
	// 	handleCreateTweet(ctx, msg.Message)
	// case pb.NatsMessage_DeleteTweet:
	// 	handleDeleteTweets(ctx, msg.Message)
	// }
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Logger = log.With().Caller().Logger()

	psqlUrl := "postgres://user:password@localhost:5432/twitter_db?pool_max_conns=2"

	var err error
	if pool, err = pgxpool.Connect(context.Background(), psqlUrl); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to postgresql server")
	}

	//TODO: add connect timeout + ping/pong
	if nc, err = nats.Connect(nats.DefaultURL); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to nats server")
	}
	nc.QueueSubscribe("tweets", "users_queue", natsCallback)
	nc.Flush()
	defer nc.Close()

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		PoolSize: 2,
	})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}
	log.Info().Msgf("Start listening on: %s", port)

	s := grpc.NewServer()
	pb.RegisterTweetsServer(s, &TweetsServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve grpc service")
	}
}
