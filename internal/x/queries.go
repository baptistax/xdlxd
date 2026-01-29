package x

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type QueryIDs struct {
	TweetDetail      string `json:"TweetDetail"`
	UserByScreenName string `json:"UserByScreenName"`
	UserMedia        string `json:"UserMedia"`
}

func LoadQueryIDsOrDefault() QueryIDs {
	ids := QueryIDs{
		TweetDetail:      "Kzfv17rukSzjT96BerOWZA",
		UserByScreenName: "-oaLodhGbbnzJBACb1kk2Q",
		UserMedia:        "8HCIrWwy4C0fBTbPnMq5aA",
	}

	cwd, err := os.Getwd()
	if err != nil {
		return ids
	}

	p := filepath.Join(cwd, "queries.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return ids
	}

	var override QueryIDs
	if err := json.Unmarshal(b, &override); err != nil {
		return ids
	}

	if override.TweetDetail != "" {
		ids.TweetDetail = override.TweetDetail
	}
	if override.UserByScreenName != "" {
		ids.UserByScreenName = override.UserByScreenName
	}
	if override.UserMedia != "" {
		ids.UserMedia = override.UserMedia
	}
	return ids
}

const DefaultBearerToken = "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs=1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
