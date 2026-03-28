package style

import (
	"fmt"
	"strings"
)

const (
	brand     = "muga"
	dash      = "\u2500"   // U+2500 BOX DRAWINGS LIGHT HORIZONTAL
	numDashes = 7          // fixed dash count
	dotSep    = " \u00b7 " // " · "
)

// SignatureLine renders "muga ─────── project · tier" with a fixed 7-dash separator.
// When tier is empty, only the project is shown. When both are empty, only the
// brand and dashes are rendered.
func (r *Renderer) SignatureLine(project, tier string) string {
	dashes := strings.Repeat(dash, numDashes)
	base := r.purple(brand) + " " + r.purple(dashes)

	suffix := buildSuffix(project, tier)
	if suffix != "" {
		return base + r.muted(" "+suffix)
	}
	return base
}

// buildSuffix joins project and tier with " · " when both are present.
func buildSuffix(project, tier string) string {
	switch {
	case project != "" && tier != "":
		return project + dotSep + tier
	case project != "":
		return project
	default:
		return ""
	}
}

// Tagline returns the branded tagline in dimmed purple.
func (r *Renderer) Tagline() string {
	return r.purpleDim("observability for developers who ship from the terminal")
}

// SectionHeader renders a section title in uppercase purple.
func (r *Renderer) SectionHeader(title string) string {
	return r.purple(strings.ToUpper(title))
}

// CommandRow renders a command name and description with fixed-width alignment.
func (r *Renderer) CommandRow(name, description string, nameWidth int) string {
	padded := fmt.Sprintf("%-*s", nameWidth, name)
	if description == "" {
		return "  " + padded
	}
	return "  " + padded + "  " + description
}

// Footer joins parts with a dot separator in muted style.
func (r *Renderer) Footer(parts ...string) string {
	return r.muted(strings.Join(parts, dotSep))
}

// QuickStartStep renders a numbered onboarding step.
func (r *Renderer) QuickStartStep(number int, command, description string) string {
	cmd := r.purple(fmt.Sprintf("%-22s", command))
	return fmt.Sprintf("  %d. %s %s", number, cmd, description)
}
