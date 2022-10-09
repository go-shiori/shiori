package core

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"math"
	"net/url"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/go-shiori/go-readability"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"

	// Add support for png
	_ "image/png"
)

// ProcessRequest is the request for processing bookmark.
type ProcessRequest struct {
	DataDir     string
	Bookmark    model.Bookmark
	Content     io.Reader
	ContentType string
	KeepTitle   bool
	KeepExcerpt bool
	LogArchival bool
}

// ProcessBookmark process the bookmark and archive it if needed.
// Return three values, is error fatal, and error value.
func ProcessBookmark(req ProcessRequest) (book model.Bookmark, isFatalErr bool, err error) {
	book = req.Bookmark
	contentType := req.ContentType

	// Make sure bookmark ID is defined
	if book.ID == 0 {
		return book, true, fmt.Errorf("bookmark ID is not valid")
	}

	// Split bookmark content so it can be processed several times
	archivalInput := bytes.NewBuffer(nil)
	readabilityInput := bytes.NewBuffer(nil)
	readabilityCheckInput := bytes.NewBuffer(nil)

	var multiWriter io.Writer
	if !strings.Contains(contentType, "text/html") {
		multiWriter = io.MultiWriter(archivalInput)
	} else {
		multiWriter = io.MultiWriter(archivalInput, readabilityInput, readabilityCheckInput)
	}

	_, err = io.Copy(multiWriter, req.Content)
	if err != nil {
		return book, false, fmt.Errorf("failed to process article: %v", err)
	}

	// If this is HTML, parse for readable content
	var imageURLs []string
	if strings.Contains(contentType, "text/html") {
		isReadable := readability.Check(readabilityCheckInput)

		nurl, err := url.Parse(book.URL)
		if err != nil {
			return book, true, fmt.Errorf("Failed to parse url: %v", err)
		}

		article, err := readability.FromReader(readabilityInput, nurl)
		if err != nil {
			return book, false, fmt.Errorf("failed to parse article: %v", err)
		}

		book.Author = article.Byline
		book.Content = article.TextContent
		book.HTML = article.Content

		// If title and excerpt doesnt have submitted value, use from article
		if !req.KeepTitle || book.Title == "" {
			book.Title = article.Title
		}

		if !req.KeepExcerpt || book.Excerpt == "" {
			book.Excerpt = article.Excerpt
		}

		// Sometimes article doesn't have any title, so make sure it is not empty
		if book.Title == "" {
			book.Title = book.URL
		}

		// Get image URL
		if article.Image != "" {
			imageURLs = append(imageURLs, article.Image)
		}

		if article.Favicon != "" {
			imageURLs = append(imageURLs, article.Favicon)
		}

		if !isReadable {
			book.Content = ""
		}

		book.HasContent = book.Content != ""
	}

	// Save article image to local disk
	strID := strconv.Itoa(book.ID)
	imgPath := fp.Join(req.DataDir, "thumb", strID)

	for _, imageURL := range imageURLs {
		err = downloadBookImage(imageURL, imgPath)
		if err == nil {
			book.ImageURL = path.Join("/", "bookmark", strID, "thumb")
			break
		}
	}

	// If needed, create offline archive as well
	if book.CreateArchive {
		archivePath := fp.Join(req.DataDir, "archive", fmt.Sprintf("%d", book.ID))
		os.Remove(archivePath)

		archivalRequest := warc.ArchivalRequest{
			URL:         book.URL,
			Reader:      archivalInput,
			ContentType: contentType,
			UserAgent:   userAgent,
			LogEnabled:  req.LogArchival,
		}

		err = warc.NewArchive(archivalRequest, archivePath)
		if err != nil {
			return book, false, fmt.Errorf("failed to create archive: %v", err)
		}

		book.HasArchive = true
	}

	return book, false, nil
}

func downloadBookImage(url, dstPath string) error {
	// Fetch data from URL
	resp, err := httpClient.Get(url)
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
	err = os.MkdirAll(fp.Dir(dstPath), model.DataDirPerm)
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
		draw.Draw(bg, imgRect, image.NewUniform(color.White), image.Point{}, draw.Src)
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
