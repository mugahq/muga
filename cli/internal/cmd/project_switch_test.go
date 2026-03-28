package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/api/models"
	"github.com/mugahq/muga/cli/internal/output"
)

// --- helpers ---

func execProjectSwitch(t *testing.T, deps *projectSwitchDeps, args ...string) (string, error) {
	t.Helper()
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	projCmd := findSubCommand(root, "project")
	if projCmd == nil {
		t.Fatal("project subcommand not found")
	}
	projCmd.RemoveCommand(findSubCommand(projCmd, "switch"))
	projCmd.AddCommand(newProjectSwitchCmd(deps))

	orig := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, a []string) error {
		if err := orig(cmd, a); err != nil {
			return err
		}
		opts := output.FromContext(cmd.Context())
		opts.IsTTY = true
		opts.NoColor = true
		return nil
	}

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs(append([]string{"project", "switch"}, args...))

	err := root.Execute()
	return buf.String(), err
}

// --- tests ---

func TestProjectSwitchSuccess(t *testing.T) {
	saver := &mockConfigSaver{}
	deps := &projectSwitchDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "Alpha", Slug: "alpha"},
				{Name: "Beta", Slug: "beta"},
			},
		},
		configSaver: saver,
	}

	out, err := execProjectSwitch(t, deps, "beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "beta") {
		t.Errorf("expected 'beta' in output, got %q", out)
	}
	if saver.slug != "beta" {
		t.Errorf("expected config saved with slug 'beta', got %q", saver.slug)
	}
}

func TestProjectSwitchByName(t *testing.T) {
	saver := &mockConfigSaver{}
	deps := &projectSwitchDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "My Cool Project", Slug: "my-cool-project"},
			},
		},
		configSaver: saver,
	}

	out, err := execProjectSwitch(t, deps, "My Cool Project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "my-cool-project") {
		t.Errorf("expected slug in output, got %q", out)
	}
	if saver.slug != "my-cool-project" {
		t.Errorf("expected config saved with slug 'my-cool-project', got %q", saver.slug)
	}
}

func TestProjectSwitchNotFound(t *testing.T) {
	deps := &projectSwitchDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "Alpha", Slug: "alpha"},
			},
		},
		configSaver: &mockConfigSaver{},
	}

	_, err := execProjectSwitch(t, deps, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %q", err.Error())
	}
}

func TestProjectSwitchMissingArgs(t *testing.T) {
	deps := &projectSwitchDeps{
		apiClient:   &mockProjectClient{},
		configSaver: &mockConfigSaver{},
	}

	_, err := execProjectSwitch(t, deps)
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestProjectSwitchJSON(t *testing.T) {
	resetViper()

	saver := &mockConfigSaver{}
	deps := &projectSwitchDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "Alpha", Slug: "alpha"},
			},
		},
		configSaver: saver,
	}

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	projCmd := findSubCommand(root, "project")
	projCmd.RemoveCommand(findSubCommand(projCmd, "switch"))
	projCmd.AddCommand(newProjectSwitchCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root.SetArgs([]string{"--json", "project", "switch", "alpha"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"active_project"`) {
		t.Errorf("expected JSON output with active_project field, got %q", out)
	}
}

func TestProjectSwitchConfigSaveError(t *testing.T) {
	deps := &projectSwitchDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "Alpha", Slug: "alpha"},
			},
		},
		configSaver: &mockConfigSaver{
			saveErr: fmt.Errorf("disk full"),
		},
	}

	_, err := execProjectSwitch(t, deps, "alpha")
	if err == nil {
		t.Fatal("expected error for config save failure")
	}
	if !strings.Contains(err.Error(), "setting active project") {
		t.Errorf("expected 'setting active project' in error, got %q", err.Error())
	}
}

func TestProjectSubcommandsRegistered(t *testing.T) {
	resetViper()
	root := NewRootCmd("dev", "abc123", "2025-01-01")
	projCmd := findSubCommand(root, "project")
	if projCmd == nil {
		t.Fatal("expected project subcommand on root")
	}

	for _, name := range []string{"create", "ls", "switch"} {
		if findSubCommand(projCmd, name) == nil {
			t.Errorf("expected %q subcommand on project", name)
		}
	}
}
