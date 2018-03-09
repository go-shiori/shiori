package cmd

import (
	"strings"
	"testing"

	"github.com/RadhiFadlillah/shiori/model"
)

func TestAddBookMark(t *testing.T) {
	tests := []struct {
		bookmark model.Bookmark
		offline  bool
		want     string
	}{
		{
			model.Bookmark{},
			true, "URL is not valid",
		},
		{
			model.Bookmark{
				URL: "https://github.com/RadhiFadlillah/shiori",
			},
			true, "Title must not be empty",
		},
		{
			model.Bookmark{
				URL:   "https://github.com/RadhiFadlillah/shiori",
				Title: "Shiori",
			},
			true, "",
		},
		{
			model.Bookmark{
				URL: "https://github.com/RadhiFadlillah/shiori/issues",
			},
			false, "",
		},
	}
	for _, tt := range tests {
		bk, err := addBookmark(tt.bookmark, tt.offline)
		if err != nil {
			if tt.want == "" {
				t.Errorf("got unexpected error: '%v'", err)
				continue
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("expected error '%s', got '%v'", tt.want, err)
			}
			continue
		}
		if tt.bookmark.URL == "" {
			t.Errorf("expected error '%s', got '%v'", tt.want, err)
			continue
		}
		if tt.offline && tt.bookmark.Title == "" {
			t.Error("expected error 'Title must not be empty', got no error")
			continue
		}

		if tt.want != "" {
			t.Errorf("expected error '%s', got no error", tt.want)
			continue
		}
		if tt.offline && bk.Title != tt.bookmark.Title {
			t.Errorf("expected title '%s', got '%s'", tt.bookmark.Title, bk.Title)
		}
		if !tt.offline && bk.Title == "" {
			t.Error("expected title, got empty string ''")
		}
	}
}
