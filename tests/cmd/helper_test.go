package cmd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"openlist/cmd"
	"openlist/config"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	fn()

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func setFlag(t *testing.T, cmdFlags interface {
	Set(string, string) error
}, name, value string) {
	t.Helper()
	if err := cmdFlags.Set(name, value); err != nil {
		t.Fatalf("set flag %s: %v", name, err)
	}
}

func newJSONServer(t *testing.T, wantPath string, wantToken string, wantBody interface{}, resp interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
}

func resetViperAndRebind(t *testing.T) {
	t.Helper()
	viper.Reset()
	rootCmd := cmd.RootCmd
	if err := viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url")); err != nil {
		t.Fatalf("bind url flag: %v", err)
	}
	if err := viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token")); err != nil {
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

func resetConfigFlags() {
	_ = cmd.ConfigGetCmd.Flags().Set("key", "")
	_ = cmd.ConfigSetCmd.Flags().Set("key", "")
	_ = cmd.ConfigSetCmd.Flags().Set("value", "")
	cmd.ConfigGetCmd.SetArgs([]string{})
	cmd.ConfigSetCmd.SetArgs([]string{})
}

func prepareBatchRenameCmd(t *testing.T) {
	t.Helper()
	resetBatchRenameFlags(t)
	t.Cleanup(func() { resetBatchRenameFlags(t) })
}

func resetBatchRenameFlags(t *testing.T) {
	t.Helper()
	flags := cmd.BatchRenameCmd.Flags()
	stringDefaults := map[string]string{
		"dir":               "/",
		"password":          "",
		"case":              "",
		"chinese":           "",
		"page":              "1",
		"per-page":          "0",
		"all":               "false",
		"refresh":           "false",
		"include-extension": "false",
		"dry-run":           "false",
	}
	for name, value := range stringDefaults {
		flag := flags.Lookup(name)
		if flag == nil {
			t.Fatalf("missing batch rename flag %s", name)
		}
		if err := flag.Value.Set(value); err != nil {
			t.Fatalf("reset flag %s: %v", name, err)
		}
		flag.Changed = false
	}
	for _, name := range []string{"names", "paths", "replace", "regex-replace", "insert", "delete"} {
		flag := flags.Lookup(name)
		if flag == nil {
			t.Fatalf("missing batch rename flag %s", name)
		}
		value, ok := flag.Value.(pflag.SliceValue)
		if !ok {
			t.Fatalf("batch rename flag %s is not a slice value", name)
		}
		if err := value.Replace([]string{}); err != nil {
			t.Fatalf("reset flag %s: %v", name, err)
		}
		flag.Changed = false
	}
}
