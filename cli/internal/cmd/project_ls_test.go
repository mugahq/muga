package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/mugahq/muga/api/models"
	"github.com/mugahq/muga/cli/internal/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// --- helpers ---

func execProjectLs(t *testing.T, deps *projectLsDeps, args ...string) (string, error) {
	t.Helper()
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	projCmd := findSubCommand(root, "project")
	if projCmd == nil {
		t.Fatal("project subcommand not found")
	}
	projCmd.RemoveCommand(findSubCommand(projCmd, "ls"))
	projCmd.AddCommand(newProjectLsCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs(append([]string{"project", "ls"}, args...))

	err := root.Execute()
	return buf.String(), err
}

// --- tests ---

func TestProjectLsSuccess(t *testing.T) {
	deps := &projectLsDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{
					Id:        openapi_types.UUID{},
					Name:      "Project Alpha",
					Slug:      "project-alpha",
					CreatedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Id:        openapi_types.UUID{},
					Name:      "Project Beta",
					Slug:      "project-beta",
					CreatedAt: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	out, err := execProjectLs(t, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Project Alpha") {
		t.Errorf("expected 'Project Alpha' in output, got %q", out)
	}
	if !strings.Contains(out, "project-beta") {
		t.Errorf("expected 'project-beta' in output, got %q", out)
	}
}

func TestProjectLsActiveMarker(t *testing.T) {
	resetViper()

	deps := &projectLsDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "Alpha", Slug: "alpha"},
				{Name: "Beta", Slug: "beta"},
			},
		},
	}

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	projCmd := findSubCommand(root, "project")
	projCmd.RemoveCommand(findSubCommand(projCmd, "ls"))
	projCmd.AddCommand(newProjectLsCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root.SetArgs([]string{"--project", "alpha", "project", "ls"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "*") {
		t.Errorf("expected active marker '*' in output, got %q", out)
	}
}

func TestProjectLsEmpty(t *testing.T) {
	deps := &projectLsDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{},
		},
	}

	out, err := execProjectLs(t, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should render a table (possibly empty) without error.
	if strings.Contains(out, "error") {
		t.Errorf("unexpected error text in output: %q", out)
	}
}

func TestProjectLsJSON(t *testing.T) {
	resetViper()

	deps := &projectLsDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "Test", Slug: "test"},
			},
		},
	}

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	projCmd := findSubCommand(root, "project")
	projCmd.RemoveCommand(findSubCommand(projCmd, "ls"))
	projCmd.AddCommand(newProjectLsCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root.SetArgs([]string{"--json", "project", "ls"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"slug"`) {
		t.Errorf("expected JSON output with slug field, got %q", out)
	}
}

func TestProjectLsAPIError(t *testing.T) {
	deps := &projectLsDeps{
		apiClient: &mockProjectClient{
			listErr: &api.APIError{Code: "unauthorized", Message: "invalid token"},
		},
	}

	_, err := execProjectLs(t, deps)
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("expected 'unauthorized' in error, got %q", err.Error())
	}
}
