package cmd

import (
	"hingmy/internal/auth"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show yer current session",
	Long:  `Display information aboot yer currently stored auth token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		u := resolveUI(cmd)
		store, err := resolveTokenStore(cmd)
		if err != nil {
			return err
		}

		tok, err := store.Load()
		if err != nil {
			u.Warning("Ye're no logged in — run `hingmy auth login` first.")
			return err
		}

		u.Section("Session Status")
		u.Table([][]string{
			{"Field", "Value"},
			{"Token", auth.MaskToken(tok.AccessToken)},
			{"Expires", tok.ExpiresAt.String()},
		})
		return nil
	},
}
