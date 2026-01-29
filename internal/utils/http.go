package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type cancelOnCloseReadCloser struct {
	rc     io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnCloseReadCloser) Read(p []byte) (int, error) {
	return c.rc.Read(p)
}

func (c *cancelOnCloseReadCloser) Close() error {
	err := c.rc.Close()
	c.cancel()
	return err
}

// DoWithTimeout executes an HTTP request with a timeout.
// The request context is canceled when the response body is closed to avoid premature "context canceled" during body reads.
func DoWithTimeout(client *http.Client, req *http.Request, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), timeout)
	req = req.WithContext(ctx)

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if debugEnabled() {
		safeURL := sanitizeURL(req.URL)
		if resp != nil {
			fmt.Printf(
				"HTTP %s %s -> %d in %s (timeout %s), proto=%s, ct=%s, rl_rem=%s, rl_reset=%s\n",
				req.Method,
				safeURL,
				resp.StatusCode,
				elapsed,
				timeout,
				resp.Proto,
				resp.Header.Get("content-type"),
				resp.Header.Get("x-rate-limit-remaining"),
				resp.Header.Get("x-rate-limit-reset"),
			)
		} else {
			fmt.Printf(
				"HTTP %s %s failed in %s (timeout %s), err=%v\n",
				req.Method,
				safeURL,
				elapsed,
				timeout,
				err,
			)
		}
	}

	if err != nil {
		cancel()
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		cancel()
		return nil, fmt.Errorf("empty response")
	}

	resp.Body = &cancelOnCloseReadCloser{rc: resp.Body, cancel: cancel}
	return resp, nil
}

func ReadAllWithLimit(r io.Reader, limit int64) ([]byte, error) {
	lr := &io.LimitedReader{R: r, N: limit}
	b, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func sanitizeURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	uu := *u
	q := uu.Query()
	for _, k := range []string{"variables", "features", "fieldToggles"} {
		if q.Has(k) {
			q.Set(k, "<redacted>")
		}
	}
	uu.RawQuery = q.Encode()
	return uu.String()
}

func debugEnabled() bool {
	return strings.TrimSpace(os.Getenv("XDL_DEBUG")) != ""
}
