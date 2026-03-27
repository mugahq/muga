package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestDeviceCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/auth/device" {
			t.Errorf("path = %s, want /auth/device", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DeviceCodeResponse{
			DeviceCode:      "dev123",
			UserCode:        "ABCD-1234",
			VerificationURI: "https://github.com/login/device",
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
		if r.URL.Path != "/auth/token" {
			t.Errorf("path = %s, want /auth/token", r.URL.Path)
		}
		if r.URL.Query().Get("device_code") != "dev123" {
			t.Errorf("device_code = %q, want dev123", r.URL.Query().Get("device_code"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "tok_abc",
			UserName:    "alice",
			UserEmail:   "alice@example.com",
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	resp, err := client.PollToken("dev123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "tok_abc" {
		t.Errorf("access_token = %q, want tok_abc", resp.AccessToken)
	}
}

func TestPollTokenPending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(TokenResponse{
			Error: "authorization_pending",
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	resp, err := client.PollToken("dev123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Error != "authorization_pending" {
		t.Errorf("error = %q, want authorization_pending", resp.Error)
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
		w.Write([]byte("{invalid"))
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
		w.Write([]byte("{invalid"))
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
