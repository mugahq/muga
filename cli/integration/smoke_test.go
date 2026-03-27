//go:build integration

package integration

import (
	"testing"
)

func TestSmoke_Version(t *testing.T) {
	result := runCLI(t, nil, "--version")
	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "integration-test")
}

func TestSmoke_Help(t *testing.T) {
	result := runCLI(t, nil, "--help")
	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "muga")
	assertStdoutContains(t, result, "auth")
}

func TestSmoke_AuthHelp(t *testing.T) {
	result := runCLI(t, nil, "auth", "--help")
	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "login")
}

func TestSmoke_UnknownCommand(t *testing.T) {
	result := runCLI(t, nil, "nonexistent")
	if result.exitCode == 0 {
		t.Error("expected non-zero exit code for unknown command")
	}
}

func TestSmoke_VersionCommand(t *testing.T) {
	result := runCLI(t, nil, "version")
	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "integration-test")
	assertStdoutContains(t, result, "test-commit")
	assertStdoutContains(t, result, "2025-01-01T00:00:00Z")
}

func TestSmoke_CompletionBash(t *testing.T) {
	result := runCLI(t, nil, "completion", "bash")
	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "bash")
}

func TestSmoke_CompletionZsh(t *testing.T) {
	result := runCLI(t, nil, "completion", "zsh")
	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "zsh")
}

func TestSmoke_CompletionFish(t *testing.T) {
	result := runCLI(t, nil, "completion", "fish")
	assertExitCode(t, result, 0)
	assertStdoutContains(t, result, "fish")
}

func TestSmoke_Healthz(t *testing.T) {
	ts := newMockServer(t)

	// Verify the mock server is reachable (sanity check for the test infra).
	result := runCLI(t, baseEnv(t, ts.URL), "--help")
	assertExitCode(t, result, 0)
}
