package style

import (
	"fmt"
	"strings"
)

const (
	brand     = "muga"
	dash      = "\u2500" // U+2500 BOX DRAWINGS LIGHT HORIZONTAL
	numDashes = 7        // fixed dash count
	dotSep    = " \u00b7 "                 // " · "
)

// SignatureLine renders "muga ─────── suffix" with a fixed 7-dash separator.
func (r *Renderer) SignatureLine(suffix string) string {
	dashes := strings.Repeat(dash, numDashes)
	if suffix != "" {
		return r.purple(brand) + " " +
			r.purple(dashes) +
			r.muted(" "+suffix)
	}
	return r.purple(brand) + " " + r.purple(dashes)
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
