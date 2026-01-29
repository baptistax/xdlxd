package x

import (
	"net/url"
	"path"
	"strings"
)

func bestMP4VariantURL(variants []any) string {
	var bestURL string
	bestBitrate := float64(-1)

	for _, v := range variants {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}

		contentType, _ := m["content_type"].(string)
		if !strings.Contains(contentType, "mp4") {
			continue
		}

		urlStr, _ := m["url"].(string)
		if urlStr == "" {
			continue
		}

		br, _ := m["bitrate"].(float64)
		if br > bestBitrate {
			bestBitrate = br
			bestURL = urlStr
		}
	}

	return bestURL
}

func normalizePhotoURL(raw string) (string, string) {
	if raw == "" {
		return "", ""
	}

	u, err := url.Parse(raw)
	if err != nil {
		ext := path.Ext(strings.Split(raw, "?")[0])
		if ext == "" {
			ext = ".jpg"
		}
		return raw, ext
	}

	q := u.Query()
	if q.Get("name") == "" {
		q.Set("name", "orig")
	}

	ext := path.Ext(u.Path)
	format := q.Get("format")
	if format == "" {
		if ext != "" {
			format = strings.TrimPrefix(ext, ".")
		} else {
			format = "jpg"
		}
		q.Set("format", format)
	}

	if ext == "" {
		ext = "." + format
	}

	u.RawQuery = q.Encode()
	return u.String(), ext
}
