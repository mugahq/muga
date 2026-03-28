package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/output"
)

func cmdWithOpts(opts *output.Opts) *cobra.Command {
	root := &cobra.Command{Use: "muga"}
	child := &cobra.Command{
		Use:   "auth",
		Short: "Sign in, sign out, check status",
		Run:   func(*cobra.Command, []string) {},
	}
	root.AddCommand(child)
	ctx := output.WithOpts(context.Background(), opts)
	root.SetContext(ctx)
	return root
}

func TestRenderBannerJSON(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	store := auth.NewCredentialStoreWithDir(dir)

	opts := &output.Opts{JSON: true}
	cmd := cmdWithOpts(opts)

	var buf bytes.Buffer
	err := renderBanner(&buf, cmd, "0.1.0", &bannerDeps{credStore: store})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if v, ok := result["version"].(string); !ok || v != "0.1.0" {
		t.Errorf("expected version=0.1.0, got %v", result["version"])
	}
	if v, ok := result["authenticated"].(bool); !ok || v {
		t.Errorf("expected authenticated=false, got %v", result["authenticated"])
	}
}

func TestRenderBannerJSONAuthenticated(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	store := auth.NewCredentialStoreWithDir(dir)
	_ = store.Save(&auth.Credential{AccessToken: "tok"})

	opts := &output.Opts{JSON: true, Project: "my-saas"}
	cmd := cmdWithOpts(opts)

	var buf bytes.Buffer
	err := renderBanner(&buf, cmd, "0.1.0", &bannerDeps{credStore: store})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if v, ok := result["authenticated"].(bool); !ok || !v {
		t.Errorf("expected authenticated=true, got %v", result["authenticated"])
	}
	if v, ok := result["project"].(string); !ok || v != "my-saas" {
		t.Errorf("expected project=my-saas, got %v", result["project"])
	}
}

func TestRenderBannerPlain(t *testing.T) {
	resetViper()

	opts := &output.Opts{IsTTY: false}
	cmd := cmdWithOpts(opts)

	var buf bytes.Buffer
	err := renderBanner(&buf, cmd, "0.1.0", &bannerDeps{credStore: auth.NewCredentialStoreWithDir(t.TempDir())})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "muga ") && strings.Contains(out, "\u2500") {
		t.Error("expected no signature line in plain output")
	}
	if strings.Contains(out, "\x1b[") {
		t.Error("expected no ANSI codes in plain output")
	}
	if !strings.Contains(out, "auth") {
		t.Error("expected command name 'auth' in plain output")
	}
}

func TestRenderBannerTTYUnauthenticated(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	store := auth.NewCredentialStoreWithDir(dir)

	opts := &output.Opts{IsTTY: true, NoColor: true}
	cmd := cmdWithOpts(opts)

	var buf bytes.Buffer
	err := renderBanner(&buf, cmd, "0.1.0", &bannerDeps{credStore: store})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "muga") {
		t.Error("expected 'muga' in TTY output")
	}
	if !strings.Contains(out, "Quick start:") {
		t.Error("expected quick start section for unauthenticated user")
	}
	if !strings.Contains(out, "muga auth login") {
		t.Error("expected auth login step in quick start")
	}
}

func TestRenderBannerTTYAuthenticated(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	store := auth.NewCredentialStoreWithDir(dir)
	_ = store.Save(&auth.Credential{AccessToken: "tok"})

	opts := &output.Opts{IsTTY: true, NoColor: true, Project: "my-saas"}
	cmd := cmdWithOpts(opts)

	var buf bytes.Buffer
	err := renderBanner(&buf, cmd, "0.1.0", &bannerDeps{credStore: store})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "OBSERVABILITY") {
		t.Error("expected OBSERVABILITY section header")
	}
	if !strings.Contains(out, "SETUP") {
		t.Error("expected SETUP section header")
	}
	if !strings.Contains(out, "my-saas") {
		t.Error("expected project name in signature line")
	}
}

func TestMugaOutputEnvVar(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MUGA_OUTPUT", "json")

	cmd := NewRootCmd("0.1.0", "abc123", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("expected JSON output from MUGA_OUTPUT=json, got: %s", buf.String())
	}
}

func TestJsonFlagPrecedenceOverEnvVar(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MUGA_OUTPUT", "json")

	cmd := NewRootCmd("0.1.0", "abc123", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}
}

func TestCompletionHiddenFromPlainOutput(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In non-TTY (test), output is plain command names.
	if strings.Contains(buf.String(), "completion") {
		t.Error("expected completion to be hidden from root output")
	}
}

func TestCompletionStillWorks(t *testing.T) {
	viper.Reset()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"completion", "bash"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("completion bash should work: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected completion output")
	}
}

func TestNoColorRendersBrandedWithoutANSI(t *testing.T) {
	resetViper()

	dir := t.TempDir()
	store := auth.NewCredentialStoreWithDir(dir)
	_ = store.Save(&auth.Credential{AccessToken: "tok"})

	opts := &output.Opts{IsTTY: true, NoColor: true}
	cmd := cmdWithOpts(opts)

	var buf bytes.Buffer
	err := renderBanner(&buf, cmd, "0.1.0", &bannerDeps{credStore: store})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Error("expected no ANSI escape codes with --no-color")
	}
	if !strings.Contains(out, "muga") {
		t.Error("expected branded layout even with --no-color")
	}
}

// --- Noun help tests ---

func newTestNounCmd() *cobra.Command {
	parent := &cobra.Command{
		Use:   "things",
		Short: "Manage things",
	}
	parent.AddCommand(&cobra.Command{Use: "add", Short: "Create a thing"})
	parent.AddCommand(&cobra.Command{Use: "ls", Short: "List things"})
	parent.AddCommand(&cobra.Command{Use: "rm", Short: "Delete a thing"})
	return parent
}

func TestRenderNounHelp_TTY(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestNounCmd()
	opts := &output.Opts{IsTTY: true, NoColor: true, Project: "my-saas"}

	if err := renderNounHelp(&buf, cmd, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// Must contain signature line with project suffix.
	if !strings.Contains(out, "muga") {
		t.Error("expected 'muga' in signature line")
	}
	if !strings.Contains(out, "my-saas") {
		t.Error("expected project suffix 'my-saas' in signature line")
	}

	// Must contain section header.
	if !strings.Contains(out, "THINGS") {
		t.Error("expected section header 'THINGS'")
	}

	// Must contain subcommand rows.
	if !strings.Contains(out, "add") {
		t.Error("expected 'add' subcommand")
	}
	if !strings.Contains(out, "Create a thing") {
		t.Error("expected description 'Create a thing'")
	}
	if !strings.Contains(out, "ls") {
		t.Error("expected 'ls' subcommand")
	}
	if !strings.Contains(out, "rm") {
		t.Error("expected 'rm' subcommand")
	}

	// Must contain footer.
	if !strings.Contains(out, "muga things [cmd] --help for details") {
		t.Error("expected footer with noun help hint")
	}
}

func TestRenderNounHelp_NoTTY(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestNounCmd()
	opts := &output.Opts{IsTTY: false}

	if err := renderNounHelp(&buf, cmd, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	// Should be plain subcommand names only.
	expected := []string{"add", "ls", "rm"}
	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d: %q", len(expected), len(lines), out)
	}
	for i, want := range expected {
		if strings.TrimSpace(lines[i]) != want {
			t.Errorf("line %d: expected %q, got %q", i, want, lines[i])
		}
	}

	// Should not contain ANSI, signature, or descriptions.
	if strings.Contains(out, "muga") {
		t.Error("no-TTY output should not contain signature line")
	}
	if strings.Contains(out, "THINGS") {
		t.Error("no-TTY output should not contain section header")
	}
}

func TestRenderNounHelp_NilOpts(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestNounCmd()

	// nil opts should default to non-TTY behavior.
	if err := renderNounHelp(&buf, cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	if len(lines) != 3 {
		t.Errorf("expected 3 lines for non-TTY default, got %d", len(lines))
	}
}

func TestRenderNounHelp_HiddenCommand(t *testing.T) {
	parent := &cobra.Command{Use: "things", Short: "Manage things"}
	parent.AddCommand(&cobra.Command{Use: "add", Short: "Create a thing"})
	hidden := &cobra.Command{Use: "internal", Short: "Internal command", Hidden: true}
	parent.AddCommand(hidden)

	var buf bytes.Buffer
	opts := &output.Opts{IsTTY: true, NoColor: true}

	if err := renderNounHelp(&buf, parent, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "internal") {
		t.Error("hidden command should not appear in help output")
	}
}

func TestRenderNounHelp_NoTagline(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTestNounCmd()
	opts := &output.Opts{IsTTY: true, NoColor: true}

	if err := renderNounHelp(&buf, cmd, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// Noun help should NOT contain the root tagline.
	if strings.Contains(out, "observability") {
		t.Error("noun help should not contain root tagline")
	}
}

func TestMaxSubcommandWidth(t *testing.T) {
	parent := &cobra.Command{Use: "things"}
	parent.AddCommand(&cobra.Command{Use: "add", Short: "Create"})
	parent.AddCommand(&cobra.Command{Use: "history", Short: "Show history"})
	parent.AddCommand(&cobra.Command{Use: "rm", Short: "Delete"})

	got := maxSubcommandWidth(parent)
	if got != 7 { // "history" = 7 chars
		t.Errorf("expected max width 7, got %d", got)
	}
}

func TestMaxSubcommandWidth_ExcludesHidden(t *testing.T) {
	parent := &cobra.Command{Use: "things"}
	parent.AddCommand(&cobra.Command{Use: "add", Short: "Create"})
	parent.AddCommand(&cobra.Command{Use: "verylongname", Short: "Hidden", Hidden: true})

	got := maxSubcommandWidth(parent)
	if got != 3 { // "add" = 3 chars
		t.Errorf("expected max width 3, got %d", got)
	}
}

func TestRenderFullHelp_TTY(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("0.2.0", "abc123", "2025-01-01")
	opts := &output.Opts{IsTTY: true, NoColor: true}
	root.SetContext(output.WithOpts(context.Background(), opts))

	var buf bytes.Buffer
	if err := renderFullHelp(&buf, root, "0.2.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "muga") {
		t.Error("expected 'muga' in full help output")
	}
	if !strings.Contains(out, "COMMANDS") {
		t.Error("expected 'COMMANDS' section header")
	}
	if !strings.Contains(out, "GLOBAL FLAGS") {
		t.Error("expected 'GLOBAL FLAGS' section")
	}
}

func TestRenderFullHelp_NoTTY(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("0.2.0", "abc123", "2025-01-01")
	opts := &output.Opts{IsTTY: false}
	root.SetContext(output.WithOpts(context.Background(), opts))

	var buf bytes.Buffer
	if err := renderFullHelp(&buf, root, "0.2.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "muga") {
		t.Error("expected 'muga' in non-TTY full help output")
	}
	if !strings.Contains(out, "auth") {
		t.Error("expected command names in non-TTY output")
	}
}

func TestStubCommandsReturnNotImplemented(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	stubCmds := [][]string{
		{"monitor", "add"},
		{"monitor", "ls"},
		{"monitor", "rm", "abc123"},
		{"cron", "add"},
		{"cron", "ls"},
		{"cron", "rm", "abc123"},
		{"cron", "ping", "abc123"},
		{"alerts", "add"},
		{"alerts", "ls"},
		{"alerts", "history"},
		{"alerts", "rm", "abc123"},
		{"logs", "search"},
		{"logs", "tail"},
		{"logs", "send", "hello"},
		{"errors", "ls"},
		{"errors", "show", "abc123"},
		{"config", "set", "key", "value"},
		{"config", "get", "key"},
		{"config", "ls"},
		{"plan", "status"},
		{"plan", "upgrade"},
		{"project", "rm", "my-project"},
	}

	for _, args := range stubCmds {
		t.Run(strings.Join(args, "/"), func(t *testing.T) {
			root := NewRootCmd("dev", "abc123", "2025-01-01")
			root.SetOut(new(bytes.Buffer))
			root.SetErr(new(bytes.Buffer))
			root.SetArgs(args)
			err := root.Execute()
			if err == nil {
				t.Error("expected errNotImplemented, got nil")
			}
			if err != errNotImplemented {
				t.Errorf("expected errNotImplemented, got %v", err)
			}
		})
	}
}
