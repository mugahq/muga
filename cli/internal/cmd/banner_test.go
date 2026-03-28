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
