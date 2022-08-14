package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/spf13/cobra"
)

func printCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print [indices]",
		Short: "Print the saved bookmarks",
		Long: "Show the saved bookmarks by its database index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), " +
			"hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, all records with actual index from database are shown.",
		Aliases: []string{"list", "ls"},
		Run:     printHandler,
	}

	cmd.Flags().BoolP("json", "j", false, "Output data in JSON format")
	cmd.Flags().BoolP("latest", "l", false, "Sort bookmark by latest instead of ID")
	cmd.Flags().BoolP("index-only", "i", false, "Only print the index of bookmarks")
	cmd.Flags().StringP("search", "s", "", "Search bookmark with specified keyword")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Print bookmarks with matching tag(s)")
	cmd.Flags().StringSliceP("exclude-tags", "e", []string{}, "Print bookmarks without these tag(s)")

	return cmd
}

func printHandler(cmd *cobra.Command, args []string) {
	// Read flags
	tags, _ := cmd.Flags().GetStringSlice("tags")
	keyword, _ := cmd.Flags().GetString("search")
	useJSON, _ := cmd.Flags().GetBool("json")
	indexOnly, _ := cmd.Flags().GetBool("index-only")
	orderLatest, _ := cmd.Flags().GetBool("latest")
	excludedTags, _ := cmd.Flags().GetStringSlice("exclude-tags")

	// Convert args to ids
	ids, err := parseStrIndices(args)
	if err != nil {
		cError.Printf("Failed to parse args: %v\n", err)
		return
	}

	// Read bookmarks from database
	orderMethod := database.DefaultOrder
	if orderLatest {
		orderMethod = database.ByLastModified
	}

	searchOptions := database.GetBookmarksOptions{
		IDs:          ids,
		Tags:         tags,
		ExcludedTags: excludedTags,
		Keyword:      keyword,
		OrderMethod:  orderMethod,
	}

	bookmarks, err := db.GetBookmarks(cmd.Context(), searchOptions)
	if err != nil {
		cError.Printf("Failed to get bookmarks: %v\n", err)
		return
	}

	if len(bookmarks) == 0 {
		switch {
		case len(ids) > 0:
			cError.Println("No matching index found")
		case keyword != "", len(tags) > 0:
			cError.Println("No matching bookmarks found")
		default:
			cError.Println("No bookmarks saved yet")
		}
		return
	}

	// Print data
	if useJSON {
		bt, err := json.MarshalIndent(&bookmarks, "", "    ")
		if err != nil {
			cError.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(bt))
		return
	}

	if indexOnly {
		for _, bookmark := range bookmarks {
			fmt.Printf("%d ", bookmark.ID)
		}

		fmt.Println()
		return
	}

	printBookmarks(bookmarks...)
}
