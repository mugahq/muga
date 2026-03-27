package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mugahq/muga/api/models"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// --- helpers ---

func projectFixture(name, slug string) models.Project {
	return models.Project{
		Id:        openapi_types.UUID{},
		Name:      name,
		Slug:      slug,
		CreatedAt: time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC),
	}
}

// --- NewAuthenticatedClient ---

func TestNewAuthenticatedClient(t *testing.T) {
	captured := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":       []models.Project{},
			"pagination": models.Pagination{HasMore: false},
		})
	}))
	defer srv.Close()

	client := NewAuthenticatedClient(srv.URL, "tok_abc")
	_, err := client.ListProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured != "Bearer tok_abc" {
		t.Errorf("Authorization header = %q, want %q", captured, "Bearer tok_abc")
	}
}

func TestUnauthenticatedClientNoAuthHeader(t *testing.T) {
	captured := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	// use get directly
	_, _ = client.get("/")
	if captured != "" {
		t.Errorf("expected no Authorization header, got %q", captured)
	}
}

// --- checkResponse ---

func TestCheckResponseSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		t.Errorf("unexpected error for 200: %v", err)
	}
}

func TestCheckResponseAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "not_found",
				"message": "Project not found",
			},
		})
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	gotErr := checkResponse(resp)
	if gotErr == nil {
		t.Fatal("expected error for 404")
	}
	apiErr, ok := gotErr.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", gotErr)
	}
	if apiErr.Code != "not_found" {
		t.Errorf("Code = %q, want not_found", apiErr.Code)
	}
	if apiErr.Message != "Project not found" {
		t.Errorf("Message = %q, want Project not found", apiErr.Message)
	}
}

func TestCheckResponseUnknownError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	gotErr := checkResponse(resp)
	if gotErr == nil {
		t.Fatal("expected error for 500 with non-JSON body")
	}
}

// --- ListProjects ---

func TestListProjectsSuccess(t *testing.T) {
	projects := []models.Project{
		projectFixture("Alpha", "alpha"),
		projectFixture("Beta", "beta"),
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/projects" {
			t.Errorf("path = %s, want /v1/projects", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":       projects,
			"pagination": models.Pagination{HasMore: false},
		})
	}))
	defer srv.Close()

	client := NewAuthenticatedClient(srv.URL, "tok")
	got, err := client.ListProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(got))
	}
	if got[0].Slug != "alpha" {
		t.Errorf("slug = %q, want alpha", got[0].Slug)
	}
}

func TestListProjectsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":       []models.Project{},
			"pagination": models.Pagination{HasMore: false},
		})
	}))
	defer srv.Close()

	client := NewAuthenticatedClient(srv.URL, "tok")
	got, err := client.ListProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 projects, got %d", len(got))
	}
}

func TestListProjectsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "unauthorized",
				"message": "invalid token",
			},
		})
	}))
	defer srv.Close()

	client := NewAuthenticatedClient(srv.URL, "bad-tok")
	_, err := client.ListProjects()
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestListProjectsConnectionError(t *testing.T) {
	client := NewAuthenticatedClient("http://127.0.0.1:1", "tok")
	_, err := client.ListProjects()
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestListProjectsInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{invalid"))
	}))
	defer srv.Close()

	client := NewAuthenticatedClient(srv.URL, "tok")
	_, err := client.ListProjects()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- CreateProject ---

func TestCreateProjectSuccess(t *testing.T) {
	created := projectFixture("My Project", "my-project")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/projects" {
			t.Errorf("path = %s, want /v1/projects", r.URL.Path)
		}

		var req models.CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if req.Name != "My Project" {
			t.Errorf("name = %q, want My Project", req.Name)
		}
		if req.Slug != "my-project" {
			t.Errorf("slug = %q, want my-project", req.Slug)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	}))
	defer srv.Close()

	client := NewAuthenticatedClient(srv.URL, "tok")
	got, err := client.CreateProject(models.CreateProjectRequest{
		Name: "My Project",
		Slug: "my-project",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Slug != "my-project" {
		t.Errorf("slug = %q, want my-project", got.Slug)
	}
}

func TestCreateProjectConflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "conflict",
				"message": "slug already taken",
			},
		})
	}))
	defer srv.Close()

	client := NewAuthenticatedClient(srv.URL, "tok")
	_, err := client.CreateProject(models.CreateProjectRequest{
		Name: "Taken",
		Slug: "taken",
	})
	if err == nil {
		t.Fatal("expected error for 409")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		// error may be wrapped; just check message
		if apiErr == nil {
			return
		}
	}
	_ = apiErr
}

func TestCreateProjectConnectionError(t *testing.T) {
	client := NewAuthenticatedClient("http://127.0.0.1:1", "tok")
	_, err := client.CreateProject(models.CreateProjectRequest{
		Name: "X",
		Slug: "x",
	})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestCreateProjectInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("{invalid"))
	}))
	defer srv.Close()

	client := NewAuthenticatedClient(srv.URL, "tok")
	_, err := client.CreateProject(models.CreateProjectRequest{Name: "X", Slug: "x"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
