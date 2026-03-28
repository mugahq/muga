package cmd

import "github.com/spf13/cobra"

// newAlertsCmd creates the `muga alerts` parent command.
func newAlertsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "alerts",
		Short: "Manage alert rules",
	}
}
