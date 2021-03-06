package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/examples/data"
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
		return nil, tools.GrpcError(codes.InvalidArgument, "'username' field can't be empty")
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
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	if cmdTag.RowsAffected() == 0 {
		return nil, tools.GrpcError(codes.AlreadyExists, fmt.Sprintf("user '%s' already exists", in.Username))
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) DeleteUser(ctx context.Context, in *pb.DeleteRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetUsername() == "" {
		return nil, tools.GrpcError(codes.InvalidArgument, "'username' field can't be empty")
	}

	var followers, following []string
	var tweets []int64

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	query := "select followers, following, tweets from users where username = $1"

	if err := pool.QueryRow(psqlCtx, query, in.Username).
		Scan(&followers, &following, &tweets); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Username).Msg(query)
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	if len(tweets) > 0 {
		cmdProto := &pb.NatsDeleteUserMessage{Username: in.Username, Tweets: tweets}
		serializedCmd, err := proto.Marshal(cmdProto)
		if err != nil {
			sublogger.Error().Err(err).Str("protobuf message", "NatsDeleteUserMessage").Send()
			return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
		}

		msgProto := &pb.NatsMessage{Command: pb.NatsMessage_DeleteUser, Message: serializedCmd}
		serializedMsg, err := proto.Marshal(msgProto)
		if err != nil {
			sublogger.Error().Err(err).Str("protobuf message", "NatsMessage").Send()
			return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
		}

		rdbCtx, rdbCtxCancel := context.WithTimeout(ctx, rdbCtxTimeout)
		defer rdbCtxCancel()

		key := fmt.Sprintf("%s:delete", in.Username)
		pipe := rdb.Pipeline()
		pipe.HSet(rdbCtx, key, "followers", strings.Join(followers, ","),
			"following", strings.Join(following, ","))
		pipe.Expire(rdbCtx, key, 10*time.Second)

		if _, err = pipe.Exec(rdbCtx); err != nil {
			// key in redis will be cleaned up after expiration time if error happened
			sublogger.Error().Err(err).Msgf(
				"pipeline hset %[1]v 'followers' '%s' 'following' '%s'; expire %[1]v 10",
				key, strings.Join(followers[:], ", "), strings.Join(following[:], ", "))
			return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
		}

		if err := nc.Publish("tweets", serializedMsg); err != nil {
			sublogger.Error().Err(err).Str("nats command",
				pb.NatsMessage_Command_name[int32(pb.NatsMessage_DeleteUser)]).Send()
			return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
		}
	} else {
		query := "delete from users where username = $1"

		if _, err := pool.Exec(psqlCtx, query, in.Username); err != nil {
			sublogger.Error().Err(err).Str("$1", in.Username).Msg(query)
			return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
		}
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) GetUserInfoSummary(ctx context.Context, in *pb.GetSummaryRequest) (*pb.GetSummaryReply, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetUsername() == "" {
		return nil, tools.GrpcError(codes.InvalidArgument, "'username' field can't be empty")
	}

	var timestamp time.Time

	summary := &pb.GetSummaryReply{}
	query := `select username, firstname, lastname, description,
		registration_timestamp, cardinality(followers), cardinality(following), tweets
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
		&summary.Tweets,
	); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Username).Msg(query)
		if err == pgx.ErrNoRows {
			return nil, tools.GrpcError(codes.NotFound, fmt.Sprintf("user '%s' doesn't exist", in.Username))
		}
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	summary.RegistrationTimestamp = timestamp.Unix()
	return summary, nil
}

func (*UsersServerImpl) GetUsers(in *pb.GetUsersRequest, stream pb.Users_GetUsersServer) error {
	if in.GetUsername() == "" {
		return tools.GrpcError(codes.InvalidArgument, "'username' field can't be empty")
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
		log.Error().Err(err).Str("$1", in.Username).Msg(query)
		return tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	query = fmt.Sprintf(`select username, firstname, lastname, description, 
		registration_timestamp, followers, following, tweets 
		from users where username in ('%s')`, strings.Join(users[:], "', '"))
	rows, err := pool.Query(psqlCtx, query)

	if err != nil {
		log.Error().Err(err).Msg(query)
		return tools.GrpcError(codes.Internal, "INTERNAL ERROR")
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

func (*UsersServerImpl) GetTimeline(ctx context.Context, in *pb.GetTimelineRequest) (*pb.GetTimelineResponse, error) {
	if in.GetUsername() == "" {
		return nil, tools.GrpcError(codes.InvalidArgument, "'username' field can't be empty")
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(context.Background(), psqlCtxTimeout)
	defer psqlCtxCancel()

	var timeline []int64
	query := "select timeline from users where username = $1"

	if err := pool.QueryRow(psqlCtx, query, in.Username).Scan(&timeline); err != nil {
		log.Error().Err(err).Str("$1", in.Username).Msg(query)
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	return &pb.GetTimelineResponse{Timeline: timeline}, nil
}

func (*UsersServerImpl) Follow(ctx context.Context, in *pb.FollowRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetFollower() == "" {
		return nil, tools.GrpcError(codes.InvalidArgument, "'follower' field can't be empty")
	}
	if in.GetFollowed() == "" {
		return nil, tools.GrpcError(codes.InvalidArgument, "'followed' field can't be empty")
	}
	if in.GetFollower() == in.GetFollowed() {
		return nil, tools.GrpcError(codes.InvalidArgument, "user can't follow himself")
	}

	followerKey := fmt.Sprintf("delete:%s", in.Follower)
	followeeKey := fmt.Sprintf("delete:%s", in.Followed)

	rdbCtx, rdbCtxCancel := context.WithTimeout(ctx, rdbCtxTimeout)
	defer rdbCtxCancel()

	pipe := rdb.Pipeline()
	followerTxExists := pipe.Exists(rdbCtx, followerKey)
	followeeTxExists := pipe.Exists(rdbCtx, followeeKey)

	if _, err := pipe.Exec(rdbCtx); err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	if followerTxExists.Val() == 1 || followeeTxExists.Val() == 1 {
		sublogger.Warn().
			Str("follower", strconv.FormatInt(followerTxExists.Val(), 10)).
			Str("followed", strconv.FormatInt(followeeTxExists.Val(), 10)).
			Msg("delete tx in progress")
		return nil, tools.GrpcError(codes.Canceled, "can't proceed, delete tx in progress")
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	tx, err := pool.BeginTx(psqlCtx, pgx.TxOptions{})
	if err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
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
			return nil, tools.GrpcError(codes.NotFound, fmt.Sprintf("user '%s' doesn't exist", in.Follower))
		}
		sublogger.Error().Err(err).Msg(strQuery)
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	var tweets []int64
	strQuery = fmt.Sprintf(query("followers")) + " returning tweets"
	if err = tx.QueryRow(psqlCtx, strQuery, in.Follower, in.Followed).Scan(&tweets); err != nil {
		if err == pgx.ErrNoRows {
			return nil, tools.GrpcError(codes.NotFound, fmt.Sprintf("user '%s' doesn't exist", in.Followed))
		}
		sublogger.Error().Err(err).Msg(strQuery)
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	timeline = append(timeline, tweets...)

	if len(timeline) != 0 {
		// sort by tweets id, cause greater id means that tweet was created later
		// id is stored in bigserial which is 2^63-1, so we don't have to worry about renewing it
		sort.Slice(timeline, func(i, j int) bool {
			return timeline[i] < timeline[j]
		})

		strQuery = fmt.Sprintf("update users set timeline = array[%s]::integer[] where username = $1",
			tools.IntArrayToString(timeline))

		if _, err = tx.Exec(psqlCtx, strQuery, in.Follower); err != nil {
			sublogger.Error().Err(err).Msg(strQuery)
			return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
		}
	}

	if err = tx.Commit(psqlCtx); err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	return &pb.Empty{}, nil
}

func (*UsersServerImpl) Unfollow(ctx context.Context, in *pb.FollowRequest) (*pb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	sublogger := log.With().
		Str("client ip", p.Addr.String()).
		Logger()

	if in.GetFollower() == "" {
		return nil, tools.GrpcError(codes.InvalidArgument, "'follower' field can't be empty")
	}
	if in.GetFollowed() == "" {
		return nil, tools.GrpcError(codes.InvalidArgument, "'followed' field can't be empty")
	}
	if in.GetFollower() == in.GetFollowed() {
		return nil, tools.GrpcError(codes.InvalidArgument, "user can't unfollow himself")
	}

	followerKey := fmt.Sprintf("delete:%s", in.Follower)
	followeeKey := fmt.Sprintf("delete:%s", in.Followed)

	rdbCtx, rdbCtxCancel := context.WithTimeout(ctx, rdbCtxTimeout)
	defer rdbCtxCancel()

	pipe := rdb.Pipeline()
	followerTxExists := pipe.Exists(rdbCtx, followerKey)
	followeeTxExists := pipe.Exists(rdbCtx, followeeKey)

	if _, err := pipe.Exec(rdbCtx); err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	if followerTxExists.Val() == 1 || followeeTxExists.Val() == 1 {
		sublogger.Warn().
			Str("follower", strconv.FormatInt(followerTxExists.Val(), 10)).
			Str("followed", strconv.FormatInt(followeeTxExists.Val(), 10)).
			Msg("delete tx in progress")
		return nil, tools.GrpcError(codes.Canceled, "can't proceed, delete tx in progress")
	}

	psqlCtx, psqlCtxCancel := context.WithTimeout(ctx, psqlCtxTimeout)
	defer psqlCtxCancel()

	tx, err := pool.BeginTx(psqlCtx, pgx.TxOptions{})
	if err != nil {
		sublogger.Error().Err(err).Send()
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
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
			return nil, tools.GrpcError(codes.NotFound, fmt.Sprintf("user '%s' doesn't exist", in.Follower))
		}
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	strQuery = query("followers") + " returning tweets"
	var tweets []int64

	if err = tx.QueryRow(psqlCtx, strQuery, in.Follower, in.Followed).Scan(&tweets); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Follower).Str("$2", in.Followed).Msg(strQuery)
		if err == pgx.ErrNoRows {
			return nil, tools.GrpcError(codes.NotFound, fmt.Sprintf("user '%s' doesn't exist", in.Followed))
		}
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	newTimeline := tools.Difference(timeline, tweets)
	// sort by tweets id, cause greater id means that tweet was created later
	// id is stored in bigserial which is 2^63-1, so we don't have to worry about renewing it
	sort.Slice(newTimeline, func(i, j int) bool {
		return newTimeline[i] < newTimeline[j]
	})

	strQuery = fmt.Sprintf("update users set timeline = array[%s]::integer[] where username = $1",
		tools.IntArrayToString(newTimeline))

	if _, err = tx.Exec(psqlCtx, strQuery, in.Follower); err != nil {
		sublogger.Error().Err(err).Str("$1", in.Follower).Msg(strQuery)
		return nil, tools.GrpcError(codes.Internal, "INTERNAL ERROR")
	}

	if err = tx.Commit(psqlCtx); err != nil {
		sublogger.Error().Err(err).Send()
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
	hmgetRes, err := rdb.HMGet(rdbCtx, key, "followers", "following").Result()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	if hmgetRes == nil {
		log.Error().Msgf("transaction '%s' doesn't exist", key)
		return
	}

	followers := strings.Split(hmgetRes[0].(string), ",")
	following := strings.Split(hmgetRes[1].(string), ",")

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
			newTimeline := tools.Difference(timeline, msgProto.Tweets)
			// sort by tweets id, cause greater id means that tweet was created later
			// id is stored in bigserial which is 2^63-1, so we don't have to worry about renewing it
			sort.Slice(newTimeline, func(i, j int) bool {
				return newTimeline[i] < newTimeline[j]
			})

			queryBuilder.WriteString(fmt.Sprintf("('%s', array[%s]::integer[]),",
				username, tools.IntArrayToString(newTimeline)))
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

	strQuery = "delete from users where username = $1"

	if _, err := tx.Exec(psqlCtx, strQuery, msgProto.Username); err != nil {
		log.Error().Err(err).Str("$1", msgProto.Username).Msg(strQuery)
		return
	}

	if err = tx.Commit(psqlCtx); err != nil {
		log.Error().Err(err).Send()
		return
	}

	if err := rdb.Del(rdbCtx, key).Err(); err != nil {
		log.Error().Err(err).Send()
	}
}

func handleCreateTweet(ctx context.Context, serializedMsg []byte) {
	msgProto := &pb.NatsCreateTweetMessage{}
	if err := proto.Unmarshal(serializedMsg, msgProto); err != nil {
		log.Error().Err(err).Send()
		return
	}

	if msgProto.GetCreator() == "" {
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

	if err = tx.QueryRow(psqlCtx, query, msgProto.TweetId, msgProto.Creator).Scan(&followers); err != nil {
		log.Error().Err(err).
			Str("$1", strconv.FormatInt(msgProto.TweetId, 10)).
			Str("$2", msgProto.Creator).
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

	ncMsg := &pb.NatsMessage{Command: pb.NatsMessage_CreateTweet, Message: serializedMsg}
	msg, err := proto.Marshal(ncMsg)
	if err != nil {
		log.Error().Err(err).Str("protobuf message", "NatsMessage").Send()
		return
	}

	if err := nc.Publish("tweets", msg); err != nil {
		log.Error().Err(err).Str("nats command",
			pb.NatsMessage_Command_name[int32(pb.NatsMessage_CreateTweet)]).Send()
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
			username, tools.IntArrayToString(newTweets), tools.IntArrayToString(newTimeline)))
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
	case pb.NatsMessage_DeleteTweets:
		handleDeleteTweets(ctx, msg.Message)
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Logger = log.With().Caller().Logger()

	var configPath string
	flag.StringVar(&configPath, "config", "conf.json", "config file path")

	flag.Parse()

	err := tools.ParseConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if pool, err = pgxpool.Connect(context.Background(), tools.Conf.PostgresUrl); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to postgresql server")
	}

	if nc, err = nats.Connect(tools.Conf.NatsUrl,
		nats.UserInfo(tools.Conf.NatsUser, tools.Conf.NatsPassword)); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to nats server")
	}
	nc.QueueSubscribe("users", "tweets_queue", natsCallback)
	nc.Flush()
	defer nc.Close()

	rdb = redis.NewClient(&redis.Options{
		Addr:     tools.Conf.RedisUrl,
		PoolSize: tools.Conf.RedisPoolSize,
	})

	cert, err := tls.LoadX509KeyPair(data.Path("x509/server_cert.pem"), data.Path("x509/server_key.pem"))
	if err != nil {
		log.Fatal().Msgf("failed to load key pair: %s", err)
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(tools.EnsureValidToken),
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}

	s := grpc.NewServer(opts...)
	pb.RegisterUsersServer(s, &UsersServerImpl{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", tools.Conf.UsersPort))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}

	log.Info().Msgf("Start listening on: %d", tools.Conf.UsersPort)
	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve grpc service")
	}
}
