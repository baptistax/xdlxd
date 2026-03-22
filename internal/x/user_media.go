package x

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
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

	all := make([]MediaItem, 0, 256)
	seenURLs := map[string]bool{}
	seenCursors := map[string]bool{"": true}
	progress := &progressLine{}

	cursor := ""
	stagnantPages := 0

	for page := 1; shouldFetchProfilePage(page, maxPages); page++ {
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
			progress.Finish("Scan failed")
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
		progress.Update("Scan page %d | total %d | +%d new", page, len(all), added)

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

	progress.Finish("Scan complete | %d item(s) found", len(all))
	return all, nil
}

func shouldFetchProfilePage(page int, maxPages int) bool {
	return maxPages <= 0 || page <= maxPages
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
	normalizedScreenName := normalizeScreenName(screenName)
	endpoint := fmt.Sprintf("https://x.com/i/api/graphql/%s/UserByScreenName", c.queryIDs.UserByScreenName)
	variables := map[string]any{
		"screen_name":           normalizedScreenName,
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
	c.applyAuthHeaders(req, "https://x.com/"+normalizedScreenName)

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

	userID := extractUserRestID(payload, normalizedScreenName)
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
	nextCursor := extractNextCursor(payload)

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
	return s[:18] + "..."
}

func normalizeScreenName(screenName string) string {
	return strings.TrimPrefix(strings.TrimSpace(screenName), "@")
}

func extractUserRestID(payload map[string]any, screenName string) string {
	if payload == nil {
		return ""
	}

	normalizedScreenName := normalizeScreenName(screenName)
	for _, path := range [][]string{
		{"data", "user", "result"},
		{"data", "user", "result", "result"},
		{"user", "result"},
		{"user", "result", "result"},
	} {
		if node := lookupMapPath(payload, path...); node != nil {
			if restID := restIDFromUserNode(node, normalizedScreenName); restID != "" {
				return restID
			}
		}
	}

	bestRestID := ""
	bestScore := -1

	var walk func(any)
	walk = func(obj any) {
		switch v := obj.(type) {
		case map[string]any:
			if score, restID := scoreUserNode(v, normalizedScreenName); score > bestScore {
				bestScore = score
				bestRestID = restID
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

	walk(payload)
	return bestRestID
}

func lookupMapPath(root map[string]any, keys ...string) map[string]any {
	current := root
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

func restIDFromUserNode(node map[string]any, screenName string) string {
	if score, restID := scoreUserNode(node, screenName); score >= 0 {
		return restID
	}
	if result, ok := node["result"].(map[string]any); ok {
		if score, restID := scoreUserNode(result, screenName); score >= 0 {
			return restID
		}
	}
	return ""
}

func scoreUserNode(node map[string]any, screenName string) (int, string) {
	if node == nil {
		return -1, ""
	}

	restID, _ := node["rest_id"].(string)
	if restID == "" {
		return -1, ""
	}

	score := 0
	if typename, _ := node["__typename"].(string); strings.EqualFold(typename, "User") {
		score += 4
	}

	if legacy, ok := node["legacy"].(map[string]any); ok {
		score++
		score += scoreScreenNameMatch(legacy["screen_name"], screenName)
	} else {
		score += scoreScreenNameMatch(node["screen_name"], screenName)
	}

	if score < 0 {
		return -1, ""
	}

	return score, restID
}

func scoreScreenNameMatch(raw any, expected string) int {
	if expected == "" {
		return 0
	}

	screenName, _ := raw.(string)
	if screenName == "" {
		return 0
	}

	if strings.EqualFold(screenName, expected) {
		return 10
	}

	return -5
}

func extractNextCursor(payload map[string]any) string {
	candidates := collectTimelineCursorCandidates(payload)
	if len(candidates) == 0 {
		return ""
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].order > candidates[j].order
		}
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].value
}

type timelineCursorCandidate struct {
	value string
	score int
	order int
}

func collectTimelineCursorCandidates(obj any) []timelineCursorCandidate {
	candidates := make([]timelineCursorCandidate, 0, 8)
	order := 0

	addCandidate := func(candidate timelineCursorCandidate) {
		candidate.order = order
		order++
		candidates = append(candidates, candidate)
	}

	var walk func(any)
	walk = func(obj any) {
		switch v := obj.(type) {
		case map[string]any:
			if instructions, ok := v["instructions"].([]any); ok {
				for _, instructionAny := range instructions {
					instruction, _ := instructionAny.(map[string]any)
					if instruction == nil {
						continue
					}

					if entries, ok := instruction["entries"].([]any); ok {
						for _, entryAny := range entries {
							entry, _ := entryAny.(map[string]any)
							if candidate, ok := timelineCursorFromEntry(entry); ok {
								addCandidate(candidate)
							}
						}
					}

					if entry, ok := instruction["entry"].(map[string]any); ok {
						if candidate, ok := timelineCursorFromEntry(entry); ok {
							addCandidate(candidate)
						}
					}
				}
			}

			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}

	walk(obj)
	if len(candidates) > 0 {
		return candidates
	}

	order = 0
	var fallbackWalk func(any)
	fallbackWalk = func(obj any) {
		switch v := obj.(type) {
		case map[string]any:
			if candidate, ok := timelineCursorFromMap(v); ok {
				addCandidate(candidate)
			}
			for _, child := range v {
				fallbackWalk(child)
			}
		case []any:
			for _, child := range v {
				fallbackWalk(child)
			}
		}
	}

	fallbackWalk(obj)
	return candidates
}

func timelineCursorFromEntry(entry map[string]any) (timelineCursorCandidate, bool) {
	if entry == nil {
		return timelineCursorCandidate{}, false
	}

	entryID, _ := entry["entryId"].(string)
	content, _ := entry["content"].(map[string]any)
	if content == nil {
		return timelineCursorFromMap(entry)
	}

	candidate, ok := timelineCursorFromMap(content)
	if !ok {
		return timelineCursorCandidate{}, false
	}

	entryIDLower := strings.ToLower(entryID)
	if strings.Contains(entryIDLower, "cursor-bottom") {
		candidate.score += 25
	} else if strings.Contains(entryIDLower, "bottom") {
		candidate.score += 10
	}

	return candidate, true
}

func timelineCursorFromMap(node map[string]any) (timelineCursorCandidate, bool) {
	if node == nil {
		return timelineCursorCandidate{}, false
	}

	entryType, _ := node["entryType"].(string)
	cursorType, _ := node["cursorType"].(string)
	value, _ := node["value"].(string)

	if itemContent, ok := node["itemContent"].(map[string]any); ok {
		if entryType == "" {
			entryType, _ = itemContent["entryType"].(string)
		}
		if cursorType == "" {
			cursorType, _ = itemContent["cursorType"].(string)
		}
		if value == "" {
			value, _ = itemContent["value"].(string)
		}
	}

	if value == "" || cursorType == "" {
		return timelineCursorCandidate{}, false
	}

	score := 0
	if strings.EqualFold(entryType, "TimelineTimelineCursor") {
		score += 40
	}

	switch {
	case strings.EqualFold(cursorType, "Bottom"):
		score += 100
	case strings.EqualFold(cursorType, "ShowMore"), strings.EqualFold(cursorType, "ShowMoreThreads"):
		score += 60
	case strings.EqualFold(cursorType, "Top"):
		score += 10
	default:
		score += 20
	}

	return timelineCursorCandidate{
		value: value,
		score: score,
	}, true
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
