package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-shiori/shiori/internal/database"
)

func Test_parseCsvExport_old_format(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
	}{
		{
			name:     "Test old file format",
			fileName: "pocket-old.csv",
		},
		{
			name:     "Test new file format",
			fileName: "pocket-new.csv",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Open("../../testdata/" + tt.fileName)
			if err != nil {
				t.Error(err.Error())
			}
			defer file.Close()
			ctx := context.TODO()

			tmpDir, err := os.MkdirTemp("", "shiori-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			dbPath := filepath.Join(tmpDir, "shiori.db")
			db, err := database.OpenSQLiteDatabase(ctx, dbPath)
			if err != nil {
				t.Fatalf("failed to open sqlite database: %v", err)
			}

			if err := db.Migrate(ctx); err != nil {
				t.Fatalf("failed to migrate sqlite database: %v", err)
			}

			bookmarks := parseCsvExport(ctx, db, file)
			if len(bookmarks) != 1 {
				t.Errorf("Expected 1 bookmarks, got %d", len(bookmarks))
			}
			bm := bookmarks[0]
			if bm.Title != "Shiori" {
				t.Errorf("Expected Title Shiori got %s", bm.URL)
			}
			if bm.URL != "https://github.com/go-shiori/shiori" {
				t.Errorf("Expected URL https://github.com/go-shiori/shiori, got %s", bm.URL)
			}
			if len(bm.Tags) != 1 {
				t.Errorf("Expected 1 tags, got %d", len(bm.Tags))
			}
			if bm.Tags[0].Name != "shiori" {
				t.Errorf("Expected tag shiori, got %s", bm.Tags[0].Name)
			}
			if bm.CreatedAt == "" {
				t.Error("Expected CreatedAt to be not empty")
			}
			if bm.ModifiedAt == "" {
				t.Error("Expected CreatedAt to be not empty")
			}
		})
	}
}
