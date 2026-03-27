package browser

import (
	"runtime"
	"testing"
)

func TestCommandDarwin(t *testing.T) {
	cmd, args, err := command("darwin", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "open" {
		t.Errorf("cmd = %q, want open", cmd)
	}
	if len(args) != 1 || args[0] != "https://example.com" {
		t.Errorf("args = %v, want [https://example.com]", args)
	}
}

func TestCommandLinux(t *testing.T) {
	cmd, args, err := command("linux", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "xdg-open" {
		t.Errorf("cmd = %q, want xdg-open", cmd)
	}
	if len(args) != 1 || args[0] != "https://example.com" {
		t.Errorf("args = %v, want [https://example.com]", args)
	}
}

func TestCommandWindows(t *testing.T) {
	cmd, args, err := command("windows", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "cmd" {
		t.Errorf("cmd = %q, want cmd", cmd)
	}
	if len(args) != 3 || args[0] != "/c" || args[1] != "start" || args[2] != "https://example.com" {
		t.Errorf("args = %v, want [/c start https://example.com]", args)
	}
}

func TestCommandUnsupported(t *testing.T) {
	_, _, err := command("plan9", "https://example.com")
	if err == nil {
		t.Fatal("expected error for unsupported platform")
	}
}

func TestOpenOnCurrentPlatform(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		// Verify Open doesn't panic. We use an invalid URL to avoid side effects.
		_ = Open("http://localhost:0/test")
	default:
		err := Open("http://localhost:0/test")
		if err == nil {
			t.Error("expected error on unsupported platform")
		}
	}
}
