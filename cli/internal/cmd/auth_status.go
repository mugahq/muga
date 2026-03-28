package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/output"
)

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication state",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := output.FromContext(cmd.Context())
			store := auth.NewCredentialStore()
			w := cmd.OutOrStdout()

			cred, err := store.Load()
			if err != nil {
				return fmt.Errorf("loading credentials: %w", err)
			}

			if cred == nil {
				if opts.JSON {
					return output.RenderJSON(w, map[string]any{
						"authenticated": false,
					})
				}
				_, _ = fmt.Fprintln(w, "Not logged in. Run muga auth login to authenticate.")
				return nil
			}

			if opts.JSON {
				result := map[string]any{
					"authenticated": true,
					"user":          cred.UserName,
					"email":         cred.UserEmail,
				}
				if !cred.ExpiresAt.IsZero() {
					result["expires_at"] = cred.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z")
				}
				return output.RenderJSON(w, result)
			}

			rows := []output.DetailRow{
				{Key: "User", Value: cred.UserName},
				{Key: "Email", Value: cred.UserEmail},
			}
			if !cred.ExpiresAt.IsZero() {
				rows = append(rows, output.DetailRow{
					Key:   "Expires",
					Value: cred.ExpiresAt.UTC().Format("2006-01-02 15:04 UTC"),
				})
			}
			return output.RenderDetail(w, rows)
		},
	}
}
