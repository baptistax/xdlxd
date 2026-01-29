package x

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yourname/xdl2/internal/utils"
)

// FetchUserProfileMedia downloads as much as the UserMedia timeline will provide.
// If maxPages <= 0, it will keep paging until stop-conditions are met (cursor end / cursor loop / no progress).
func (c *Client) FetchUserProfileMedia(screenName string, maxPages int) ([]MediaItem, error) {
	if c.cookieHeader == "" || c.ct0 == "" {
		return nil, fmt.Errorf("authenticated session not initialized")
	}

	userID, err := c.ResolveUserIDByScreenName(screenName)
	if err != nil {
		return nil, err
	}

	const hardMaxPages = 200
	if maxPages <= 0 {
		maxPages = hardMaxPages
	}

	all := make([]MediaItem, 0, 256)
	seenURLs := map[string]bool{}
	seenCursors := map[string]bool{"": true}

	cursor := ""
	stagnantPages := 0

	for page := 1; page <= maxPages; page++ {
		if cursor != "" {
			if seenCursors[cursor] {
				if debugEnabled() {
					fmt.Printf("profile: stop (cursor loop) at page %d, cursor=%s\n", page, short(cursor))
				}
				break
			}
			seenCursors[cursor] = true
		}

		items, nextCursor, meta, err := c.FetchUserMediaPage(userID, cursor, 20)
		if err != nil {
			return nil, err
		}

		added := 0
		for _, it := range items {
			if it.URL == "" || it.FileName == "" {
				continue
			}
			if seenURLs[it.URL] {
				continue
			}
			seenURLs[it.URL] = true
			all = append(all, it)
			added++
		}

		if added == 0 {
			stagnantPages++
		} else {
			stagnantPages = 0
		}

		if debugEnabled() {
			fmt.Printf(
				"profile: page %d, +%d, total %d, stagnant %d, cursor=%s, next=%s, status=%d, bytes=%d, rl_rem=%s\n",
				page, added, len(all), stagnantPages,
				short(cursor), short(nextCursor),
				meta.StatusCode, meta.BodyBytes, meta.RateLimitRemaining,
			)
		}

		if nextCursor == "" {
			if debugEnabled() {
				fmt.Printf("profile: stop (no next cursor) at page %d\n", page)
			}
			break
		}
		if nextCursor == cursor {
			if debugEnabled() {
				fmt.Printf("profile: stop (cursor did not advance) at page %d, cursor=%s\n", page, short(cursor))
			}
			break
		}
		if seenCursors[nextCursor] {
			if debugEnabled() {
				fmt.Printf("profile: stop (next cursor already seen) at page %d, next=%s\n", page, short(nextCursor))
			}
			break
		}
		if stagnantPages >= 3 {
			if debugEnabled() {
				fmt.Printf("profile: stop (no progress for 3 pages) at page %d\n", page)
			}
			break
		}

		cursor = nextCursor
		time.Sleep(200 * time.Millisecond)
	}

	return all, nil
}

type PageMeta struct {
	StatusCode          int
	BodyBytes           int
	RateLimitRemaining  string
	RateLimitReset      string
	RateLimitLimit      string
	ResponseContentType string
	ResponseProto       string
}

func (c *Client) ResolveUserIDByScreenName(screenName string) (string, error) {
	endpoint := fmt.Sprintf("https://x.com/i/api/graphql/%s/UserByScreenName", c.queryIDs.UserByScreenName)
	variables := map[string]any{
		"screen_name":           strings.TrimPrefix(strings.TrimSpace(screenName), "@"),
		"withGrokTranslatedBio": false,
	}
	features := defaultFeaturesUserByScreenName()
	fieldToggles := defaultFieldTogglesUserByScreenName()

	reqURL, err := buildGraphQLURL(endpoint, variables, features, fieldToggles)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	c.applyAuthHeaders(req, "https://x.com/"+strings.TrimPrefix(screenName, "@"))

	resp, err := utils.DoWithTimeout(c.httpClient, req, 25*time.Second)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := utils.ReadAllWithLimit(resp.Body, 128*1024)
		return "", fmt.Errorf("UserByScreenName status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	body, err := utils.ReadAllWithLimit(resp.Body, 6*1024*1024)
	if err != nil {
		return "", err
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode UserByScreenName json: %w", err)
	}

	userID := findFirstStringValue(payload, "rest_id")
	if userID == "" {
		return "", fmt.Errorf("UserByScreenName: rest_id not found for %s", screenName)
	}
	return userID, nil
}

func (c *Client) FetchUserMediaPage(userID string, cursor string, count int) ([]MediaItem, string, PageMeta, error) {
	if count <= 0 || count > 20 {
		count = 20
	}

	endpoint := fmt.Sprintf("https://x.com/i/api/graphql/%s/UserMedia", c.queryIDs.UserMedia)
	variables := map[string]any{
		"userId":                 userID,
		"count":                  count,
		"includePromotedContent": false,
		"withClientEventToken":   false,
		"withBirdwatchNotes":     false,
		"withVoice":              true,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	features := defaultFeaturesTweetAndUserMedia()
	fieldToggles := defaultFieldTogglesUserMedia()

	reqURL, err := buildGraphQLURL(endpoint, variables, features, fieldToggles)
	if err != nil {
		return nil, "", PageMeta{}, err
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, "", PageMeta{}, err
	}
	c.applyAuthHeaders(req, "https://x.com/i/user/"+userID+"/media")

	resp, err := utils.DoWithTimeout(c.httpClient, req, 25*time.Second)
	if err != nil {
		return nil, "", PageMeta{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := utils.ReadAllWithLimit(resp.Body, 256*1024)
		return nil, "", PageMeta{}, fmt.Errorf("UserMedia status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	body, err := utils.ReadAllWithLimit(resp.Body, 25*1024*1024)
	if err != nil {
		return nil, "", PageMeta{}, err
	}

	meta := PageMeta{
		StatusCode:          resp.StatusCode,
		BodyBytes:           len(body),
		RateLimitRemaining:  resp.Header.Get("x-rate-limit-remaining"),
		RateLimitReset:      resp.Header.Get("x-rate-limit-reset"),
		RateLimitLimit:      resp.Header.Get("x-rate-limit-limit"),
		ResponseContentType: resp.Header.Get("content-type"),
		ResponseProto:       resp.Proto,
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, "", meta, fmt.Errorf("decode UserMedia json: %w", err)
	}

	items := extractMediaItemsFromAny(payload)
	nextCursor := extractBottomCursor(payload)
	if nextCursor == "" {
		nextCursor = extractAnyCursor(payload)
	}

	return items, nextCursor, meta, nil
}

func debugEnabled() bool {
	return strings.TrimSpace(os.Getenv("XDL_DEBUG")) != ""
}

func short(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	if len(s) <= 18 {
		return s
	}
	return s[:18] + "…"
}

func findFirstStringValue(obj any, key string) string {
	switch v := obj.(type) {
	case map[string]any:
		if s, ok := v[key].(string); ok && s != "" {
			return s
		}
		for _, vv := range v {
			if found := findFirstStringValue(vv, key); found != "" {
				return found
			}
		}
	case []any:
		for _, vv := range v {
			if found := findFirstStringValue(vv, key); found != "" {
				return found
			}
		}
	}
	return ""
}

func extractBottomCursor(obj any) string {
	last := ""
	var walk func(any)
	walk = func(o any) {
		switch v := o.(type) {
		case map[string]any:
			if ct, ok := v["cursorType"].(string); ok && strings.EqualFold(ct, "Bottom") {
				if val, ok := v["value"].(string); ok && val != "" {
					last = val
				}
			}
			for _, vv := range v {
				walk(vv)
			}
		case []any:
			for _, vv := range v {
				walk(vv)
			}
		}
	}
	walk(obj)
	return last
}

func extractAnyCursor(obj any) string {
	var found string
	var walk func(any)
	walk = func(o any) {
		if found != "" {
			return
		}
		switch v := o.(type) {
		case map[string]any:
			for k, vv := range v {
				if strings.Contains(strings.ToLower(k), "cursor") {
					if s, ok := vv.(string); ok && s != "" {
						found = s
						return
					}
				}
			}
			for _, vv := range v {
				walk(vv)
			}
		case []any:
			for _, vv := range v {
				walk(vv)
			}
		}
	}
	walk(obj)
	return found
}

// extractMediaItemsFromAny mirrors legacy xdl behavior:
// walk the full JSON tree and collect nodes that contain "media_url_https".
func extractMediaItemsFromAny(root any) []MediaItem {
	out := make([]MediaItem, 0, 128)
	seenURL := map[string]bool{}
	photoIndex := map[string]int{}
	videoIndex := map[string]int{}

	var walk func(any, string)
	walk = func(v any, currentTweetID string) {
		switch t := v.(type) {
		case map[string]any:
			if id, ok := t["rest_id"].(string); ok && id != "" {
				currentTweetID = id
			}
			if currentTweetID == "" {
				currentTweetID = "unknown"
			}

			rawURL, _ := t["media_url_https"].(string)
			if rawURL == "" {
				if u2, _ := t["media_url"].(string); u2 != "" {
					rawURL = u2
				}
			}

			if rawURL != "" {
				mediaType := "photo"
				if tp, ok := t["type"].(string); ok && tp != "" {
					mediaType = strings.ToLower(tp)
				}

				switch mediaType {
				case "photo":
					bestURL, ext := normalizePhotoURL(rawURL)
					if bestURL != "" && !seenURL[bestURL] {
						seenURL[bestURL] = true
						photoIndex[currentTweetID]++
						out = append(out, MediaItem{
							URL:      bestURL,
							FileName: fmt.Sprintf("%s_photo_%02d%s", currentTweetID, photoIndex[currentTweetID], ext),
						})
					}

				case "video", "animated_gif":
					if vi, ok := t["video_info"].(map[string]any); ok {
						if variants, ok := vi["variants"].([]any); ok {
							best := bestMP4VariantURL(variants)
							if best != "" && !seenURL[best] {
								seenURL[best] = true
								videoIndex[currentTweetID]++
								out = append(out, MediaItem{
									URL:      best,
									FileName: fmt.Sprintf("%s_video_%02d.mp4", currentTweetID, videoIndex[currentTweetID]),
								})
							}
						}
					}
				default:
					bestURL, ext := normalizePhotoURL(rawURL)
					if bestURL != "" && !seenURL[bestURL] {
						seenURL[bestURL] = true
						photoIndex[currentTweetID]++
						out = append(out, MediaItem{
							URL:      bestURL,
							FileName: fmt.Sprintf("%s_photo_%02d%s", currentTweetID, photoIndex[currentTweetID], ext),
						})
					}
				}
			}

			for _, child := range t {
				walk(child, currentTweetID)
			}

		case []any:
			for _, child := range t {
				walk(child, currentTweetID)
			}
		}
	}

	walk(root, "")
	return out
}
