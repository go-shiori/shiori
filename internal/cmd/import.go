package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func importCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import source-file",
		Short: "Import bookmarks from HTML file in Netscape Bookmark format",
		Args:  cobra.ExactArgs(1),
		Run:   importHandler,
	}

	cmd.Flags().BoolP("generate-tag", "t", false, "Auto generate tag from bookmark's category")

	return cmd
}

func importHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	generateTag := cmd.Flags().Changed("generate-tag")

	// If user doesn't specify, ask if tag need to be generated
	if !generateTag {
		var submit string
		fmt.Print("Add parents folder as tag? (y/N): ")
		fmt.Scanln(&submit)

		generateTag = submit == "y"
	}

	// Prepare bookmark's ID
	bookID, err := db.CreateNewID("bookmark")
	if err != nil {
		cError.Printf("Failed to create ID: %v\n", err)
		os.Exit(1)
	}

	// Open bookmark's file
	srcFile, err := os.Open(args[0])
	if err != nil {
		cError.Printf("Failed to open %s: %v\n", args[0], err)
		os.Exit(1)
	}
	defer srcFile.Close()

	// Parse bookmark's file
	bookmarks := []model.Bookmark{}
	mapURL := make(map[string]struct{})

	doc, err := goquery.NewDocumentFromReader(srcFile)
	if err != nil {
		cError.Printf("Failed to parse bookmark: %v\n", err)
		os.Exit(1)
	}

	doc.Find("dt>a").Each(func(_ int, a *goquery.Selection) {
		// Get related elements
		dt := a.Parent()
		dl := dt.Parent()
		h3 := dl.Parent().Find("h3").First()

		// Get metadata
		title := a.Text()
		url, _ := a.Attr("href")
		strTags, _ := a.Attr("tags")

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

		if _, exist := db.GetBookmark(0, url); exist {
			cError.Printf("Skip %s: URL already exists\n", url)
			mapURL[url] = struct{}{}
			return
		}

		// Get bookmark tags
		tags := []model.Tag{}
		for _, strTag := range strings.Split(strTags, ",") {
			strTag = normalizeSpace(strTag)
			if strTag != "" {
				tags = append(tags, model.Tag{Name: strTag})
			}
		}

		// Get category name for this bookmark
		// and add it as tags (if necessary)
		category := normalizeSpace(h3.Text())
		if category != "" && generateTag {
			tags = append(tags, model.Tag{Name: category})
		}

		// Add item to list
		bookmark := model.Bookmark{
			ID:    bookID,
			URL:   url,
			Title: title,
			Tags:  tags,
		}

		bookID++
		mapURL[url] = struct{}{}
		bookmarks = append(bookmarks, bookmark)
	})

	// Save bookmark to database
	bookmarks, err = db.SaveBookmarks(bookmarks...)
	if err != nil {
		cError.Printf("Failed to save bookmarks: %v\n", err)
		os.Exit(1)
	}

	// Print imported bookmark
	fmt.Println()
	printBookmarks(bookmarks...)
}
