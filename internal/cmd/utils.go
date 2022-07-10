package cmd

import (
	"errors"
	"fmt"
	nurl "net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/go-shiori/shiori/internal/model"
	"golang.org/x/term"
)

var (
	cIndex   = color.New(color.FgHiCyan)
	cSymbol  = color.New(color.FgHiMagenta)
	cTitle   = color.New(color.FgHiGreen).Add(color.Bold)
	cURL     = color.New(color.FgHiYellow)
	cExcerpt = color.New(color.FgHiWhite)
	cTag     = color.New(color.FgHiBlue)

	cInfo  = color.New(color.FgHiCyan)
	cError = color.New(color.FgHiRed)

	errInvalidIndex = errors.New("index is not valid")
)

func normalizeSpace(str string) string {
	str = strings.TrimSpace(str)
	return strings.Join(strings.Fields(str), " ")
}

func isURLValid(s string) bool {
	tmp, err := nurl.Parse(s)
	return err == nil && tmp.Scheme != "" && tmp.Hostname() != ""
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

// openBrowser tries to open the URL in a browser,
// and returns any error if it happened.
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

func getTerminalWidth() int {
	width, _, _ := term.GetSize(int(os.Stdin.Fd()))
	return width
}

func validateTitle(title, fallback string) string {
	// Normalize spaces before we begin
	title = normalizeSpace(title)
	title = strings.TrimSpace(title)

	// If at this point title already empty, just uses fallback
	if title == "" {
		return fallback
	}

	// Check if it's already valid UTF-8 string
	if valid := utf8.ValidString(title); valid {
		return title
	}

	// Remove invalid runes to get the valid UTF-8 title
	fixUtf := func(r rune) rune {
		if r == utf8.RuneError {
			return -1
		}
		return r
	}
	validUtf := strings.Map(fixUtf, title)

	// If it's empty use fallback string
	validUtf = strings.TrimSpace(validUtf)
	if validUtf == "" {
		return fallback
	}

	return validUtf
}
