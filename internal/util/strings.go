package util

import "strings"

func RepeatJoin(s string, count int, sep string) string {
	return strings.Repeat(s+sep, count-1) + s
}
