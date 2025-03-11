package model

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBookmarkToDTO(t *testing.T) {
	// Create a test bookmark
	bookmark := Bookmark{
		ID:         123,
		URL:        "https://example.com",
		Title:      "Example Title",
		Excerpt:    "This is an excerpt",
		Author:     "John Doe",
		Public:     1,
		CreatedAt:  "2023-01-01 12:00:00",
		ModifiedAt: "2023-01-02 12:00:00",
		HasContent: true,
	}

	// Convert to DTO
	dto := bookmark.ToDTO()

	// Verify all fields are correctly transferred
	assert.Equal(t, bookmark.ID, dto.ID, "ID should match")
	assert.Equal(t, bookmark.URL, dto.URL, "URL should match")
	assert.Equal(t, bookmark.Title, dto.Title, "Title should match")
	assert.Equal(t, bookmark.Excerpt, dto.Excerpt, "Excerpt should match")
	assert.Equal(t, bookmark.Author, dto.Author, "Author should match")
	assert.Equal(t, bookmark.Public, dto.Public, "Public should match")
	assert.Equal(t, bookmark.CreatedAt, dto.CreatedAt, "CreatedAt should match")
	assert.Equal(t, bookmark.ModifiedAt, dto.ModifiedAt, "ModifiedAt should match")
	assert.Equal(t, bookmark.HasContent, dto.HasContent, "HasContent should match")

	// Verify default values for fields not in Bookmark
	assert.Empty(t, dto.Content, "Content should be empty")
	assert.Empty(t, dto.HTML, "HTML should be empty")
	assert.Empty(t, dto.ImageURL, "ImageURL should be empty")
	assert.Empty(t, dto.Tags, "Tags should be empty")
	assert.False(t, dto.HasArchive, "HasArchive should be false")
	assert.False(t, dto.HasEbook, "HasEbook should be false")
	assert.False(t, dto.CreateArchive, "CreateArchive should be false")
	assert.False(t, dto.CreateEbook, "CreateEbook should be false")
}

func TestBookmarkDTOToBookmark(t *testing.T) {
	// Create a test BookmarkDTO with all fields populated
	dto := BookmarkDTO{
		ID:            123,
		URL:           "https://example.com",
		Title:         "Example Title",
		Excerpt:       "This is an excerpt",
		Author:        "John Doe",
		Public:        1,
		CreatedAt:     "2023-01-01 12:00:00",
		ModifiedAt:    "2023-01-02 12:00:00",
		Content:       "This is the content",
		HTML:          "<p>This is HTML</p>",
		ImageURL:      "https://example.com/image.jpg",
		HasContent:    true,
		Tags:          []TagDTO{{Tag: Tag{ID: 1, Name: "tag1"}}, {Tag: Tag{ID: 2, Name: "tag2"}}},
		HasArchive:    true,
		HasEbook:      true,
		CreateArchive: true,
		CreateEbook:   true,
	}

	// Convert to Bookmark
	bookmark := dto.ToBookmark()

	// Verify all fields are correctly transferred
	assert.Equal(t, dto.ID, bookmark.ID, "ID should match")
	assert.Equal(t, dto.URL, bookmark.URL, "URL should match")
	assert.Equal(t, dto.Title, bookmark.Title, "Title should match")
	assert.Equal(t, dto.Excerpt, bookmark.Excerpt, "Excerpt should match")
	assert.Equal(t, dto.Author, bookmark.Author, "Author should match")
	assert.Equal(t, dto.Public, bookmark.Public, "Public should match")
	assert.Equal(t, dto.CreatedAt, bookmark.CreatedAt, "CreatedAt should match")
	assert.Equal(t, dto.ModifiedAt, bookmark.ModifiedAt, "ModifiedAt should match")
	assert.Equal(t, dto.HasContent, bookmark.HasContent, "HasContent should match")

	// Fields that should not be transferred
	// These fields are only in BookmarkDTO and not in Bookmark
}

func TestGetThumbnailPath(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		bookmark BookmarkDTO
		expected string
	}{
		{
			name: "With ID",
			bookmark: BookmarkDTO{
				ID: 123,
			},
			expected: filepath.Join("thumb", "123"),
		},
		{
			name: "With zero ID",
			bookmark: BookmarkDTO{
				ID: 0,
			},
			expected: filepath.Join("thumb", "0"),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := GetThumbnailPath(&tc.bookmark)
			assert.Equal(t, tc.expected, path, "Thumbnail path should match expected value")
		})
	}
}

func TestGetEbookPath(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		bookmark BookmarkDTO
		expected string
	}{
		{
			name: "With ID",
			bookmark: BookmarkDTO{
				ID: 123,
			},
			expected: filepath.Join("ebook", "123.epub"),
		},
		{
			name: "With zero ID",
			bookmark: BookmarkDTO{
				ID: 0,
			},
			expected: filepath.Join("ebook", "0.epub"),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := GetEbookPath(&tc.bookmark)
			assert.Equal(t, tc.expected, path, "Ebook path should match expected value")
		})
	}
}

func TestGetArchivePath(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		bookmark BookmarkDTO
		expected string
	}{
		{
			name: "With ID",
			bookmark: BookmarkDTO{
				ID: 123,
			},
			expected: filepath.Join("archive", "123"),
		},
		{
			name: "With zero ID",
			bookmark: BookmarkDTO{
				ID: 0,
			},
			expected: filepath.Join("archive", "0"),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := GetArchivePath(&tc.bookmark)
			assert.Equal(t, tc.expected, path, "Archive path should match expected value")
		})
	}
}

func TestBookmarkRoundTrip(t *testing.T) {
	// Test that converting from Bookmark to DTO and back preserves data
	original := Bookmark{
		ID:         123,
		URL:        "https://example.com",
		Title:      "Example Title",
		Excerpt:    "This is an excerpt",
		Author:     "John Doe",
		Public:     1,
		CreatedAt:  "2023-01-01 12:00:00",
		ModifiedAt: "2023-01-02 12:00:00",
		HasContent: true,
	}

	// Convert to DTO and back
	dto := original.ToDTO()
	roundTrip := dto.ToBookmark()

	// Verify all fields are preserved
	assert.Equal(t, original.ID, roundTrip.ID, "ID should be preserved")
	assert.Equal(t, original.URL, roundTrip.URL, "URL should be preserved")
	assert.Equal(t, original.Title, roundTrip.Title, "Title should be preserved")
	assert.Equal(t, original.Excerpt, roundTrip.Excerpt, "Excerpt should be preserved")
	assert.Equal(t, original.Author, roundTrip.Author, "Author should be preserved")
	assert.Equal(t, original.Public, roundTrip.Public, "Public should be preserved")
	assert.Equal(t, original.CreatedAt, roundTrip.CreatedAt, "CreatedAt should be preserved")
	assert.Equal(t, original.ModifiedAt, roundTrip.ModifiedAt, "ModifiedAt should be preserved")
	assert.Equal(t, original.HasContent, roundTrip.HasContent, "HasContent should be preserved")
}
