package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mugahq/muga/api/models"
)

func ptr[T any](v T) *T { return &v }

func TestRequestDeviceCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/auth/device" {
			t.Errorf("path = %s, want /v1/auth/device", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.DeviceFlowResponse{
			DeviceCode:      "dev123",
			UserCode:        "ABCD-1234",
			VerificationUri: "https://github.com/login/device",
			ExpiresIn:       900,
			Interval:        5,
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	resp, err := client.RequestDeviceCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserCode != "ABCD-1234" {
		t.Errorf("user_code = %q, want ABCD-1234", resp.UserCode)
	}
	if resp.DeviceCode != "dev123" {
		t.Errorf("device_code = %q, want dev123", resp.DeviceCode)
	}
	if resp.Interval != 5 {
		t.Errorf("interval = %d, want 5", resp.Interval)
	}
}

func TestRequestDeviceCodeServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.RequestDeviceCode()
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestPollTokenSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token" {
			t.Errorf("path = %s, want /v1/auth/token", r.URL.Path)
		}

		var req models.PollTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if req.DeviceCode != "dev123" {
			t.Errorf("device_code = %q, want dev123", req.DeviceCode)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.PollTokenResponse{
			Status:      models.Authorized,
			AccessToken: ptr("tok_abc"),
			User:        &models.AuthUser{Name: "alice", Email: "alice@example.com"},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	resp, err := client.PollToken("dev123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == nil || *resp.AccessToken != "tok_abc" {
		t.Errorf("access_token = %v, want tok_abc", resp.AccessToken)
	}
}

func TestPollTokenPending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.PollTokenResponse{
			Status: models.Pending,
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	resp, err := client.PollToken("dev123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != models.Pending {
		t.Errorf("status = %q, want pending", resp.Status)
	}
}

func TestPollTokenServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.PollToken("dev123")
	if err == nil {
		t.Fatal("expected error for 502 response")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.example.com")
	if client.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %q, want https://api.example.com", client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

func TestAPIErrorFormat(t *testing.T) {
	e := &APIError{Code: "not_found", Message: "Monitor not found"}
	got := e.Error()
	want := "not_found: Monitor not found"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestRequestDeviceCodeConnectionError(t *testing.T) {
	client := NewClient("http://127.0.0.1:1")
	_, err := client.RequestDeviceCode()
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestRequestDeviceCodeInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{invalid"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.RequestDeviceCode()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestPollTokenInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{invalid"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.PollToken("dev123")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestPollTokenConnectionError(t *testing.T) {
	client := NewClient("http://127.0.0.1:1")
	_, err := client.PollToken("dev123")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
