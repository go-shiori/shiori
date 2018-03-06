package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	openCmd = &cobra.Command{
		Use:   "open [indices]",
		Short: "Open the saved bookmarks",
		Long: "Open bookmarks in browser. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be opened.",
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			cacheOnly, _ := cmd.Flags().GetBool("cache")
			trimSpace, _ := cmd.Flags().GetBool("trim-space")
			skipConfirmation, _ := cmd.Flags().GetBool("yes")

			// If no arguments, confirm to user
			if len(args) == 0 && !skipConfirmation {
				confirmOpen := ""
				fmt.Print("Open ALL bookmarks? (y/n): ")
				fmt.Scanln(&confirmOpen)

				if confirmOpen != "y" {
					return
				}
			}

			if cacheOnly {
				openBookmarksCache(trimSpace, args...)
			} else {
				openBookmarks(args...)
			}
		},
	}
)

func init() {
	openCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and open ALL bookmarks")
	openCmd.Flags().BoolP("cache", "c", false, "Open the bookmark's cache in text-only mode")
	openCmd.Flags().Bool("trim-space", false, "Trim all spaces and newlines from the bookmark's cache")
	rootCmd.AddCommand(openCmd)
}

func openBookmarks(args ...string) {
	// Read bookmarks from database
	bookmarks, err := DB.GetBookmarks(false, args...)
	if err != nil {
		cError.Println(err)
		return
	}

	if len(bookmarks) == 0 {
		if len(args) > 0 {
			cError.Println("No matching index found")
		} else {
			cError.Println("No saved bookmarks yet")
		}
		return
	}

	// Open in browser
	for _, book := range bookmarks {
		err = openBrowser(book.URL)
		if err != nil {
			cError.Printf("Failed to open %s: %v\n", book.URL, err)
		}
	}
}

func openBookmarksCache(trimSpace bool, args ...string) {
	// Read bookmark content from database
	bookmarks, err := DB.GetBookmarks(true, args...)
	if err != nil {
		cError.Println(err)
		return
	}

	// Get terminal width
	termWidth := getTerminalWidth()
	if termWidth < 50 {
		termWidth = 50
	}

	// Show bookmarks content
	for _, book := range bookmarks {
		if trimSpace {
			words := strings.Fields(book.Content)
			book.Content = strings.Join(words, " ")
		}

		cIndex.Printf("%d. ", book.ID)
		cTitle.Println(book.Title)
		fmt.Println()

		if book.Content == "" {
			cError.Println("This bookmark doesn't have any cached content")
		} else {
			fmt.Println(book.Content)
		}

		fmt.Println()
		cSymbol.Println(strings.Repeat("-", termWidth))
		fmt.Println()
	}
}

// openBrowser tries to open the URL in a browser,
// and returns whether it succeed in doing so.
func openBrowser(url string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}

	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Run()
}
