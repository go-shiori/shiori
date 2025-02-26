package core

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	fp "path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/go-shiori/go-readability"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
	"github.com/pkg/errors"
	_ "golang.org/x/image/webp"

	// Add support for png
	_ "image/png"
)

// ProcessRequest is the request for processing bookmark.
type ProcessRequest struct {
	DataDir     string
	Bookmark    model.BookmarkDTO
	Content     io.Reader
	ContentType string
	KeepTitle   bool
	KeepExcerpt bool
	LogArchival bool
}

var ErrNoSupportedImageType = errors.New("unsupported image type")

// ProcessBookmark process the bookmark and archive it if needed.
// Return three values, is error fatal, and error value.
func ProcessBookmark(deps model.Dependencies, req ProcessRequest) (book model.BookmarkDTO, isFatalErr bool, err error) {
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
	strID := strconv.Itoa(book.ID)
	imgPath := model.GetThumbnailPath(&book)
	var imageURLs []string
	if strings.Contains(contentType, "text/html") {
		isReadable := readability.Check(readabilityCheckInput)

		nurl, err := url.Parse(book.URL)
		if err != nil {
			return book, true, fmt.Errorf("failed to parse url: %v", err)
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
		} else {
			deps.Domains().Storage().FS().Remove(imgPath)
		}

		if article.Favicon != "" {
			imageURLs = append(imageURLs, article.Favicon)
		}

		if !isReadable {
			book.Content = ""
		}

		book.HasContent = book.Content != ""
		book.ModifiedAt = ""
	}

	// Save article image to local disk
	for i, imageURL := range imageURLs {
		err = DownloadBookImage(deps, imageURL, imgPath)
		if err != nil && errors.Is(err, ErrNoSupportedImageType) {
			log.Printf("%s: %s", err, imageURL)
			if i == len(imageURLs)-1 {
				deps.Domains().Storage().FS().Remove(imgPath)
			}
		}
		if err != nil {
			log.Printf("File download not successful for image URL: %s", imageURL)
			continue
		}
		if err == nil {
			book.ImageURL = fp.Join("/", "bookmark", strID, "thumb")
			book.ModifiedAt = ""
			break
		}
	}

	// If needed, create ebook as well
	if book.CreateEbook {
		ebookPath := model.GetEbookPath(&book)
		req.Bookmark = book

		if strings.Contains(contentType, "application/pdf") {
			return book, false, errors.Wrap(err, "can't create ebook from pdf")
		} else {
			_, err = GenerateEbook(deps, req, ebookPath)
			if err != nil {
				return book, true, errors.Wrap(err, "failed to create ebook")
			}
			book.HasEbook = true
			book.ModifiedAt = ""
		}
	}

	// If needed, create offline archive as well
	if book.CreateArchive {
		tmpFile, err := os.CreateTemp("", "archive")
		if err != nil {
			return book, false, fmt.Errorf("failed to create temp archive: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		archivalRequest := warc.ArchivalRequest{
			URL:         book.URL,
			Reader:      archivalInput,
			ContentType: contentType,
			UserAgent:   userAgent,
			LogEnabled:  req.LogArchival,
		}

		err = warc.NewArchive(archivalRequest, tmpFile.Name())
		if err != nil {
			return book, false, fmt.Errorf("failed to create archive: %v", err)
		}

		dstPath := model.GetArchivePath(&book)
		err = deps.Domains().Storage().WriteFile(dstPath, tmpFile)
		if err != nil {
			return book, false, fmt.Errorf("failed move archive to destination `: %v", err)
		}

		book.HasArchive = true
		book.ModifiedAt = ""
	}

	return book, false, nil
}

func DownloadBookImage(deps model.Dependencies, url, dstPath string) error {
	// Fetch data from URL
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure it's JPG or PNG image
	cp := resp.Header.Get("Content-Type")
	if !strings.Contains(cp, "image/jpeg") &&
		!strings.Contains(cp, "image/pjpeg") &&
		!strings.Contains(cp, "image/jpg") &&
		!strings.Contains(cp, "image/webp") &&
		!strings.Contains(cp, "image/png") {
		return ErrNoSupportedImageType
	}

	// At this point, the download has finished successfully.
	// Create tmpFile
	tmpFile, err := os.CreateTemp("", "image")
	if err != nil {
		return fmt.Errorf("failed to create temporary image file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

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
		err = jpeg.Encode(tmpFile, img, nil)
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
		err = jpeg.Encode(tmpFile, bg, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to save image %s: %v", url, err)
	}

	err = deps.Domains().Storage().WriteFile(dstPath, tmpFile)
	if err != nil {
		return err
	}

	return nil
}
