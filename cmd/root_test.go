package cmd_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRootCmdConfigFileUsed(t *testing.T) {
	var gotAuth string
	srv := newAuthListServer(t, func(auth string) { gotAuth = auth })

	tmp := t.TempDir()
	writeConfigFile(t, tmp, srv.URL, "file-token")
	t.Setenv("OPENLIST_URL", "")
	t.Setenv("OPENLIST_TOKEN", "")

	_, _, err := execCLI(t, "fs", "list", "--path", "/")
	mustOK(t, err)
	if gotAuth != "file-token" {
		t.Fatalf("expected Authorization from config file, got %q", gotAuth)
	}
}

func TestRootCmdEnvOverridesConfig(t *testing.T) {
	var gotAuth string
	srv := newAuthListServer(t, func(auth string) { gotAuth = auth })

	tmp := t.TempDir()
	writeConfigFile(t, tmp, "http://config.example", "config-token")
	t.Setenv("OPENLIST_URL", srv.URL)
	t.Setenv("OPENLIST_TOKEN", "env-token")

	_, _, err := execCLI(t, "fs", "list", "--path", "/")
	mustOK(t, err)
	if gotAuth != "env-token" {
		t.Fatalf("expected Authorization from env, got %q", gotAuth)
	}
}

func TestRootCmdFlagOverridesEnv(t *testing.T) {
	var gotAuth string
	srv := newAuthListServer(t, func(auth string) { gotAuth = auth })

	isolateConfig(t)
	t.Setenv("OPENLIST_URL", "http://env.example")
	t.Setenv("OPENLIST_TOKEN", "env-token")

	_, _, err := execCLI(t, "--url", srv.URL, "--token", "flag-token", "fs", "list", "--path", "/")
	mustOK(t, err)
	if gotAuth != "flag-token" {
		t.Fatalf("expected Authorization from flag, got %q", gotAuth)
	}
}

func TestGetClientWarnsNoToken(t *testing.T) {
	srv := newAuthListServer(t, nil)
	isolateConfig(t)
	t.Setenv("OPENLIST_URL", srv.URL)
	t.Setenv("OPENLIST_TOKEN", "")

	_, stderr, err := execCLI(t, "fs", "list", "--path", "/")
	mustOK(t, err)
	if !strings.Contains(stderr, "Warning") {
		t.Fatalf("expected warning on stderr, got %q", stderr)
	}
}

func TestHelpIsAvailable(t *testing.T) {
	isolateConfig(t)
	out, _, err := execCLI(t, "--help")
	mustOK(t, err)
	if !strings.Contains(out, "OpenList") {
		t.Fatalf("help output: %q", out)
	}
}

func newAuthListServer(t *testing.T, onAuth func(string)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if onAuth != nil {
			onAuth(r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    200,
			"message": "ok",
			"data": map[string]any{
				"content": []any{},
				"total":   0,
			},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}
