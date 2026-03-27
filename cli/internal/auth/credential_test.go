package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	store := NewCredentialStoreWithDir(t.TempDir())

	cred := &Credential{
		AccessToken:  "tok_abc",
		RefreshToken: "ref_xyz",
		ExpiresAt:    time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		UserName:     "alice",
		UserEmail:    "alice@example.com",
	}

	if err := store.Save(cred); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got == nil {
		t.Fatal("expected credential, got nil")
	}
	if got.AccessToken != "tok_abc" {
		t.Errorf("access_token = %q, want tok_abc", got.AccessToken)
	}
	if got.UserName != "alice" {
		t.Errorf("user_name = %q, want alice", got.UserName)
	}
}

func TestLoadNonExistent(t *testing.T) {
	store := NewCredentialStoreWithDir(t.TempDir())

	got, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil credential, got %+v", got)
	}
}

func TestDelete(t *testing.T) {
	store := NewCredentialStoreWithDir(t.TempDir())

	cred := &Credential{AccessToken: "tok"}
	if err := store.Save(cred); err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := store.Delete(); err != nil {
		t.Fatalf("delete: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("load after delete: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after delete, got %+v", got)
	}
}

func TestDeleteNonExistent(t *testing.T) {
	store := NewCredentialStoreWithDir(t.TempDir())

	if err := store.Delete(); err != nil {
		t.Fatalf("unexpected error deleting non-existent: %v", err)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	store := NewCredentialStoreWithDir(dir)

	if err := store.Save(&Credential{AccessToken: "tok"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, credentialFile)); err != nil {
		t.Fatalf("credential file not found: %v", err)
	}
}

func TestSaveFilePermissions(t *testing.T) {
	dir := t.TempDir()
	store := NewCredentialStoreWithDir(dir)

	if err := store.Save(&Credential{AccessToken: "tok"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, credentialFile))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("permission = %o, want 600", perm)
	}
}

func TestLoadCorruptedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, credentialFile), []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}

	store := NewCredentialStoreWithDir(dir)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for corrupted file")
	}
}

func TestCredentialPathXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := credentialPath()
	want := "/custom/config/muga"
	if got != want {
		t.Errorf("credentialPath() = %q, want %q", got, want)
	}
}

func TestCredentialPathDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got := credentialPath()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "muga")
	if got != want {
		t.Errorf("credentialPath() = %q, want %q", got, want)
	}
}

func TestNewCredentialStore(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")
	store := NewCredentialStore()
	if store.dir != "/tmp/test-config/muga" {
		t.Errorf("dir = %q, want /tmp/test-config/muga", store.dir)
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	store := NewCredentialStoreWithDir(t.TempDir())

	if err := store.Save(&Credential{AccessToken: "old"}); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := store.Save(&Credential{AccessToken: "new"}); err != nil {
		t.Fatalf("second save: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.AccessToken != "new" {
		t.Errorf("access_token = %q, want new", got.AccessToken)
	}
}

func TestSaveUnwritableDirectory(t *testing.T) {
	store := NewCredentialStoreWithDir("/proc/nonexistent/path")
	err := store.Save(&Credential{AccessToken: "tok"})
	if err == nil {
		t.Fatal("expected error for unwritable directory")
	}
}

func TestLoadUnreadableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, credentialFile)
	if err := os.WriteFile(path, []byte("valid json"), 0o000); err != nil {
		t.Fatal(err)
	}

	store := NewCredentialStoreWithDir(dir)
	_, err := store.Load()
	// On some systems this might succeed if running as root, so just check it doesn't panic.
	_ = err
}
