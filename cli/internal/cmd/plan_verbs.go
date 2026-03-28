package cmd

import "github.com/spf13/cobra"

func newPlanStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "View your plan and usage",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newPlanUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Open the billing portal",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}
