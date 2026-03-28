package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/output"
)

// newErrorsCmd creates the `muga errors` parent command.
func newErrorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "errors",
		Short: "Manage error groups",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderNounHelp(cmd.OutOrStdout(), cmd, output.FromContext(cmd.Context()))
		},
	}
}
