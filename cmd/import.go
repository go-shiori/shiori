package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"../model"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

var (
	importCmd = &cobra.Command{
		Use:   "import source-file",
		Short: "Import bookmarks from HTML file in Netscape Bookmark format",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generateTag := cmd.Flags().Changed("generate-tag")
			if !generateTag {
				var submitGenerateTag string
				fmt.Print("Add parents folder as tag? (y/n): ")
				fmt.Scanln(&submitGenerateTag)

				generateTag = submitGenerateTag == "y"
			}

			err := importBookmarks(args[0], generateTag)
			if err != nil {
				cError.Println(err)
				return
			}
		},
	}
)

func init() {
	importCmd.Flags().BoolP("generate-tag", "t", false, "Auto generate tag from bookmark's category")
	rootCmd.AddCommand(importCmd)
}

func importBookmarks(pth string, generateTag bool) error {
	// Open file
	srcFile, err := os.Open(pth)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Parse file
	doc, err := goquery.NewDocumentFromReader(srcFile)
	if err != nil {
		return err
	}

	// Loop each bookmark item
	bookmarks := []model.Bookmark{}
	doc.Find("dt>a").Each(func(_ int, a *goquery.Selection) {
		// Get related elements
		dt := a.Parent()
		dl := dt.Parent()

		// Get metadata
		title := a.Text()
		url, _ := a.Attr("href")
		strTags, _ := a.Attr("tags")
		strModified, _ := a.Attr("last_modified")
		intModified, _ := strconv.ParseInt(strModified, 10, 64)
		modified := time.Unix(intModified, 0)

		// Get bookmark tags
		tags := []model.Tag{}
		for _, strTag := range strings.Split(strTags, ",") {
			if strTag != "" {
				tags = append(tags, model.Tag{Name: strTag})
			}
		}

		// Get bookmark excerpt
		excerpt := ""
		if dd := dt.Next(); dd.Is("dd") {
			excerpt = dd.Text()
		}

		// Get category name for this bookmark
		// and add it as tags (if necessary)
		category := ""
		if dtCategory := dl.Prev(); dtCategory.Is("h3") {
			category = dtCategory.Text()
			category = normalizeSpace(category)
			category = strings.ToLower(category)
			category = strings.Replace(category, " ", "-", -1)
		}

		if category != "" && generateTag {
			tags = append(tags, model.Tag{Name: category})
		}

		// Add item to list
		bookmark := model.Bookmark{
			URL:      url,
			Title:    normalizeSpace(title),
			Excerpt:  normalizeSpace(excerpt),
			Modified: modified.Format("2006-01-02 15:04:05"),
			Tags:     tags,
		}

		bookmarks = append(bookmarks, bookmark)
	})

	// Save bookmarks to database
	for _, book := range bookmarks {
		result, err := addBookmark(book, true)
		if err != nil {
			cError.Printf("URL %s already exists\n\n", book.URL)
			continue
		}

		printBookmark(result)
	}

	return nil
}
