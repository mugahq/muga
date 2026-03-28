package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
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

// TestGoldenCoverage ensures every registered CLI command has a corresponding
// golden file and test entry, and that no orphaned golden files exist.
func TestGoldenCoverage(t *testing.T) {
	root := NewRootCmd("dev", "abc123", "2025-01-01")

	// Build expected golden names from the command tree.
	commandGoldens := make(map[string]bool)
	commandGoldens["root_authenticated"] = true
	commandGoldens["root_unauthenticated"] = true

	for _, child := range root.Commands() {
		if child.Hidden || child.Name() == "completion" || child.Name() == "help" {
			continue
		}

		if !child.HasSubCommands() {
			// Standalone command (e.g., version).
			commandGoldens[child.Name()] = true
			continue
		}

		// Noun command — expect a noun help golden.
		commandGoldens["noun_"+child.Name()] = true

		// Verb commands under this noun.
		for _, verb := range child.Commands() {
			if verb.Hidden {
				continue
			}
			commandGoldens[child.Name()+"_"+verb.Name()] = true
		}
	}

	t.Run("all_commands_covered", func(t *testing.T) {
		for name := range commandGoldens {
			path := filepath.Join("testdata", name+".golden")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("missing golden file for command: %s\n"+
					"  Add a TestGolden_* test to golden_screens_test.go and run:\n"+
					"  UPDATE_GOLDEN=1 go test ./internal/cmd/...", name)
			}
		}
	})

	t.Run("no_orphaned_files", func(t *testing.T) {
		src, err := os.ReadFile("golden_screens_test.go")
		if err != nil {
			t.Fatalf("reading golden_screens_test.go: %v", err)
		}

		// Extract all golden names referenced in the test source:
		// 1. String literals in assertGolden calls: assertGolden(t, ..., "name")
		// 2. Function names: TestGolden_PascalName → snake_name
		referencedByLiteral := extractAssertGoldenNames(src)
		referencedByFunc := extractTestGoldenFuncNames(src)

		entries, err := os.ReadDir("testdata")
		if err != nil {
			t.Fatalf("reading testdata: %v", err)
		}

		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".golden") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".golden")

			// Referenced as a string literal (active test)?
			if referencedByLiteral[name] {
				continue
			}

			// Referenced via function name (possibly skipped test)?
			if referencedByFunc[name] {
				continue
			}

			// Is it a variant of a known command? (e.g., project_ls_empty → prefix project_ls)
			if isGoldenVariant(name, referencedByLiteral, referencedByFunc) {
				continue
			}

			t.Errorf("orphaned golden file: testdata/%s.golden\n"+
				"  Not referenced in golden_screens_test.go.\n"+
				"  Add a TestGolden_* entry or remove the file.", name)
		}
	})
}

// extractAssertGoldenNames finds all string literals passed to assertGolden.
func extractAssertGoldenNames(src []byte) map[string]bool {
	re := regexp.MustCompile(`assertGolden\([^,]+,[^,]+,\s*"([^"]+)"\)`)
	matches := re.FindAllSubmatch(src, -1)
	names := make(map[string]bool, len(matches))
	for _, m := range matches {
		names[string(m[1])] = true
	}
	return names
}

// extractTestGoldenFuncNames finds all TestGolden_* function declarations
// and converts the PascalCase suffix to snake_case golden names.
func extractTestGoldenFuncNames(src []byte) map[string]bool {
	re := regexp.MustCompile(`func TestGolden_(\w+)\(`)
	matches := re.FindAllSubmatch(src, -1)
	names := make(map[string]bool, len(matches))
	for _, m := range matches {
		names[pascalToSnake(string(m[1]))] = true
	}
	return names
}

// pascalToSnake converts PascalCase to snake_case.
func pascalToSnake(s string) string {
	var buf bytes.Buffer
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			buf.WriteByte('_')
		}
		buf.WriteRune(unicode.ToLower(r))
	}
	return buf.String()
}

// isGoldenVariant returns true if name is a variant of a known golden
// (e.g., "project_ls_empty" is a variant of "project_ls").
func isGoldenVariant(name string, sets ...map[string]bool) bool {
	for _, s := range sets {
		for known := range s {
			if strings.HasPrefix(name, known+"_") {
				return true
			}
		}
	}
	return false
}
