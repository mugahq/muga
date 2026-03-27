package cmd

import (
	"strings"
	"testing"

	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/config"
)

func TestRequireAuthNotLoggedIn(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no credentials file

	cfg := &config.Config{APIURL: "https://api.muga.sh"}
	_, err := requireAuth(cfg, nil)
	if err == nil {
		t.Fatal("expected error when no credentials exist")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in' in error, got %q", err.Error())
	}
}

func TestRequireAuthWithCredentials(t *testing.T) {
	resetViper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Pre-seed credentials.
	store := auth.NewCredentialStoreWithDir(dir + "/muga")
	if err := store.Save(&auth.Credential{AccessToken: "tok_test"}); err != nil {
		t.Fatalf("saving credentials: %v", err)
	}

	cfg := &config.Config{APIURL: "https://api.muga.sh"}
	client, err := requireAuth(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestRequireAuthNilCredStoreDefaultsToNewStore(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{APIURL: "https://api.muga.sh"}
	// Passing nil credStore should not panic — it defaults to NewCredentialStore.
	_, err := requireAuth(cfg, nil)
	if err == nil {
		t.Fatal("expected error (not logged in) but got nil")
	}
}

// TestProjectCreateNotLoggedIn exercises the nil-deps production path in
// newProjectCreateCmd — it should fail with "not logged in".
func TestProjectCreateNotLoggedIn(t *testing.T) {
	_, err := execProjectCreate(t, nil, "My Project")
	if err == nil {
		t.Fatal("expected error when not logged in")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in', got %q", err.Error())
	}
}

// TestProjectLsNotLoggedIn exercises the nil-deps production path in newProjectLsCmd.
func TestProjectLsNotLoggedIn(t *testing.T) {
	_, err := execProjectLs(t, nil)
	if err == nil {
		t.Fatal("expected error when not logged in")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in', got %q", err.Error())
	}
}

// TestProjectSwitchNotLoggedIn exercises the nil-deps production path in newProjectSwitchCmd.
func TestProjectSwitchNotLoggedIn(t *testing.T) {
	_, err := execProjectSwitch(t, nil, "alpha")
	if err == nil {
		t.Fatal("expected error when not logged in")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in', got %q", err.Error())
	}
}
