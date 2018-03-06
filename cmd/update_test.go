package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/RadhiFadlillah/shiori/model"
)

func TestUpdateBookMark(t *testing.T) {
	testbks := []model.Bookmark{
		{
			URL:   "https://github.com/RadhiFadlillah/shiori/releases",
			Title: "Releases",
		},
		{
			URL:   "https://github.com/RadhiFadlillah/shiori/projects",
			Title: "Projects",
		},
	}
	for i, tb := range testbks {
		bk, err := addBookmark(tb, true)
		if err != nil {
			t.Fatalf("failed to create testing bookmarks: %v", err)
		}
		testbks[i].ID = bk.ID
	}

	tests := []struct {
		indices []string
		url     string
		title   string
		excerpt string
		tags    []string
		offline bool
		want    string
	}{
		{
			indices: []string{"9000"},
			want:    "No matching index found",
		},
		{
			indices: []string{"-1"},
			want:    "Index is not valid",
		},
		{
			indices: []string{"3", "-1"},
			want:    "Index is not valid",
		},
		{
			indices: []string{fmt.Sprintf("%d", testbks[0].ID)},
			url:     testbks[0].URL,
			title:   testbks[0].Title + " updated",
			offline: true,
		},
		{
			indices: []string{fmt.Sprintf("%d", testbks[0].ID)},
			offline: false,
		},
		{
			indices: []string{fmt.Sprintf("%d", testbks[1].ID)},
			offline: true,
		},
	}
	for _, tt := range tests {
		bks, err := updateBookmarks(tt.indices, tt.url, tt.title, tt.excerpt, tt.tags, tt.offline)
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
		if tt.want != "" {
			t.Errorf("expected error '%s', got no errors", tt.want)
			continue
		}
		if len(bks) == 0 {
			t.Error("expected at least 1 bookmark, got 0")
			continue
		}
		bk := bks[0]
		if tt.title == "" && bk.Title == tt.title {
			t.Errorf("expected title as '%s', got '%s'", tt.title, bk.Title)
		}
	}
}
