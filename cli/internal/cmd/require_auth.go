package cmd

import (
	"fmt"

	"github.com/mugahq/muga/cli/internal/api"
	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/config"
)

// requireAuth loads stored credentials and returns an authenticated API client.
func requireAuth(cfg *config.Config, credStore *auth.CredentialStore) (*api.Client, error) {
	cred, err := credStore.Load()
	if err != nil {
		return nil, fmt.Errorf("reading credentials: %w", err)
	}
	if cred == nil {
		return nil, fmt.Errorf("not logged in — run 'muga auth login' first")
	}
	return api.NewAuthenticatedClient(cfg.APIURL, cred.AccessToken), nil
}
