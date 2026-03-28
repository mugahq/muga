package style

import (
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
