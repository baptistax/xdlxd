package x

import (
	"regexp"
	"strings"
)

var tweetURLRe = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:x\.com|twitter\.com)/[^/]+/status/(\d+)`)
var tweetIDRe = regexp.MustCompile(`^\d{5,}$`)

func LooksLikeTweet(s string) bool {
	s = strings.TrimSpace(s)
	if tweetIDRe.MatchString(s) {
		return true
	}
	return tweetURLRe.MatchString(s)
}

func ExtractTweetID(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if tweetIDRe.MatchString(s) {
		return s, true
	}
	m := tweetURLRe.FindStringSubmatch(s)
	if len(m) == 2 {
		return m[1], true
	}
	return "", false
}

func sanitizeTarget(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "https://", "")
	s = strings.ReplaceAll(s, "http://", "")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "&", "_")
	return s
}
