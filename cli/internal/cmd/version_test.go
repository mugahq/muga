package cmd

import (
	"bytes"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	resetViper()

	cmd := NewRootCmd("1.0.0", "abc1234", "2025-06-15T10:00:00Z")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"1.0.0", "abc1234", "2025-06-15T10:00:00Z"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("expected %q in output, got %q", want, out)
		}
	}
}

func TestVersionCmdFormat(t *testing.T) {
	resetViper()

	cmd := NewRootCmd("dev", "unknown", "unknown")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "muga dev (commit: unknown, built: unknown)\n\n"
	if got := buf.String(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
