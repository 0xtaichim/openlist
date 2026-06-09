package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"openlist/config"

	"github.com/spf13/viper"
)

func resetViper() {
	viper.Reset()
}

func TestLoadDefaults(t *testing.T) {
	t.Helper()
	resetViper()
	os.Unsetenv("OPENLIST_URL")
	os.Unsetenv("OPENLIST_TOKEN")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

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
	t.Helper()
	resetViper()

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
	t.Helper()
	resetViper()

	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := &config.Config{URL: "http://saved.local:5244", Token: "saved-token"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	resetViper()
	t.Setenv("XDG_CONFIG_HOME", tmp)

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
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected config file at %s, stat error: %v", expectedPath, err)
	}
}

func TestSetURLTrimsTrailingSlash(t *testing.T) {
	t.Helper()
	resetViper()

	tests := []struct {
		input    string
		expected string
	}{
		{"http://example.com/", "http://example.com"},
		{"http://example.com:5244/", "http://example.com:5244"},
		{"https://example.com/subpath/", "https://example.com/subpath"},
		{"http://localhost:5244", "http://localhost:5244"},
		{"http://localhost:5244///", "http://localhost:5244"},
	}

	for _, tc := range tests {
		cfg := &config.Config{}
		cfg.SetURL(tc.input)
		if cfg.URL != tc.expected {
			t.Errorf("SetURL(%q): got %q, want %q", tc.input, cfg.URL, tc.expected)
		}
	}
}

func TestSettersUpdateViper(t *testing.T) {
	t.Helper()
	resetViper()

	cfg := &config.Config{}
	cfg.SetURL("http://setter.local:5244")
	cfg.SetToken("setter-token")

	if cfg.URL != "http://setter.local:5244" {
		t.Fatalf("SetURL did not update config URL")
	}
	if cfg.Token != "setter-token" {
		t.Fatalf("SetToken did not update config token")
	}

	if got := viper.GetString("url"); got != "http://setter.local:5244" {
		t.Fatalf("viper url not updated, got %q", got)
	}
	if got := viper.GetString("token"); got != "setter-token" {
		t.Fatalf("viper token not updated, got %q", got)
	}
}
