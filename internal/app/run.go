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

	isTweet := x.LooksLikeTweet(target)
	printBanner()
	statusf("Target : %s", target)
	if isTweet {
		statusf("Mode   : single tweet")
	} else {
		statusf("Mode   : profile media")
	}

	loadConfig := config.LoadDefaultRequireCookies
	if isTweet {
		loadConfig = config.LoadDefault
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	statusf("Output : %s", cfg.OutputDir)

	client := x.NewClient(cfg)

	// Auth is optional for single-tweet mode (public fallback exists), but required for profile mode.
	if cfg.CookiesPath != "" {
		if err := client.InitAuthenticatedSession(); err != nil {
			if !isTweet {
				return err
			}
			statusf("Auth   : public fallback")
			// Tweet mode can continue without auth and will fall back to public fetch.
		} else {
			statusf("Auth   : cookies loaded")
		}
	} else if isTweet {
		statusf("Auth   : public mode")
	}

	if isTweet {
		items, err := client.FetchTweetMedia(target)
		if err != nil {
			return err
		}
		statusf("Media  : %d item(s)", len(items))
		return client.DownloadMedia(target, items)
	}

	items, err := client.FetchUserProfileMedia(target, 0)
	if err != nil {
		return err
	}
	return client.DownloadMedia(target, items)
}
