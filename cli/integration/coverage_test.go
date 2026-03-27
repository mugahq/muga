//go:build integration

package integration

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// TestIntegration_CommandCoverage verifies that every leaf command registered
// in the CLI has at least one integration test. This prevents adding new
// commands without corresponding integration tests.
//
// It works by asking the binary itself for its command tree and then checking
// that a test function named Test<Noun>_<Verb>_* exists in this package.
func TestIntegration_CommandCoverage(t *testing.T) {
	commands := discoverLeafCommands(t)
	tests := discoverTestFunctions(t)

	for _, cmd := range commands {
		prefix := testPrefix(cmd)
		if !hasMatchingTest(prefix, tests) {
			t.Errorf("command %q has no integration test (expected Test%s_*)", cmd, prefix)
		}
	}
}

// discoverLeafCommands returns all leaf commands (e.g., "auth login") by
// parsing the output of `muga --help` recursively.
func discoverLeafCommands(t *testing.T) []string {
	t.Helper()
	return walkCommands(t, nil)
}

// walkCommands recursively discovers commands by calling `muga <prefix> --help`.
func walkCommands(t *testing.T, prefix []string) []string {
	t.Helper()
	args := append(prefix, "--help")
	out, err := exec.Command(cliBin, args...).Output()
	if err != nil {
		t.Fatalf("running %s %s: %v", cliBin, strings.Join(args, " "), err)
	}

	subs := parseSubcommands(string(out))
	if len(subs) == 0 && len(prefix) > 0 {
		// Leaf command.
		return []string{strings.Join(prefix, " ")}
	}

	var leaves []string
	for _, sub := range subs {
		leaves = append(leaves, walkCommands(t, append(prefix, sub))...)
	}
	return leaves
}

// parseSubcommands extracts subcommand names from cobra's help output.
// It looks for the "Available Commands:" section and extracts the first word
// of each indented line.
var subcommandRe = regexp.MustCompile(`(?m)^\s{2}(\w[\w-]*)`)

func parseSubcommands(helpOutput string) []string {
	// Find the "Available Commands:" section.
	idx := strings.Index(helpOutput, "Available Commands:")
	if idx == -1 {
		return nil
	}
	section := helpOutput[idx:]

	// The section ends at the next blank line or "Flags:" header.
	if end := strings.Index(section, "\nFlags:"); end != -1 {
		section = section[:end]
	}
	if end := strings.Index(section, "\nUse \""); end != -1 {
		section = section[:end]
	}

	matches := subcommandRe.FindAllStringSubmatch(section, -1)
	var cmds []string
	for _, m := range matches {
		name := m[1]
		// Skip meta-commands.
		if name == "help" || name == "completion" || name == "version" {
			continue
		}
		cmds = append(cmds, name)
	}
	return cmds
}

// testPrefix converts a command path like "auth login" to "Auth_Login"
// which is the expected test function prefix.
func testPrefix(cmd string) string {
	parts := strings.Fields(cmd)
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "_")
}

// discoverTestFunctions returns all test function names in this package by
// running `go test -list '.*'`.
func discoverTestFunctions(t *testing.T) []string {
	t.Helper()
	cmd := exec.Command("go", "test", "-tags", "integration", "-list", ".*", "./integration/")
	cmd.Dir = ".."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("listing test functions: %v", err)
	}

	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Test") {
			names = append(names, line)
		}
	}
	return names
}

// hasMatchingTest checks if any test function starts with "Test" + prefix.
func hasMatchingTest(prefix string, tests []string) bool {
	target := "Test" + prefix
	for _, name := range tests {
		if strings.HasPrefix(name, target) {
			return true
		}
	}
	return false
}
