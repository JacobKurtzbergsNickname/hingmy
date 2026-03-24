package cmd

import "github.com/spf13/cobra"

// authCmd is the parent for login, logout, and status subcommands.
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage yer authentication",
	Long:  `Log in, log oot, or check the status o' yer session.`,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)
}
