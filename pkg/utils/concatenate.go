package utils

import "strings"

func ConcatStrings(parts ...string) string {
	var sb strings.Builder
	sz := 0
	for _, part := range parts {
		sz += len(part)
	}
	sb.Grow(sz)
	for _, part := range parts {
		if len(part) != 0 {
			sb.WriteString(part)
		}
	}
	return sb.String()
}

func ConcatSlices[T any](start []T, slices ...[]T) []T {
	r := start
	for _, slice := range slices {
		r = append(r, slice...)
	}
	return r
}

func BuildString(sb *strings.Builder, parts ...string) {
	for _, part := range parts {
		sb.WriteString(part)
	}
}

func Substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}
