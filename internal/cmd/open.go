package cmd

import (
	"fmt"
	"strings"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/spf13/cobra"
)

func openCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [indices]",
		Short: "Open the saved bookmarks",
		Long: "Open bookmarks in browser. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), " +
			"hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be opened.",
		Run: openHandler,
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and open ALL bookmarks")
	cmd.Flags().BoolP("text-cache", "t", false, "Open the bookmark's text cache in terminal")

	return cmd
}

func openHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	textCacheMode, _ := cmd.Flags().GetBool("text-cache")

	// If no arguments (i.e all bookmarks will be opened),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmOpen := ""
		fmt.Print("Open ALL bookmarks? (y/N): ")
		fmt.Scanln(&confirmOpen)

		if confirmOpen != "y" {
			return
		}
	}

	// Convert args to ids
	ids, err := parseStrIndices(args)
	if err != nil {
		cError.Println(err)
		return
	}

	// Read bookmarks from database
	getOptions := database.GetBookmarksOptions{
		IDs:         ids,
		WithContent: true,
	}

	bookmarks, err := DB.GetBookmarks(getOptions)
	if err != nil {
		cError.Printf("Failed to get bookmarks: %v\n", err)
		return
	}

	if len(bookmarks) == 0 {
		switch {
		case len(ids) > 0:
			cError.Println("No matching index found")
		default:
			cError.Println("No bookmarks saved yet")
		}
		return
	}

	// If not text cache mode, open bookmarks in browser
	if !textCacheMode {
		for _, book := range bookmarks {
			err = openBrowser(book.URL)
			if err != nil {
				cError.Printf("Failed to open %s: %v\n", book.URL, err)
			}
		}
		return
	}

	// Show bookmarks content in terminal
	termWidth := getTerminalWidth()

	for _, book := range bookmarks {
		cIndex.Printf("%d. ", book.ID)
		cTitle.Println(book.Title)
		fmt.Println()

		if book.Content == "" {
			cError.Println("This bookmark doesn't have any cached content")
		} else {
			book.Content = strings.Join(strings.Fields(book.Content), " ")
			fmt.Println(book.Content)
		}

		fmt.Println()
		cSymbol.Println(strings.Repeat("=", termWidth))
		fmt.Println()
	}
}
