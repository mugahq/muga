package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/api"
	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/browser"
	"github.com/mugahq/muga/cli/internal/config"
	"github.com/mugahq/muga/cli/internal/output"
)

const (
	defaultPollInterval = 5 * time.Second
	slowDownIncrement   = 5 * time.Second
)

// loginDeps holds injectable dependencies for the login command.
type loginDeps struct {
	credStore    credentialStore
	apiClient    deviceFlowClient
	openBrowser  func(string) error
	stdin        *bufio.Reader
	pollInterval time.Duration // override for testing; 0 means use server value
}

// credentialStore abstracts credential persistence for testing.
type credentialStore interface {
	Load() (*auth.Credential, error)
	Save(cred *auth.Credential) error
}

// deviceFlowClient abstracts the API calls needed for device flow.
type deviceFlowClient interface {
	RequestDeviceCode() (*api.DeviceCodeResponse, error)
	PollToken(deviceCode string) (*api.TokenResponse, error)
}

func newLoginCmd(deps *loginDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with GitHub via device flow",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if deps == nil {
				cfg := config.FromContext(cmd.Context())
				deps = newDefaultLoginDeps(cfg)
			}
			return runLogin(cmd.Context(), cmd, deps)
		},
	}
}

func runLogin(ctx context.Context, cmd *cobra.Command, deps *loginDeps) error {
	opts := output.FromContext(ctx)

	// Check if already logged in.
	existing, err := deps.credStore.Load()
	if err != nil {
		return fmt.Errorf("checking existing credentials: %w", err)
	}
	if existing != nil {
		if !confirmReauth(cmd, opts, deps.stdin) {
			fmt.Fprintln(cmd.OutOrStdout(), "Login cancelled.")
			return nil
		}
	}

	// Step 1: Request device code.
	device, err := deps.apiClient.RequestDeviceCode()
	if err != nil {
		return fmt.Errorf("server unreachable — check your connection or API URL: %w", err)
	}

	w := cmd.OutOrStdout()

	// Step 2: Display code and open browser.
	fmt.Fprintln(w, "\nOpening browser to authenticate with GitHub...")
	fmt.Fprintf(w, "\n  Your code: %s\n", device.UserCode)

	if err := deps.openBrowser(device.VerificationURI); err != nil {
		fmt.Fprintf(w, "\n  If the browser didn't open, visit:\n  %s\n", device.VerificationURI)
	}

	// Step 3: Poll for token.
	interval := deps.pollInterval
	if interval == 0 {
		interval = time.Duration(device.Interval) * time.Second
	}
	if interval == 0 {
		interval = defaultPollInterval
	}

	fmt.Fprintf(w, "\nWaiting for authorization...")

	token, err := pollForToken(ctx, deps.apiClient, device.DeviceCode, interval)
	if err != nil {
		return err
	}

	// Step 4: Save credentials.
	cred := &auth.Credential{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		UserName:     token.UserName,
		UserEmail:    token.UserEmail,
	}
	if token.ExpiresIn > 0 {
		cred.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	if err := deps.credStore.Save(cred); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	// Step 5: Success message.
	identity := token.UserName
	if identity == "" {
		identity = token.UserEmail
	}

	if opts.JSON {
		return output.RenderJSON(w, map[string]string{
			"status": "logged_in",
			"user":   identity,
		})
	}

	fmt.Fprintf(w, " ✓\n\nLogged in as %s\n", identity)
	return nil
}

func pollForToken(ctx context.Context, client deviceFlowClient, deviceCode string, interval time.Duration) (*api.TokenResponse, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			token, err := client.PollToken(deviceCode)
			if err != nil {
				return nil, fmt.Errorf("polling for authorization: %w", err)
			}

			switch token.Error {
			case "":
				return token, nil
			case "authorization_pending":
				continue
			case "slow_down":
				interval += slowDownIncrement
				ticker.Reset(interval)
			case "expired_token":
				return nil, fmt.Errorf("device code expired — please run `muga auth login` again")
			default:
				return nil, fmt.Errorf("authorization failed: %s", token.Error)
			}
		}
	}
}

func confirmReauth(cmd *cobra.Command, opts *output.Opts, reader *bufio.Reader) bool {
	if !opts.IsTTY {
		return false
	}

	fmt.Fprint(cmd.OutOrStdout(), "You are already logged in. Re-authenticate? [y/N] ")

	line, _ := reader.ReadString('\n')
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes"
}

// newDefaultLoginDeps creates the real dependencies for the login command.
func newDefaultLoginDeps(cfg *config.Config) *loginDeps {
	return &loginDeps{
		credStore:   auth.NewCredentialStore(),
		apiClient:   api.NewClient(cfg.APIURL),
		openBrowser: browser.Open,
		stdin:       bufio.NewReader(os.Stdin),
	}
}
