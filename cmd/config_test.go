package cmd_test

import (
	"testing"
)

func TestConfigGetSetList(t *testing.T) {
	isolateConfig(t)

	out, _, err := execCLI(t, "config", "set", "--key", "url", "--value", "http://example:5244")
	mustOK(t, err)
	got := decodeJSON[map[string]string](t, out)
	if got["message"] != "ok" {
		t.Fatalf("unexpected set output: %v", got)
	}

	out, _, err = execCLI(t, "config", "get", "--key", "url")
	mustOK(t, err)
	got = decodeJSON[map[string]string](t, out)
	if got["url"] != "http://example:5244" {
		t.Fatalf("unexpected get output: %v", got)
	}

	out, _, err = execCLI(t, "config", "list")
	mustOK(t, err)
	got = decodeJSON[map[string]string](t, out)
	if got["url"] != "http://example:5244" {
		t.Fatalf("unexpected list output: %v", got)
	}
	if got["token"] != "" {
		t.Fatalf("expected empty token, got %q", got["token"])
	}
}

func TestConfigUnknownKey(t *testing.T) {
	isolateConfig(t)

	if _, _, err := execCLI(t, "config", "get", "--key", "unknown"); err == nil {
		t.Fatalf("expected error for unknown key")
	}
	if _, _, err := execCLI(t, "config", "set", "--key", "unknown", "--value", "x"); err == nil {
		t.Fatalf("expected error for unknown key")
	}
}

func TestConfigGetSetArgs(t *testing.T) {
	isolateConfig(t)

	if _, _, err := execCLI(t, "config", "set", "url", "http://example:9999"); err != nil {
		t.Fatalf("runConfigSet error: %v", err)
	}

	out, _, err := execCLI(t, "config", "get", "url")
	mustOK(t, err)
	got := decodeJSON[map[string]string](t, out)
	if got["url"] != "http://example:9999" {
		t.Fatalf("unexpected get output: %v", got)
	}
}

func TestConfigGetRequiresKey(t *testing.T) {
	isolateConfig(t)
	if _, _, err := execCLI(t, "config", "get"); err == nil {
		t.Fatalf("expected error when key is missing")
	}
}
