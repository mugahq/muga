package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/config"
	"github.com/mugahq/muga/cli/internal/output"
)

// projectLsDeps holds injectable dependencies for the ls command.
type projectLsDeps struct {
	apiClient projectClient
}

func newProjectLsCmd(deps *projectLsDeps) *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if deps == nil {
				cfg := config.FromContext(cmd.Context())
				client, err := requireAuth(cfg, nil)
				if err != nil {
					return err
				}
				deps = &projectLsDeps{apiClient: client}
			}
			return runProjectLs(cmd, deps)
		},
	}
}

func runProjectLs(cmd *cobra.Command, deps *projectLsDeps) error {
	opts := output.FromContext(cmd.Context())
	cfg := config.FromContext(cmd.Context())
	w := cmd.OutOrStdout()

	projects, err := deps.apiClient.ListProjects()
	if err != nil {
		return err
	}

	if opts.JSON {
		return output.RenderJSON(w, projects)
	}

	if len(projects) == 0 {
		fmt.Fprintln(w, "No projects yet. Run muga project create NAME to get started.")
		return nil
	}

	headers := []string{"Name", "Slug", "Created", "Active"}
	rows := make([][]string, 0, len(projects))
	for _, p := range projects {
		active := ""
		if p.Slug == cfg.Project || p.Slug == opts.Project {
			active = "*"
		}
		rows = append(rows, []string{
			p.Name,
			p.Slug,
			p.CreatedAt.Format("2006-01-02"),
			active,
		})
	}

	return output.RenderTable(w, headers, rows)
}
