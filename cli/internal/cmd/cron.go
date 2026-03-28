package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/output"
)

// newCronCmd creates the `muga cron` parent command.
func newCronCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cron",
		Short: "Manage cron monitors",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderNounHelp(cmd.OutOrStdout(), cmd, output.FromContext(cmd.Context()))
		},
	}
}
