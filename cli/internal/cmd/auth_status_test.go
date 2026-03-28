package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mugahq/muga/cli/internal/auth"
)

func TestAuthStatusNotLoggedIn(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "Not logged in") {
		t.Errorf("expected 'Not logged in' in output, got %q", buf.String())
	}
}

func TestAuthStatusNotLoggedInJSON(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--json", "auth", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"authenticated"`) {
		t.Errorf("expected 'authenticated' key in JSON, got %q", out)
	}
	if !strings.Contains(out, "false") {
		t.Errorf("expected false in JSON output, got %q", out)
	}
}

func TestAuthStatusLoggedIn(t *testing.T) {
	resetViper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	store := auth.NewCredentialStoreWithDir(filepath.Join(dir, "muga"))
	if err := store.Save(&auth.Credential{
		AccessToken: "tok",
		UserName:    "alice",
		UserEmail:   "alice@example.com",
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "alice") {
		t.Errorf("expected username 'alice' in output, got %q", out)
	}
}

func TestAuthStatusLoggedInJSON(t *testing.T) {
	resetViper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	expires := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	store := auth.NewCredentialStoreWithDir(filepath.Join(dir, "muga"))
	if err := store.Save(&auth.Credential{
		AccessToken: "tok",
		UserName:    "alice",
		UserEmail:   "alice@example.com",
		ExpiresAt:   expires,
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--json", "auth", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"authenticated"`) {
		t.Errorf("expected 'authenticated' in JSON, got %q", out)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected username in JSON, got %q", out)
	}
	if !strings.Contains(out, "expires_at") {
		t.Errorf("expected 'expires_at' in JSON, got %q", out)
	}
}
