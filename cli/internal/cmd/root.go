package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mugahq/muga/cli/internal/config"
	"github.com/mugahq/muga/cli/internal/output"
)

// NewRootCmd creates the root cobra command with all global flags.
func NewRootCmd(version, commit, date string) *cobra.Command {
	var outputOpts output.Opts

	rootCmd := &cobra.Command{
		Use:     "muga",
		Short:   "Muga — observability for AI-assisted development",
		Version: version,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return renderBanner(cmd.OutOrStdout(), cmd, version, nil)
		},
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
			if outputOpts.Tier == "" {
				outputOpts.Tier = cfg.Tier
			}

			// MUGA_OUTPUT=json overrides default output format.
			if os.Getenv("MUGA_OUTPUT") == "json" && !cmd.Flags().Changed("json") {
				outputOpts.JSON = true
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
	authCmd.AddCommand(newLogoutCmd())
	authCmd.AddCommand(newAuthStatusCmd())
	rootCmd.AddCommand(authCmd)

	projectCmd := newProjectCmd()
	projectCmd.AddCommand(newProjectCreateCmd(nil))
	projectCmd.AddCommand(newProjectLsCmd(nil))
	projectCmd.AddCommand(newProjectSwitchCmd(nil))
	projectCmd.AddCommand(newProjectRmCmd())
	rootCmd.AddCommand(projectCmd)

	monitorCmd := newMonitorCmd()
	monitorCmd.AddCommand(newMonitorAddCmd())
	monitorCmd.AddCommand(newMonitorLsCmd())
	monitorCmd.AddCommand(newMonitorRmCmd())
	rootCmd.AddCommand(monitorCmd)

	cronCmd := newCronCmd()
	cronCmd.AddCommand(newCronAddCmd())
	cronCmd.AddCommand(newCronLsCmd())
	cronCmd.AddCommand(newCronRmCmd())
	cronCmd.AddCommand(newCronPingCmd())
	rootCmd.AddCommand(cronCmd)

	alertsCmd := newAlertsCmd()
	alertsCmd.AddCommand(newAlertsAddCmd())
	alertsCmd.AddCommand(newAlertsLsCmd())
	alertsCmd.AddCommand(newAlertsHistoryCmd())
	alertsCmd.AddCommand(newAlertsRmCmd())
	rootCmd.AddCommand(alertsCmd)

	logsCmd := newLogsCmd()
	logsCmd.AddCommand(newLogsSearchCmd())
	logsCmd.AddCommand(newLogsTailCmd())
	logsCmd.AddCommand(newLogsSendCmd())
	rootCmd.AddCommand(logsCmd)

	errorsCmd := newErrorsCmd()
	errorsCmd.AddCommand(newErrorsLsCmd())
	errorsCmd.AddCommand(newErrorsShowCmd())
	rootCmd.AddCommand(errorsCmd)

	configCmd := newConfigCmd()
	configCmd.AddCommand(newConfigSetCmd())
	configCmd.AddCommand(newConfigGetCmd())
	configCmd.AddCommand(newConfigLsCmd())
	rootCmd.AddCommand(configCmd)

	planCmd := newPlanCmd()
	planCmd.AddCommand(newPlanStatusCmd())
	planCmd.AddCommand(newPlanUpgradeCmd())
	rootCmd.AddCommand(planCmd)

	rootCmd.AddCommand(newVersionCmd(VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}))

	// Hide Cobra's built-in completion from root/help output.
	// The command still works via `muga completion bash|zsh|fish`.
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Custom help: root shows the full command reference, noun commands
	// show branded noun help, leaf commands keep Cobra's default.
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		if cmd.Root() == cmd {
			_ = renderFullHelp(cmd.OutOrStdout(), cmd, version)
			return
		}
		if cmd.HasSubCommands() {
			opts := output.FromContext(cmd.Context())
			_ = renderNounHelp(cmd.OutOrStdout(), cmd, opts)
			return
		}
		_ = cmd.Usage()
	})

	return rootCmd
}

// Execute runs the root command.
func Execute(version, commit, date string) error {
	return NewRootCmd(version, commit, date).Execute()
}
