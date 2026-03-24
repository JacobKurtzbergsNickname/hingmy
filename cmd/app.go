package cmd

import (
	"os"

	"hingmy/internal/auth"
	"hingmy/internal/ui"

	"github.com/spf13/cobra"
)

// resolveUI selects PtermUI or PlainUI based on --no-color or the CI
// environment variable. Must be called inside a RunE/PersistentPreRunE
// handler — not in init() — so that flags are fully parsed before we select
// the UI implementation.
func resolveUI(cmd *cobra.Command) ui.UI {
	noColor, _ := cmd.Root().PersistentFlags().GetBool("no-color")
	if noColor || os.Getenv("CI") == "true" || os.Getenv("NO_COLOR") != "" {
		return ui.NewPlainUI()
	}
	return ui.NewPtermUI()
}

// resolveTokenStore returns a KeyringStore if the keychain is available,
// otherwise falls back to FileStore. The --no-keyring flag or CI=true always
// selects FileStore.
func resolveTokenStore(cmd *cobra.Command) (auth.TokenStore, error) {
	noKeyring, _ := cmd.Root().PersistentFlags().GetBool("no-keyring")
	if !noKeyring && os.Getenv("CI") != "true" {
		ks := auth.NewKeyringStore()
		// Probe: if the keychain tool isn't available, fall back gracefully.
		if _, err := ks.Load(); err != auth.ErrKeychainUnavailable {
			return ks, nil
		}
	}
	return auth.NewFileStore("")
}
