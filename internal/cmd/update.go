package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func updateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [indices]",
		Short: "Update the saved bookmarks",
		Long: "Update fields and archive of an existing bookmark. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), " +
			"hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be updated. Update works differently depending on the flags:\n" +
			"- If indices are passed without any flags (--url, --title, --tag and --excerpt), read the URLs from database and update titles from web.\n" +
			"- If --url is passed (and --title is omitted), update the title from web using the URL. While using this flag, update only accept EXACTLY one index.\n" +
			"While updating bookmark's tags, you can use - to remove tag (e.g. -nature to remove nature tag from this bookmark).",
		Run: updateHandler,
	}

	cmd.Flags().StringP("url", "u", "", "New URL for this bookmark")
	cmd.Flags().StringP("title", "i", "", "New title for this bookmark")
	cmd.Flags().StringP("excerpt", "e", "", "New excerpt for this bookmark")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark")
	cmd.Flags().BoolP("offline", "o", false, "Update bookmark without fetching data from internet")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and update ALL bookmarks")
	cmd.Flags().Bool("keep-metadata", false, "Keep existing metadata. Useful when only want to update bookmark's content")
	cmd.Flags().BoolP("no-archival", "a", false, "Update bookmark without updating offline archive")
	cmd.Flags().Bool("log-archival", false, "Log the archival process")

	return cmd
}

func updateHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	url, _ := cmd.Flags().GetString("url")
	title, _ := cmd.Flags().GetString("title")
	excerpt, _ := cmd.Flags().GetString("excerpt")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	offline, _ := cmd.Flags().GetBool("offline")
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	noArchival, _ := cmd.Flags().GetBool("no-archival")
	logArchival, _ := cmd.Flags().GetBool("log-archival")
	keepMetadata := cmd.Flags().Changed("keep-metadata")

	// If no arguments (i.e all bookmarks going to be updated), confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmUpdate := ""
		fmt.Print("Update ALL bookmarks? (y/N): ")
		fmt.Scanln(&confirmUpdate)

		if confirmUpdate != "y" {
			fmt.Println("No bookmarks updated")
			return
		}
	}

	// Convert args to ids
	ids, err := parseStrIndices(args)
	if err != nil {
		cError.Printf("Failed to parse args: %v\n", err)
		os.Exit(1)
	}

	// Clean up new parameter from flags
	title = validateTitle(title, "")
	excerpt = normalizeSpace(excerpt)

	if cmd.Flags().Changed("url") {
		// Clean up bookmark URL
		url, err = core.RemoveUTMParams(url)
		if err != nil {
			panic(fmt.Errorf("failed to clean URL: %v", err))
		}

		// Since user uses custom URL, make sure there is only one ID to update
		if len(ids) != 1 {
			cError.Println("Update only accepts one index while using --url flag")
			os.Exit(1)
		}
	}

	// Fetch bookmarks from database
	filterOptions := database.GetBookmarksOptions{
		IDs: ids,
	}

	bookmarks, err := db.GetBookmarks(cmd.Context(), filterOptions)
	if err != nil {
		cError.Printf("Failed to get bookmarks: %v\n", err)
		os.Exit(1)
	}

	if len(bookmarks) == 0 {
		cError.Println("No matching index found")
		os.Exit(1)
	}

	// Check if user really want to batch update archive
	if nBook := len(bookmarks); nBook > 5 && !offline && !noArchival && !skipConfirm {
		fmt.Printf("This update will generate offline archive for %d bookmark(s).\n", nBook)
		fmt.Println("This might take a long time and uses lot of your network bandwidth.")

		confirmUpdate := ""
		fmt.Printf("Continue update and archival process ? (y/N): ")
		fmt.Scanln(&confirmUpdate)

		if confirmUpdate != "y" {
			fmt.Println("No bookmarks updated")
			return

		}
	}

	// If it's not offline mode, fetch data from internet
	idWithProblems := []int{}

	if !offline {
		mx := sync.RWMutex{}
		wg := sync.WaitGroup{}
		chDone := make(chan struct{})
		chProblem := make(chan int, 10)
		chMessage := make(chan interface{}, 10)
		semaphore := make(chan struct{}, 10)

		cInfo.Println("Downloading article(s)...")

		for i, book := range bookmarks {
			wg.Add(1)

			// Mark whether book will be archived
			book.CreateArchive = !noArchival

			// If used, use submitted URL
			if url != "" {
				book.URL = url
			}

			go func(i int, book model.Bookmark) {
				// Make sure to finish the WG
				defer wg.Done()

				// Register goroutine to semaphore
				semaphore <- struct{}{}
				defer func() {
					<-semaphore
				}()

				// Download data from internet
				content, contentType, err := core.DownloadBookmark(book.URL)
				if err != nil {
					chProblem <- book.ID
					chMessage <- fmt.Errorf("failed to download %s: %v", book.URL, err)
					return
				}

				request := core.ProcessRequest{
					DataDir:     dataDir,
					Bookmark:    book,
					Content:     content,
					ContentType: contentType,
					KeepTitle:   keepMetadata,
					KeepExcerpt: keepMetadata,
					LogArchival: logArchival,
				}

				book, _, err = core.ProcessBookmark(request)
				content.Close()

				if err != nil {
					chProblem <- book.ID
					chMessage <- fmt.Errorf("failed to process %s: %v", book.URL, err)
					return
				}

				// Send success message
				chMessage <- fmt.Sprintf("Downloaded %s", book.URL)

				// Save parse result to bookmark
				mx.Lock()
				bookmarks[i] = book
				mx.Unlock()
			}(i, book)
		}

		// Print log message
		go func(nBookmark int) {
			logIndex := 0

			for {
				select {
				case <-chDone:
					cInfo.Println("Download finished")
					return
				case id := <-chProblem:
					idWithProblems = append(idWithProblems, id)
				case msg := <-chMessage:
					logIndex++

					switch msg.(type) {
					case error:
						cError.Printf("[%d/%d] %v\n", logIndex, nBookmark, msg)
					case string:
						cInfo.Printf("[%d/%d] %s\n", logIndex, nBookmark, msg)
					}
				}
			}
		}(len(bookmarks))

		// Wait until all download finished
		wg.Wait()
		close(chDone)
	}

	// Map which tags is new or deleted from flag --tags
	addedTags := make(map[string]struct{})
	deletedTags := make(map[string]struct{})
	for _, tag := range tags {
		tagName := strings.ToLower(tag)
		tagName = strings.TrimSpace(tagName)

		if strings.HasPrefix(tagName, "-") {
			tagName = strings.TrimPrefix(tagName, "-")
			deletedTags[tagName] = struct{}{}
		} else {
			addedTags[tagName] = struct{}{}
		}
	}

	// Attach user submitted value to the bookmarks
	for i, book := range bookmarks {
		// If user submit his own title or excerpt, use it
		if title != "" {
			book.Title = title
		}

		if excerpt != "" {
			book.Excerpt = excerpt
		}

		// If user submits url, use it
		if url != "" {
			book.URL = url
		}

		// Make sure title is valid and not empty
		book.Title = validateTitle(book.Title, book.URL)

		// Generate new tags
		tmpAddedTags := make(map[string]struct{})
		for key, value := range addedTags {
			tmpAddedTags[key] = value
		}

		newTags := []model.Tag{}
		for _, tag := range book.Tags {
			if _, isDeleted := deletedTags[tag.Name]; isDeleted {
				tag.Deleted = true
			}

			if _, alreadyExist := addedTags[tag.Name]; alreadyExist {
				delete(tmpAddedTags, tag.Name)
			}

			newTags = append(newTags, tag)
		}

		for tag := range tmpAddedTags {
			newTags = append(newTags, model.Tag{Name: tag})
		}

		book.Tags = newTags

		// Set bookmark's new data
		bookmarks[i] = book
	}

	// Save bookmarks to database
	bookmarks, err = db.SaveBookmarks(cmd.Context(), false, bookmarks...)
	if err != nil {
		cError.Printf("Failed to save bookmark: %v\n", err)
		os.Exit(1)
	}

	// Print updated bookmarks
	fmt.Println()
	printBookmarks(bookmarks...)

	var code int
	if len(idWithProblems) > 0 {
		code = 1
		sort.Ints(idWithProblems)

		cError.Println("Encountered error while downloading some bookmark(s):")
		for _, id := range idWithProblems {
			cError.Printf("%d ", id)
		}
		fmt.Println()
	}
	os.Exit(code)
}
