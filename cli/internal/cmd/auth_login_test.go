package cmd

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/api"
	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/output"
)

// --- mocks ---

type mockCredStore struct {
	cred    *auth.Credential
	loadErr error
	saveErr error
	saved   *auth.Credential
}

func (m *mockCredStore) Load() (*auth.Credential, error) {
	return m.cred, m.loadErr
}

func (m *mockCredStore) Save(c *auth.Credential) error {
	m.saved = c
	return m.saveErr
}

type mockAPIClient struct {
	deviceResp *api.DeviceCodeResponse
	deviceErr  error
	tokenResp  *api.TokenResponse
	tokenErr   error
	pollCount  int
}

func (m *mockAPIClient) RequestDeviceCode() (*api.DeviceCodeResponse, error) {
	return m.deviceResp, m.deviceErr
}

func (m *mockAPIClient) PollToken(_ string) (*api.TokenResponse, error) {
	m.pollCount++
	return m.tokenResp, m.tokenErr
}

// --- helpers ---

func runAuthLogin(t *testing.T, deps *loginDeps, args ...string) (string, error) {
	t.Helper()
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev")
	// Replace the login command registered by NewRootCmd with one using our deps.
	authCmd := findSubCommand(root, "auth")
	if authCmd == nil {
		t.Fatal("auth subcommand not found")
	}
	// Remove existing login and add one with test deps.
	authCmd.RemoveCommand(findSubCommand(authCmd, "login"))
	authCmd.AddCommand(newLoginCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs(append([]string{"auth", "login"}, args...))

	err := root.Execute()
	return buf.String(), err
}

func findSubCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// --- tests ---

func TestLoginSuccess(t *testing.T) {
	deps := &loginDeps{
		credStore: &mockCredStore{},
		apiClient: &mockAPIClient{
			deviceResp: &api.DeviceCodeResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://github.com/login/device",
				Interval:        0,
			},
			tokenResp: &api.TokenResponse{
				Status:      "authorized",
				AccessToken: "tok_abc",
				User:        &api.TokenUser{Name: "Alice", Email: "alice@example.com"},
			},
		},
		openBrowser:  func(_ string) error { return nil },
		stdin:        bufio.NewReader(strings.NewReader("")),
		pollInterval: time.Millisecond,
	}

	out, err := runAuthLogin(t, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ABCD-1234") {
		t.Errorf("expected user code in output, got %q", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected user name in output, got %q", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected checkmark in output, got %q", out)
	}

	store := deps.credStore.(*mockCredStore)
	if store.saved == nil {
		t.Fatal("expected credentials to be saved")
	}
	if store.saved.AccessToken != "tok_abc" {
		t.Errorf("saved access_token = %q, want tok_abc", store.saved.AccessToken)
	}
}

func TestLoginBrowserFallback(t *testing.T) {
	deps := &loginDeps{
		credStore: &mockCredStore{},
		apiClient: &mockAPIClient{
			deviceResp: &api.DeviceCodeResponse{
				DeviceCode:      "dev123",
				UserCode:        "WXYZ-5678",
				VerificationURI: "https://github.com/login/device",
				Interval:        0,
			},
			tokenResp: &api.TokenResponse{
				Status:      "authorized",
				AccessToken: "tok",
				User:        &api.TokenUser{Name: "Bob"},
			},
		},
		openBrowser: func(_ string) error {
			return context.DeadlineExceeded
		},
		stdin:        bufio.NewReader(strings.NewReader("")),
		pollInterval: time.Millisecond,
	}

	out, err := runAuthLogin(t, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "browser didn't open") {
		t.Errorf("expected browser fallback message, got %q", out)
	}
	if !strings.Contains(out, "https://github.com/login/device") {
		t.Errorf("expected verification URI in output, got %q", out)
	}
}

func TestLoginServerUnreachable(t *testing.T) {
	deps := &loginDeps{
		credStore: &mockCredStore{},
		apiClient: &mockAPIClient{
			deviceErr: context.DeadlineExceeded,
		},
		openBrowser: func(_ string) error { return nil },
		stdin:       bufio.NewReader(strings.NewReader("")),
	}

	_, err := runAuthLogin(t, deps)
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
	if !strings.Contains(err.Error(), "server unreachable") {
		t.Errorf("expected 'server unreachable' in error, got %q", err.Error())
	}
}

func TestLoginExpiredToken(t *testing.T) {
	deps := &loginDeps{
		credStore: &mockCredStore{},
		apiClient: &mockAPIClient{
			deviceResp: &api.DeviceCodeResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://github.com/login/device",
				Interval:        0,
			},
			tokenResp: &api.TokenResponse{
				Status: "expired_token",
			},
		},
		openBrowser:  func(_ string) error { return nil },
		stdin:        bufio.NewReader(strings.NewReader("")),
		pollInterval: time.Millisecond,
	}

	_, err := runAuthLogin(t, deps)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected 'expired' in error, got %q", err.Error())
	}
}

func TestLoginAlreadyLoggedInDecline(t *testing.T) {
	deps := &loginDeps{
		credStore: &mockCredStore{
			cred: &auth.Credential{AccessToken: "existing"},
		},
		apiClient:   &mockAPIClient{},
		openBrowser: func(_ string) error { return nil },
		stdin:       bufio.NewReader(strings.NewReader("n\n")),
	}

	// IsTTY is false by default in tests, so it should decline automatically.
	out, err := runAuthLogin(t, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cancelled") {
		t.Errorf("expected 'cancelled' in output, got %q", out)
	}
}

func TestLoginAlreadyLoggedInConfirm(t *testing.T) {
	// We need to force IsTTY=true via the context.
	resetViper()

	deps := &loginDeps{
		credStore: &mockCredStore{
			cred: &auth.Credential{AccessToken: "existing"},
		},
		apiClient: &mockAPIClient{
			deviceResp: &api.DeviceCodeResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://github.com/login/device",
			},
			tokenResp: &api.TokenResponse{
				Status:      "authorized",
				AccessToken: "new_tok",
				User:        &api.TokenUser{Name: "Alice"},
			},
		},
		openBrowser:  func(_ string) error { return nil },
		stdin:        bufio.NewReader(strings.NewReader("y\n")),
		pollInterval: time.Millisecond,
	}

	// Build the command manually to inject IsTTY.
	root := NewRootCmd("dev")
	authCmd := findSubCommand(root, "auth")
	authCmd.RemoveCommand(findSubCommand(authCmd, "login"))

	loginCmd := newLoginCmd(deps)
	authCmd.AddCommand(loginCmd)

	var buf bytes.Buffer
	root.SetOut(&buf)

	// Override PersistentPreRunE to set IsTTY=true.
	originalPreRun := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := originalPreRun(cmd, args); err != nil {
			return err
		}
		opts := output.FromContext(cmd.Context())
		opts.IsTTY = true
		return nil
	}

	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	root.SetArgs([]string{"auth", "login"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected 'Alice' in output, got %q", out)
	}

	store := deps.credStore.(*mockCredStore)
	if store.saved == nil || store.saved.AccessToken != "new_tok" {
		t.Error("expected new credentials to be saved")
	}
}

func TestLoginJSONOutput(t *testing.T) {
	resetViper()

	deps := &loginDeps{
		credStore: &mockCredStore{},
		apiClient: &mockAPIClient{
			deviceResp: &api.DeviceCodeResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://github.com/login/device",
			},
			tokenResp: &api.TokenResponse{
				Status:      "authorized",
				AccessToken: "tok",
				User:        &api.TokenUser{Name: "Alice"},
			},
		},
		openBrowser:  func(_ string) error { return nil },
		stdin:        bufio.NewReader(strings.NewReader("")),
		pollInterval: time.Millisecond,
	}

	root := NewRootCmd("dev")
	authCmd := findSubCommand(root, "auth")
	authCmd.RemoveCommand(findSubCommand(authCmd, "login"))
	authCmd.AddCommand(newLoginCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root.SetArgs([]string{"--json", "auth", "login"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"logged_in"`) {
		t.Errorf("expected JSON output with status, got %q", out)
	}
}

func TestLoginEmailFallback(t *testing.T) {
	deps := &loginDeps{
		credStore: &mockCredStore{},
		apiClient: &mockAPIClient{
			deviceResp: &api.DeviceCodeResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://github.com/login/device",
			},
			tokenResp: &api.TokenResponse{
				Status:      "authorized",
				AccessToken: "tok",
				User:        &api.TokenUser{Email: "alice@example.com"},
			},
		},
		openBrowser:  func(_ string) error { return nil },
		stdin:        bufio.NewReader(strings.NewReader("")),
		pollInterval: time.Millisecond,
	}

	out, err := runAuthLogin(t, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "alice@example.com") {
		t.Errorf("expected email in output, got %q", out)
	}
}

func TestAuthSubcommandRegistered(t *testing.T) {
	resetViper()
	root := NewRootCmd("dev")
	authCmd := findSubCommand(root, "auth")
	if authCmd == nil {
		t.Fatal("expected auth subcommand on root")
	}
	loginCmd := findSubCommand(authCmd, "login")
	if loginCmd == nil {
		t.Fatal("expected login subcommand on auth")
	}
}

func TestLoginSaveError(t *testing.T) {
	deps := &loginDeps{
		credStore: &mockCredStore{
			saveErr: context.DeadlineExceeded,
		},
		apiClient: &mockAPIClient{
			deviceResp: &api.DeviceCodeResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://github.com/login/device",
			},
			tokenResp: &api.TokenResponse{
				Status:      "authorized",
				AccessToken: "tok",
				User:        &api.TokenUser{Name: "Alice"},
			},
		},
		openBrowser:  func(_ string) error { return nil },
		stdin:        bufio.NewReader(strings.NewReader("")),
		pollInterval: time.Millisecond,
	}

	_, err := runAuthLogin(t, deps)
	if err == nil {
		t.Fatal("expected error for save failure")
	}
	if !strings.Contains(err.Error(), "saving credentials") {
		t.Errorf("expected 'saving credentials' in error, got %q", err.Error())
	}
}

func TestLoginCredentialLoadError(t *testing.T) {
	deps := &loginDeps{
		credStore: &mockCredStore{
			loadErr: context.DeadlineExceeded,
		},
		apiClient:    &mockAPIClient{},
		openBrowser:  func(_ string) error { return nil },
		stdin:        bufio.NewReader(strings.NewReader("")),
		pollInterval: time.Millisecond,
	}

	_, err := runAuthLogin(t, deps)
	if err == nil {
		t.Fatal("expected error for credential load failure")
	}
}

func TestPollForTokenAuthorizationPending(t *testing.T) {
	calls := 0
	client := &sequenceClient{
		responses: []*api.TokenResponse{
			{Status: "pending"},
			{Status: "pending"},
			{Status: "authorized", AccessToken: "final_tok", User: &api.TokenUser{Name: "Alice"}},
		},
	}

	ctx := context.Background()
	token, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "final_tok" {
		t.Errorf("access_token = %q, want final_tok", token.AccessToken)
	}
	if client.calls != 3 {
		t.Errorf("calls = %d, want 3", client.calls)
	}
	_ = calls
}

type sequenceClient struct {
	calls     int
	responses []*api.TokenResponse
}

func (s *sequenceClient) RequestDeviceCode() (*api.DeviceCodeResponse, error) {
	return nil, nil
}

func (s *sequenceClient) PollToken(_ string) (*api.TokenResponse, error) {
	idx := s.calls
	s.calls++
	if idx < len(s.responses) {
		return s.responses[idx], nil
	}
	return s.responses[len(s.responses)-1], nil
}

func TestPollForTokenContextCancelled(t *testing.T) {
	client := &mockAPIClient{
		tokenResp: &api.TokenResponse{Status: "pending"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestPollForTokenSlowDown(t *testing.T) {
	origIncrement := slowDownIncrement
	slowDownIncrement = time.Millisecond
	t.Cleanup(func() { slowDownIncrement = origIncrement })

	client := &sequenceClient{
		responses: []*api.TokenResponse{
			{Status: "slow_down"},
			{Status: "authorized", AccessToken: "tok_ok", User: &api.TokenUser{Name: "Alice"}},
		},
	}

	ctx := context.Background()
	token, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "tok_ok" {
		t.Errorf("access_token = %q, want tok_ok", token.AccessToken)
	}
	if client.calls != 2 {
		t.Errorf("calls = %d, want 2", client.calls)
	}
}

func TestPollForTokenSlowDownStacks(t *testing.T) {
	origIncrement := slowDownIncrement
	slowDownIncrement = time.Millisecond
	t.Cleanup(func() { slowDownIncrement = origIncrement })

	// Multiple slow_down responses should stack: each adds the increment to the interval.
	client := &sequenceClient{
		responses: []*api.TokenResponse{
			{Status: "slow_down"},
			{Status: "slow_down"},
			{Status: "slow_down"},
			{Status: "authorized", AccessToken: "tok_stacked", User: &api.TokenUser{Name: "Bob"}},
		},
	}

	ctx := context.Background()
	token, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "tok_stacked" {
		t.Errorf("access_token = %q, want tok_stacked", token.AccessToken)
	}
	if client.calls != 4 {
		t.Errorf("calls = %d, want 4 (3 slow_down + 1 success)", client.calls)
	}
}

func TestPollForTokenSlowDownThenPending(t *testing.T) {
	origIncrement := slowDownIncrement
	slowDownIncrement = time.Millisecond
	t.Cleanup(func() { slowDownIncrement = origIncrement })

	// After a slow_down, subsequent pending polls should continue with the
	// increased interval and eventually succeed.
	client := &sequenceClient{
		responses: []*api.TokenResponse{
			{Status: "pending"},
			{Status: "slow_down"},
			{Status: "pending"},
			{Status: "authorized", AccessToken: "tok_mixed", User: &api.TokenUser{Name: "Carol"}},
		},
	}

	ctx := context.Background()
	token, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "tok_mixed" {
		t.Errorf("access_token = %q, want tok_mixed", token.AccessToken)
	}
	if client.calls != 4 {
		t.Errorf("calls = %d, want 4", client.calls)
	}
}

func TestPollForTokenUnknownError(t *testing.T) {
	client := &mockAPIClient{
		tokenResp: &api.TokenResponse{Status: "access_denied"},
	}

	ctx := context.Background()
	_, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err == nil {
		t.Fatal("expected error for access_denied")
	}
	if !strings.Contains(err.Error(), "access_denied") {
		t.Errorf("expected 'access_denied' in error, got %q", err.Error())
	}
}
