package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"openlist/cmd"
	"openlist/config"
)

func execCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	root := cmd.NewRootCmd()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs(args)
	err = root.Execute()
	return out.String(), errBuf.String(), err
}

func isolateConfig(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENLIST_URL", "")
	t.Setenv("OPENLIST_TOKEN", "")
}

func withServer(t *testing.T, token string, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	isolateConfig(t)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	t.Setenv("OPENLIST_URL", srv.URL)
	t.Setenv("OPENLIST_TOKEN", token)
	return srv
}

func newJSONServer(t *testing.T, wantPath, wantToken string, wantBody any, resp any) *httptest.Server {
	t.Helper()
	return withServer(t, wantToken, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != wantPath {
			t.Fatalf("expected path %s, got %s", wantPath, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != wantToken {
			t.Fatalf("expected Authorization %q, got %q", wantToken, got)
		}
		if wantBody != nil {
			gotBody := reflect.New(reflect.TypeOf(wantBody)).Interface()
			if err := json.NewDecoder(r.Body).Decode(gotBody); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if !reflect.DeepEqual(reflect.ValueOf(gotBody).Elem().Interface(), wantBody) {
				t.Fatalf("request body mismatch: %#v", gotBody)
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
}

func errorServer(t *testing.T) *httptest.Server {
	t.Helper()
	return withServer(t, "tok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	})
}

func writeConfigFile(t *testing.T, dir, url, token string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", dir)
	configDir := filepath.Join(dir, config.AppName)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	content := []byte("url: " + url + "\n" + "token: " + token + "\n")
	configPath := filepath.Join(configDir, config.ConfigFileName+"."+config.ConfigFileType)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
}

func decodeJSON[T any](t *testing.T, raw string) T {
	t.Helper()
	var got T
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("decode output %q: %v", raw, err)
	}
	return got
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
}
