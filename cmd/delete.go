package cmd

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a hingmy interactively",
	Long:  `Delete an existing hingmy. Run 'hingmy' for the full interactive experience.`,
	Run: func(cmd *cobra.Command, args []string) {
		accessor, err := getAccessor()
		if err != nil {
			return
		}
		doDelete(accessor)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
