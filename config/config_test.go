package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func resetViper() {
	viper.Reset()
}

func TestLoadDefaults(t *testing.T) {
	t.Helper()
	resetViper()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.URL != DefaultURL {
		t.Fatalf("expected default URL %q, got %q", DefaultURL, cfg.URL)
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

	cfg, err := Load()
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

	cfg := &Config{URL: "http://saved.local:5244", Token: "saved-token"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	resetViper()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.URL != cfg.URL {
		t.Fatalf("expected URL %q, got %q", cfg.URL, loaded.URL)
	}
	if loaded.Token != cfg.Token {
		t.Fatalf("expected token %q, got %q", cfg.Token, loaded.Token)
	}

	expectedPath := filepath.Join(tmp, AppName, ConfigFileName+"."+ConfigFileType)
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected config file at %s, stat error: %v", expectedPath, err)
	}
}

func TestSettersUpdateViper(t *testing.T) {
	t.Helper()
	resetViper()

	cfg := &Config{}
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
