package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/output"
)

// newAuthCmd creates the `muga auth` parent command.
func newAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderNounHelp(cmd.OutOrStdout(), cmd, output.FromContext(cmd.Context()))
		},
	}
}
