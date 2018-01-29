package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete [indices]",
		Short: "Delete the saved bookmarks.",
		Long: "Delete bookmarks. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45) and hyphenated range (e.g. 100-200). " +
			"If no arguments, all records will be deleted",
		Run: func(cmd *cobra.Command, args []string) {
			// If no arguments, confirm to user
			if len(args) == 0 {
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
			if len(oldIndices) > 0 {
				for i, oldIndex := range oldIndices {
					newIndex := newIndices[i]
					fmt.Printf("Index %d moved to %d\n", oldIndex, newIndex)
				}
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}
