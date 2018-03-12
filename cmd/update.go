package cmd

import (
	"fmt"
	nurl "net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RadhiFadlillah/go-readability"
	"github.com/RadhiFadlillah/shiori/model"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update [indices]",
		Short: "Update the saved bookmarks",
		Long: "Update fields of an existing bookmark. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be updated. Update works differently depending on the flags:\n" +
			"- If indices are passed without any flags (--url, --title, --tag and --excerpt), read the URLs from DB and update titles from web.\n" +
			"- If --url is passed (and --title is omitted), update the title from web using the URL. While using this flag, update only accept EXACTLY one index.\n" +
			"While updating bookmark's tags, you can use - to remove tag (e.g. -nature to remove nature tag from this bookmark).",
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			url, _ := cmd.Flags().GetString("url")
			title, _ := cmd.Flags().GetString("title")
			excerpt, _ := cmd.Flags().GetString("excerpt")
			tags, _ := cmd.Flags().GetStringSlice("tags")
			offline, _ := cmd.Flags().GetBool("offline")
			skipConfirmation, _ := cmd.Flags().GetBool("yes")
			overwriteMetadata := !cmd.Flags().Changed("dont-overwrite")

			// Check if --url flag is used
			if cmd.Flags().Changed("url") {
				if len(args) != 1 {
					cError.Println("Update only accepts one index while using --url flag")
					return
				}

				idx, err := strconv.Atoi(args[0])
				if err != nil || idx < -1 {
					cError.Println("Index is not valid")
					return
				}
			}

			// Check if --excerpt flag is used
			if !cmd.Flags().Changed("excerpt") {
				excerpt = "empty"
			}

			// If no arguments, confirm to user
			if len(args) == 0 && !skipConfirmation {
				confirmUpdate := ""
				fmt.Print("Update ALL bookmarks? (y/n): ")
				fmt.Scanln(&confirmUpdate)

				if confirmUpdate != "y" {
					fmt.Println("No bookmarks updated")
					return
				}
			}

			// Update bookmarks
			base := model.Bookmark{
				URL:     url,
				Title:   title,
				Excerpt: excerpt,
			}

			base.Tags = make([]model.Tag, len(tags))
			for i, tag := range tags {
				base.Tags[i] = model.Tag{Name: tag}
			}

			bookmarks, err := updateBookmarks(args, base, offline, overwriteMetadata)
			if err != nil {
				cError.Println(err)
				return
			}

			printBookmark(bookmarks...)
		},
	}
)

func init() {
	updateCmd.Flags().StringP("url", "u", "", "New URL for this bookmark.")
	updateCmd.Flags().StringP("title", "i", "", "New title for this bookmark.")
	updateCmd.Flags().StringP("excerpt", "e", "", "New excerpt for this bookmark.")
	updateCmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark.")
	updateCmd.Flags().BoolP("offline", "o", false, "Update bookmark without fetching data from internet.")
	updateCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and update ALL bookmarks")
	updateCmd.Flags().Bool("dont-overwrite", false, "Don't overwrite existing metadata. Useful when only want to update bookmark's content.")
	rootCmd.AddCommand(updateCmd)
}

func updateBookmarks(indices []string, base model.Bookmark, offline, overwrite bool) ([]model.Bookmark, error) {
	// Prepare wait group
	waitGroup := sync.WaitGroup{}

	// Check if URL is not empty
	if base.URL != "" {
		// Make sure URL valid
		parsedURL, err := nurl.ParseRequestURI(base.URL)
		if err != nil || parsedURL.Host == "" {
			return []model.Bookmark{}, fmt.Errorf("URL is not valid")
		}

		// Clear UTM parameters from URL
		base.URL, err = clearUTMParams(parsedURL)
		if err != nil {
			return []model.Bookmark{}, err
		}
	}

	// Read bookmarks from database
	bookmarks, err := DB.GetBookmarks(true, indices...)
	if err != nil {
		return []model.Bookmark{}, err
	}

	if len(bookmarks) == 0 {
		return []model.Bookmark{}, fmt.Errorf("No matching index found")
	}

	if base.URL != "" && len(bookmarks) == 1 {
		bookmarks[0].URL = base.URL
	}

	// If not offline, fetch articles from internet
	if !offline {
		fmt.Println("Fetching new bookmarks data")
		uiprogress.Start()
		bar := uiprogress.AddBar(len(bookmarks)).AppendCompleted().PrependElapsed()

		for i, book := range bookmarks {
			waitGroup.Add(1)

			go func(pos int, book model.Bookmark) {
				defer func() {
					bar.Incr()
					waitGroup.Done()
				}()

				article, err := readability.Parse(book.URL, 10*time.Second)
				if err != nil {
					return
				}

				if overwrite {
					book.Title = article.Meta.Title
					book.Excerpt = article.Meta.Excerpt
				}

				book.ImageURL = article.Meta.Image
				book.Author = article.Meta.Author
				book.MinReadTime = article.Meta.MinReadTime
				book.MaxReadTime = article.Meta.MaxReadTime
				book.Content = article.Content
				book.HTML = article.RawContent

				bookmarks[pos] = book
			}(i, book)
		}

		time.Sleep(1 * time.Second)
		waitGroup.Wait()

		uiprogress.Stop()
		fmt.Println("\nSaving new data")
	}

	// Map the tags to be deleted
	addedTags := make(map[string]struct{})
	deletedTags := make(map[string]struct{})
	for _, tag := range base.Tags {
		tagName := strings.ToLower(tag.Name)
		tagName = strings.TrimSpace(tagName)

		if strings.HasPrefix(tagName, "-") {
			tagName = strings.TrimPrefix(tagName, "-")
			deletedTags[tagName] = struct{}{}
		} else {
			addedTags[tagName] = struct{}{}
		}
	}

	// Set default title, excerpt and tags
	for i := range bookmarks {
		if base.Title != "" && overwrite {
			bookmarks[i].Title = base.Title
		}

		if base.Excerpt != "empty" && overwrite {
			bookmarks[i].Excerpt = base.Excerpt
		}

		tempAddedTags := make(map[string]struct{})
		for key, value := range addedTags {
			tempAddedTags[key] = value
		}

		newTags := []model.Tag{}
		for _, tag := range bookmarks[i].Tags {
			if _, isDeleted := deletedTags[tag.Name]; isDeleted {
				tag.Deleted = true
			}

			if _, alreadyExist := addedTags[tag.Name]; alreadyExist {
				delete(tempAddedTags, tag.Name)
			}

			newTags = append(newTags, tag)
		}

		for tag := range tempAddedTags {
			newTags = append(newTags, model.Tag{Name: tag})
		}

		bookmarks[i].Tags = newTags
		bookmarks[i].Modified = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	result, err := DB.UpdateBookmarks(bookmarks)
	if err != nil {
		return []model.Bookmark{}, fmt.Errorf("Failed to update bookmarks: %v", err)
	}

	return result, nil
}
