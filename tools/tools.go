package tools

import (
	"fmt"
	"strings"

	codes "google.golang.org/grpc/codes"
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
