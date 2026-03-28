package cmd

import "github.com/spf13/cobra"

func newAlertsAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create an alert rule",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	cmd.Flags().String("name", "", "Alert rule name")
	cmd.Flags().String("type", "", "Alert type (e.g., monitor_down, error_spike, cron_missed)")
	cmd.Flags().StringSlice("channel", nil, "Notification channels (e.g., slack, email, webhook:URL)")
	return cmd
}

func newAlertsLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List alert rules",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newAlertsHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history",
		Short: "Show alert timeline",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newAlertsRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm [id]",
		Short: "Delete an alert rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}
