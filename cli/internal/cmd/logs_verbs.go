package cmd

import "github.com/spf13/cobra"

func newLogsSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Query log entries with filters",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	cmd.Flags().String("service", "", "Filter by service name")
	cmd.Flags().String("level", "", "Filter by log level (debug, info, warn, error)")
	cmd.Flags().String("since", "", "Time range (e.g., 1h, 24h, 7d)")
	return cmd
}

func newLogsTailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Stream log entries in real time",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	cmd.Flags().String("service", "", "Filter by service name")
	return cmd
}

func newLogsSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [message]",
		Short: "Send a log entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	cmd.Flags().String("service", "", "Service name")
	cmd.Flags().String("level", "info", "Log level (debug, info, warn, error)")
	return cmd
}
