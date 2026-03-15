package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a hingmy interactively",
	Long:  `Update an existing hingmy. Run 'hingmy' for the full interactive experience.`,
	Run: func(cmd *cobra.Command, args []string) {
		accessor, err := getAccessor()
		if err != nil {
			return
		}
		doUpdate(accessor)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
