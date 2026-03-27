package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/config"
	"github.com/mugahq/muga/cli/internal/output"
)

// projectSwitchDeps holds injectable dependencies for the switch command.
type projectSwitchDeps struct {
	apiClient   projectClient
	configSaver configSaver
}

func newProjectSwitchCmd(deps *projectSwitchDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "switch SLUG",
		Short: "Set the active project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps == nil {
				cfg := config.FromContext(cmd.Context())
				client, err := requireAuth(cfg, nil)
				if err != nil {
					return err
				}
				deps = &projectSwitchDeps{
					apiClient:   client,
					configSaver: &defaultConfigSaver{},
				}
			}
			return runProjectSwitch(cmd, args[0], deps)
		},
	}
}

func runProjectSwitch(cmd *cobra.Command, slug string, deps *projectSwitchDeps) error {
	opts := output.FromContext(cmd.Context())
	w := cmd.OutOrStdout()

	projects, err := deps.apiClient.ListProjects()
	if err != nil {
		return err
	}

	var found bool
	for _, p := range projects {
		if p.Slug == slug || p.Name == slug {
			slug = p.Slug
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("project %q not found", slug)
	}

	if err := deps.configSaver.SetProject(slug); err != nil {
		return fmt.Errorf("setting active project: %w", err)
	}

	if opts.JSON {
		return output.RenderJSON(w, map[string]string{
			"active_project": slug,
		})
	}

	_, _ = fmt.Fprintf(w, "Switched to project %q\n", slug)
	return nil
}
