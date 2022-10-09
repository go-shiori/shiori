package cmd

import (
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func exportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export target-file",
		Short: "Export bookmarks into HTML file in Netscape Bookmark format",
		Args:  cobra.ExactArgs(1),
		Run:   exportHandler,
	}

	return cmd
}

func exportHandler(cmd *cobra.Command, args []string) {
	// Fetch bookmarks from database
	bookmarks, err := db.GetBookmarks(cmd.Context(), database.GetBookmarksOptions{})
	if err != nil {
		cError.Printf("Failed to get bookmarks: %v\n", err)
		os.Exit(1)
	}

	if len(bookmarks) == 0 {
		cError.Println("No saved bookmarks yet")
		return
	}

	// Make sure destination directory exist
	dstDir := fp.Dir(args[0])
	if err := os.MkdirAll(dstDir, model.DataDirPerm); err != nil {
		cError.Printf("Error crating destination directory: %s", err)
	}

	// Create destination file
	dstFile, err := os.Create(args[0])
	if err != nil {
		cError.Printf("Failed to create destination file: %v\n", err)
		os.Exit(1)
	}
	defer dstFile.Close()

	// Write exported bookmark to file
	fmt.Fprintln(dstFile, ``+
		`<!DOCTYPE NETSCAPE-Bookmark-file-1>`+
		`<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">`+
		`<TITLE>Bookmarks</TITLE>`+
		`<H1>Bookmarks</H1>`+
		`<DL>`)

	for _, book := range bookmarks {
		// Create Unix timestamp for bookmark
		modifiedTime, err := time.Parse(model.DatabaseDateFormat, book.Modified)
		if err != nil {
			modifiedTime = time.Now()
		}
		unixTimestamp := modifiedTime.Unix()

		// Create tags for bookmarks
		tags := []string{}
		for _, tag := range book.Tags {
			tags = append(tags, tag.Name)
		}
		strTags := strings.Join(tags, ",")

		// Make sure title is valid
		book.Title = validateTitle(book.Title, book.URL)

		// Write to file
		exportLine := fmt.Sprintf(`<DT><A HREF="%s" ADD_DATE="%d" LAST_MODIFIED="%d" TAGS="%s">%s</A>`,
			book.URL, unixTimestamp, unixTimestamp, strTags, book.Title)
		fmt.Fprintln(dstFile, exportLine)
	}

	fmt.Fprintln(dstFile, "</DL>")

	// Flush data to storage
	err = dstFile.Sync()
	if err != nil {
		cError.Printf("Failed to export the bookmarks: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Export finished")
}
