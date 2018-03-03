package cmd

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/RadhiFadlillah/shiori/model"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	importCmd = &cobra.Command{
		Use:   "import source-file",
		Short: "Import bookmarks from HTML file in Netscape Bookmark format.",
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

	// Fetch each bookmark categories
	bookmarks := []model.Bookmark{}
	doc.Find("body>dl>dt").Each(func(_ int, el *goquery.Selection) {
		// Create category title
		category := el.Find("h3").First().Text()
		category = normalizeSpace(category)
		category = strings.ToLower(category)
		category = strings.Replace(category, " ", "-", -1)

		// Fetch all link in this categories
		el.Find("dl>dt").Each(func(_ int, dt *goquery.Selection) {
			// Get bookmark link
			a := dt.Find("a").First()
			title := a.Text()
			url, _ := a.Attr("href")
			strModified, _ := a.Attr("last_modified")
			strTags, _ := a.Attr("tags")
			intModified, _ := strconv.ParseInt(strModified, 10, 64)
			modified := time.Unix(intModified, 0)

			// Get bookmark excerpt
			excerpt := ""
			if nxt := dt.Next(); nxt.Is("dd") {
				excerpt = nxt.Text()
			}

			// Create bookmark item
			bookmark := model.Bookmark{
				URL:      url,
				Title:    normalizeSpace(title),
				Excerpt:  normalizeSpace(excerpt),
				Modified: modified.Format("2006-01-02 15:04:05"),
				Tags:     []model.Tag{},
			}

			if generateTag {
				bookmark.Tags = []model.Tag{
					{Name: category},
				}
			}


			//
			// Add any tags from the bookmark itself.
			//
			tags := strings.Split(strTags, ",")
			for _,entry := range(tags) {
				bookmark.Tags = append(bookmark.Tags, model.Tag{Name: entry} )
			}
			bookmarks = append(bookmarks, bookmark)
		})
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
