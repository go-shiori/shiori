package cmd

import (
	"fmt"
	nurl "net/url"
	fp "path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-shiori/go-readability"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func updateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [indices]",
		Short: "Update the saved bookmarks",
		Long: "Update fields of an existing bookmark. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), " +
			"hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be updated. Update works differently depending on the flags:\n" +
			"- If indices are passed without any flags (--url, --title, --tag and --excerpt), read the URLs from DB and update titles from web.\n" +
			"- If --url is passed (and --title is omitted), update the title from web using the URL. While using this flag, update only accept EXACTLY one index.\n" +
			"While updating bookmark's tags, you can use - to remove tag (e.g. -nature to remove nature tag from this bookmark).",
		Run: updateHandler,
	}

	cmd.Flags().StringP("url", "u", "", "New URL for this bookmark.")
	cmd.Flags().StringP("title", "i", "", "New title for this bookmark.")
	cmd.Flags().StringP("excerpt", "e", "", "New excerpt for this bookmark.")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark.")
	cmd.Flags().BoolP("offline", "o", false, "Update bookmark without fetching data from internet.")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and update ALL bookmarks")
	cmd.Flags().Bool("dont-overwrite", false, "Don't overwrite existing metadata. Useful when only want to update bookmark's content.")

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
	dontOverwrite := cmd.Flags().Changed("dont-overwrite")

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
		return
	}

	// Clean up new parameter from flags
	title = normalizeSpace(title)
	excerpt = normalizeSpace(excerpt)

	if cmd.Flags().Changed("url") {
		// Clean up URL by removing its fragment and UTM parameters
		tmp, err := nurl.Parse(url)
		if err != nil || tmp.Scheme == "" || tmp.Hostname() == "" {
			cError.Println("URL is not valid")
			return
		}

		tmp.Fragment = ""
		clearUTMParams(tmp)
		url = tmp.String()

		// Since user uses custom URL, make sure there is only one ID to update
		if len(ids) != 1 {
			cError.Println("Update only accepts one index while using --url flag")
			return
		}
	}

	// Fetch bookmarks from database
	filterOptions := database.GetBookmarksOptions{
		IDs: ids,
	}

	bookmarks, err := DB.GetBookmarks(filterOptions)
	if err != nil {
		cError.Printf("Failed to get bookmarks: %v\n", err)
		return
	}

	if len(bookmarks) == 0 {
		cError.Println("No matching index found")
		return
	}

	// If it's not offline mode, fetch data from internet
	if !offline {
		mx := sync.RWMutex{}
		wg := sync.WaitGroup{}
		semaphore := make(chan struct{}, 10)

		for i, book := range bookmarks {
			wg.Add(1)

			// If used, use submitted URL
			if url != "" {
				book.URL = url
			}

			go func(i int, book model.Bookmark, nData int) {
				// Make sure to finish the WG
				defer wg.Done()

				// Register goroutine to semaphore
				semaphore <- struct{}{}
				defer func() {
					<-semaphore
				}()

				// Download article
				cInfo.Printf("[ %d / %d ] Downloading %s\n", i+1, nData, book.URL)

				article, err := readability.FromURL(book.URL, time.Minute)
				if err != nil {
					cError.Printf("[ %d / %d ] Failed to download article: %v\n", i+1, nData, err)
					return
				}

				book.Author = article.Byline
				book.Content = article.TextContent
				book.HTML = article.Content

				if !dontOverwrite {
					book.Title = article.Title
					book.Excerpt = article.Excerpt
				}

				// Get image URL and save it to local disk
				var imageURL string
				if article.Image != "" {
					imageURL = article.Image
				} else if article.Favicon != "" {
					imageURL = article.Favicon
				}

				if imageURL != "" {
					imgPath := fp.Join(DataDir, "thumb", fmt.Sprintf("%d", book.ID))

					err = downloadFile(imageURL, imgPath, time.Minute)
					if err != nil {
						cError.Printf("Failed to download image: %v\n", err)
						return
					}
				}

				// Save parse result to bookmark
				mx.Lock()
				bookmarks[i] = book
				mx.Unlock()
			}(i, book, len(bookmarks))
		}

		// Wait until all download finished
		wg.Wait()
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

		// Make sure title is not empty
		if book.Title == "" {
			book.Title = book.URL
		}

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
	bookmarks, err = DB.SaveBookmarks(bookmarks...)
	if err != nil {
		cError.Printf("Failed to save bookmark: %v\n", err)
		return
	}

	// Print updated bookmarks
	fmt.Println()
	printBookmarks(bookmarks...)
}
