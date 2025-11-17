/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"taedae/database"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var title string
var description string
var dueDate string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		accessor, err := database.NewAccessor()
		if err != nil {
			pterm.Error.Println("Failed to create database accessor:", err)
			return
		}

		todo, err := accessor.CreateTodo(title, description, dueDate)
		if err != nil {
			pterm.Error.Println("Failed to create todo:", err)
			return
		}
		pterm.Success.Printf("Todo created successfully: %s\n", todo.ToString())
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Flags for the create command
	createCmd.Flags().StringVarP(&title, "title", "t", "", "Title of the todo item")
	createCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the todo item")
	createCmd.Flags().StringVarP(&dueDate, "due", "u", "", "Due date of the todo item (YYYY-MM-DD)")
}
