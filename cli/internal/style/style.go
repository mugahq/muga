package style

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/mugahq/muga/cli/internal/output"
)

// Brand palette — the single source of truth for CLI colors.
var (
	Purple    = lipgloss.Color("#7C3AED")
	PurpleDim = lipgloss.Color("#A78BFA")
	Muted     = lipgloss.Color("#6B7280")
	Success   = lipgloss.Color("#10B981")
	Warning   = lipgloss.Color("#F59E0B")
	Error     = lipgloss.Color("#EF4444")
)

// Renderer provides styled text output that respects TTY/color settings.
type Renderer struct {
	isTTY   bool
	noColor bool
}

// NewRenderer creates a Renderer from the existing output.Opts.
func NewRenderer(opts output.Opts) *Renderer {
	return &Renderer{
		isTTY:   opts.IsTTY,
		noColor: opts.NoColor,
	}
}

// styled returns a lipgloss style when color is enabled, or an empty style otherwise.
func (r *Renderer) styled() bool {
	return r.isTTY && !r.noColor
}

// purple returns text in bold purple when styled.
func (r *Renderer) purple(text string) string {
	if !r.styled() {
		return text
	}
	return lipgloss.NewStyle().Bold(true).Foreground(Purple).Render(text)
}

// purpleDim returns text in dimmed purple when styled.
func (r *Renderer) purpleDim(text string) string {
	if !r.styled() {
		return text
	}
	return lipgloss.NewStyle().Foreground(PurpleDim).Render(text)
}

// muted returns text in muted gray when styled.
func (r *Renderer) muted(text string) string {
	if !r.styled() {
		return text
	}
	return lipgloss.NewStyle().Foreground(Muted).Render(text)
}

// TerminalWidth returns the current terminal width, defaulting to 80.
func TerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// IsNarrow returns true when the terminal width is below 40 columns.
func IsNarrow() bool {
	return TerminalWidth() < 40
}
