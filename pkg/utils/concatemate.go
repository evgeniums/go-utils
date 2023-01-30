package utils

import "strings"

func ConcatStrings(parts ...string) string {
	var sb strings.Builder
	for _, part := range parts {
		sb.WriteString(part)
	}
	return sb.String()
}
