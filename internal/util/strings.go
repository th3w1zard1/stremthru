package util

import (
	"strconv"
	"strings"
)

func RepeatJoin(s string, count int, sep string) string {
	if count == 0 {
		return ""
	}
	return strings.Repeat(s+sep, count-1) + s
}

func MustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
