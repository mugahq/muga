package cmd

import "github.com/spf13/cobra"

// newProjectCmd creates the `muga project` parent command.
func newProjectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}
}
