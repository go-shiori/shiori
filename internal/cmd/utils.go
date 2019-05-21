package cmd

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	nurl "net/url"
	"os"
	fp "path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-shiori/shiori/internal/model"
)

var (
	cIndex    = color.New(color.FgHiCyan)
	cSymbol   = color.New(color.FgHiMagenta)
	cTitle    = color.New(color.FgHiGreen).Add(color.Bold)
	cReadTime = color.New(color.FgHiMagenta)
	cURL      = color.New(color.FgHiYellow)
	cError    = color.New(color.FgHiRed)
	cExcerpt  = color.New(color.FgHiWhite)
	cTag      = color.New(color.FgHiBlue)

	errInvalidIndex = errors.New("Index is not valid")
)

func normalizeSpace(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

func isURLValid(s string) bool {
	tmp, err := nurl.Parse(s)
	return err == nil && tmp.Scheme != "" && tmp.Hostname() != ""
}

func clearUTMParams(url *nurl.URL) {
	queries := url.Query()

	for key := range queries {
		if strings.HasPrefix(key, "utm_") {
			queries.Del(key)
		}
	}

	url.RawQuery = queries.Encode()
}

func downloadFile(url, dstPath string, timeout time.Duration) error {
	// Fetch data from URL
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure destination directory exist
	err = os.MkdirAll(fp.Dir(dstPath), os.ModePerm)
	if err != nil {
		return err
	}

	// If destination path doesn't have extension, create it
	if fp.Ext(dstPath) == "" {
		cp := resp.Header.Get("Content-Type")
		exts, err := mime.ExtensionsByType(cp)
		if err != nil {
			return fmt.Errorf("failed to create extension: %v", err)
		}

		if len(exts) == 0 {
			return fmt.Errorf("unknown content type")
		}

		dstPath += exts[0]
	}

	// Create destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Write response body to the file
	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func printBookmarks(bookmarks ...model.Bookmark) {
	for _, bookmark := range bookmarks {
		// Create bookmark index
		strBookmarkIndex := fmt.Sprintf("%d. ", bookmark.ID)
		strSpace := strings.Repeat(" ", len(strBookmarkIndex))

		// Print bookmark title
		cIndex.Print(strBookmarkIndex)
		cTitle.Println(bookmark.Title)

		// Print bookmark URL
		cSymbol.Print(strSpace + "> ")
		cURL.Println(bookmark.URL)

		// Print bookmark excerpt
		if bookmark.Excerpt != "" {
			cSymbol.Print(strSpace + "+ ")
			cExcerpt.Println(bookmark.Excerpt)
		}

		// Print bookmark tags
		if len(bookmark.Tags) > 0 {
			cSymbol.Print(strSpace + "# ")
			for i, tag := range bookmark.Tags {
				if i == len(bookmark.Tags)-1 {
					cTag.Println(tag.Name)
				} else {
					cTag.Print(tag.Name + ", ")
				}
			}
		}

		// Append new line
		fmt.Println()
	}
}

// parseStrIndices converts a list of indices to their integer values
func parseStrIndices(indices []string) ([]int, error) {
	var listIndex []int
	for _, strIndex := range indices {
		if !strings.Contains(strIndex, "-") {
			index, err := strconv.Atoi(strIndex)
			if err != nil || index < 1 {
				return nil, errInvalidIndex
			}

			listIndex = append(listIndex, index)
			continue
		}

		parts := strings.Split(strIndex, "-")
		if len(parts) != 2 {
			return nil, errInvalidIndex
		}

		minIndex, errMin := strconv.Atoi(parts[0])
		maxIndex, errMax := strconv.Atoi(parts[1])
		if errMin != nil || errMax != nil || minIndex < 1 || minIndex > maxIndex {
			return nil, errInvalidIndex
		}

		for i := minIndex; i <= maxIndex; i++ {
			listIndex = append(listIndex, i)
		}
	}

	return listIndex, nil
}
