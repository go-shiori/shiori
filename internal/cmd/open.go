package cmd

import (
	"fmt"
	"net"
	"net/http"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/pkg/warc"
	"github.com/julienschmidt/httprouter"
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
	cmd.Flags().BoolP("archive", "a", false, "Open the bookmark's archived content")
	cmd.Flags().BoolP("text-cache", "t", false, "Open the bookmark's text cache in terminal")

	return cmd
}

func openHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	archiveMode, _ := cmd.Flags().GetBool("archive")
	textCacheMode, _ := cmd.Flags().GetBool("text-cache")

	// Convert args to ids
	ids, err := parseStrIndices(args)
	if err != nil {
		cError.Println(err)
		return
	}

	// If in archive mode, only one bookmark allowed
	if len(ids) > 1 && archiveMode {
		cError.Println("In archive mode, only one bookmark allowed")
		return
	}

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
		if len(ids) > 0 {
			cError.Println("No matching index found")
		} else {
			cError.Println("No bookmarks saved yet")
		}
		return
	}

	// If not text cache mode nor archive mode, open bookmarks in browser
	if !textCacheMode && !archiveMode {
		for _, book := range bookmarks {
			err = openBrowser(book.URL)
			if err != nil {
				cError.Printf("Failed to open %s: %v\n", book.URL, err)
			}
		}
		return
	}

	// Show bookmarks content in terminal
	if textCacheMode {
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

	// Open archive
	id := strconv.Itoa(bookmarks[0].ID)
	archivePath := fp.Join(DataDir, "archive", id)

	archive, err := warc.Open(archivePath)
	if err != nil {
		cError.Printf("Failed to open archive: %v\n", err)
		return
	}
	defer archive.Close()

	// Create simple server
	router := httprouter.New()
	router.GET("/*filename", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		filename := ps.ByName("filename")
		resourceName := fp.Base(filename)
		if resourceName == "/" {
			resourceName = ""
		}

		content, contentType, err := archive.Read(resourceName)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", contentType)
		if _, err = w.Write(content); err != nil {
			panic(err)
		}
	})

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
		http.Error(w, fmt.Sprint(arg), 500)
	}

	// Choose random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		cError.Printf("Failed to serve archive: %v\n", err)
		return
	}

	portNumber := listener.Addr().(*net.TCPAddr).Port
	cInfo.Printf("Archive served in http://localhost:%d\n", portNumber)

	err = http.Serve(listener, router)
	if err != nil {
		cError.Printf("Failed to serve archive: %v\n", err)
	}
}
