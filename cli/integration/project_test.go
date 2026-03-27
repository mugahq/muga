//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// projectMockServer creates a test server with both auth endpoints and project
// endpoints. It also pre-seeds credentials in the XDG_CONFIG_HOME.
func newProjectMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Projects list.
	mux.HandleFunc("GET /v1/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{"code": "unauthorized", "message": "missing token"},
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "00000000-0000-4000-8000-000000000001", "name": "Alpha", "slug": "alpha", "created_at": "2026-01-01T00:00:00Z"},
				{"id": "00000000-0000-4000-8000-000000000002", "name": "Beta", "slug": "beta", "created_at": "2026-02-01T00:00:00Z"},
			},
			"pagination": map[string]any{"has_more": false},
		})
	})

	// Create project.
	mux.HandleFunc("POST /v1/projects", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "00000000-0000-4000-8000-000000000099",
			"name":       req.Name,
			"slug":       req.Slug,
			"created_at": "2026-03-27T00:00:00Z",
		})
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

// seedCredentials writes a mock credentials file so commands don't require a
// real login flow.
func seedCredentials(t *testing.T, xdgHome, token string) {
	t.Helper()
	dir := filepath.Join(xdgHome, "muga")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("creating credentials dir: %v", err)
	}
	cred := map[string]string{
		"access_token": token,
		"user_name":    "Test User",
	}
	data, err := json.Marshal(cred)
	if err != nil {
		t.Fatalf("encoding credentials: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0o600); err != nil {
		t.Fatalf("writing credentials: %v", err)
	}
}

// projectEnv returns env for project commands: server URL, pre-seeded
// credentials, and an isolated config dir.
func projectEnv(t *testing.T, serverURL string) map[string]string {
	t.Helper()
	xdgHome := newTempHome(t)
	seedCredentials(t, xdgHome, "muga_integration_test_token")
	return map[string]string{
		"MUGA_API_URL":    serverURL,
		"XDG_CONFIG_HOME": xdgHome,
		"HOME":            t.TempDir(),
		"BROWSER":         "",
	}
}

// --- project ls ---

func TestProject_Ls_TableOutput(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "project", "ls")

	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "Alpha")
	assertStdoutContains(t, result, "Beta")
	assertStdoutContains(t, result, "alpha")
	assertStdoutContains(t, result, "beta")
}

func TestProject_Ls_JSONOutput(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "--json", "project", "ls")

	assertExitCode(t, result, 0)

	var projects []map[string]any
	if err := json.Unmarshal([]byte(result.stdout), &projects); err != nil {
		t.Fatalf("decoding JSON output: %v\nraw: %s", err, result.stdout)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestProject_Ls_NotLoggedIn(t *testing.T) {
	ts := newProjectMockServer(t)
	// Use baseEnv — no credentials seeded.
	env := baseEnv(t, ts.URL)

	result := runCLI(t, env, "project", "ls")

	if result.exitCode == 0 {
		t.Error("expected non-zero exit code when not logged in")
	}
	combined := result.stdout + result.stderr
	if !strings.Contains(combined, "not logged in") {
		t.Errorf("expected 'not logged in' in output, got stdout=%q stderr=%q", result.stdout, result.stderr)
	}
}

// --- project create ---

func TestProject_Create_Success(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "project", "create", "New Project")

	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "New Project")
	assertStdoutContains(t, result, "new-project")
}

func TestProject_Create_JSONOutput(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "--json", "project", "create", "JSON Project")

	assertExitCode(t, result, 0)

	var project map[string]any
	if err := json.Unmarshal([]byte(result.stdout), &project); err != nil {
		t.Fatalf("decoding JSON output: %v\nraw: %s", err, result.stdout)
	}
	if project["slug"] != "json-project" {
		t.Errorf("slug = %v, want json-project", project["slug"])
	}
}

func TestProject_Create_SetsActiveProject(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "project", "create", "Active Test")

	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "active project")

	// Verify that after creating, switching to a known project and then
	// listing shows the active marker.
	switchResult := runCLI(t, env, "project", "switch", "alpha")
	assertExitCode(t, switchResult, 0)

	lsResult := runCLI(t, env, "project", "ls")
	assertExitCode(t, lsResult, 0)
	if !strings.Contains(lsResult.stdout, "*") {
		t.Errorf("expected active marker '*' after switch, got: %s", lsResult.stdout)
	}
}

func TestProject_Create_MissingName(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "project", "create")

	if result.exitCode == 0 {
		t.Error("expected non-zero exit code for missing name arg")
	}
}

// --- project switch ---

func TestProject_Switch_Success(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "project", "switch", "beta")

	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "beta")
}

func TestProject_Switch_JSONOutput(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "--json", "project", "switch", "alpha")

	assertExitCode(t, result, 0)

	var out map[string]string
	if err := json.Unmarshal([]byte(result.stdout), &out); err != nil {
		t.Fatalf("decoding JSON output: %v\nraw: %s", err, result.stdout)
	}
	if out["active_project"] != "alpha" {
		t.Errorf("active_project = %q, want alpha", out["active_project"])
	}
}

func TestProject_Switch_NotFound(t *testing.T) {
	ts := newProjectMockServer(t)
	env := projectEnv(t, ts.URL)

	result := runCLI(t, env, "project", "switch", "nonexistent")

	if result.exitCode == 0 {
		t.Error("expected non-zero exit code for nonexistent project")
	}
	combined := result.stdout + result.stderr
	if !strings.Contains(combined, "not found") {
		t.Errorf("expected 'not found' in output, got stdout=%q stderr=%q", result.stdout, result.stderr)
	}
}
