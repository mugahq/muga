//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuth_Login_DeviceFlow(t *testing.T) {
	ts := newMockServer(t)
	env := baseEnv(t, ts.URL)

	result := runCLI(t, env, "auth", "login")

	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "INTG-1234")
	assertStdoutContains(t, result, "Logged in as Integration User")
}

func TestAuth_Login_SavesCredentials(t *testing.T) {
	ts := newMockServer(t)
	xdgHome := newTempHome(t)
	env := map[string]string{
		"MUGA_API_URL":    ts.URL,
		"XDG_CONFIG_HOME": xdgHome,
		"HOME":            t.TempDir(),
		"BROWSER":         "",
	}

	result := runCLI(t, env, "auth", "login")
	assertExitCode(t, result, 0)

	// Verify credentials file was written.
	credPath := filepath.Join(xdgHome, "muga", "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("reading credentials file: %v", err)
	}

	var cred map[string]any
	if err := json.Unmarshal(data, &cred); err != nil {
		t.Fatalf("decoding credentials: %v", err)
	}

	if got := cred["access_token"]; got != "muga_integration_test_token" {
		t.Errorf("access_token: got %v, want muga_integration_test_token", got)
	}
	if got := cred["user_name"]; got != "Integration User" {
		t.Errorf("user_name: got %v, want Integration User", got)
	}
}

func TestAuth_Login_JSONOutput(t *testing.T) {
	ts := newMockServer(t)
	env := baseEnv(t, ts.URL)

	result := runCLI(t, env, "auth", "login", "--json")

	assertExitCode(t, result, 0)

	// The CLI prints human-readable text before the JSON object.
	// Extract the JSON portion starting from the first '{'.
	raw := result.stdout
	idx := strings.Index(raw, "{")
	if idx == -1 {
		t.Fatalf("no JSON found in stdout:\n%s", raw)
	}

	var output map[string]string
	if err := json.Unmarshal([]byte(raw[idx:]), &output); err != nil {
		t.Fatalf("decoding JSON output: %v\nraw: %s", err, raw)
	}

	if output["status"] != "logged_in" {
		t.Errorf("status: got %q, want logged_in", output["status"])
	}
	if output["user"] != "Integration User" {
		t.Errorf("user: got %q, want Integration User", output["user"])
	}
}

func TestAuth_Login_ServerUnreachable(t *testing.T) {
	env := map[string]string{
		"MUGA_API_URL":    "http://127.0.0.1:1", // Nothing listening.
		"XDG_CONFIG_HOME": newTempHome(t),
		"HOME":            t.TempDir(),
		"BROWSER":         "",
	}

	result := runCLI(t, env, "auth", "login")

	if result.exitCode == 0 {
		t.Error("expected non-zero exit code when server is unreachable")
	}
}
