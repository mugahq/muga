package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VersionInfo holds build-time metadata injected via ldflags.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// newVersionCmd creates the `muga version` command.
func newVersionCmd(info VersionInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version, commit, and build date",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "muga %s (commit: %s, built: %s)\n",
				info.Version, info.Commit, info.Date)
		},
	}
}
