package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/output"
)

// newConfigCmd creates the `muga config` parent command.
func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderNounHelp(cmd.OutOrStdout(), cmd, output.FromContext(cmd.Context()))
		},
	}
}
