package core

import (
	"fmt"
	"os"
	fp "path/filepath"
	"strconv"
	"strings"

	epub "github.com/go-shiori/go-epub"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/pkg/errors"
)

// GenerateEbook receives a `ProcessRequest` and generates an ebook file in the destination path specified.
// The destination path `dstPath` should include file name with ".epub" extension
// The bookmark model will be used to update the UI based on whether this function is successful or not.
func GenerateEbook(req ProcessRequest, dstPath string) (book model.Bookmark, err error) {

	book = req.Bookmark

	// Make sure bookmark ID is defined
	if book.ID == 0 {
		return book, errors.New("bookmark ID is not valid")
	}

	// Get current state of bookmark cheak archive and thumb
	strID := strconv.Itoa(book.ID)

	imagePath := fp.Join(req.DataDir, "thumb", fmt.Sprintf("%d", book.ID))
	archivePath := fp.Join(req.DataDir, "archive", fmt.Sprintf("%d", book.ID))

	if _, err := os.Stat(imagePath); err == nil {
		book.ImageURL = fp.Join("/", "bookmark", strID, "thumb")
	}

	if _, err := os.Stat(archivePath); err == nil {
		book.HasArchive = true
	}

	// This function create ebook from reader mode of bookmark so
	// we can't create ebook from PDF so we return error here if bookmark is a pdf
	contentType := req.ContentType
	if strings.Contains(contentType, "application/pdf") {
		return book, errors.New("can't create ebook for pdf")
	}

	// Create temporary epub file
	tmpFile, err := os.CreateTemp("", "ebook")
	if err != nil {
		return book, errors.Wrap(err, "can't create temporary EPUB file")
	}
	defer os.Remove(tmpFile.Name())

	// Create last line of ebook
	lastline := `<hr/><p style="text-align:center">Generated By <a href="https://github.com/go-shiori/shiori">Shiori</a> From <a href="` + book.URL + `">This Page</a></p>`

	// Create ebook
	ebook, err := epub.NewEpub(book.Title)
	if err != nil {
		return book, errors.Wrap(err, "can't create EPUB")
	}

	ebook.SetTitle(book.Title)
	ebook.SetAuthor(book.Author)
	ebook.SetDescription(book.Excerpt)
	_, err = ebook.AddSection(`<h1 style="text-align:center"> `+book.Title+` </h1>`+book.HTML+lastline, book.Title, "", "")
	if err != nil {
		return book, errors.Wrap(err, "can't add ebook Section")
	}
	ebook.EmbedImages()
	err = ebook.Write(tmpFile.Name())
	if err != nil {
		return book, errors.Wrap(err, "can't create ebook file")
	}

	defer tmpFile.Close()

	// If everitings go well we start move ebook to dstPath
	err = MoveFileToDestination(dstPath, tmpFile)
	if err != nil {
		return book, errors.Wrap(err, "failed move ebook to destination")
	}

	book.HasEbook = true
	return book, nil
}
