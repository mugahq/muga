package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/output"
)

// newAlertsCmd creates the `muga alerts` parent command.
func newAlertsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "alerts",
		Short: "Manage alert rules",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderNounHelp(cmd.OutOrStdout(), cmd, output.FromContext(cmd.Context()))
		},
	}
}
