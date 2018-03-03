package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete [indices]",
		Short: "Delete the saved bookmarks",
		Long: "Delete bookmarks. " +
			"When a record is deleted, the last record is moved to the removed index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, all records will be deleted.",
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			skipConfirmation, _ := cmd.Flags().GetBool("yes")

			// If no arguments, confirm to user
			if len(args) == 0 && !skipConfirmation {
				confirmDelete := ""
				fmt.Print("Remove ALL bookmarks? (y/n): ")
				fmt.Scanln(&confirmDelete)

				if confirmDelete != "y" {
					fmt.Println("No bookmarks deleted")
					return
				}
			}

			// Delete bookmarks from database
			oldIndices, newIndices, err := DB.DeleteBookmarks(args...)
			if err != nil {
				cError.Println(err)
				os.Exit(1)
			}

			fmt.Println("Bookmarks has been deleted")
			for i, oldIndex := range oldIndices {
				newIndex := newIndices[i]
				fmt.Printf("Index %d moved to %d\n", oldIndex, newIndex)
			}
		},
	}
)

func init() {
	deleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and delete ALL bookmarks")
	rootCmd.AddCommand(deleteCmd)
}
