package cmd

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func pocketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pocket source-file",
		Short: "Import bookmarks from Pocket's data export file",
		Args:  cobra.ExactArgs(1),
		Run:   pocketHandler,
	}

	return cmd
}

func pocketHandler(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	_, deps := initShiori(ctx, cmd)

	// Open pocket's file
	filePath := args[0]
	srcFile, err := os.Open(filePath)
	if err != nil {
		cError.Println(err)
		os.Exit(1)
	}
	defer srcFile.Close()

	var bookmarks []model.BookmarkDTO
	switch filepath.Ext(filePath) {
	case ".html":
		bookmarks = parseHtmlExport(ctx, deps.Database(), srcFile)
	case ".csv":
		bookmarks = parseCsvExport(ctx, deps.Database(), srcFile)
	default:
		cError.Println("Invalid file format. Only HTML and CSV are supported.")
		os.Exit(1)
	}

	// Save bookmark to database
	bookmarks, err = deps.Database().SaveBookmarks(ctx, true, bookmarks...)
	if err != nil {
		cError.Printf("Failed to save bookmarks: %v\n", err)
		os.Exit(1)
	}

	// Print imported bookmarks
	fmt.Println()
	printBookmarks(bookmarks...)
}

// Parse bookmarks from HTML file
func parseHtmlExport(ctx context.Context, db model.DB, srcFile *os.File) []model.BookmarkDTO {
	bookmarks := []model.BookmarkDTO{}
	mapURL := make(map[string]struct{})

	doc, err := goquery.NewDocumentFromReader(srcFile)
	if err != nil {
		cError.Println(err)
		os.Exit(1)
	}

	doc.Find("a").Each(func(_ int, a *goquery.Selection) {
		// Get metadata
		title := a.Text()
		url, _ := a.Attr("href")
		tagsStr, _ := a.Attr("tags")
		timeAddedStr, _ := a.Attr("time_added")

		title, url, timeAdded, tags, err := verifyMetadata(title, url, timeAddedStr, tagsStr)
		if err != nil {
			cError.Printf("Skip %s: %v\n", url, err)
			return
		}

		if err = handleDuplicates(ctx, db, mapURL, url); err != nil {
			cError.Printf("Skip %s: %v\n", url, err)
			return
		}

		// Add item to list
		bookmark := model.BookmarkDTO{
			URL:        url,
			Title:      title,
			ModifiedAt: timeAdded.Format(model.DatabaseDateFormat),
			CreatedAt:  timeAdded.Format(model.DatabaseDateFormat),
			Tags:       tags,
		}

		mapURL[url] = struct{}{}
		bookmarks = append(bookmarks, bookmark)
	})

	return bookmarks
}

// Parse bookmarks from CSV file
func parseCsvExport(ctx context.Context, db model.DB, srcFile *os.File) []model.BookmarkDTO {
	bookmarks := []model.BookmarkDTO{}
	mapURL := make(map[string]struct{})

	reader := csv.NewReader(srcFile)
	records, err := reader.ReadAll()
	if err != nil {
		cError.Println(err)
		os.Exit(1)
	}

	for i, cols := range records {
		// Check and skip header
		if i == 0 {
			expected := []string{"title", "url", "time_added", "cursor", "tags", "status"}
			if slices.Compare(cols, expected) != 0 {
				cError.Printf("Invalid CSV format. Header must be: %s\n", strings.Join(expected, ","))
				os.Exit(1)
			}
			continue
		}

		// Get metadata
		title, url, timeAdded, tags, err := verifyMetadata(cols[0], cols[1], cols[2], cols[4])
		if err != nil {
			cError.Printf("Skip %s: %v\n", url, err)
			continue
		}

		if err = handleDuplicates(ctx, db, mapURL, url); err != nil {
			cError.Printf("Skip %s: %v\n", url, err)
			continue
		}

		// Add item to list
		bookmark := model.BookmarkDTO{
			URL:        url,
			Title:      title,
			ModifiedAt: timeAdded.Format(model.DatabaseDateFormat),
			CreatedAt:  timeAdded.Format(model.DatabaseDateFormat),
			Tags:       tags,
		}

		mapURL[url] = struct{}{}
		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks
}

// Parse metadata and verify it's validity
func verifyMetadata(title, url, timeAddedStr, tags string) (string, string, time.Time, []model.TagDTO, error) {
	// Clean up URL
	var err error
	url, err = core.RemoveUTMParams(url)
	if err != nil {
		err = fmt.Errorf("URL is not valid, %w", err)
		return "", "", time.Time{}, nil, err
	}

	// Make sure title is valid Utf-8
	title = validateTitle(title, url)

	// Parse time added
	timeAddedInt, err := strconv.ParseInt(timeAddedStr, 10, 64)
	if err != nil {
		err = fmt.Errorf("Invalid time added, %w", err)
		return "", "", time.Time{}, nil, err
	}
	timeAdded := time.Unix(timeAddedInt, 0)

	// Get bookmark tags
	tagsList := []model.TagDTO{}
	// We need to split tags by both comma or pipe,
	// because Pocket's CSV export use pipe as separator,
	// while HTML export use comma.
	for _, tag := range regexp.MustCompile(`[,|]`).Split(tags, -1) {
		if tag != "" {
			tagsList = append(tagsList, model.TagDTO{
				Tag: model.Tag{Name: tag},
			})
		}
	}

	return title, url, timeAdded, tagsList, nil
}

// Checks if the URL already exist, both in bookmark
// file or in database
func handleDuplicates(ctx context.Context, db model.DB, mapURL map[string]struct{}, url string) error {
	if _, exists := mapURL[url]; exists {
		return errors.New("URL already exists")
	}

	_, exists, err := db.GetBookmark(ctx, 0, url)
	if err != nil {
		return fmt.Errorf("Failed getting bookmark, %w", err)
	}

	if exists {
		return errors.New("URL already exists")
	}

	return nil
}
