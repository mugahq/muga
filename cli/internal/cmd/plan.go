package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/output"
)

// newPlanCmd creates the `muga plan` parent command.
func newPlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Manage plan and billing",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderNounHelp(cmd.OutOrStdout(), cmd, output.FromContext(cmd.Context()))
		},
	}
}
