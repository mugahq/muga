package cmd

import "github.com/spf13/cobra"

func newMonitorAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create an uptime monitor",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
	cmd.Flags().String("name", "", "Monitor name")
	cmd.Flags().String("url", "", "URL to monitor")
	cmd.Flags().Int("interval", 60, "Check interval in seconds")
	return cmd
}

func newMonitorLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List uptime monitors",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newMonitorRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm [id]",
		Short: "Delete an uptime monitor",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}
