package core_test

import (
	"fmt"
	"os"
	fp "path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerateEbook_ValidBookmarkID_ReturnsBookmarkWithHasEbookTrue(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)
	parentDir := t.TempDir()
	defer os.RemoveAll(parentDir)

	mockRequest := core.ProcessRequest{
		Bookmark: model.Bookmark{
			ID:       1,
			Title:    "Example Bookmark",
			HTML:     "<html><body>Example HTML</body></html>",
			HasEbook: false,
		},
		DataDir:     parentDir,
		ContentType: "text/html",
	}

	bookmark, err := core.GenerateEbook(mockRequest, fp.Join(tempDir, "1"))

	assert.True(t, bookmark.HasEbook)
	assert.NoError(t, err)
}

func TestGenerateEbook_InvalidBookmarkID_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)
	mockRequest := core.ProcessRequest{
		Bookmark: model.Bookmark{
			ID:       0,
			HasEbook: false,
		},
		DataDir:     tempDir,
		ContentType: "text/html",
	}

	bookmark, err := core.GenerateEbook(mockRequest, tempDir)

	assert.Equal(t, model.Bookmark{
		ID:       0,
		HasEbook: false,
	}, bookmark)
	assert.Error(t, err)
}

func TestGenerateEbook_ValidBookmarkID_EbookExist_EbookExist_ReturnWithHasEbookTrue(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)
	parentDir := t.TempDir()
	defer os.RemoveAll(parentDir)

	mockRequest := core.ProcessRequest{
		Bookmark: model.Bookmark{
			ID:       1,
			HasEbook: false,
		},
		DataDir:     parentDir,
		ContentType: "text/html",
	}
	// Create the ebook directory
	ebookDir := fp.Join(mockRequest.DataDir, "ebook")
	err := os.MkdirAll(ebookDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	// Create the ebook file
	ebookfile := fp.Join(mockRequest.DataDir, "ebook", fmt.Sprintf("%d.epub", mockRequest.Bookmark.ID))
	file, err := os.Create(ebookfile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	bookmark, err := core.GenerateEbook(mockRequest, fp.Join(tempDir, "1"))

	assert.True(t, bookmark.HasEbook)
	assert.NoError(t, err)
}

func TestGenerateEbook_ValidBookmarkID_EbookExist_ImagePathExist_ReturnWithHasEbookTrue(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)
	parentDir := t.TempDir()
	defer os.RemoveAll(parentDir)

	mockRequest := core.ProcessRequest{
		Bookmark: model.Bookmark{
			ID:       1,
			HasEbook: false,
		},
		DataDir:     parentDir,
		ContentType: "text/html",
	}
	// Create the image directory
	imageDir := fp.Join(mockRequest.DataDir, "thumb")
	err := os.MkdirAll(imageDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	// Create the image file
	imagePath := fp.Join(mockRequest.DataDir, "thumb", fmt.Sprintf("%d", mockRequest.Bookmark.ID))
	file, err := os.Create(imagePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	bookmark, err := core.GenerateEbook(mockRequest, fp.Join(tempDir, "1"))
	expectedimagePath := "/bookmark/1/thumb"
	if expectedimagePath != bookmark.ImageURL {
		t.Errorf("Expected imageURL %s, but got %s", bookmark.ImageURL, expectedimagePath)
	}
	assert.True(t, bookmark.HasEbook)
	assert.NoError(t, err)
}

func TestGenerateEbook_ValidBookmarkID_EbookExist_ReturnWithHasArchiveTrue(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)
	parentDir := t.TempDir()
	defer os.RemoveAll(parentDir)

	mockRequest := core.ProcessRequest{
		Bookmark: model.Bookmark{
			ID:       1,
			HasEbook: false,
		},
		DataDir:     parentDir,
		ContentType: "text/html",
	}
	// Create the archive directory
	archiveDir := fp.Join(mockRequest.DataDir, "archive")
	err := os.MkdirAll(archiveDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	// Create the archive file
	archivePath := fp.Join(mockRequest.DataDir, "archive", fmt.Sprintf("%d", mockRequest.Bookmark.ID))
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	bookmark, err := core.GenerateEbook(mockRequest, fp.Join(tempDir, "1"))
	assert.True(t, bookmark.HasArchive)
	assert.NoError(t, err)
}

func TestGenerateEbook_ValidBookmarkID_RetuenError_PDF(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)

	mockRequest := core.ProcessRequest{
		Bookmark: model.Bookmark{
			ID:       1,
			HasEbook: false,
		},
		DataDir:     tempDir,
		ContentType: "application/pdf",
	}

	bookmark, err := core.GenerateEbook(mockRequest, tempDir)

	assert.False(t, bookmark.HasEbook)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can't create ebook for pdf")
}

// Add more unit tests for other scenarios that missing specialy
// can't create ebook directory and can't write situatuin
// writing inside zip file
// html variable that not export and image download loop

func TestGetImages(t *testing.T) {
	// Test case 1: HTML with no image tags
	html1 := `<html><body><h1>Hello, World!</h1></body></html>`
	expected1 := make(map[string]string)
	result1, err1 := core.GetImages(html1)
	if err1 != nil {
		t.Errorf("Unexpected error: %v", err1)
	}
	if len(result1) != len(expected1) {
		t.Errorf("Expected %d images, but got %d", len(expected1), len(result1))
	}

	// Test case 2: HTML with one image tag
	html2 := `<html><body><img src="image1.jpg"></body></html>`
	expected2 := map[string]string{"image1.jpg": "<img src=\"image1.jpg\">"}
	result2, err2 := core.GetImages(html2)
	if err2 != nil {
		t.Errorf("Unexpected error: %v", err2)
	}
	if len(result2) != len(expected2) {
		t.Errorf("Expected %d images, but got %d", len(expected2), len(result2))
	}
	for key, value := range expected2 {
		if result2[key] != value {
			t.Errorf("Expected image URL %s with tag %s, but got %s", key, value, result2[key])
		}
	}

	// Test case 3: HTML with multiple image tags
	html3 := `<html><body><img src="image1.jpg"><img src="image2.jpg"></body></html>`
	expected3 := map[string]string{
		"image1.jpg": "<img src=\"image1.jpg\">",
		"image2.jpg": "<img src=\"image2.jpg\">",
	}
	result3, err3 := core.GetImages(html3)
	if err3 != nil {
		t.Errorf("Unexpected error: %v", err3)
	}
	if len(result3) != len(expected3) {
		t.Errorf("Expected %d images, but got %d", len(expected3), len(result3))
	}
	for key, value := range expected3 {
		if result3[key] != value {
			t.Errorf("Expected image URL %s with tag %s, but got %s", key, value, result3[key])
		}
	}
	// Test case 4: HTML with multiple image tags with duplicayr
	html4 := `<html><body><img src="image1.jpg"><img src="image2.jpg"><img src="image2.jpg"></body></html>`
	expected4 := map[string]string{
		"image1.jpg": "<img src=\"image1.jpg\">",
		"image2.jpg": "<img src=\"image2.jpg\">",
	}
	result4, err4 := core.GetImages(html4)
	if err4 != nil {
		t.Errorf("Unexpected error: %v", err4)
	}
	if len(result4) != len(expected4) {
		t.Errorf("Expected %d images, but got %d", len(expected4), len(result4))
	}
	for key, value := range expected4 {
		if result4[key] != value {
			t.Errorf("Expected image URL %s with tag %s, but got %s", key, value, result4[key])
		}
	}
}
