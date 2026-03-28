package cmd

import "github.com/spf13/cobra"

// newErrorsCmd creates the `muga errors` parent command.
func newErrorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "errors",
		Short: "Manage error groups",
	}
}
