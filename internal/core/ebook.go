package core

import (
	"fmt"
	"os"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"

	epub "github.com/bmaupin/go-epub"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/pkg/errors"
)

// GenerateEbook receives a `ProcessRequest` and generates an ebook file in the destination path specified.
// The destination path `dstPath` should include file name with ".epub" extension
// The bookmark model will be used to update the UI based on whether this function is successful or not.
func GenerateEbook(req ProcessRequest, dstPath string) (book model.Bookmark, err error) {
	// variable for store generated html code
	//var html string

	book = req.Bookmark

	// Make sure bookmark ID is defined
	if book.ID == 0 {
		return book, errors.New("bookmark ID is not valid")
	}

	// get current state of bookmark
	// cheak archive and thumb
	strID := strconv.Itoa(book.ID)

	imagePath := fp.Join(req.DataDir, "thumb", fmt.Sprintf("%d", book.ID))
	archivePath := fp.Join(req.DataDir, "archive", fmt.Sprintf("%d", book.ID))

	if _, err := os.Stat(imagePath); err == nil {
		book.ImageURL = fp.Join("/", "bookmark", strID, "thumb")
	}

	if _, err := os.Stat(archivePath); err == nil {
		book.HasArchive = true
	}

	// this function create ebook from reader mode of bookmark so
	// we can't create ebook from PDF so we return error here if bookmark is a pdf
	contentType := req.ContentType
	if strings.Contains(contentType, "application/pdf") {
		return book, errors.New("can't create ebook for pdf")
	}

	// create temporary epub file
	tmpFile, err := os.CreateTemp("", "ebook")
	if err != nil {
		return book, errors.Wrap(err, "can't create temporary EPUB file")
	}
	defer os.Remove(tmpFile.Name())

	// Create zip archive
	//epubWriter := zip.NewWriter(tmpFile)
	ebook := epub.NewEpub(book.Title)
	ebook.SetTitle(book.Title)
	ebook.AddSection(string(book.HTML), book.Title, "", "")
	ebook.EmbedImages()
	ebook.Write(tmpFile.Name())

	defer tmpFile.Close()
	// if everitings go well we start move ebook to dstPath
	err = MoveFileToDestination(dstPath, tmpFile)
	if err != nil {
		return book, errors.Wrap(err, "failed move ebook to destination")
	}

	book.HasEbook = true
	return book, nil
}

// function get html and return list of image url inside html file
func GetImages(html string) (map[string]string, error) {
	// Regular expression to match image tags and their URLs
	imageTagRegex := regexp.MustCompile(`<img.*?src="(.*?)".*?>`)

	// Find all matches in the HTML string
	imageTagMatches := imageTagRegex.FindAllStringSubmatch(html, -1)
	// Create a dictionary to store the image URLs
	images := make(map[string]string)

	// Check if there are any matches
	if len(imageTagMatches) == 0 {
		return nil, nil
	}

	// Loop through all the matches and add them to the dictionary
	for _, match := range imageTagMatches {
		imageURL := match[1]
		if !strings.HasPrefix(imageURL, "data:image/") {
			images[imageURL] = match[0]
		}
	}

	return images, nil
}
