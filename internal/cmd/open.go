package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	fp "path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/warc"
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
	cmd.Flags().IntP("archive-port", "p", 0, "Port number that used to serve archive")
	cmd.Flags().BoolP("text-cache", "t", false, "Open the bookmark's text cache in terminal")

	return cmd
}

func openHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	archiveMode, _ := cmd.Flags().GetBool("archive")
	archivePort, _ := cmd.Flags().GetInt("archive-port")
	textCacheMode, _ := cmd.Flags().GetBool("text-cache")

	// Convert args to ids
	ids, err := parseStrIndices(args)
	if err != nil {
		cError.Println(err)
		os.Exit(1)
	}

	// If in archive mode, only one bookmark allowed
	if len(ids) > 1 && archiveMode {
		cError.Println("In archive mode, only one bookmark allowed")
		os.Exit(1)
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

	bookmarks, err := db.GetBookmarks(cmd.Context(), getOptions)
	if err != nil {
		cError.Printf("Failed to get bookmarks: %v\n", err)
		os.Exit(1)
	}

	if len(bookmarks) == 0 {
		if len(ids) > 0 {
			cError.Println("No matching index found")
			os.Exit(1)
		} else {
			cError.Println("No bookmarks saved yet")
			os.Exit(1)
		}
		return
	}

	// If not text cache mode nor archive mode, open bookmarks in browser
	if !textCacheMode && !archiveMode {
		var code int
		for _, book := range bookmarks {
			err = openBrowser(book.URL)
			if err != nil {
				cError.Printf("Failed to open %s: %v\n", book.URL, err)
				code = 1
			}
		}
		os.Exit(code)
	}

	// Show bookmarks content in terminal
	if textCacheMode {
		termWidth := getTerminalWidth()

		var code int
		for _, book := range bookmarks {
			cIndex.Printf("%d. ", book.ID)
			cTitle.Println(book.Title)
			fmt.Println()

			if book.Content == "" {
				cError.Println("This bookmark doesn't have any cached content")
				code = 1
			} else {
				book.Content = strings.Join(strings.Fields(book.Content), " ")
				fmt.Println(book.Content)
			}

			fmt.Println()
			cSymbol.Println(strings.Repeat("=", termWidth))
			fmt.Println()
		}
		os.Exit(code)
	}

	// Open archive
	id := strconv.Itoa(bookmarks[0].ID)
	archivePath := fp.Join(dataDir, "archive", id)

	archive, err := warc.Open(archivePath)
	if err != nil {
		cError.Printf("Failed to open archive: %v\n", err)
		os.Exit(1)
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
		w.Header().Set("Content-Encoding", "gzip")
		if _, err = w.Write(content); err != nil {
			panic(err)
		}
	})

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
		http.Error(w, fmt.Sprint(arg), 500)
	}

	// Choose random port
	listenerAddr := fmt.Sprintf(":%d", archivePort)
	listener, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		cError.Printf("Failed to serve archive: %v\n", err)
		os.Exit(1)
	}

	portNumber := listener.Addr().(*net.TCPAddr).Port
	localhostAddr := fmt.Sprintf("http://localhost:%d", portNumber)
	cInfo.Printf("Archive served in %s\n", localhostAddr)

	// Open browser
	go func() {
		time.Sleep(time.Second)

		err := openBrowser(localhostAddr)
		if err != nil {
			cError.Printf("Failed to open browser: %v\n", err)
			os.Exit(1)
		}
	}()

	// Serve archive
	err = http.Serve(listener, router)
	if err != nil {
		cError.Printf("Failed to serve archive: %v\n", err)
		os.Exit(1)
	}
}
