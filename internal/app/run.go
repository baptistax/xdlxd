package app

import (
	"fmt"
	"strings"

	"github.com/yourname/xdl2/internal/config"
	"github.com/yourname/xdl2/internal/x"
)

func Run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: xdl <tweet_url_or_id|username>")
	}

	target := strings.TrimSpace(args[1])
	if target == "" {
		return fmt.Errorf("empty target")
	}

	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}

	client := x.NewClient(cfg)

	// Auth is optional for single-tweet mode (public fallback exists), but required for profile mode.
	if err := client.InitAuthenticatedSession(); err != nil {
		if !x.LooksLikeTweet(target) {
			return err
		}
		// Tweet mode can continue without auth and will fall back to public fetch.
	}

	if x.LooksLikeTweet(target) {
		items, err := client.FetchTweetMedia(target)
		if err != nil {
			return err
		}
		return client.DownloadMedia(target, items)
	}

	items, err := client.FetchUserProfileMedia(target, 0)
	if err != nil {
		return err
	}
	return client.DownloadMedia(target, items)
}
