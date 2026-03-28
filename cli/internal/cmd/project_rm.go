package cmd

import "github.com/spf13/cobra"

func newProjectRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm [slug]",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}
