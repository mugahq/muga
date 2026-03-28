package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mugahq/muga/cli/internal/auth"
)

func TestLogoutSuccess(t *testing.T) {
	resetViper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Create a credential so delete has something to remove.
	store := auth.NewCredentialStoreWithDir(dir)
	if err := store.Save(&auth.Credential{AccessToken: "tok"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "logout"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "Logged out") {
		t.Errorf("expected 'Logged out' in output, got %q", buf.String())
	}
}

func TestLogoutJSON(t *testing.T) {
	resetViper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--json", "auth", "logout"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), `"ok"`) {
		t.Errorf("expected JSON {\"ok\":true} in output, got %q", buf.String())
	}
}
