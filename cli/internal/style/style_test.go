package style

import (
	"strings"
	"testing"

	"github.com/mugahq/muga/cli/internal/output"
)

func TestNewRendererNoColor(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: true})
	if r.styled() {
		t.Error("expected styled() to return false when NoColor is set")
	}
}

func TestNewRendererNotTTY(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: false, NoColor: false})
	if r.styled() {
		t.Error("expected styled() to return false when not a TTY")
	}
}

func TestNewRendererTTYWithColor(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: false})
	if !r.styled() {
		t.Error("expected styled() to return true for TTY with color")
	}
}

func TestTerminalWidthDefault(t *testing.T) {
	// In test environments, GetSize usually fails — should fall back to 80.
	w := TerminalWidth()
	if w <= 0 {
		t.Errorf("expected positive terminal width, got %d", w)
	}
}

func TestSectionHeaderColor(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: false})
	got := r.SectionHeader("auth")

	if !strings.Contains(got, "AUTH") {
		t.Errorf("expected 'AUTH' in output, got %q", got)
	}
}

func TestDetailRow(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: true})
	got := r.DetailRow("Plan", "pro", 10)

	if !strings.HasPrefix(got, "  Plan") {
		t.Errorf("expected 2-space indent with 'Plan', got %q", got)
	}
	if !strings.HasSuffix(got, "pro") {
		t.Errorf("expected 'pro' at end, got %q", got)
	}
}

func TestEmptyHint_NoColor(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: true})
	msg := "No projects yet. Run muga project create NAME to get started."
	got := r.EmptyHint(msg)

	if got != msg {
		t.Errorf("expected plain hint, got %q", got)
	}
}

func TestErrorMessage_NoColor(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: true})
	got := r.ErrorMessage("monitor not found")

	if got != "Error: monitor not found" {
		t.Errorf("expected 'Error: monitor not found', got %q", got)
	}
}

func TestErrorMessage_Color(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: false})
	got := r.ErrorMessage("monitor not found")

	if !strings.Contains(got, "Error: monitor not found") {
		t.Errorf("expected error message in output, got %q", got)
	}
}

func TestDetailRow_Color(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: false})
	got := r.DetailRow("Plan", "pro", 10)

	if !strings.Contains(got, "pro") {
		t.Errorf("expected value 'pro' in output, got %q", got)
	}
}

func TestEmptyHint_Color(t *testing.T) {
	r := NewRenderer(output.Opts{IsTTY: true, NoColor: false})
	got := r.EmptyHint("No items found.")

	if !strings.Contains(got, "No items found.") {
		t.Errorf("expected hint text in output, got %q", got)
	}
}

func TestIsNarrow(t *testing.T) {
	// In test environments TerminalWidth() returns 80 (default), so IsNarrow should be false.
	got := IsNarrow()
	if got {
		t.Error("expected IsNarrow() to be false in test environment (width defaults to 80)")
	}
}
