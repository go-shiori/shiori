package cmd

import (
	"github.com/RadhiFadlillah/go-readability"
	"github.com/RadhiFadlillah/shiori/model"
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
	// Fetch data from internet
	article := readability.Article{}
	if !offline {
		article, err = readability.Parse(url, 10*time.Second)
		if err != nil {
			cError.Println("Failed to fetch article from internet:", err)
			article.URL = url
			article.Meta.Title = "Untitled"
		}
	}

	// Prepare bookmark
	bookmark := model.Bookmark{
		URL:         article.URL,
		Title:       article.Meta.Title,
		ImageURL:    article.Meta.Image,
		Excerpt:     article.Meta.Excerpt,
		Author:      article.Meta.Author,
		Language:    article.Meta.Language,
		MinReadTime: article.Meta.MinReadTime,
		MaxReadTime: article.Meta.MaxReadTime,
		Content:     article.Content,
	}

	bookTags := make([]model.Tag, len(tags))
	for i, tag := range tags {
		bookTags[i].Name = tag
	}

	bookmark.Tags = bookTags

	// Set custom value
	if title != "" {
		bookmark.Title = title
	}

	if excerpt != "" {
		bookmark.Excerpt = excerpt
	}

	// Save to database
	bookmark.ID, err = DB.SaveBookmark(bookmark)
	if err != nil {
		return err
	}

	printBookmark(bookmark)

	return nil
}
