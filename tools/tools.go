package tools

import (
	"fmt"
	"strings"
)

//FIXME: move unused to end and then remove by slicing
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

func InterfaceToStringArray(data interface{}) []string {
	arr := data.([]interface{})
	result := make([]string, len(arr))
	for i, v := range arr {
		result[i] = v.(string)
	}
	return result
}

func IntArrayToString(arr []int64) string {
	return strings.Trim(strings.Replace(fmt.Sprint(arr), " ", ", ", -1), "[]")
}
