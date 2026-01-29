package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourname/xdl2/internal/utils"
)

type Config struct {
	CookiesPath string
	OutputDir   string
	UserAgent   string
}

func LoadDefault() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	cookiesPath := filepath.Join(cwd, "cookies.txt")
	if _, err := os.Stat(cookiesPath); err != nil {
		return nil, fmt.Errorf("cookies.txt not found (expected %s)", cookiesPath)
	}

	outDir := filepath.Join(cwd, "out")
	if err := utils.EnsureDir(outDir); err != nil {
		return nil, fmt.Errorf("ensure out dir: %w", err)
	}

	return &Config{
		CookiesPath: cookiesPath,
		OutputDir:   outDir,
		UserAgent:   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36",
	}, nil
}
