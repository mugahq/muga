package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/output"
)

// newLogsCmd creates the `muga logs` parent command.
func newLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Manage log entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderNounHelp(cmd.OutOrStdout(), cmd, output.FromContext(cmd.Context()))
		},
	}
}
