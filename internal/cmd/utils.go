package cmd

import (
	"errors"
	"fmt"
	"image"
	clr "image/color"
	"image/draw"
	"image/jpeg"
	"math"
	"net/http"
	nurl "net/url"
	"os"
	"os/exec"
	fp "path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/disintegration/imaging"
	"github.com/fatih/color"
	"github.com/go-shiori/shiori/internal/model"
	"golang.org/x/crypto/ssh/terminal"

	// Add supports for PNG image
	_ "image/png"
)

var (
	cIndex    = color.New(color.FgHiCyan)
	cSymbol   = color.New(color.FgHiMagenta)
	cTitle    = color.New(color.FgHiGreen).Add(color.Bold)
	cReadTime = color.New(color.FgHiMagenta)
	cURL      = color.New(color.FgHiYellow)
	cExcerpt  = color.New(color.FgHiWhite)
	cTag      = color.New(color.FgHiBlue)

	cInfo    = color.New(color.FgHiCyan)
	cError   = color.New(color.FgHiRed)
	cWarning = color.New(color.FgHiYellow)

	errInvalidIndex = errors.New("Index is not valid")
)

func normalizeSpace(str string) string {
	str = strings.TrimSpace(str)
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

func downloadBookImage(url, dstPath string, timeout time.Duration) error {
	// Fetch data from URL
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure it's JPG or PNG image
	cp := resp.Header.Get("Content-Type")
	if !strings.Contains(cp, "image/jpeg") && !strings.Contains(cp, "image/png") {
		return fmt.Errorf("%s is not a supported image", url)
	}

	// At this point, the download has finished successfully.
	// Prepare destination file.
	err = os.MkdirAll(fp.Dir(dstPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create image dir: %v", err)
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %v", err)
	}
	defer dstFile.Close()

	// Parse image and process it.
	// If image is smaller than 600x400 or its ratio is less than 4:3, resize.
	// Else, save it as it is.
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse image %s: %v", url, err)
	}

	imgRect := img.Bounds()
	imgWidth := imgRect.Dx()
	imgHeight := imgRect.Dy()
	imgRatio := float64(imgWidth) / float64(imgHeight)

	if imgWidth >= 600 && imgHeight >= 400 && imgRatio > 1.3 {
		err = jpeg.Encode(dstFile, img, nil)
	} else {
		// Create background
		bg := image.NewNRGBA(imgRect)
		draw.Draw(bg, imgRect, image.NewUniform(clr.White), image.Point{}, draw.Src)
		draw.Draw(bg, imgRect, img, image.Point{}, draw.Over)

		bg = imaging.Fill(bg, 600, 400, imaging.Center, imaging.Lanczos)
		bg = imaging.Blur(bg, 150)
		bg = imaging.AdjustBrightness(bg, 30)

		// Create foreground
		fg := imaging.Fit(img, 600, 400, imaging.Lanczos)

		// Merge foreground and background
		bgRect := bg.Bounds()
		fgRect := fg.Bounds()
		fgPosition := image.Point{
			X: bgRect.Min.X - int(math.Round(float64(bgRect.Dx()-fgRect.Dx())/2)),
			Y: bgRect.Min.Y - int(math.Round(float64(bgRect.Dy()-fgRect.Dy())/2)),
		}

		draw.Draw(bg, bgRect, fg, fgPosition, draw.Over)

		// Save to file
		err = jpeg.Encode(dstFile, bg, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to save image %s: %v", url, err)
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
	width, _, _ := terminal.GetSize(int(os.Stdin.Fd()))
	return width
}

func toValidUtf8(src, fallback string) string {
	// Check if it's already valid
	if valid := utf8.ValidString(src); valid {
		return src
	}

	// Remove invalid runes
	fixUtf := func(r rune) rune {
		if r == utf8.RuneError {
			return -1
		}
		return r
	}
	validUtf := strings.Map(fixUtf, src)

	// If it's empty use fallback string
	validUtf = strings.TrimSpace(validUtf)
	if validUtf == "" {
		return fallback
	}

	return validUtf
}
