package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add url",
		Short: "Bookmark the specified URL",
		Args:  cobra.ExactArgs(1),
		Run:   addHandler,
	}

	cmd.Flags().StringP("title", "i", "", "Custom title for this bookmark")
	cmd.Flags().StringP("excerpt", "e", "", "Custom excerpt for this bookmark")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark")
	cmd.Flags().BoolP("offline", "o", false, "Save bookmark without fetching data from internet")
	cmd.Flags().BoolP("no-archival", "a", false, "Save bookmark without creating offline archive")
	cmd.Flags().Bool("log-archival", false, "Log the archival process")

	return cmd
}

func addHandler(cmd *cobra.Command, args []string) {
	// Read flag and arguments
	url := args[0]
	title, _ := cmd.Flags().GetString("title")
	excerpt, _ := cmd.Flags().GetString("excerpt")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	offline, _ := cmd.Flags().GetBool("offline")
	noArchival, _ := cmd.Flags().GetBool("no-archival")
	logArchival, _ := cmd.Flags().GetBool("log-archival")

	// Normalize input
	title = validateTitle(title, "")
	excerpt = normalizeSpace(excerpt)

	// Create bookmark item
	book := model.Bookmark{
		URL:           url,
		Title:         title,
		Excerpt:       excerpt,
		CreateArchive: !noArchival,
	}

	// Set bookmark tags
	book.Tags = make([]model.Tag, len(tags))
	for i, tag := range tags {
		book.Tags[i].Name = strings.TrimSpace(tag)
	}

	// Clean up bookmark URL
	var err error
	book.URL, err = core.RemoveUTMParams(book.URL)
	if err != nil {
		cError.Printf("Failed to clean URL: %v\n", err)
		os.Exit(1)
	}

	// Make sure bookmark's title not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Save bookmark to database
	books, err := db.SaveBookmarks(cmd.Context(), true, book)
	if err != nil {
		cError.Printf("Failed to save bookmark: %v\n", err)
		os.Exit(1)
	}

	book = books[0]

	// If it's not offline mode, fetch data from internet.
	if !offline {
		cInfo.Println("Downloading article...")

		var isFatalErr bool
		content, contentType, err := core.DownloadBookmark(book.URL)
		if err != nil {
			cError.Printf("Failed to download: %v\n", err)
		}

		if err == nil && content != nil {
			request := core.ProcessRequest{
				DataDir:     dataDir,
				Bookmark:    book,
				Content:     content,
				ContentType: contentType,
				LogArchival: logArchival,
				KeepTitle:   title != "",
				KeepExcerpt: excerpt != "",
			}

			book, isFatalErr, err = core.ProcessBookmark(request)
			content.Close()

			if err != nil {
				cError.Printf("Failed: %v\n", err)
			}

			if isFatalErr {
				os.Exit(1)
			}
		}

		// Save bookmark to database
		_, err = db.SaveBookmarks(cmd.Context(), false, book)
		if err != nil {
			cError.Printf("Failed to save bookmark with content: %v\n", err)
			os.Exit(1)
		}
	}

	// Print added bookmark
	fmt.Println()
	printBookmarks(book)
}
