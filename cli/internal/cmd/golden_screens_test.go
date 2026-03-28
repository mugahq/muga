package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/api/models"
	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/output"
)

// --- helpers ----------------------------------------------------------------

func noColorOpts(project, tier string) *output.Opts {
	return &output.Opts{IsTTY: true, NoColor: true, Project: project, Tier: tier}
}

func nounTestCmd(t *testing.T, noun, project, tier string) string {
	t.Helper()
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev", "abc123", "2025-01-01")

	// Override PersistentPreRunE to inject test opts instead of DetectTTY.
	originalPreRun := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := originalPreRun(cmd, args); err != nil {
			return err
		}
		opts := output.FromContext(cmd.Context())
		opts.IsTTY = true
		opts.NoColor = true
		opts.Project = project
		opts.Tier = tier
		return nil
	}

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{noun})

	if err := root.Execute(); err != nil {
		t.Fatalf("executing %s: %v", noun, err)
	}
	return buf.String()
}

// --- noun help screens ------------------------------------------------------

func TestGolden_NounAuth(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "auth", "spedr", "pro"), "noun_auth")
}

func TestGolden_NounProject(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "project", "spedr", "pro"), "noun_project")
}

func TestGolden_NounMonitor(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "monitor", "spedr", "pro"), "noun_monitor")
}

func TestGolden_NounCron(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "cron", "spedr", "pro"), "noun_cron")
}

func TestGolden_NounAlerts(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "alerts", "spedr", "pro"), "noun_alerts")
}

func TestGolden_NounLogs(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "logs", "spedr", "pro"), "noun_logs")
}

func TestGolden_NounErrors(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "errors", "spedr", "pro"), "noun_errors")
}

func TestGolden_NounConfig(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "config", "spedr", "pro"), "noun_config")
}

func TestGolden_NounPlan(t *testing.T) {
	assertGolden(t, nounTestCmd(t, "plan", "spedr", "pro"), "noun_plan")
}

// --- root banner screens ----------------------------------------------------

func TestGolden_RootAuthenticated(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	opts := noColorOpts("spedr", "pro")

	// Create credential store with valid creds.
	dir := t.TempDir()
	store := auth.NewCredentialStoreWithDir(dir)
	_ = store.Save(&auth.Credential{AccessToken: "tok_test"})

	var buf bytes.Buffer
	deps := &bannerDeps{credStore: store}

	if err := renderBannerTTY(&buf, "dev", deps, opts); err != nil {
		t.Fatalf("rendering banner: %v", err)
	}
	assertGolden(t, buf.String(), "root_authenticated")
}

func TestGolden_RootUnauthenticated(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	opts := noColorOpts("", "")

	// Empty credential store.
	dir := t.TempDir()
	store := auth.NewCredentialStoreWithDir(dir)
	deps := &bannerDeps{credStore: store}

	var buf bytes.Buffer
	if err := renderBannerTTY(&buf, "dev", deps, opts); err != nil {
		t.Fatalf("rendering banner: %v", err)
	}
	assertGolden(t, buf.String(), "root_unauthenticated")
}

// --- implemented verb screens -----------------------------------------------

func TestGolden_ProjectLs(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	created := time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC)
	deps := &projectLsDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "audiarc", Slug: "audiarc", CreatedAt: created},
				{Name: "spedr", Slug: "spedr", CreatedAt: created},
			},
		},
	}

	out, err := execProjectLs(t, deps, "--project", "spedr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, out, "project_ls")
}

func TestGolden_ProjectLsEmpty(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	deps := &projectLsDeps{
		apiClient: &mockProjectClient{projects: []models.Project{}},
	}

	out, err := execProjectLs(t, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, out, "project_ls_empty")
}

func TestGolden_ProjectCreate(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	deps := &projectCreateDeps{
		apiClient: &mockProjectClient{
			created: &models.Project{Name: "my-api", Slug: "my-api"},
		},
		configSaver: &mockConfigSaver{},
	}

	out, err := execProjectCreate(t, deps, "my-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, out, "project_create")
}

func TestGolden_ProjectSwitch(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	deps := &projectSwitchDeps{
		apiClient: &mockProjectClient{
			projects: []models.Project{
				{Name: "spedr", Slug: "spedr"},
			},
		},
		configSaver: &mockConfigSaver{},
	}

	out, err := execProjectSwitch(t, deps, "spedr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, out, "project_switch")
}

func TestGolden_AuthLogout(t *testing.T) {
	resetViper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Create credential to delete.
	store := auth.NewCredentialStoreWithDir(dir + "/muga")
	_ = store.Save(&auth.Credential{AccessToken: "tok_test"})

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "logout"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, buf.String(), "auth_logout")
}

func TestGolden_AuthStatusUnauth(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, buf.String(), "auth_status_unauth")
}

// --- standalone verb screens ------------------------------------------------

func TestGolden_Version(t *testing.T) {
	resetViper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	root := NewRootCmd("dev", "abc123", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertGolden(t, buf.String(), "version")
}

// --- stub verb screens (golden file is the spec) ----------------------------

func TestGolden_AuthLogin(t *testing.T) {
	t.Skip("not yet implemented — auth login is interactive (device code flow)")
}

func TestGolden_ProjectRm(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_AuthStatus(t *testing.T) {
	t.Skip("not yet implemented — auth status does not yet return tier/full user info from API")
}

func TestGolden_MonitorLs(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_MonitorLsEmpty(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_MonitorAdd(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_MonitorRm(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_CronLs(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_CronLsEmpty(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_CronAdd(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_CronPing(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_CronRm(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_AlertsLs(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_AlertsLsEmpty(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_AlertsAdd(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_AlertsHistory(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_AlertsHistoryEmpty(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_AlertsRm(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_LogsSearch(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_LogsSearchEmpty(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_LogsTail(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_LogsSend(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ErrorsLs(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ErrorsLsEmpty(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ErrorsShow(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ConfigLs(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ConfigSet(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ConfigGet(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_PlanStatus(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_PlanUpgrade(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ErrorNotAuthenticated(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ErrorNotFound(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ErrorTierLimit(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}

func TestGolden_ErrorMissingFlag(t *testing.T) {
	t.Skip("not yet implemented — golden file is the spec")
}
