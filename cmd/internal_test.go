package cmd

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetClientMissingConfig(t *testing.T) {
	c := &cobra.Command{}
	c.SetContext(context.Background())
	if _, err := getClient(c); err == nil {
		t.Fatalf("expected error when config is nil")
	}
}

type badJSON struct{}

func (badJSON) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshal failed")
}

func TestWriteJSONError(t *testing.T) {
	var buf bytes.Buffer
	if err := writeJSON(&buf, badJSON{}); err == nil {
		t.Fatalf("expected marshal error")
	}
}
