package cmd

import "github.com/spf13/cobra"

// newAuthCmd creates the `muga auth` parent command.
func newAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}
}
