package cmd

import (
	"hingmy/internal/auth"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log oot o' hingmy",
	Long:  `Revoke yer session token an' remove it frae yer local store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		u := resolveUI(cmd)
		store, err := resolveTokenStore(cmd)
		if err != nil {
			return err
		}

		// Use OAuthAuth's Logout so that provider-side revocation is attempted.
		// For password-based sessions the base Logout (store.Delete) is sufficient.
		svc := auth.NewOAuthAuth(store, auth.OAuthConfig{}, auth.WithUI(u))

		if err := svc.Logout(); err != nil {
			u.Warning("Logout encountered an error: " + err.Error())
			return err
		}

		u.Info("Cheerio! Ye've been logged oot.")
		return nil
	},
}
