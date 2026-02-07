package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"openlist/config"

	"github.com/spf13/viper"
)

func resetViperAndRebind(t *testing.T) {
	t.Helper()
	viper.Reset()
	if err := viper.BindPFlag("url", RootCmd.PersistentFlags().Lookup("url")); err != nil {
		t.Fatalf("bind url flag: %v", err)
	}
	if err := viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token")); err != nil {
		t.Fatalf("bind token flag: %v", err)
	}
}

func writeConfigFile(t *testing.T, dir, url, token string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", dir)
	configDir := filepath.Join(dir, config.AppName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	content := []byte("url: " + url + "\n" + "token: " + token + "\n")
	configPath := filepath.Join(configDir, config.ConfigFileName+"."+config.ConfigFileType)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
}

func TestRootCmdConfigFileUsed(t *testing.T) {
	resetViperAndRebind(t)
	t.Setenv("OPENLIST_URL", "")
	t.Setenv("OPENLIST_TOKEN", "")
	t.Cleanup(func() { config.GlobalConfig = nil })

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    200,
			"message": "ok",
			"data": map[string]interface{}{
				"content": []interface{}{},
				"total":   0,
			},
		})
	}))
	defer srv.Close()

	tmp := t.TempDir()
	writeConfigFile(t, tmp, srv.URL, "file-token")

	RootCmd.SetArgs([]string{"fs", "list", "--path", "/"})
	_, err := RootCmd.ExecuteC()
	if err != nil {
		t.Fatalf("ExecuteC error: %v", err)
	}
	if gotAuth != "file-token" {
		t.Fatalf("expected Authorization from config file, got %q", gotAuth)
	}
}

func TestRootCmdEnvOverridesConfig(t *testing.T) {
	resetViperAndRebind(t)
	t.Cleanup(func() { config.GlobalConfig = nil })

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    200,
			"message": "ok",
			"data": map[string]interface{}{
				"content": []interface{}{},
				"total":   0,
			},
		})
	}))
	defer srv.Close()

	tmp := t.TempDir()
	writeConfigFile(t, tmp, "http://config.example", "config-token")
	t.Setenv("OPENLIST_URL", srv.URL)
	t.Setenv("OPENLIST_TOKEN", "env-token")

	RootCmd.SetArgs([]string{"fs", "list", "--path", "/"})
	_, err := RootCmd.ExecuteC()
	if err != nil {
		t.Fatalf("ExecuteC error: %v", err)
	}
	if gotAuth != "env-token" {
		t.Fatalf("expected Authorization from env, got %q", gotAuth)
	}
}
