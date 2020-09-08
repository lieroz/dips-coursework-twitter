package tools

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

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

func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	return token == "some-secret-token"
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
