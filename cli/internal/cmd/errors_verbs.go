package cmd

import "github.com/spf13/cobra"

func newErrorsLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List error groups",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newErrorsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [fingerprint]",
		Short: "Show error group details",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}
