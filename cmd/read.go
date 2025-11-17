/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"taedae/database"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get a database accessor
		accessor, err := database.NewAccessor()
		if err != nil {
			fmt.Println("Error creating database accessor:", err)
			return
		}

		// Read all todos
		todos, err := accessor.GetAllTodos()
		if err != nil {
			fmt.Println("Error reading todos:", err)
			return
		}

		// Print todos
		for _, todo := range todos {
			pterm.Info.Println(todo.ToString())
			pterm.Print("\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(readCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// readCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// readCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
