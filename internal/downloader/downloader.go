package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/yourname/xdl2/internal/utils"
)

type Downloader struct {
	client    *http.Client
	userAgent string
}

func New(userAgent string) *Downloader {
	return &Downloader{
		client:    &http.Client{Timeout: 60 * time.Second},
		userAgent: userAgent,
	}
}

func (d *Downloader) DownloadToFile(url, outPath string) error {
	if _, err := os.Stat(outPath); err == nil {
		return nil
	}

	tmpPath := outPath + ".part"
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("user-agent", d.userAgent)

	resp, err := utils.DoWithTimeout(d.client, req, 60*time.Second)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := utils.ReadAllWithLimit(resp.Body, 64*1024)
		return fmt.Errorf("download status %d: %s", resp.StatusCode, string(b))
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write file: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close file: %w", closeErr)
	}

	return os.Rename(tmpPath, outPath)
}
