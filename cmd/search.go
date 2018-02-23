package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	searchCmd = &cobra.Command{
		Use:   "search keyword",
		Short: "Search bookmarks by submitted keyword.",
		Long: "Search bookmarks by looking for matching keyword in bookmark's title and content. " +
			"If no keyword submitted, print all saved bookmarks. " +
			"Search results will be different depending on DBMS that used by shiori :\n" +
			"- sqlite3, search works using fts4 method: https://www.sqlite.org/fts3.html.\n" +
			"- mysql or mariadb, search works using natural language mode: https://dev.mysql.com/doc/refman/5.5/en/fulltext-natural-language.html.",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			tags, _ := cmd.Flags().GetStringSlice("tags")
			useJSON, _ := cmd.Flags().GetBool("json")
			indexOnly, _ := cmd.Flags().GetBool("index-only")

			// Fetch keyword
			keyword := ""
			if len(args) > 0 {
				keyword = args[0]
			}

			// Read bookmarks from database
			bookmarks, err := DB.SearchBookmarks(false, keyword, tags...)
			if err != nil {
				cError.Println(err)
				os.Exit(1)
			}

			if len(bookmarks) == 0 {
				cError.Println("No matching bookmarks found")
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
	searchCmd.Flags().BoolP("json", "j", false, "Output data in JSON format")
	searchCmd.Flags().BoolP("index-only", "i", false, "Only print the index of bookmarks")
	searchCmd.Flags().StringSliceP("tags", "t", []string{}, "Search bookmarks with specified tag(s)")
	rootCmd.AddCommand(searchCmd)
}
