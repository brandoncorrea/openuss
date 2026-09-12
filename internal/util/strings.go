package util

import "strings"

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func TrimString[T ~string](v T) string {
	return strings.TrimSpace(string(v))
}
