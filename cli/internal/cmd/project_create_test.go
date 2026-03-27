package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mugahq/muga/api/models"
	"github.com/mugahq/muga/cli/internal/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// --- mocks ---

type mockProjectClient struct {
	projects   []models.Project
	listErr    error
	created    *models.Project
	createReq  models.CreateProjectRequest
	createErr  error
}

func (m *mockProjectClient) ListProjects() ([]models.Project, error) {
	return m.projects, m.listErr
}

func (m *mockProjectClient) CreateProject(req models.CreateProjectRequest) (*models.Project, error) {
	m.createReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.created, nil
}

type mockConfigSaver struct {
	slug    string
	saveErr error
}

func (m *mockConfigSaver) SetProject(slug string) error {
	m.slug = slug
	return m.saveErr
}

// --- helpers ---

func execProjectCreate(t *testing.T, deps *projectCreateDeps, args ...string) (string, error) {
	t.Helper()
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	projCmd := findSubCommand(root, "project")
	if projCmd == nil {
		t.Fatal("project subcommand not found")
	}
	projCmd.RemoveCommand(findSubCommand(projCmd, "create"))
	projCmd.AddCommand(newProjectCreateCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs(append([]string{"project", "create"}, args...))

	err := root.Execute()
	return buf.String(), err
}

// --- tests ---

func TestProjectCreateSuccess(t *testing.T) {
	saver := &mockConfigSaver{}
	deps := &projectCreateDeps{
		apiClient: &mockProjectClient{
			created: &models.Project{
				Id:        openapi_types.UUID{},
				Name:      "My Project",
				Slug:      "my-project",
				CreatedAt: time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC),
			},
		},
		configSaver: saver,
	}

	out, err := execProjectCreate(t, deps, "My Project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "My Project") {
		t.Errorf("expected project name in output, got %q", out)
	}
	if !strings.Contains(out, "my-project") {
		t.Errorf("expected slug in output, got %q", out)
	}
	if !strings.Contains(out, "active project") {
		t.Errorf("expected active project message, got %q", out)
	}
	if saver.slug != "my-project" {
		t.Errorf("expected config saved with slug 'my-project', got %q", saver.slug)
	}

	client := deps.apiClient.(*mockProjectClient)
	if client.createReq.Name != "My Project" {
		t.Errorf("expected create request name 'My Project', got %q", client.createReq.Name)
	}
	if client.createReq.Slug != "my-project" {
		t.Errorf("expected create request slug 'my-project', got %q", client.createReq.Slug)
	}
}

func TestProjectCreateJSON(t *testing.T) {
	resetViper()
	saver := &mockConfigSaver{}
	deps := &projectCreateDeps{
		apiClient: &mockProjectClient{
			created: &models.Project{
				Id:        openapi_types.UUID{},
				Name:      "Test",
				Slug:      "test",
				CreatedAt: time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC),
			},
		},
		configSaver: saver,
	}

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	projCmd := findSubCommand(root, "project")
	projCmd.RemoveCommand(findSubCommand(projCmd, "create"))
	projCmd.AddCommand(newProjectCreateCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root.SetArgs([]string{"--json", "project", "create", "Test"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"slug"`) {
		t.Errorf("expected JSON output with slug field, got %q", out)
	}
}

func TestProjectCreateMissingArgs(t *testing.T) {
	deps := &projectCreateDeps{
		apiClient:   &mockProjectClient{},
		configSaver: &mockConfigSaver{},
	}

	_, err := execProjectCreate(t, deps)
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestProjectCreateAPIError(t *testing.T) {
	deps := &projectCreateDeps{
		apiClient: &mockProjectClient{
			createErr: &api.APIError{Code: "conflict", Message: "slug already taken"},
		},
		configSaver: &mockConfigSaver{},
	}

	_, err := execProjectCreate(t, deps, "My Project")
	if err == nil {
		t.Fatal("expected error for API failure")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Errorf("expected 'conflict' in error, got %q", err.Error())
	}
}

func TestProjectCreateConfigSaveError(t *testing.T) {
	deps := &projectCreateDeps{
		apiClient: &mockProjectClient{
			created: &models.Project{
				Name: "Test",
				Slug: "test",
			},
		},
		configSaver: &mockConfigSaver{
			saveErr: fmt.Errorf("disk full"),
		},
	}

	_, err := execProjectCreate(t, deps, "Test")
	if err == nil {
		t.Fatal("expected error for config save failure")
	}
	if !strings.Contains(err.Error(), "setting active project") {
		t.Errorf("expected 'setting active project' in error, got %q", err.Error())
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Project", "my-project"},
		{"  Hello World  ", "hello-world"},
		{"test-slug", "test-slug"},
		{"UPPERCASE", "uppercase"},
		{"special!@#chars", "special-chars"},
		{"multiple   spaces", "multiple-spaces"},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
