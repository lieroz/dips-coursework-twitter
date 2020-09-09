package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

type Config struct {
	RedisUrl          string `json:"redis_url"`
	RedisPoolSize     int    `json:"redis_pool_size"`
	PostgresUrl       string `json:"postgres_url"`
	NatsUrl           string `json:"nats_url"`
	NatsUser          string `json:"nats_user"`
	NatsPassword      string `json:"nats_password"`
	AuthPort          int    `json:"auth_port"`
	GatewayPort       int    `json:"gateway_port"`
	UsersPort         int    `json:"users_port"`
	TweetsPort        int    `json:"tweets_port"`
	AuthServiceHost   string `json:"auth_service_host"`
	UsersServiceHost  string `json:"users_service_host"`
	TweetsServiceHost string `json:"tweets_service_host"`
}

var Conf Config

func ParseConfig(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	config := Config{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	Conf = config
	return err
}

func Difference(a, b []int64) (diff []int64) {
	m := make(map[int64]bool)

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

func IntArrayToString(arr []int64) string {
	return strings.Trim(strings.Replace(fmt.Sprint(arr), " ", ", ", -1), "[]")
}

func GrpcError(code codes.Code, msg string) error {
	return status.Error(code, msg)
}

var authToken string

func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")

	r, err := http.Get(fmt.Sprintf("http://%s:%d/service/token", Conf.AuthServiceHost, Conf.AuthPort))
	if err != nil {
		return false
	}

	if r.StatusCode == http.StatusCreated || r.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return false
		}
		authToken = string(body)
	}

	return token == authToken
}

func EnsureValidToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, GrpcError(codes.InvalidArgument, "missing metadata")
	}
	if !valid(md["authorization"]) {
		return nil, GrpcError(codes.Unauthenticated, "invalid token")
	}
	return handler(ctx, req)
}
