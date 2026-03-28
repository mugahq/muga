package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// assertGolden compares got against the golden file testdata/{name}.golden.
// Set UPDATE_GOLDEN=1 to regenerate golden files.
func assertGolden(t *testing.T, got, name string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("creating testdata dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("writing golden file %s: %v", path, err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden file %s: %v\nRun UPDATE_GOLDEN=1 go test ./internal/cmd/... to generate", path, err)
	}
	if got != string(want) {
		t.Errorf("output mismatch for %s\n--- want ---\n%s\n--- got ---\n%s", name, want, got)
	}
}
