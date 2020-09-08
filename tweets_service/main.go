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
		return nil, tools.GrpcError(codes.InvalidArgument, "'creator' field can't be empty")
	}
	if in.GetContent() == "" {
		return nil, tools.GrpcError(codes.InvalidArgument, "'content' field can't be empty")
	}

	query := "select nextval('tweet_id_seq')"
	var id int64

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	if err := pool.QueryRow(psqlCtx, query).Scan(&id); err != nil {
		sublogger.Error().Err(err).
			Str("$1", strconv.FormatInt(in.ParentId, 10)).
			Str("$2", in.Creator).
			Str("$3", in.Content).
			Msg(query)
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	cmdProto := &pb.NatsCreateTweetMessage{TweetId: id, ParentTweetId: in.ParentId, Creator: in.Creator, Content: in.Content}
	serializedCmd, err := proto.Marshal(cmdProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsCreateTweetMessage").Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	msgProto := &pb.NatsMessage{Command: pb.NatsMessage_CreateTweet, Message: serializedCmd}
	serializedMsg, err := proto.Marshal(msgProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsMessage").Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	if err := nc.Publish("users", serializedMsg); err != nil {
		sublogger.Error().Err(err).Str("nats command",
			pb.NatsMessage_Command_name[int32(pb.NatsMessage_CreateTweet)]).Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	return &pb.Empty{}, nil
}

func (*TweetsServerImpl) GetTweets(in *pb.GetTweetsRequest, stream pb.Tweets_GetTweetsServer) error {
	if in.GetTweets() == nil {
		return tools.GrpcError(codes.InvalidArgument, "'tweets' field can't be empty")
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(context.Background(), psqlCtxTimeout)
	defer psqlCtxCancel()

	query := `with recursive tweets_and_retweets as (
		select * from tweets where id in (%s) 
		union
		select r.* from tweets r join tweets_and_retweets tar on r.id = tar.parent_id
	) select * from tweets_and_retweets`

	rows, err := pool.Query(psqlCtx, fmt.Sprintf(query, tools.IntArrayToString(in.Tweets)))
	if err != nil {
		log.Error().Err(err).Msg(query)
		return tools.GrpcError(codes.Internal, "INTERNAL ERROR")
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
		}

		if err = stream.Send(&reply); err != nil {
			log.Error().Err(err).Send()
			return tools.GrpcError(codes.Internal, "INTERNAL ERROR")
		}
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err()).Send()
		return tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	return nil
}

func deleteTweets(ctx context.Context, logger *zerolog.Logger, tweets []int64, cmd pb.NatsMessage_Command, msg []byte) error {
	query := `with recursive tweets_and_retweets as (
		select id from tweets where id in (%s)
		union 
		select r.id from tweets r join tweets_and_retweets tar on r.parent_id = tar.id
	) delete from tweets where id in (select * from tweets_and_retweets);`

	if _, err := pool.Exec(ctx, fmt.Sprintf(query, tools.IntArrayToString(tweets))); err != nil {
		logger.Error().Err(err).Msg(query)
		return err
	}
	msgProto := &pb.NatsMessage{Command: cmd, Message: msg}
	serializedMsg, err := proto.Marshal(msgProto)
	if err != nil {
		logger.Error().Err(err).Str("protobuf message", "NatsMessage").Send()
		return err
	}

	if err := nc.Publish("users", serializedMsg); err != nil {
		logger.Error().Err(err).Str("nats command",
			pb.NatsMessage_Command_name[int32(cmd)]).Send()
		return err
	}

	return nil
}

func (*TweetsServerImpl) DeleteTweets(ctx context.Context, in *pb.DeleteTweetsRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetTweets() == nil {
		return nil, tools.GrpcError(codes.InvalidArgument, "'tweets' field can't be omitted")
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	msgProto := &pb.NatsDeleteTweetsMessage{Tweets: in.Tweets}
	msg, err := proto.Marshal(msgProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsDeleteTweetsMessage").Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	if err := deleteTweets(psqlCtx, &sublogger, in.Tweets, pb.NatsMessage_DeleteTweets, msg); err != nil {
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	return &pb.Empty{}, nil
}

func handleDeleteUser(ctx context.Context, serializedMsg []byte) {
	msgProto := &pb.NatsDeleteUserMessage{}
	if err := proto.Unmarshal(serializedMsg, msgProto); err != nil {
		log.Error().Err(err).Send()
		return
	}

	if msgProto.GetUsername() == "" {
		log.Error().Msg("'username' field can't be empty")
		return
	}
	if msgProto.GetTweets() == nil {
		log.Error().Msg("'tweets' field can't be empty")
		return
	}

	rdbCtx, rdbCtxCancel := context.WithTimeout(ctx, rdbCtxTimeout)
	defer rdbCtxCancel()

	key := fmt.Sprintf("%s:delete", msgProto.Username)
	if res, err := rdb.Exists(rdbCtx, key).Result(); err != nil {
		log.Error().Err(err).Send()
		return
	} else {
		if res == 0 {
			log.Error().Msgf("transaction %s doesn't exist", key)
			return
		}
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(context.Background(), psqlCtxTimeout)
	defer psqlCtxCancel()

	deleteTweets(psqlCtx, &log.Logger, msgProto.Tweets, pb.NatsMessage_DeleteUser, serializedMsg)
}

func handleCreateTweet(ctx context.Context, serializedMsg []byte) {
	msgProto := &pb.NatsCreateTweetMessage{}
	if err := proto.Unmarshal(serializedMsg, msgProto); err != nil {
		log.Error().Err(err).Send()
		return
	}

	if msgProto.GetTweetId() == 0 {
		log.Error().Msg("'tweet_id' field can't be omitted")
		return
	}
	if msgProto.GetCreator() == "" {
		log.Error().Msg("'creator' field can't be empty")
		return
	}
	if msgProto.GetContent() == "" {
		log.Error().Msg("'content' field can't be empty")
		return
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	query := "insert into tweets (id, parent_id, creator, content) values ($1, $2, $3, $4)"
	if _, err := pool.Exec(psqlCtx,
		query,
		msgProto.TweetId,
		msgProto.GetParentTweetId(),
		msgProto.Creator,
		msgProto.Content,
	); err != nil {
		log.Error().Err(err).
			Str("$1", strconv.FormatInt(msgProto.TweetId, 10)).
			Str("$2", strconv.FormatInt(msgProto.GetParentTweetId(), 10)).
			Str("$3", msgProto.Creator).
			Str("$4", msgProto.Content).
			Msg(query)
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
	}
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

	if nc, err = nats.Connect(nats.DefaultURL,
		nats.UserInfo("user", "password")); err != nil {
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
