package cmd

import (
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "List yer hingmies",
	Long:  `Display all yer active todos in a braw wee table.`,
	Run: func(cmd *cobra.Command, args []string) {
		accessor, err := getAccessor()
		if err != nil {
			return
		}
		doRead(accessor)
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
