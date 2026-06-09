package cmd_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"openlist/cmd"
	"openlist/config"
)

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

	cmd.RootCmd.SetArgs([]string{"fs", "list", "--path", "/"})
	_, err := cmd.RootCmd.ExecuteC()
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

	cmd.RootCmd.SetArgs([]string{"fs", "list", "--path", "/"})
	_, err := cmd.RootCmd.ExecuteC()
	if err != nil {
		t.Fatalf("ExecuteC error: %v", err)
	}
	if gotAuth != "env-token" {
		t.Fatalf("expected Authorization from env, got %q", gotAuth)
	}
}
