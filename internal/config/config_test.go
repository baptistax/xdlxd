package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultAllowsMissingCookiesForTweetMode(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	cfg, err := LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault() error = %v", err)
	}
	if cfg.CookiesPath != "" {
		t.Fatalf("expected empty CookiesPath, got %q", cfg.CookiesPath)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "out")); err != nil {
		t.Fatalf("expected out dir to be created: %v", err)
	}
}

func TestLoadDefaultRequireCookiesFailsWhenMissing(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	if _, err := LoadDefaultRequireCookies(); err == nil {
		t.Fatalf("expected missing cookies error")
	}
}

func TestLoadDefaultRequireCookiesUsesCookiesFromWorkingDir(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	wantPath := filepath.Join(tmpDir, "cookies.txt")
	if err := os.WriteFile(wantPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := LoadDefaultRequireCookies()
	if err != nil {
		t.Fatalf("LoadDefaultRequireCookies() error = %v", err)
	}
	if cfg.CookiesPath != wantPath {
		t.Fatalf("expected CookiesPath %q, got %q", wantPath, cfg.CookiesPath)
	}
}
