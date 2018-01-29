package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/RadhiFadlillah/shiori/model"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	printCmd = &cobra.Command{
		Use:   "print [indices]",
		Short: "Print the saved bookmarks.",
		Long: "Show details of bookmark record by its DB index. " +
			"If no arguments, all records with actual index from DB are shown. " +
			"Accepts hyphenated ranges and space-separated indices.",
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			useJSON, _ := cmd.Flags().GetBool("json")

			// Read bookmarks from database
			bookmarks, err := DB.GetBookmarks(args...)
			if err != nil {
				cError.Println(err)
				os.Exit(1)
			}

			if len(bookmarks) == 0 {
				if len(args) > 0 {
					cError.Println("No matching index found")
				} else {
					cError.Println("No saved bookmarks yet")
				}

				os.Exit(1)
			}

			// Print data
			if useJSON {
				bt, err := json.MarshalIndent(&bookmarks, "", "    ")
				if err != nil {
					cError.Println(err)
					os.Exit(1)
				}
				fmt.Println(string(bt))
			} else {
				printBookmark(bookmarks...)
			}
		},
	}
)

func init() {
	printCmd.Flags().BoolP("json", "j", false, "Output data in JSON format")
	rootCmd.AddCommand(printCmd)
}

func printBookmark(bookmarks ...model.Bookmark) {
	for _, bookmark := range bookmarks {
		// Create bookmark index
		strBookmarkIndex := fmt.Sprintf("%d. ", bookmark.ID)
		strSpace := strings.Repeat(" ", len(strBookmarkIndex))

		// Print bookmark title
		cIndex.Print(strBookmarkIndex)
		cTitle.Print(bookmark.Title)

		// Print read time
		if bookmark.MinReadTime > 0 {
			readTime := fmt.Sprintf(" (%d-%d minutes)", bookmark.MinReadTime, bookmark.MaxReadTime)
			if bookmark.MinReadTime == bookmark.MaxReadTime {
				readTime = fmt.Sprintf(" (%d minutes)", bookmark.MinReadTime)
			}
			cReadTime.Println(readTime)
		} else {
			fmt.Println()
		}

		// Print bookmark URL
		cSymbol.Print(strSpace + "> ")
		cURL.Println(bookmark.URL)

		// Print bookmark excerpt
		if bookmark.Excerpt != "" {
			cSymbol.Print(strSpace + "+ ")
			cExcerpt.Println(bookmark.Excerpt)
		}

		// Print bookmark tags
		if len(bookmark.Tags) > 0 {
			cSymbol.Print(strSpace + "# ")
			for i, tag := range bookmark.Tags {
				if i == len(bookmark.Tags)-1 {
					cTag.Println(tag.Name)
				} else {
					cTag.Print(tag.Name + ", ")
				}
			}
		}

		// Append new line
		fmt.Println()
	}
}
