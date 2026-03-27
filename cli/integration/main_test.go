//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// cliBin is the path to the compiled CLI binary, set once in TestMain.
var cliBin string

func TestMain(m *testing.M) {
	bin, err := buildCLI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "building CLI binary: %v\n", err)
		os.Exit(1)
	}
	cliBin = bin

	os.Exit(m.Run())
}

// buildCLI compiles the muga binary into a temp directory.
func buildCLI() (string, error) {
	dir, err := os.MkdirTemp("", "muga-integration-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	bin := filepath.Join(dir, "muga")
	cmd := exec.Command("go", "build",
		"-ldflags", "-X main.version=integration-test -X main.commit=test-commit -X main.date=2025-01-01T00:00:00Z",
		"-o", bin, "./cmd/muga")
	cmd.Dir = filepath.Join("..", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go build: %w", err)
	}

	return bin, nil
}

// cliResult holds the output of a CLI invocation.
type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runCLI executes the muga binary with the given args and environment.
func runCLI(t *testing.T, env map[string]string, args ...string) cliResult {
	t.Helper()

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(cliBin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start with a clean environment, inheriting only PATH.
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running CLI: %v", err)
		}
	}

	return cliResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: exitCode,
	}
}

// assertExitCode fails the test if the exit code does not match.
func assertExitCode(t *testing.T, got cliResult, want int) {
	t.Helper()
	if got.exitCode != want {
		t.Errorf("exit code: got %d, want %d\nstdout: %s\nstderr: %s", got.exitCode, want, got.stdout, got.stderr)
	}
}

// assertStdoutContains fails if stdout does not contain the substring.
func assertStdoutContains(t *testing.T, got cliResult, substr string) {
	t.Helper()
	if !strings.Contains(got.stdout, substr) {
		t.Errorf("stdout missing %q\ngot: %s", substr, got.stdout)
	}
}

// newMockServer returns an httptest.Server that implements the Muga API
// contract endpoints needed for integration testing.
func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Health check.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Device flow: initiate.
	mux.HandleFunc("POST /auth/device", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_code":      "integration-device-code",
			"user_code":        "INTG-1234",
			"verification_uri": "https://github.com/login/device",
			"expires_in":       900,
			"interval":         1, // Fast polling for tests.
		})
	})

	// Device flow: poll token — always returns authorized.
	mux.HandleFunc("POST /auth/token", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":        "authorized",
			"access_token":  "muga_integration_test_token",
			"refresh_token": "mugar_integration_test_refresh",
			"expires_in":    3600,
			"user": map[string]string{
				"id":    "00000000-0000-4000-8000-000000000001",
				"email": "test@muga.sh",
				"name":  "Integration User",
				"tier":  "pro",
			},
		})
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

// newTempHome creates an isolated XDG_CONFIG_HOME for the test so the CLI
// does not read/write real user credentials or config.
func newTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// baseEnv returns the minimum env vars for running the CLI against a test server.
func baseEnv(t *testing.T, serverURL string) map[string]string {
	t.Helper()
	return map[string]string{
		"MUGA_API_URL":    serverURL,
		"XDG_CONFIG_HOME": newTempHome(t),
		"HOME":            t.TempDir(), // Fallback for any HOME-based paths.
		"BROWSER":         "",          // Prevent opening a real browser.
	}
}
