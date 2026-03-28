package cmd

import "github.com/spf13/cobra"

// newPlanCmd creates the `muga plan` parent command.
func newPlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Manage plan and billing",
	}
}
