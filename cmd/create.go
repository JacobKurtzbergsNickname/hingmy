package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var title string
var description string
var dueDate string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new hingmy",
	Long:  `Create a new todo item. Use flags tae set the title, description, and due date.`,
	Run: func(cmd *cobra.Command, args []string) {
		accessor, err := getAccessor()
		if err != nil {
			return
		}

		todo, err := accessor.CreateTodo(title, description, dueDate)
		if err != nil {
			pterm.Error.Println("Couldnae create yer hingmy:", err)
			return
		}
		pterm.Success.Printf("Weel done! Hingmy #%d '%s' added tae yer list!\n", todo.ID, todo.Title)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&title, "title", "t", "", "Title of the todo item")
	createCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the todo item")
	createCmd.Flags().StringVarP(&dueDate, "due", "u", "", "Due date of the todo item (YYYY-MM-DD)")
}
