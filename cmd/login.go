package cmd

import (
	"context"
	"os"

	"hingmy/internal/auth"

	"github.com/spf13/cobra"
)

var loginOAuth bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log intae hingmy",
	Long:  `Authenticate wi' hingmy usin' a username/password or OAuth 2.0 Device Flow.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		u := resolveUI(cmd)
		store, err := resolveTokenStore(cmd)
		if err != nil {
			return err
		}

		var svc auth.AuthService
		if loginOAuth {
			cfg := auth.OAuthConfig{
				ClientID:           os.Getenv("OAUTH_CLIENT_ID"),
				DeviceAuthEndpoint: os.Getenv("OAUTH_DEVICE_AUTH_ENDPOINT"),
				TokenEndpoint:      os.Getenv("OAUTH_TOKEN_ENDPOINT"),
				RevocationEndpoint: os.Getenv("OAUTH_REVOCATION_ENDPOINT"),
				Scopes:             []string{"read", "write"},
			}
			svc = auth.NewOAuthAuth(store, cfg, auth.WithUI(u))
		} else {
			apiBase := os.Getenv("API_BASE_URL")
			if apiBase == "" {
				apiBase = "https://api.hingmy.example.com"
			}
			svc = auth.NewPasswordAuth(store, apiBase, auth.WithPasswordUI(u))
		}

		token, err := svc.Login(context.Background())
		if err != nil {
			u.Warning("Login failed — run wi' --debug for mair details")
			return err
		}

		u.Info("Logged in! Token: " + auth.MaskToken(token))
		return nil
	},
}

func init() {
	loginCmd.Flags().BoolVar(&loginOAuth, "oauth", false, "Use OAuth 2.0 Device Flow instead o' username/password")
}
