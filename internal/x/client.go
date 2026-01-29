package x

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"time"

	"github.com/yourname/xdl2/internal/config"
	"github.com/yourname/xdl2/internal/downloader"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
	dl         *downloader.Downloader

	cookies      *CookieJar
	ct0          string
	cookieHeader string
	queryIDs     QueryIDs
}

func NewClient(cfg *config.Config) *Client {
	jar, _ := cookiejar.New(nil)
	hc := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
	dl := downloader.New(cfg.UserAgent)
	return &Client{cfg: cfg, httpClient: hc, dl: dl}
}

func (c *Client) applyAuthHeaders(req *http.Request, referer string) {
	req.Header.Set("user-agent", c.cfg.UserAgent)
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")

	if referer != "" {
		req.Header.Set("referer", referer)
	}

	if c.cookieHeader != "" {
		req.Header.Set("cookie", c.cookieHeader)
	}

	if c.ct0 != "" {
		req.Header.Set("x-csrf-token", c.ct0)
	}

	req.Header.Set("authorization", DefaultBearerToken)
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-client-language", "en")

	if req.Header.Get("x-client-transaction-id") == "" {
		req.Header.Set("x-client-transaction-id", newClientTransactionID())
	}
}

func newClientTransactionID() string {
	buf := make([]byte, 70)
	_, _ = rand.Read(buf)
	return base64.RawStdEncoding.EncodeToString(buf)
}

func (c *Client) DownloadMedia(target string, items []MediaItem) error {
	outBase := filepath.Join(c.cfg.OutputDir, sanitizeTarget(target))
	if err := os.MkdirAll(outBase, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	for _, it := range items {
		outPath := filepath.Join(outBase, it.FileName)
		err := c.dl.DownloadToFile(it.URL, outPath)
		if err != nil {
			return err
		}
	}
	return nil
}
