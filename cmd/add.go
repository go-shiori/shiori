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
			title, _ := cmd.Flags().GetString("title")
			excerpt, _ := cmd.Flags().GetString("excerpt")
			tags, _ := cmd.Flags().GetStringSlice("tags")
			offline, _ := cmd.Flags().GetBool("offline")

			// Save new bookmark
			err := addBookmark(url, title, excerpt, tags, offline)
			if err != nil {
				cError.Println(err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	addCmd.Flags().StringP("title", "i", "", "Custom title for this bookmark.")
	addCmd.Flags().StringP("excerpt", "e", "", "Custom excerpt for this bookmark.")
	addCmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark.")
	addCmd.Flags().BoolP("offline", "o", false, "Save bookmark without fetching data from internet.")
	rootCmd.AddCommand(addCmd)
}

func addBookmark(url, title, excerpt string, tags []string, offline bool) (err error) {
	// Prepare variable
	defaultArticle := readability.Article{
		URL: url,
		Meta: readability.Metadata{
			Title: "Untitled",
		},
	}
	article := defaultArticle

	// Fetch data from internet
	if !offline {
		article, err = readability.Parse(url, 10*time.Second)
		if err != nil {
			cError.Println("Failed to fetch article from internet")
			article = defaultArticle
		}
	}

	// Set custom value
	if title != "" {
		article.Meta.Title = title
	}

	if excerpt != "" {
		article.Meta.Excerpt = excerpt
	}

	bookmark, err := DB.SaveBookmark(article, tags...)
	if err != nil {
		return err
	}

	printBookmark(bookmark)

	return nil
}
