package x

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yourname/xdl2/internal/utils"
)

type PublicTweetResponse struct {
	MediaDetails []struct {
		Type          string `json:"type"`
		MediaURLHTTPS string `json:"media_url_https"`
		VideoInfo     struct {
			Variants []struct {
				Bitrate     int    `json:"bitrate,omitempty"`
				ContentType string `json:"content_type"`
				URL         string `json:"url"`
			} `json:"variants"`
		} `json:"video_info"`
		AltText string `json:"ext_alt_text"`
	} `json:"mediaDetails"`
}

func (c *Client) FetchPublicTweetMedia(tweet string) ([]MediaItem, error) {
	tweetID, ok := ExtractTweetID(tweet)
	if !ok {
		return nil, fmt.Errorf("invalid tweet id/url: %s", tweet)
	}

	endpoint := fmt.Sprintf("https://cdn.syndication.twimg.com/tweet-result?id=%s&lang=en", tweetID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", c.cfg.UserAgent)

	resp, err := utils.DoWithTimeout(c.httpClient, req, 20*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := utils.ReadAllWithLimit(resp.Body, 64*1024)
		return nil, fmt.Errorf("public tweet endpoint status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	body, err := utils.ReadAllWithLimit(resp.Body, 4*1024*1024)
	if err != nil {
		return nil, err
	}

	var parsed PublicTweetResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode public tweet json: %w", err)
	}

	items := make([]MediaItem, 0)
	photoIndex := 0
	videoIndex := 0

	for _, md := range parsed.MediaDetails {
		switch md.Type {
		case "photo":
			urlStr, ext := normalizePhotoURL(md.MediaURLHTTPS)
			if urlStr == "" {
				continue
			}
			photoIndex++
			items = append(items, MediaItem{URL: urlStr, FileName: fmt.Sprintf("photo_%02d%s", photoIndex, ext)})

		case "video", "animated_gif":
			variants := make([]any, 0, len(md.VideoInfo.Variants))
			for _, v := range md.VideoInfo.Variants {
				variants = append(variants, map[string]any{
					"bitrate":      float64(v.Bitrate),
					"content_type": v.ContentType,
					"url":          v.URL,
				})
			}
			best := bestMP4VariantURL(variants)
			if best == "" {
				continue
			}
			videoIndex++
			name := "video.mp4"
			if videoIndex > 1 {
				name = fmt.Sprintf("video_%02d.mp4", videoIndex)
			}
			items = append(items, MediaItem{URL: best, FileName: name})
		}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no media found for tweet %s", tweetID)
	}
	return items, nil
}
