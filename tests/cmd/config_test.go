package cmd_test

import (
	"encoding/json"
	"testing"

	"openlist/cmd"
	"openlist/config"
)

func TestConfigGetSetList(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	resetConfigFlags()

	cmd.ConfigSetCmd.Flags().Set("key", "url")
	cmd.ConfigSetCmd.Flags().Set("value", "http://example:5244")
	var err error
	out := captureStdout(t, func() { err = cmd.RunConfigSet(cmd.ConfigSetCmd, nil) })
	if err != nil {
		t.Fatalf("runConfigSet error: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode set output: %v", err)
	}
	if got["message"] != "ok" {
		t.Fatalf("unexpected set output: %v", got)
	}

	cmd.ConfigGetCmd.Flags().Set("key", "url")
	out = captureStdout(t, func() { err = cmd.RunConfigGet(cmd.ConfigGetCmd, nil) })
	if err != nil {
		t.Fatalf("runConfigGet error: %v", err)
	}
	got = map[string]string{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode get output: %v", err)
	}
	if got["url"] != "http://example:5244" {
		t.Fatalf("unexpected get output: %v", got)
	}

	out = captureStdout(t, func() { err = cmd.RunConfigList() })
	if err != nil {
		t.Fatalf("runConfigList error: %v", err)
	}
	got = map[string]string{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode list output: %v", err)
	}
	if got["url"] != "http://example:5244" {
		t.Fatalf("unexpected list output: %v", got)
	}
	if got["token"] != "" {
		t.Fatalf("expected empty token, got %q", got["token"])
	}
}

func TestConfigUnknownKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	config.GlobalConfig = nil
	resetConfigFlags()

	cmd.ConfigGetCmd.Flags().Set("key", "unknown")
	if err := cmd.RunConfigGet(cmd.ConfigGetCmd, nil); err == nil {
		t.Fatalf("expected error for unknown key")
	}

	cmd.ConfigSetCmd.Flags().Set("key", "unknown")
	cmd.ConfigSetCmd.Flags().Set("value", "x")
	if err := cmd.RunConfigSet(cmd.ConfigSetCmd, nil); err == nil {
		t.Fatalf("expected error for unknown key")
	}
}

func TestConfigGetSetArgs(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	resetConfigFlags()

	var err error
	out := captureStdout(t, func() { err = cmd.RunConfigSet(cmd.ConfigSetCmd, []string{"url", "http://example:9999"}) })
	if err != nil {
		t.Fatalf("runConfigSet error: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode set output: %v", err)
	}

	out = captureStdout(t, func() { err = cmd.RunConfigGet(cmd.ConfigGetCmd, []string{"url"}) })
	if err != nil {
		t.Fatalf("runConfigGet error: %v", err)
	}
	got = map[string]string{}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode get output: %v", err)
	}
	if got["url"] != "http://example:9999" {
		t.Fatalf("unexpected get output: %v", got)
	}
}
