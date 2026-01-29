package utils

import (
	"regexp"
	"strings"
)

var invalidFileChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)

func SanitizeFileName(name string) string {
	n := strings.TrimSpace(name)
	n = invalidFileChars.ReplaceAllString(n, "_")
	n = strings.Trim(n, ". ")
	if n == "" {
		return "file"
	}
	return n
}
