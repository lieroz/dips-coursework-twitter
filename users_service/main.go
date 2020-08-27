package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"github.com/jackc/pgx/v4"
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
	port           = ":8001"
	rdbCtxTimeout  = 100 * time.Millisecond
	psqlCtxTimeout = 100 * time.Millisecond
	ncWriteTimeout = 100 * time.Millisecond
)

type UsersServerImpl struct {
	pb.UnimplementedUsersServer
}

func (*UsersServerImpl) CreateUser(ctx context.Context, in *pb.CreateRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetUsername() == "" {
		sublogger.Error().Msg("'username' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	query := "insert into users (username, firstname, lastname, description) values ($1, $2, $3, $4) on conflict do nothing"
	cmdTag, err := pool.Exec(psqlCtx, query, in.Username, in.Firstname, in.Lastname, in.Description)

	if err != nil {
		sublogger.Error().Err(err).
			Str("$1", in.Username).
			Str("$2", in.Firstname).
			Str("$3", in.Lastname).
			Str("$4", in.Description).
			Msg(query)
		return nil, tools.GrpcError(codes.Internal)
	}

	if cmdTag.RowsAffected() == 0 {
		return nil, tools.GrpcError(codes.AlreadyExists)
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) DeleteUser(ctx context.Context, in *pb.DeleteRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetUsername() == "" {
		sublogger.Error().Msg("'username' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}

	var followers, following []string
	var tweets []int64

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	query := "select followers, following, tweets from users where username = $1"

	if err := pool.QueryRow(psqlCtx, in.Username).
		Scan(query, &followers, &following, &tweets); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Username).Msg(query)
		return nil, tools.GrpcError(codes.Internal)
	}

	cmdProto := &pb.NatsDeleteUserMessage{Username: in.Username, Tweets: tweets}
	serializedCmd, err := proto.Marshal(cmdProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsDeleteUserMessage").Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	msgProto := &pb.NatsMessage{Command: pb.NatsMessage_DeleteUser, Message: serializedCmd}
	serializedMsg, err := proto.Marshal(msgProto)
	if err != nil {
		sublogger.Error().Err(err).Str("protobuf message", "NatsMessage").Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	rdbCtx, rdbCtxCancel := context.WithTimeout(ctx, rdbCtxTimeout)
	defer rdbCtxCancel()

	key := fmt.Sprintf("%s:delete", in.Username)
	pipe := rdb.Pipeline()
	pipe.HSet(rdbCtx, key, "followers", followers, "following", following)
	pipe.Expire(rdbCtx, key, 10*time.Second)

	if _, err = pipe.Exec(rdbCtx); err != nil {
		// key in redis will be cleaned up after expiration time if error happened
		sublogger.Error().Err(err).Msgf(
			"pipeline hset %[1]v 'followers' '%s' 'following' '%s'; expire %[1]v 10",
			key, strings.Join(followers[:], ", "), strings.Join(following[:], ", "))
		return nil, tools.GrpcError(codes.Internal)
	}

	if err := nc.Publish("tweets", serializedMsg); err != nil {
		sublogger.Error().Err(err).Str("nats command", pb.NatsMessage_Command_name[int32(pb.NatsMessage_DeleteUser)]).Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) GetUserInfoSummary(ctx context.Context, in *pb.GetSummaryRequest) (*pb.GetSummaryReply, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetUsername() == "" {
		sublogger.Error().Msg("'username' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}

	var timestamp time.Time

	summary := &pb.GetSummaryReply{}
	query := `select username, firstname, lastname, description,
		registration_timestamp, cardinality(followers), cardinality(following), cardinality(tweets)
		from users where username = $1`

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	if err := pool.QueryRow(psqlCtx, query, in.Username).Scan(
		&summary.Username,
		&summary.Firstname,
		&summary.Lastname,
		&summary.Description,
		&timestamp,
		&summary.FollowersCount,
		&summary.FollowingCount,
		&summary.TweetsCount,
	); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Username).Msg(query)
		if err == pgx.ErrNoRows {
			return nil, tools.GrpcError(codes.NotFound)
		}
		return nil, tools.GrpcError(codes.Internal)
	}

	summary.RegistrationTimestamp = timestamp.Unix()
	return summary, nil
}

func (*UsersServerImpl) GetUsers(in *pb.GetUsersRequest, stream pb.Users_GetUsersServer) error {
	if in.GetUsername() == "" {
		log.Error().Msg("'username' field can't be empty")
		return tools.GrpcError(codes.InvalidArgument)
	}

	query := "select %s from users where username = $1"
	switch in.GetRole() {
	case pb.Role_Follower:
		query = fmt.Sprintf(query, "followers")
	case pb.Role_Followed:
		query = fmt.Sprintf(query, "following")
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(context.Background(), psqlCtxTimeout)
	defer psqlCtxCancel()

	var users []string
	if err := pool.QueryRow(psqlCtx, query,
		in.Username).Scan(&users); err != nil {
		log.Error().Err(err).Msg(query[:len(query)-2] + in.Username)
		return tools.GrpcError(codes.Internal)
	}

	query = fmt.Sprintf(`select username, firstname, lastname, description, 
		registration_timestamp, followers, following, tweets 
		from users where username in ('%s')`, strings.Join(users[:], "', '"))
	rows, err := pool.Query(psqlCtx, query)

	if err != nil {
		log.Error().Err(err).Msg(query)
		return tools.GrpcError(codes.Internal)
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
		} else {
			user.RegistrationTimestamp = timestamp.Unix()
			reply.Reply = &pb.GetUsersReply_User{User: &user}
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

func (*UsersServerImpl) Follow(ctx context.Context, in *pb.FollowRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetFollower() == "" {
		sublogger.Error().Msg("'follower' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}
	if in.GetFollowed() == "" {
		sublogger.Error().Msg("'followed' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}
	if in.GetFollower() == in.GetFollowed() {
		sublogger.Error().Msg("user can't follow himself")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	tx, err := pool.BeginTx(psqlCtx, pgx.TxOptions{})
	if err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal)
	}
	defer tx.Rollback(psqlCtx)

	query := func(v string) string {
		return fmt.Sprintf(`update users set %[1]v = array_append(%[1]v, $1::varchar) where username = $2
			and exists (select 1 from users where username = $2 and not ($1 = any(%[1]v)))`, v)
	}

	var timeline []int64
	strQuery := fmt.Sprintf(query("following")) + " returning timeline"

	if err = tx.QueryRow(psqlCtx, strQuery, in.Followed, in.Follower).Scan(&timeline); err != nil {
		if err == pgx.ErrNoRows {
			return nil, tools.GrpcError(codes.NotFound)
		}
		sublogger.Error().Err(err).Msg(strQuery)
		return nil, tools.GrpcError(codes.Internal)
	}

	var tweets []int64
	strQuery = fmt.Sprintf(query("followers")) + " returning tweets"
	if err = tx.QueryRow(psqlCtx, strQuery, in.Follower, in.Followed).Scan(&tweets); err != nil {
		if err == pgx.ErrNoRows {
			return nil, tools.GrpcError(codes.NotFound)
		}
		sublogger.Error().Err(err).Msg(strQuery)
		return nil, tools.GrpcError(codes.Internal)
	}

	timeline = append(timeline, tweets...)

	if len(timeline) != 0 {
		// sort by tweets id, cause greater id means that tweet was created later
		// id is stored in bigserial which is 2^63-1, so we don't have to worry about renewing it
		sort.Slice(timeline, func(i, j int) bool {
			return timeline[i] > timeline[j]
		})

		strQuery = fmt.Sprintf("update users set timeline = array[%s]::integer[] where username = $1",
			tools.IntArrayToString(timeline))

		if _, err = tx.Exec(psqlCtx, strQuery, in.Follower); err != nil {
			sublogger.Error().Err(err).Msg(strQuery)
			return nil, tools.GrpcError(codes.Internal)
		}
	}

	if err = tx.Commit(psqlCtx); err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal)
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) Unfollow(ctx context.Context, in *pb.FollowRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetFollower() == "" {
		sublogger.Error().Msg("'follower' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}
	if in.GetFollowed() == "" {
		sublogger.Error().Msg("'followed' field can't be empty")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}
	if in.GetFollower() == in.GetFollowed() {
		sublogger.Error().Msg("user can't unfollow himself")
		return nil, tools.GrpcError(codes.InvalidArgument)
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	tx, err := pool.BeginTx(psqlCtx, pgx.TxOptions{})
	if err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal)
	}
	defer tx.Rollback(psqlCtx)

	query := func(v string) string {
		return fmt.Sprintf(`update users set %[1]v = array_remove(%[1]v, $1::varchar) where username = $2 
			and exists (select 1 from users where username = $2 and $1 = any(%[1]v))`, v)
	}

	strQuery := query("following") + " returning timeline"
	var timeline []int64

	if err = tx.QueryRow(psqlCtx, strQuery, in.Followed, in.Follower).Scan(&timeline); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Followed).Str("$2", in.Follower).Msg(strQuery)
		if err == pgx.ErrNoRows {
			return nil, tools.GrpcError(codes.NotFound)
		}
		return nil, tools.GrpcError(codes.Internal)
	}

	strQuery = query("followers") + " returning tweets"
	var tweets []int64

	if err = tx.QueryRow(psqlCtx, strQuery, in.Follower, in.Followed).Scan(&tweets); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Follower).Str("$2", in.Followed).Msg(strQuery)
		if err == pgx.ErrNoRows {
			return nil, tools.GrpcError(codes.NotFound)
		}
		return nil, tools.GrpcError(codes.Internal)
	}

	newTimeline := tools.Difference(timeline, tweets)
	strQuery = fmt.Sprintf("update users set timeline = array[%s]::integer[] where username = $1",
		tools.IntArrayToString(newTimeline))

	if _, err = tx.Exec(psqlCtx, strQuery, in.Follower); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Follower).Msg(strQuery)
		return nil, tools.GrpcError(codes.Internal)
	}

	if err = tx.Commit(psqlCtx); err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal)
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
	pipe := rdb.Pipeline()
	hmgetRes := pipe.HMGet(rdbCtx, key, "followers", "following")
	pipe.Del(rdbCtx, key)

	if _, err := pipe.Exec(rdbCtx); err != nil {
		log.Error().Err(err).Send()
		return
	}

	vals := hmgetRes.Val()
	followers := tools.InterfaceToStringArray(vals[0])
	following := tools.InterfaceToStringArray(vals[1])

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	tx, err := pool.BeginTx(psqlCtx, pgx.TxOptions{})
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	defer tx.Rollback(psqlCtx)

	query := func(field, users string) string {
		return fmt.Sprintf("update users set %[1]v = array_remove(%[1]v , $1) where username in ('%[2]v')",
			field, users)
	}

	strQuery := query("following", strings.Join(followers[:], "', '")) + " returning username, timeline"
	rows, err := tx.Query(psqlCtx, strQuery, msgProto.Username)

	if err != nil {
		log.Error().Err(err).Msg(strQuery)
		return
	}

	var username string
	var timeline []int64
	queryBuilder := bytes.NewBufferString("update users as u1 set timeline = u2.timeline from (values")

	for rows.Next() {
		if err = rows.Scan(&username, &timeline); err != nil {
			log.Error().Err(err).Send()
		} else {
			newTimeline := tools.Difference(timeline, msgProto.Tweets)[:]
			// sort by tweets id, cause greater id means that tweet was created later
			// id is stored in bigserial which is 2^63-1, so we don't have to worry about renewing it
			sort.Slice(newTimeline, func(i, j int) bool {
				return newTimeline[i] > newTimeline[j]
			})

			queryBuilder.WriteString(fmt.Sprintf("('%s', array[%s]::integer[]),",
				username, strings.Trim(strings.Replace(fmt.Sprint(newTimeline), " ", ", ", -1), "[]")))
		}
	}

	queryBuilder.Truncate(queryBuilder.Len() - 1)
	queryBuilder.WriteString(") as u2(username, timeline) where u2.username = u1.username")
	rows.Close()

	strQuery = queryBuilder.String()

	if rows.CommandTag().RowsAffected() > 0 && err == nil {
		if _, err = tx.Exec(psqlCtx, strQuery); err != nil {
			log.Error().Err(err).Msg(strQuery)
			return
		}
	}

	strQuery = query("followers", strings.Join(following[:], "', '"))

	if _, err = tx.Exec(psqlCtx, strQuery, msgProto.Username); err != nil {
		log.Error().Err(err).Msg(strQuery)
		return
	}

	if err = tx.Commit(psqlCtx); err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func handleCreateTweet(ctx context.Context, serializedMsg []byte) {
	msgProto := &pb.NatsCreateTweetMessage{}
	if err := proto.Unmarshal(serializedMsg, msgProto); err != nil {
		log.Error().Err(err).Send()
		return
	}

	if msgProto.GetUsername() == "" {
		log.Error().Msg("'username' field can't be empty")
		return
	}
	if msgProto.GetTweetId() == 0 {
		log.Error().Msg("'tweet_id' field can't be omitted")
		return
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	tx, err := pool.BeginTx(psqlCtx, pgx.TxOptions{})
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	defer tx.Rollback(psqlCtx)

	var followers []string
	query := "update users set tweets = array_append(tweets, $1) where username = $2 returning followers"

	if err = tx.QueryRow(psqlCtx, query, msgProto.TweetId, msgProto.Username).Scan(&followers); err != nil {
		log.Error().Err(err).
			Str("$1", strconv.FormatInt(msgProto.TweetId, 10)).
			Str("$2", msgProto.Username).
			Msg(query)
		return
	}

	query = fmt.Sprintf("update users set timeline = array_append(timeline, $1) where username in ('%s')",
		strings.Join(followers[:], "', '"))

	if _, err = tx.Exec(psqlCtx, query, msgProto.TweetId); err != nil {
		log.Error().Err(err).
			Str("$1", strconv.FormatInt(msgProto.TweetId, 10)).
			Msg(query)
		return
	}

	if err = tx.Commit(psqlCtx); err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func handleDeleteTweets(ctx context.Context, serializedMsg []byte) {
	msgProto := &pb.NatsDeleteTweetsMessage{}
	if err := proto.Unmarshal(serializedMsg, msgProto); err != nil {
		log.Error().Err(err).Send()
		return
	}

	if msgProto.GetTweets() == nil {
		log.Error().Msg("'tweets' field can't be empty")
		return
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	tx, err := pool.BeginTx(psqlCtx, pgx.TxOptions{})
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	defer tx.Rollback(psqlCtx)

	strTweets := tools.IntArrayToString(msgProto.Tweets)
	query := fmt.Sprintf("select username, tweets, timeline from users where timeline && array[%[1]v] or tweets && array[%[1]v]", strTweets)
	rows, err := tx.Query(psqlCtx, query)

	if err != nil {
		log.Error().Err(err).Msg(query)
		return
	}

	var username string
	var tweets []int64
	var timeline []int64
	queryBuilder := bytes.NewBufferString("update users as u1 set tweets = u2.tweets, timeline = u2.timeline from (values")

	for rows.Next() {
		if err = rows.Scan(&username, &tweets, &timeline); err != nil {
			log.Error().Err(err).Send()
			return
		}

		newTimeline := tools.Difference(timeline, msgProto.Tweets)
		newTweets := tools.Difference(tweets, msgProto.Tweets)
		queryBuilder.WriteString(fmt.Sprintf("('%s', array[%s]::integer[], array[%s]::integer[]),",
			username, strings.Trim(strings.Replace(fmt.Sprint(newTweets), " ", ", ", -1), "[]"),
			strings.Trim(strings.Replace(fmt.Sprint(newTimeline), " ", ", ", -1), "[]")))
	}

	queryBuilder.Truncate(queryBuilder.Len() - 1)
	queryBuilder.WriteString(") as u2(username, tweets, timeline) where u2.username = u1.username")
	query = queryBuilder.String()

	if _, err := tx.Exec(psqlCtx, query); err != nil {
		log.Error().Err(err).Msg(query)
		return
	}

	if err = tx.Commit(psqlCtx); err != nil {
		log.Error().Err(err).Send()
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
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Logger = log.With().Caller().Logger()

	var err error
	if pool, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL")); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to postgresql server")
	}

	//TODO: add connect timeout + ping/pong
	if nc, err = nats.Connect(nats.DefaultURL); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to nats server")
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
		log.Fatal().Err(err).Msg("Failed to listen")
	}
	log.Info().Msgf("Start listening on: %s", port)

	s := grpc.NewServer()
	pb.RegisterUsersServer(s, &UsersServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve grpc service")
	}
}
