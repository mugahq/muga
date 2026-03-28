package cmd

import "github.com/spf13/cobra"

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Show a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}

func newConfigLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "Show all configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errNotImplemented
		},
	}
}
