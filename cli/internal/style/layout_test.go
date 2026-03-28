package style

import (
	"strings"
	"testing"

	"github.com/mugahq/muga/cli/internal/output"
)

func plainRenderer() *Renderer {
	return NewRenderer(output.Opts{IsTTY: false, NoColor: true})
}

func TestSignatureLineNoSuffix(t *testing.T) {
	r := plainRenderer()
	line := r.SignatureLine("", "")
	want := "muga " + strings.Repeat(dash, numDashes)
	if line != want {
		t.Errorf("expected %q, got %q", want, line)
	}
}

func TestSignatureLineWithProject(t *testing.T) {
	r := plainRenderer()
	line := r.SignatureLine("spedr", "")
	want := "muga " + strings.Repeat(dash, numDashes) + " spedr"
	if line != want {
		t.Errorf("expected %q, got %q", want, line)
	}
}

func TestSignatureLineWithProjectAndTier(t *testing.T) {
	r := plainRenderer()
	line := r.SignatureLine("spedr", "pro")
	want := "muga " + strings.Repeat(dash, numDashes) + " spedr · pro"
	if line != want {
		t.Errorf("expected %q, got %q", want, line)
	}
}

func TestTagline(t *testing.T) {
	r := plainRenderer()
	tag := r.Tagline()
	if tag != "observability for developers who ship from the terminal" {
		t.Errorf("unexpected tagline: %q", tag)
	}
}

func TestSectionHeader(t *testing.T) {
	r := plainRenderer()
	header := r.SectionHeader("Observability")
	if header != "OBSERVABILITY" {
		t.Errorf("expected uppercase header, got %q", header)
	}
}

func TestCommandRow(t *testing.T) {
	r := plainRenderer()
	row := r.CommandRow("logs", "Search, tail, and send log entries", 10)
	if !strings.HasPrefix(row, "  logs") {
		t.Errorf("expected leading spaces and name, got %q", row)
	}
	if !strings.Contains(row, "Search, tail") {
		t.Errorf("expected description in row, got %q", row)
	}
}

func TestCommandRowNoDescription(t *testing.T) {
	r := plainRenderer()
	row := r.CommandRow("logs", "", 10)
	if !strings.HasPrefix(row, "  logs") {
		t.Errorf("expected leading spaces and name, got %q", row)
	}
}

func TestFooter(t *testing.T) {
	r := plainRenderer()
	footer := r.Footer("v0.1.0", "muga.sh/docs", "muga [cmd] --help for details")
	if !strings.Contains(footer, "v0.1.0") {
		t.Errorf("expected version in footer, got %q", footer)
	}
	if !strings.Contains(footer, "\u00b7") {
		t.Errorf("expected dot separator in footer, got %q", footer)
	}
}

func TestQuickStartStep(t *testing.T) {
	r := plainRenderer()
	step := r.QuickStartStep(1, "muga auth login", "Sign in with GitHub")
	if !strings.Contains(step, "1.") {
		t.Errorf("expected step number, got %q", step)
	}
	if !strings.Contains(step, "muga auth login") {
		t.Errorf("expected command in step, got %q", step)
	}
	if !strings.Contains(step, "Sign in with GitHub") {
		t.Errorf("expected description in step, got %q", step)
	}
}
