package cmd

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/api/models"
	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/output"
)

// --- helpers ---

func ptr[T any](v T) *T { return &v }

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
	deviceResp *models.DeviceFlowResponse
	deviceErr  error
	tokenResp  *models.PollTokenResponse
	tokenErr   error
	pollCount  int
}

func (m *mockAPIClient) RequestDeviceCode() (*models.DeviceFlowResponse, error) {
	return m.deviceResp, m.deviceErr
}

func (m *mockAPIClient) PollToken(_ string) (*models.PollTokenResponse, error) {
	m.pollCount++
	return m.tokenResp, m.tokenErr
}

// --- helpers ---

func runAuthLogin(t *testing.T, deps *loginDeps, args ...string) (string, error) {
	t.Helper()
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev")
	authCmd := findSubCommand(root, "auth")
	if authCmd == nil {
		t.Fatal("auth subcommand not found")
	}
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
			deviceResp: &models.DeviceFlowResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationUri: "https://github.com/login/device",
				Interval:        0,
			},
			tokenResp: &models.PollTokenResponse{
				Status:      models.Authorized,
				AccessToken: ptr("tok_abc"),
				User:        &models.AuthUser{Name: "Alice", Email: "alice@example.com"},
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
			deviceResp: &models.DeviceFlowResponse{
				DeviceCode:      "dev123",
				UserCode:        "WXYZ-5678",
				VerificationUri: "https://github.com/login/device",
				Interval:        0,
			},
			tokenResp: &models.PollTokenResponse{
				Status:      models.Authorized,
				AccessToken: ptr("tok"),
				User:        &models.AuthUser{Name: "Bob"},
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
			deviceResp: &models.DeviceFlowResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationUri: "https://github.com/login/device",
				Interval:        0,
			},
			tokenResp: &models.PollTokenResponse{
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
	if !strings.Contains(err.Error(), "authorization failed") {
		t.Errorf("expected 'authorization failed' in error, got %q", err.Error())
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

	out, err := runAuthLogin(t, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cancelled") {
		t.Errorf("expected 'cancelled' in output, got %q", out)
	}
}

func TestLoginAlreadyLoggedInConfirm(t *testing.T) {
	resetViper()

	deps := &loginDeps{
		credStore: &mockCredStore{
			cred: &auth.Credential{AccessToken: "existing"},
		},
		apiClient: &mockAPIClient{
			deviceResp: &models.DeviceFlowResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationUri: "https://github.com/login/device",
			},
			tokenResp: &models.PollTokenResponse{
				Status:      models.Authorized,
				AccessToken: ptr("new_tok"),
				User:        &models.AuthUser{Name: "Alice"},
			},
		},
		openBrowser:  func(_ string) error { return nil },
		stdin:        bufio.NewReader(strings.NewReader("y\n")),
		pollInterval: time.Millisecond,
	}

	root := NewRootCmd("dev")
	authCmd := findSubCommand(root, "auth")
	authCmd.RemoveCommand(findSubCommand(authCmd, "login"))
	authCmd.AddCommand(newLoginCmd(deps))

	var buf bytes.Buffer
	root.SetOut(&buf)

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
			deviceResp: &models.DeviceFlowResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationUri: "https://github.com/login/device",
			},
			tokenResp: &models.PollTokenResponse{
				Status:      models.Authorized,
				AccessToken: ptr("tok"),
				User:        &models.AuthUser{Name: "Alice"},
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
			deviceResp: &models.DeviceFlowResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationUri: "https://github.com/login/device",
			},
			tokenResp: &models.PollTokenResponse{
				Status:      models.Authorized,
				AccessToken: ptr("tok"),
				User:        &models.AuthUser{Email: "alice@example.com"},
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
			deviceResp: &models.DeviceFlowResponse{
				DeviceCode:      "dev123",
				UserCode:        "ABCD-1234",
				VerificationUri: "https://github.com/login/device",
			},
			tokenResp: &models.PollTokenResponse{
				Status:      models.Authorized,
				AccessToken: ptr("tok"),
				User:        &models.AuthUser{Name: "Alice"},
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
		responses: []*models.PollTokenResponse{
			{Status: models.Pending},
			{Status: models.Pending},
			{Status: models.Authorized, AccessToken: ptr("final_tok"), User: &models.AuthUser{Name: "Alice"}},
		},
	}

	ctx := context.Background()
	token, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken == nil || *token.AccessToken != "final_tok" {
		t.Errorf("access_token = %v, want final_tok", token.AccessToken)
	}
	if client.calls != 3 {
		t.Errorf("calls = %d, want 3", client.calls)
	}
	_ = calls
}

type sequenceClient struct {
	calls     int
	responses []*models.PollTokenResponse
}

func (s *sequenceClient) RequestDeviceCode() (*models.DeviceFlowResponse, error) {
	return nil, nil
}

func (s *sequenceClient) PollToken(_ string) (*models.PollTokenResponse, error) {
	idx := s.calls
	s.calls++
	if idx < len(s.responses) {
		return s.responses[idx], nil
	}
	return s.responses[len(s.responses)-1], nil
}

func TestPollForTokenContextCancelled(t *testing.T) {
	client := &mockAPIClient{
		tokenResp: &models.PollTokenResponse{Status: models.Pending},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

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
		responses: []*models.PollTokenResponse{
			{Status: models.SlowDown},
			{Status: models.Authorized, AccessToken: ptr("tok_ok"), User: &models.AuthUser{Name: "Alice"}},
		},
	}

	ctx := context.Background()
	token, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken == nil || *token.AccessToken != "tok_ok" {
		t.Errorf("access_token = %v, want tok_ok", token.AccessToken)
	}
	if client.calls != 2 {
		t.Errorf("calls = %d, want 2", client.calls)
	}
}

func TestPollForTokenSlowDownStacks(t *testing.T) {
	origIncrement := slowDownIncrement
	slowDownIncrement = time.Millisecond
	t.Cleanup(func() { slowDownIncrement = origIncrement })

	client := &sequenceClient{
		responses: []*models.PollTokenResponse{
			{Status: models.SlowDown},
			{Status: models.SlowDown},
			{Status: models.SlowDown},
			{Status: models.Authorized, AccessToken: ptr("tok_stacked"), User: &models.AuthUser{Name: "Bob"}},
		},
	}

	ctx := context.Background()
	token, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken == nil || *token.AccessToken != "tok_stacked" {
		t.Errorf("access_token = %v, want tok_stacked", token.AccessToken)
	}
	if client.calls != 4 {
		t.Errorf("calls = %d, want 4 (3 slow_down + 1 success)", client.calls)
	}
}

func TestPollForTokenSlowDownThenPending(t *testing.T) {
	origIncrement := slowDownIncrement
	slowDownIncrement = time.Millisecond
	t.Cleanup(func() { slowDownIncrement = origIncrement })

	client := &sequenceClient{
		responses: []*models.PollTokenResponse{
			{Status: models.Pending},
			{Status: models.SlowDown},
			{Status: models.Pending},
			{Status: models.Authorized, AccessToken: ptr("tok_mixed"), User: &models.AuthUser{Name: "Carol"}},
		},
	}

	ctx := context.Background()
	token, err := pollForToken(ctx, client, "dev123", time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken == nil || *token.AccessToken != "tok_mixed" {
		t.Errorf("access_token = %v, want tok_mixed", token.AccessToken)
	}
	if client.calls != 4 {
		t.Errorf("calls = %d, want 4", client.calls)
	}
}

func TestPollForTokenUnknownError(t *testing.T) {
	client := &mockAPIClient{
		tokenResp: &models.PollTokenResponse{Status: "access_denied"},
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
