package cmd

import "github.com/spf13/cobra"

// newConfigCmd creates the `muga config` parent command.
func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}
}
