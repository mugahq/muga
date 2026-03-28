package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/output"
	"github.com/mugahq/muga/cli/internal/style"
)

// bannerDeps holds injectable dependencies for banner rendering.
type bannerDeps struct {
	credStore *auth.CredentialStore
}

// commandGroup defines a section of commands shown in the authenticated banner.
type commandGroup struct {
	title    string
	commands []commandEntry
}

type commandEntry struct {
	name string
	desc string
}

var observabilityGroup = commandGroup{
	title: "Observability",
	commands: []commandEntry{
		{"logs", "Search, tail, and send log entries"},
		{"monitor", "Check if your stuff is up"},
		{"cron", "Know when a job goes silent"},
		{"errors", "Track and group runtime errors"},
		{"alerts", "Get notified when things break"},
	},
}

var setupGroup = commandGroup{
	title: "Setup",
	commands: []commandEntry{
		{"auth", "Sign in, sign out, check status"},
		{"project", "Switch between projects"},
		{"config", "Tune CLI preferences"},
		{"plan", "View your plan and usage"},
	},
}

// renderBanner writes the branded root output to w based on output mode.
func renderBanner(w io.Writer, cmd *cobra.Command, version string, deps *bannerDeps) error {
	opts := output.FromContext(cmd.Context())

	if opts.JSON {
		return renderBannerJSON(w, version, deps, opts)
	}

	if !opts.IsTTY {
		return renderBannerPlain(w, cmd)
	}

	return renderBannerTTY(w, version, deps, opts)
}

// renderBannerJSON outputs machine-readable JSON with version and auth state.
func renderBannerJSON(w io.Writer, version string, deps *bannerDeps, opts *output.Opts) error {
	cred := loadCredential(deps)

	data := map[string]any{
		"version":       version,
		"authenticated": cred != nil,
		"project":       nil,
		"tier":          nil,
	}

	if cred != nil {
		if opts.Project != "" {
			data["project"] = opts.Project
		}
		if opts.Tier != "" {
			data["tier"] = opts.Tier
		}
	}

	return output.RenderJSON(w, data)
}

// renderBannerPlain outputs command names only, one per line — for pipes and CI.
func renderBannerPlain(w io.Writer, cmd *cobra.Command) error {
	for _, c := range visibleSubcommands(cmd) {
		_, _ = fmt.Fprintln(w, c.Name())
	}
	return nil
}

// renderBannerTTY outputs the full branded banner with styling.
func renderBannerTTY(w io.Writer, version string, deps *bannerDeps, opts *output.Opts) error {
	r := style.NewRenderer(*opts)
	narrow := style.IsNarrow()

	cred := loadCredential(deps)
	authenticated := cred != nil

	// Signature line with optional project/tier suffix.
	var project, tier string
	if authenticated && opts.Project != "" {
		project = opts.Project
		tier = opts.Tier
	}
	_, _ = fmt.Fprintln(w, r.SignatureLine(project, tier))

	_, _ = fmt.Fprintln(w, r.Tagline())

	_, _ = fmt.Fprintln(w)

	if !authenticated {
		return renderQuickStart(w, r)
	}

	return renderCommandGroups(w, r, narrow, version)
}

func renderQuickStart(w io.Writer, r *style.Renderer) error {
	_, _ = fmt.Fprintln(w, "Quick start:")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, r.QuickStartStep(1, "muga auth login", "Sign in with GitHub"))
	_, _ = fmt.Fprintln(w, r.QuickStartStep(2, "muga project create", "Create your first project"))
	_, _ = fmt.Fprintln(w, r.QuickStartStep(3, `muga logs send "hello"`, "Send a test log entry"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Run muga help for all commands.")
	return nil
}

func renderCommandGroups(w io.Writer, r *style.Renderer, narrow bool, version string) error {
	nameWidth := maxCommandNameWidth()

	for i, group := range []commandGroup{observabilityGroup, setupGroup} {
		_, _ = fmt.Fprintln(w, r.SectionHeader(group.title))
		for _, entry := range group.commands {
			if narrow {
				_, _ = fmt.Fprintln(w, r.CommandRow(entry.name, "", nameWidth))
			} else {
				_, _ = fmt.Fprintln(w, r.CommandRow(entry.name, entry.desc, nameWidth))
			}
		}
		if i == 0 {
			_, _ = fmt.Fprintln(w)
		}
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, r.Footer("v"+strings.TrimPrefix(version, "v"), "muga.sh/docs", "muga [cmd] --help for details"))
	return nil
}

func maxCommandNameWidth() int {
	max := 0
	for _, group := range []commandGroup{observabilityGroup, setupGroup} {
		for _, entry := range group.commands {
			if len(entry.name) > max {
				max = len(entry.name)
			}
		}
	}
	return max
}

func loadCredential(deps *bannerDeps) *auth.Credential {
	var store *auth.CredentialStore
	if deps != nil && deps.credStore != nil {
		store = deps.credStore
	} else {
		store = auth.NewCredentialStore()
	}

	cred, _ := store.Load()
	return cred
}

// renderFullHelp writes the extended help reference used by `muga help`.
func renderFullHelp(w io.Writer, cmd *cobra.Command, version string) error {
	opts := output.FromContext(cmd.Context())

	if !opts.IsTTY {
		_, _ = fmt.Fprintln(w, "muga")
		return renderBannerPlain(w, cmd)
	}

	r := style.NewRenderer(*opts)

	_, _ = fmt.Fprintln(w, r.SignatureLine("", ""))
	_, _ = fmt.Fprintln(w, r.Tagline())
	_, _ = fmt.Fprintln(w)

	// Expanded commands: show subcommands.
	nameWidth := 0
	type flatCmd struct {
		name string
		desc string
	}
	var commands []flatCmd

	for _, c := range cmd.Commands() {
		if c.Hidden || !c.IsAvailableCommand() {
			continue
		}
		if c.HasSubCommands() {
			for _, sub := range c.Commands() {
				if sub.Hidden || !sub.IsAvailableCommand() {
					continue
				}
				full := c.Name() + " " + sub.Name()
				if len(full) > nameWidth {
					nameWidth = len(full)
				}
				commands = append(commands, flatCmd{full, sub.Short})
			}
		} else {
			if len(c.Name()) > nameWidth {
				nameWidth = len(c.Name())
			}
			commands = append(commands, flatCmd{c.Name(), c.Short})
		}
	}

	_, _ = fmt.Fprintln(w, r.SectionHeader("Commands"))
	for _, c := range commands {
		_, _ = fmt.Fprintln(w, r.CommandRow(c.name, c.desc, nameWidth))
	}

	// Global flags section.
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, r.SectionHeader("Global flags"))

	flags := []struct{ flag, desc string }{
		{"--json", "Output in JSON format"},
		{"--project, -p", "Project slug (env: MUGA_PROJECT)"},
		{"--no-color", "Disable colored output"},
		{"--verbose, -v", "Enable verbose output"},
	}

	flagWidth := 0
	for _, f := range flags {
		if len(f.flag) > flagWidth {
			flagWidth = len(f.flag)
		}
	}
	for _, f := range flags {
		_, _ = fmt.Fprintf(w, "  %-*s  %s\n", flagWidth, f.flag, f.desc)
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, r.Footer("v"+strings.TrimPrefix(version, "v"), "muga.sh/docs", "muga [cmd] --help for details"))
	return nil
}

// renderNounHelp renders a branded help screen for noun commands (commands
// with subcommands, excluding root). Pattern:
//
//	muga ──────────────────────────────────── my-saas · pro
//
//	AUTH
//	  login       Sign in with GitHub
//	  logout      Sign out and clear credentials
//
//	muga auth [cmd] --help for details
func renderNounHelp(w io.Writer, cmd *cobra.Command, opts *output.Opts) error {
	if opts == nil {
		opts = &output.Opts{}
	}

	if !opts.IsTTY {
		return renderBannerPlain(w, cmd)
	}

	r := style.NewRenderer(*opts)
	narrow := style.IsNarrow()

	// Signature line.
	_, _ = fmt.Fprintln(w, r.SignatureLine(opts.Project, opts.Tier))
	_, _ = fmt.Fprintln(w)

	// Section header = noun name in uppercase.
	_, _ = fmt.Fprintln(w, r.SectionHeader(cmd.Name()))

	// Subcommand rows.
	nameWidth := maxSubcommandWidth(cmd)
	for _, sub := range visibleSubcommands(cmd) {
		if narrow {
			_, _ = fmt.Fprintln(w, r.CommandRow(sub.Name(), "", nameWidth))
		} else {
			_, _ = fmt.Fprintln(w, r.CommandRow(sub.Name(), sub.Short, nameWidth))
		}
	}

	// Footer (skip for narrow terminals).
	if !narrow {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, r.Footer(fmt.Sprintf("muga %s [cmd] --help for details", cmd.Name())))
	}

	return nil
}

// renderSignatureHeader writes the signature line followed by a blank line to w.
// It is used by verb commands to prefix their output with the branded header.
func renderSignatureHeader(w io.Writer, opts *output.Opts) {
	r := style.NewRenderer(*opts)
	_, _ = fmt.Fprintln(w, r.SignatureLine(opts.Project, opts.Tier))
	_, _ = fmt.Fprintln(w)
}

// visibleSubcommands returns non-hidden, non-auto-generated subcommands.
func visibleSubcommands(cmd *cobra.Command) []*cobra.Command {
	var result []*cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Hidden {
			continue
		}
		// Skip auto-generated cobra commands.
		if sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		result = append(result, sub)
	}
	return result
}

// maxSubcommandWidth returns the length of the longest visible subcommand name.
func maxSubcommandWidth(cmd *cobra.Command) int {
	max := 0
	for _, sub := range visibleSubcommands(cmd) {
		if len(sub.Name()) > max {
			max = len(sub.Name())
		}
	}
	return max
}
