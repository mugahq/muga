package cmd

import "github.com/spf13/cobra"

// newLogsCmd creates the `muga logs` parent command.
func newLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Manage log entries",
	}
}
