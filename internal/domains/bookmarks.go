package domains

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
)

type BookmarksDomain struct {
	deps *dependencies.Dependencies
}

func (d *BookmarksDomain) HasEbook(b *model.BookmarkDTO) bool {
	ebookPath := filepath.Join("ebook", strconv.Itoa(b.ID)+".epub")
	return d.deps.Domains.Storage.FileExists(ebookPath)
}

func (d *BookmarksDomain) HasArchive(b *model.BookmarkDTO) bool {
	archivePath := filepath.Join(d.deps.Config.Storage.DataDir, "archive", strconv.Itoa(b.ID))
	return d.deps.Domains.Storage.FileExists(archivePath)
}

func (d *BookmarksDomain) GetThumbnailPath(b *model.BookmarkDTO) string {
	return filepath.Join("thumb", strconv.Itoa(b.ID))
}

func (d *BookmarksDomain) HasThumbnail(b *model.BookmarkDTO) bool {
	return d.deps.Domains.Storage.FileExists(d.GetThumbnailPath(b))
}

func (d *BookmarksDomain) GetBookmark(ctx context.Context, id model.DBID) (*model.BookmarkDTO, error) {
	bookmark, _, err := d.deps.Database.GetBookmark(ctx, int(id), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	// Check if it has ebook and archive.
	bookmark.HasEbook = d.HasEbook(&bookmark)
	bookmark.HasArchive = d.HasArchive(&bookmark)

	return &bookmark, nil
}

// GetBookmarkContentsFromArchive gets the HTML contents of a bookmark linking assets to the ones
// archived if the bookmark has an archived version.
func (d *BookmarksDomain) GetBookmarkContentsFromArchive(bookmark *model.BookmarkDTO) (template.HTML, error) {
	if !bookmark.HasArchive {
		return template.HTML(bookmark.HTML), nil
	}

	// Open archive, look in cache first
	archivePath := filepath.Join(d.deps.Config.Storage.DataDir, "archive", fmt.Sprintf("%d", bookmark.ID))
	// TODO: Move to archiver domain
	// TODO: Use storagedomain to operate with the file
	archive, err := warc.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}

	// Find all image and convert its source to use the archive URL.
	createArchivalURL := func(archivalName string) string {
		var archivalURL url.URL
		archivalURL.Path = path.Join(d.deps.Config.Http.RootPath, "bookmark", fmt.Sprintf("%d", bookmark.ID), "archive", archivalName)
		return archivalURL.String()
	}

	buffer := strings.NewReader(bookmark.HTML)
	doc, err := goquery.NewDocumentFromReader(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	doc.Find("img, picture, figure, source").Each(func(_ int, node *goquery.Selection) {
		// Get the needed attributes
		src, _ := node.Attr("src")
		strSrcSets, _ := node.Attr("srcset")

		// Convert `src` attributes
		if src != "" {
			archivalName := getArchiveFileBasename(src)
			if archivalName != "" && archive.HasResource(archivalName) {
				node.SetAttr("src", createArchivalURL(archivalName))
			}
		}

		// Split srcset by comma, then process it like any URLs
		srcSets := strings.Split(strSrcSets, ",")
		for i, srcSet := range srcSets {
			srcSet = strings.TrimSpace(srcSet)
			parts := strings.SplitN(srcSet, " ", 2)
			if parts[0] == "" {
				continue
			}

			archivalName := getArchiveFileBasename(parts[0])
			if archivalName != "" && archive.HasResource(archivalName) {
				archivalURL := createArchivalURL(archivalName)
				srcSets[i] = strings.Replace(srcSets[i], parts[0], archivalURL, 1)
			}
		}

		if len(srcSets) > 0 {
			node.SetAttr("srcset", strings.Join(srcSets, ","))
		}
	})

	bookmark.HTML, err = goquery.OuterHtml(doc.Selection)
	if err != nil {
		return template.HTML(bookmark.HTML), fmt.Errorf("failed to get HTML: %w", err)
	}

	return template.HTML(bookmark.HTML), nil
}

func NewBookmarksDomain(deps *dependencies.Dependencies) *BookmarksDomain {
	return &BookmarksDomain{
		deps: deps,
	}
}
