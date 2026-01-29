package x

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yourname/xdl2/internal/utils"
)

func (c *Client) FetchTweetMedia(tweet string) ([]MediaItem, error) {
	if c.cookieHeader == "" || c.ct0 == "" {
		return c.FetchPublicTweetMedia(tweet)
	}

	items, errDetail := c.FetchTweetDetailMedia(tweet)
	if errDetail == nil && len(items) > 0 {
		return items, nil
	}

	publicItems, errPublic := c.FetchPublicTweetMedia(tweet)
	if errPublic == nil && len(publicItems) > 0 {
		return publicItems, nil
	}

	if errDetail != nil && errPublic != nil {
		return nil, fmt.Errorf("TweetDetail failed: %v; public fetch failed: %v", errDetail, errPublic)
	}
	if errDetail != nil {
		return nil, errDetail
	}
	return nil, errPublic
}

func (c *Client) FetchTweetDetailMedia(tweet string) ([]MediaItem, error) {
	tweetID, ok := ExtractTweetID(tweet)
	if !ok {
		return nil, fmt.Errorf("invalid tweet id/url: %s", tweet)
	}

	variables := map[string]any{
		"focalTweetId":                           tweetID,
		"rankingMode":                            "Relevance",
		"includePromotedContent":                 true,
		"withCommunity":                          true,
		"withQuickPromoteEligibilityTweetFields": true,
		"withBirdwatchNotes":                     true,
		"withVoice":                              true,
	}

	features := defaultFeaturesTweetAndUserMedia()

	fieldToggles := defaultFieldTogglesTweetDetail()

	endpoint := fmt.Sprintf("https://x.com/i/api/graphql/%s/TweetDetail", c.queryIDs.TweetDetail)
	reqURL, err := buildGraphQLURL(endpoint, variables, features, fieldToggles)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	c.applyAuthHeaders(req, "https://x.com/i/status/"+tweetID)

	resp, err := utils.DoWithTimeout(c.httpClient, req, 25*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := utils.ReadAllWithLimit(resp.Body, 128*1024)
		return nil, fmt.Errorf("TweetDetail status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	body, err := utils.ReadAllWithLimit(resp.Body, 8*1024*1024)
	if err != nil {
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode TweetDetail json: %w", err)
	}

	tweetNode := findNodeWithRestID(payload, tweetID)
	if tweetNode == nil {
		return nil, fmt.Errorf("TweetDetail: tweet not found in response (%s)", tweetID)
	}

	legacy, _ := tweetNode["legacy"].(map[string]any)
	if legacy == nil {
		return nil, fmt.Errorf("TweetDetail: legacy node not found (%s)", tweetID)
	}

	items := make([]MediaItem, 0)
	seen := map[string]bool{}

	photoIndex := 0
	videoIndex := 0

	extractMediaList := func(mediaList []any) {
		for _, m := range mediaList {
			mm, _ := m.(map[string]any)
			if mm == nil {
				continue
			}
			mediaType, _ := mm["type"].(string)
			switch mediaType {
			case "photo":
				rawURL, _ := mm["media_url_https"].(string)
				bestURL, ext := normalizePhotoURL(rawURL)
				if bestURL == "" {
					continue
				}
				if seen[bestURL] {
					continue
				}
				seen[bestURL] = true
				photoIndex++
				items = append(items, MediaItem{URL: bestURL, FileName: fmt.Sprintf("photo_%02d%s", photoIndex, ext)})

			case "video", "animated_gif":
				vi, _ := mm["video_info"].(map[string]any)
				if vi == nil {
					continue
				}
				variants, _ := vi["variants"].([]any)
				best := bestMP4VariantURL(variants)
				if best == "" {
					continue
				}
				if seen[best] {
					continue
				}
				seen[best] = true
				videoIndex++
				name := "video.mp4"
				if videoIndex > 1 {
					name = fmt.Sprintf("video_%02d.mp4", videoIndex)
				}
				items = append(items, MediaItem{URL: best, FileName: name})
			}
		}
	}

	if ee, ok := legacy["extended_entities"].(map[string]any); ok {
		if media, ok := ee["media"].([]any); ok {
			extractMediaList(media)
		}
	}

	if len(items) == 0 {
		if ent, ok := legacy["entities"].(map[string]any); ok {
			if media, ok := ent["media"].([]any); ok {
				extractMediaList(media)
			}
		}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("TweetDetail: no media found (%s)", tweetID)
	}
	return items, nil
}

func findNodeWithRestID(obj any, restID string) map[string]any {
	switch v := obj.(type) {
	case map[string]any:
		if rid, ok := v["rest_id"].(string); ok && rid == restID {
			return v
		}
		for _, vv := range v {
			if found := findNodeWithRestID(vv, restID); found != nil {
				return found
			}
		}
	case []any:
		for _, vv := range v {
			if found := findNodeWithRestID(vv, restID); found != nil {
				return found
			}
		}
	}
	return nil
}

func buildGraphQLURL(endpoint string, variables, features, fieldToggles map[string]any) (string, error) {
	vb, err := json.Marshal(variables)
	if err != nil {
		return "", err
	}
	fb, err := json.Marshal(features)
	if err != nil {
		return "", err
	}
	tb, err := json.Marshal(fieldToggles)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("variables", string(vb))
	q.Set("features", string(fb))
	q.Set("fieldToggles", string(tb))
	u.RawQuery = q.Encode()
	return u.String(), nil
}
