package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func resetViper() {
	viper.Reset()
}

func TestLoadDefaults(t *testing.T) {
	resetViper()

	// Point to a non-existent config dir so no file is loaded.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIURL != "https://api.muga.sh" {
		t.Errorf("expected default api_url, got %q", cfg.APIURL)
	}
	if cfg.Project != "" {
		t.Errorf("expected empty project, got %q", cfg.Project)
	}
}

func TestLoadFromFile(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	mugaDir := filepath.Join(dir, "muga")
	if err := os.MkdirAll(mugaDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("api_url = \"https://custom.api\"\nproject = \"my-project\"\n")
	if err := os.WriteFile(filepath.Join(mugaDir, "config.toml"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIURL != "https://custom.api" {
		t.Errorf("expected custom api_url, got %q", cfg.APIURL)
	}
	if cfg.Project != "my-project" {
		t.Errorf("expected my-project, got %q", cfg.Project)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	resetViper()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MUGA_PROJECT", "env-project")

	_ = viper.BindEnv("project", "MUGA_PROJECT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Project != "env-project" {
		t.Errorf("expected env-project, got %q", cfg.Project)
	}
}

func TestConfigPathFallback(t *testing.T) {
	// Unset XDG_CONFIG_HOME to hit the fallback path.
	t.Setenv("XDG_CONFIG_HOME", "")

	got := configPath()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "muga")

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestLoadInvalidConfigFile(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	mugaDir := filepath.Join(dir, "muga")
	if err := os.MkdirAll(mugaDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write invalid TOML.
	if err := os.WriteFile(filepath.Join(mugaDir, "config.toml"), []byte("[invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", dir)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid config file")
	}
}
