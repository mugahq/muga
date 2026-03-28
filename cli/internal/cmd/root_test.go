package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mugahq/muga/cli/internal/output"
)

func resetViper() {
	viper.Reset()
}

func TestRootHelp(t *testing.T) {
	resetViper()
	cmd := NewRootCmd("test-version", "abc123", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatal("expected help output, got empty string")
	}
}

func TestRootVersion(t *testing.T) {
	resetViper()
	cmd := NewRootCmd("1.2.3", "abc123", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatal("expected version output")
	}
	if !bytes.Contains([]byte(out), []byte("1.2.3")) {
		t.Errorf("expected version 1.2.3 in output, got %q", out)
	}
}

func TestGlobalFlagsRegistered(t *testing.T) {
	resetViper()
	cmd := NewRootCmd("dev", "abc123", "2025-01-01")

	flags := []string{"json", "project", "no-color", "verbose"}
	for _, name := range flags {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("expected persistent flag %q to be registered", name)
		}
	}
}

func TestProjectShortFlag(t *testing.T) {
	resetViper()
	cmd := NewRootCmd("dev", "abc123", "2025-01-01")

	f := cmd.PersistentFlags().ShorthandLookup("p")
	if f == nil {
		t.Fatal("expected -p shorthand for --project")
	}
	if f.Name != "project" {
		t.Errorf("expected -p to map to project, got %q", f.Name)
	}
}

func TestExecute(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := Execute("dev", "abc123", "2025-01-01"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPersistentPreRunSetsContext(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewRootCmd("dev", "abc123", "2025-01-01")

	// Add a child command that checks the context for output opts.
	var gotOpts *output.Opts
	child := &cobra.Command{
		Use: "check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			gotOpts = output.FromContext(cmd.Context())
			return nil
		},
	}
	cmd.AddCommand(child)
	cmd.SetArgs([]string{"check"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotOpts == nil {
		t.Fatal("expected output opts in context")
	}
}

func TestProjectFlagPrecedence(t *testing.T) {
	resetViper()

	// Set up config file with project.
	dir := t.TempDir()
	mugaDir := filepath.Join(dir, "muga")
	if err := os.MkdirAll(mugaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("project = \"from-config\"\n")
	if err := os.WriteFile(filepath.Join(mugaDir, "config.toml"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)

	cmd := NewRootCmd("dev", "abc123", "2025-01-01")

	var gotOpts *output.Opts
	child := &cobra.Command{
		Use: "check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			gotOpts = output.FromContext(cmd.Context())
			return nil
		},
	}
	cmd.AddCommand(child)

	// Flag should win over config file.
	cmd.SetArgs([]string{"--project", "from-flag", "check"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotOpts.Project != "from-flag" {
		t.Errorf("expected project=from-flag, got %q", gotOpts.Project)
	}
}

func TestProjectFromConfig(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	mugaDir := filepath.Join(dir, "muga")
	if err := os.MkdirAll(mugaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("project = \"from-config\"\n")
	if err := os.WriteFile(filepath.Join(mugaDir, "config.toml"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)

	cmd := NewRootCmd("dev", "abc123", "2025-01-01")

	var gotOpts *output.Opts
	child := &cobra.Command{
		Use: "check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			gotOpts = output.FromContext(cmd.Context())
			return nil
		},
	}
	cmd.AddCommand(child)
	cmd.SetArgs([]string{"check"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotOpts.Project != "from-config" {
		t.Errorf("expected project=from-config, got %q", gotOpts.Project)
	}
}

func TestAllNounCommandsRegistered(t *testing.T) {
	resetViper()
	cmd := NewRootCmd("dev", "abc123", "2025-01-01")

	expected := []string{
		"auth", "project", "monitor", "cron", "alerts",
		"logs", "errors", "config", "plan", "version",
	}

	registered := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		registered[sub.Name()] = true
	}

	for _, name := range expected {
		if !registered[name] {
			t.Errorf("expected noun command %q to be registered", name)
		}
	}
}

func TestNounCommandsHaveSubcommands(t *testing.T) {
	resetViper()
	root := NewRootCmd("dev", "abc123", "2025-01-01")

	nouns := map[string][]string{
		"auth":    {"login", "logout", "status"},
		"project": {"create", "ls", "switch", "rm"},
		"monitor": {"add", "ls", "rm"},
		"cron":    {"add", "ls", "rm", "ping"},
		"alerts":  {"add", "ls", "history", "rm"},
		"logs":    {"search", "tail", "send"},
		"errors":  {"ls", "show"},
		"config":  {"set", "get", "ls"},
		"plan":    {"status", "upgrade"},
	}

	for noun, verbs := range nouns {
		var nounCmd *cobra.Command
		for _, sub := range root.Commands() {
			if sub.Name() == noun {
				nounCmd = sub
				break
			}
		}
		if nounCmd == nil {
			t.Errorf("noun command %q not found", noun)
			continue
		}

		registered := make(map[string]bool)
		for _, sub := range nounCmd.Commands() {
			registered[sub.Name()] = true
		}

		for _, verb := range verbs {
			if !registered[verb] {
				t.Errorf("expected verb %q under %q", verb, noun)
			}
		}
	}
}
