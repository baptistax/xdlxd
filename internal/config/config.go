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
	return loadDefault(false)
}

func LoadDefaultRequireCookies() (*Config, error) {
	return loadDefault(true)
}

func loadDefault(requireCookies bool) (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	cookiesPath, err := resolveCookiesPath(cwd)
	if err != nil {
		if requireCookies {
			return nil, err
		}
		cookiesPath = ""
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

func resolveCookiesPath(cwd string) (string, error) {
	exePath, err := os.Executable()
	if err == nil {
		if resolvedExePath, evalErr := filepath.EvalSymlinks(exePath); evalErr == nil {
			exePath = resolvedExePath
		}
		exeDir := filepath.Dir(exePath)
		exeCandidate := filepath.Join(exeDir, "cookies.txt")
		if _, statErr := os.Stat(exeCandidate); statErr == nil {
			return exeCandidate, nil
		}
	}

	cwdCandidate := filepath.Join(cwd, "cookies.txt")
	if _, statErr := os.Stat(cwdCandidate); statErr == nil {
		return cwdCandidate, nil
	}

	return "", fmt.Errorf("cookies.txt not found (checked executable dir and cwd: %s)", cwdCandidate)
}
