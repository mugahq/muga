package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mugahq/muga/cli/internal/auth"
	"github.com/mugahq/muga/cli/internal/output"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Sign out and clear credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := output.FromContext(cmd.Context())
			store := auth.NewCredentialStore()

			if err := store.Delete(); err != nil {
				return fmt.Errorf("clearing credentials: %w", err)
			}

			w := cmd.OutOrStdout()
			if opts.JSON {
				return output.RenderJSON(w, map[string]bool{"ok": true})
			}

			fmt.Fprintln(w, "Logged out.")
			return nil
		},
	}
}
