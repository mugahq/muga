package cmd

import "github.com/spf13/cobra"

func newCronAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a cron monitor",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	cmd.Flags().String("name", "", "Cron monitor name")
	cmd.Flags().Int("interval", 0, "Expected interval in seconds")
	cmd.Flags().Int("grace", 0, "Grace period in seconds")
	return cmd
}

func newCronLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List cron monitors",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newCronRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm [id]",
		Short: "Delete a cron monitor",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newCronPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping [id]",
		Short: "Send a heartbeat ping",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}
