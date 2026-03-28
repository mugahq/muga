package cmd

import "github.com/spf13/cobra"

// newCronCmd creates the `muga cron` parent command.
func newCronCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cron",
		Short: "Manage cron monitors",
	}
}
