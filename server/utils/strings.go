package utils

import (
	"strconv"
	"strings"
)

func ParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func ParseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func TrimSpace(str string) string {
	return strings.TrimSpace(str)
}
