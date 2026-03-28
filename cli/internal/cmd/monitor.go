package cmd

import "github.com/spf13/cobra"

// newMonitorCmd creates the `muga monitor` parent command.
func newMonitorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Manage uptime monitors",
	}
}
