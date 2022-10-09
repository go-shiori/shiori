package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func pocketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pocket source-file",
		Short: "Import bookmarks from Pocket's exported HTML file",
		Args:  cobra.ExactArgs(1),
		Run:   pocketHandler,
	}

	return cmd
}

func pocketHandler(cmd *cobra.Command, args []string) {
	// Prepare bookmark's ID
	bookID, err := db.CreateNewID(cmd.Context(), "bookmark")
	if err != nil {
		cError.Printf("Failed to create ID: %v\n", err)
		return
	}

	// Open pocket's file
	srcFile, err := os.Open(args[0])
	if err != nil {
		cError.Println(err)
		os.Exit(1)
	}
	defer srcFile.Close()

	// Parse pocket's file
	bookmarks := []model.Bookmark{}
	mapURL := make(map[string]struct{})

	doc, err := goquery.NewDocumentFromReader(srcFile)
	if err != nil {
		cError.Println(err)
		os.Exit(1)
	}

	doc.Find("a").Each(func(_ int, a *goquery.Selection) {
		// Get metadata
		title := a.Text()
		url, _ := a.Attr("href")
		strTags, _ := a.Attr("tags")
		strModified, _ := a.Attr("time_added")
		intModified, _ := strconv.ParseInt(strModified, 10, 64)
		modified := time.Unix(intModified, 0)

		// Clean up URL
		var err error
		url, err = core.RemoveUTMParams(url)
		if err != nil {
			cError.Printf("Skip %s: URL is not valid\n", url)
			return
		}

		// Make sure title is valid Utf-8
		title = validateTitle(title, url)

		// Check if the URL already exist before, both in bookmark
		// file or in database
		if _, exist := mapURL[url]; exist {
			cError.Printf("Skip %s: URL already exists\n", url)
			return
		}

		_, exist, err := db.GetBookmark(cmd.Context(), 0, url)
		if err != nil {
			cError.Printf("Skip %s: Get Bookmark fail, %v", url, err)
			return
		}

		if exist {
			cError.Printf("Skip %s: URL already exists\n", url)
			mapURL[url] = struct{}{}
			return
		}

		// Get bookmark tags
		tags := []model.Tag{}
		for _, strTag := range strings.Split(strTags, ",") {
			if strTag != "" {
				tags = append(tags, model.Tag{Name: strTag})
			}
		}

		// Add item to list
		bookmark := model.Bookmark{
			ID:       bookID,
			URL:      url,
			Title:    title,
			Modified: modified.Format(model.DatabaseDateFormat),
			Tags:     tags,
		}

		bookID++
		mapURL[url] = struct{}{}
		bookmarks = append(bookmarks, bookmark)
	})

	// Save bookmark to database
	bookmarks, err = db.SaveBookmarks(cmd.Context(), bookmarks...)
	if err != nil {
		cError.Printf("Failed to save bookmarks: %v\n", err)
		os.Exit(1)
	}

	// Print imported bookmark
	fmt.Println()
	printBookmarks(bookmarks...)
}
