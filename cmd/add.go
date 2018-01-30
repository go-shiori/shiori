package cmd

import (
	"github.com/RadhiFadlillah/go-readability"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	addCmd = &cobra.Command{
		Use:   "add url",
		Short: "Bookmark the specified URL.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Read flag and arguments
			url := args[0]
			tags, _ := cmd.Flags().GetStringSlice("tags")

			// Save new bookmark
			err := addBookmark(url, tags...)
			if err != nil {
				cError.Println(err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	addCmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark.")
	rootCmd.AddCommand(addCmd)
}

func addBookmark(url string, tags ...string) error {
	article, err := readability.Parse(url, 10*time.Second)
	if err != nil {
		return err
	}

	bookmark, err := DB.SaveBookmark(article, tags...)
	if err != nil {
		return err
	}

	printBookmark(bookmark)

	return nil
}
