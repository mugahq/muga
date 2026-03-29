package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/api/models"
	"github.com/mugahq/muga/cli/internal/config"
	"github.com/mugahq/muga/cli/internal/output"
)

// projectClient abstracts the API calls needed for project commands.
type projectClient interface {
	ListProjects() ([]models.Project, error)
	CreateProject(req models.CreateProjectRequest) (*models.Project, error)
}

// configSaver abstracts config persistence for testing.
type configSaver interface {
	SetProject(slug string) error
}

// defaultConfigSaver uses the real config package.
type defaultConfigSaver struct{}

func (d *defaultConfigSaver) SetProject(slug string) error {
	return config.SetProject(slug)
}

// projectCreateDeps holds injectable dependencies for the create command.
type projectCreateDeps struct {
	apiClient   projectClient
	configSaver configSaver
}

var slugRe = regexp.MustCompile(`[^a-z0-9-]+`)

// slugify converts a project name to a URL-safe slug.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func newProjectCreateCmd(deps *projectCreateDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps == nil {
				cfg := config.FromContext(cmd.Context())
				client, err := requireAuth(cfg, nil)
				if err != nil {
					return err
				}
				deps = &projectCreateDeps{
					apiClient:   client,
					configSaver: &defaultConfigSaver{},
				}
			}
			return runProjectCreate(cmd, args[0], deps)
		},
	}
}

func runProjectCreate(cmd *cobra.Command, name string, deps *projectCreateDeps) error {
	opts := output.FromContext(cmd.Context())
	w := cmd.OutOrStdout()

	slug := slugify(name)
	if slug == "" {
		return fmt.Errorf("invalid project name: cannot generate a slug from %q", name)
	}

	project, err := deps.apiClient.CreateProject(models.CreateProjectRequest{
		Name: name,
		Slug: slug,
	})
	if err != nil {
		return err
	}

	if err := deps.configSaver.SetProject(project.Slug); err != nil {
		return fmt.Errorf("setting active project: %w", err)
	}

	if opts.JSON {
		return output.RenderJSON(w, project)
	}

	renderSignatureHeader(w, opts)
	_, _ = fmt.Fprintf(w, "Project created: %s (%s)\n", project.Name, project.Slug)
	_, _ = fmt.Fprintf(w, "Set as active project.\n")
	_, _ = fmt.Fprintln(w)
	return nil
}
