package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"openlist/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENLIST_URL", "")
	t.Setenv("OPENLIST_TOKEN", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.URL != config.DefaultURL {
		t.Fatalf("expected default URL %q, got %q", config.DefaultURL, cfg.URL)
	}
	if cfg.Token != "" {
		t.Fatalf("expected empty default token, got %q", cfg.Token)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENLIST_URL", "http://example.com:5244")
	t.Setenv("OPENLIST_TOKEN", "env-token")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.URL != "http://example.com:5244" {
		t.Fatalf("expected env URL override, got %q", cfg.URL)
	}
	if cfg.Token != "env-token" {
		t.Fatalf("expected env token override, got %q", cfg.Token)
	}
}

func TestSaveAndLoadConfigFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("OPENLIST_URL", "")
	t.Setenv("OPENLIST_TOKEN", "")

	cfg := &config.Config{URL: "http://saved.local:5244", Token: "saved-token"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	info, err := os.Stat(config.FilePath())
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.URL != cfg.URL {
		t.Fatalf("expected URL %q, got %q", cfg.URL, loaded.URL)
	}
	if loaded.Token != cfg.Token {
		t.Fatalf("expected token %q, got %q", cfg.Token, loaded.Token)
	}

	expectedPath := filepath.Join(tmp, config.AppName, config.ConfigFileName+"."+config.ConfigFileType)
	if expectedPath != config.FilePath() {
		t.Fatalf("FilePath = %q, want %q", config.FilePath(), expectedPath)
	}
}

func TestSetURLTrimsTrailingSlash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://example.com/", "http://example.com"},
		{"http://example.com:5244/", "http://example.com:5244"},
		{"https://example.com/subpath/", "https://example.com/subpath"},
		{"http://localhost:5244", "http://localhost:5244"},
		{"http://localhost:5244///", "http://localhost:5244"},
		{"  https://cdn.example.com/  ", "https://cdn.example.com"},
	}

	for _, tc := range tests {
		cfg := &config.Config{}
		cfg.SetURL(tc.input)
		if cfg.URL != tc.expected {
			t.Errorf("SetURL(%q): got %q, want %q", tc.input, cfg.URL, tc.expected)
		}
	}
}

func TestContextRoundTrip(t *testing.T) {
	cfg := &config.Config{URL: "http://ctx.local", Token: "tok"}
	ctx := config.WithContext(context.Background(), cfg)
	got := config.FromContext(ctx)
	if got != cfg {
		t.Fatalf("FromContext did not return the stored config")
	}
	if config.FromContext(context.Background()) != nil {
		t.Fatalf("expected nil config on empty context")
	}
}

func TestEnvOverridesConfigFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := &config.Config{URL: "http://file.local:5244", Token: "file-token"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	t.Setenv("OPENLIST_URL", "http://env.local:5244")
	t.Setenv("OPENLIST_TOKEN", "env-token")

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.URL != "http://env.local:5244" {
		t.Fatalf("expected env URL, got %q", loaded.URL)
	}
	if loaded.Token != "env-token" {
		t.Fatalf("expected env token, got %q", loaded.Token)
	}
}
