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
		Long: "Show the saved bookmarks by its DB index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, all records with actual index from DB are shown.",
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			useJSON, _ := cmd.Flags().GetBool("json")
			indexOnly, _ := cmd.Flags().GetBool("index-only")

			// Read bookmarks from database
			bookmarks, err := DB.GetBookmarks(false, args...)
			if err != nil {
				cError.Println(err)
				os.Exit(1)
			}

			if len(bookmarks) == 0 {
				if len(args) > 0 {
					cError.Println("No matching index found")
				} else {
					cError.Println("No bookmarks saved yet")
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
			} else if indexOnly {
				printBookmarkIndex(bookmarks...)
			} else {
				printBookmark(bookmarks...)
			}
		},
	}
)

func init() {
	printCmd.Flags().BoolP("json", "j", false, "Output data in JSON format")
	printCmd.Flags().BoolP("index-only", "i", false, "Only print the index of bookmarks")
	rootCmd.AddCommand(printCmd)
}

func printBookmarkIndex(bookmarks ...model.Bookmark) {
	for _, bookmark := range bookmarks {
		fmt.Printf("%d ", bookmark.ID)
	}
	fmt.Println()
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
