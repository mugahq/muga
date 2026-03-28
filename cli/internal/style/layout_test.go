package style

import (
	"strings"
	"testing"

	"github.com/mugahq/muga/cli/internal/output"
)

func plainRenderer() *Renderer {
	return NewRenderer(output.Opts{IsTTY: false, NoColor: true})
}

func TestSignatureLineMinWidth(t *testing.T) {
	r := plainRenderer()
	line := r.SignatureLine(5, "")
	if !strings.HasPrefix(line, "muga ") {
		t.Errorf("expected line to start with 'muga ', got %q", line)
	}
	dashCount := strings.Count(line, dash)
	if dashCount < minDashes {
		t.Errorf("expected at least %d dashes, got %d", minDashes, dashCount)
	}
}

func TestSignatureLineNoSuffix(t *testing.T) {
	r := plainRenderer()
	line := r.SignatureLine(60, "")
	if !strings.HasPrefix(line, "muga ") {
		t.Errorf("expected line to start with 'muga ', got %q", line)
	}
	if strings.Contains(line, " · ") {
		t.Error("expected no suffix separator in line without suffix")
	}
}

func TestSignatureLineWithSuffix(t *testing.T) {
	r := plainRenderer()
	line := r.SignatureLine(60, "my-saas · pro")
	if !strings.HasPrefix(line, "muga ") {
		t.Errorf("expected line to start with 'muga ', got %q", line)
	}
	if !strings.HasSuffix(line, " my-saas · pro") {
		t.Errorf("expected suffix at end, got %q", line)
	}
}

func TestSignatureLineSuffixOmittedWhenTooNarrow(t *testing.T) {
	r := plainRenderer()
	// "muga " (5) + " my-saas · pro" (14) = 19, leaves only 1 dash at width 20
	line := r.SignatureLine(20, "my-saas · pro")
	if strings.Contains(line, "my-saas") {
		t.Errorf("expected suffix to be omitted at narrow width, got %q", line)
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
