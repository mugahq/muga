package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mugahq/muga/cli/internal/config"
	"github.com/mugahq/muga/cli/internal/output"
)

// NewRootCmd creates the root cobra command with all global flags.
func NewRootCmd(version string) *cobra.Command {
	var outputOpts output.Opts

	rootCmd := &cobra.Command{
		Use:     "muga",
		Short:   "Muga — observability for AI-assisted development",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if p := viper.GetString("project"); p != "" && !cmd.Flags().Changed("project") {
				outputOpts.Project = p
			}
			if outputOpts.Project == "" {
				outputOpts.Project = cfg.Project
			}

			outputOpts.DetectTTY()
			ctx := output.WithOpts(cmd.Context(), &outputOpts)
			ctx = config.WithConfig(ctx, cfg)
			cmd.SetContext(ctx)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := rootCmd.PersistentFlags()
	flags.BoolVar(&outputOpts.JSON, "json", false, "Output in JSON format")
	flags.StringVarP(&outputOpts.Project, "project", "p", "", "Project slug (env: MUGA_PROJECT)")
	flags.BoolVar(&outputOpts.NoColor, "no-color", false, "Disable colored output")
	flags.BoolVarP(&outputOpts.Verbose, "verbose", "v", false, "Enable verbose output")

	// Register subcommands.
	authCmd := newAuthCmd()
	authCmd.AddCommand(newLoginCmd(nil))
	rootCmd.AddCommand(authCmd)

	return rootCmd
}

// Execute runs the root command.
func Execute(version string) error {
	return NewRootCmd(version).Execute()
}
