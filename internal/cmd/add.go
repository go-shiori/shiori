package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/go-shiori/shiori/pkg/warc"

	"github.com/go-shiori/go-readability"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add url",
		Short: "Bookmark the specified URL",
		Args:  cobra.ExactArgs(1),
		Run:   addHandler,
	}

	cmd.Flags().StringP("title", "i", "", "Custom title for this bookmark")
	cmd.Flags().StringP("excerpt", "e", "", "Custom excerpt for this bookmark")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for this bookmark")
	cmd.Flags().BoolP("offline", "o", false, "Save bookmark without fetching data from internet")
	cmd.Flags().BoolP("no-archival", "a", false, "Save bookmark without creating offline archive")
	cmd.Flags().Bool("log-archival", false, "Log the archival process")

	return cmd
}

func addHandler(cmd *cobra.Command, args []string) {
	// Read flag and arguments
	url := args[0]
	title, _ := cmd.Flags().GetString("title")
	excerpt, _ := cmd.Flags().GetString("excerpt")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	offline, _ := cmd.Flags().GetBool("offline")
	noArchival, _ := cmd.Flags().GetBool("no-archival")
	logArchival, _ := cmd.Flags().GetBool("log-archival")

	// Clean up URL by removing its fragment and UTM parameters
	tmp, err := nurl.Parse(url)
	if err != nil || tmp.Scheme == "" || tmp.Hostname() == "" {
		cError.Println("URL is not valid")
		return
	}

	tmp.Fragment = ""
	clearUTMParams(tmp)

	// Create bookmark item
	book := model.Bookmark{
		URL:     tmp.String(),
		Title:   normalizeSpace(title),
		Excerpt: normalizeSpace(excerpt),
	}

	// Create bookmark ID
	book.ID, err = DB.CreateNewID("bookmark")
	if err != nil {
		cError.Printf("Failed to create ID: %v\n", err)
		return
	}

	// Set bookmark tags
	book.Tags = make([]model.Tag, len(tags))
	for i, tag := range tags {
		book.Tags[i].Name = strings.TrimSpace(tag)
	}

	// If it's not offline mode, fetch data from internet
	var imageURLs []string

	if !offline {
		func() {
			cInfo.Println("Downloading article...")

			// Prepare download request
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				cError.Printf("Failed to download article: %v\n", err)
				return
			}

			// Send download request
			req.Header.Set("User-Agent", "Shiori/2.0.0 (+https://github.com/go-shiori/shiori)")
			resp, err := httpClient.Do(req)
			if err != nil {
				cError.Printf("Failed to download article: %v\n", err)
				return
			}
			defer resp.Body.Close()

			// Split response body so it can be processed twice
			archivalInput := bytes.NewBuffer(nil)
			readabilityInput := bytes.NewBuffer(nil)
			multiWriter := io.MultiWriter(archivalInput, readabilityInput)

			_, err = io.Copy(multiWriter, resp.Body)
			if err != nil {
				cError.Printf("Failed to process article: %v\n", err)
				return
			}

			// If this is HTML, parse for readable content
			contentType := resp.Header.Get("Content-Type")
			if strings.Contains(contentType, "text/html") {
				article, err := readability.FromReader(readabilityInput, url)
				if err != nil {
					cError.Printf("Failed to parse article: %v\n", err)
					return
				}

				book.Author = article.Byline
				book.Content = article.TextContent
				book.HTML = article.Content

				// If title and excerpt doesnt have submitted value, use from article
				if book.Title == "" {
					book.Title = article.Title
				}

				if book.Excerpt == "" {
					book.Excerpt = article.Excerpt
				}

				// Get image URL
				if article.Image != "" {
					imageURLs = append(imageURLs, article.Image)
				}

				if article.Favicon != "" {
					imageURLs = append(imageURLs, article.Favicon)
				}
			}

			// If needed, create offline archive as well
			if !noArchival {
				archivePath := fp.Join(DataDir, "archive", fmt.Sprintf("%d", book.ID))
				archivalRequest := warc.ArchivalRequest{
					URL:         url,
					Reader:      archivalInput,
					ContentType: contentType,
					LogEnabled:  logArchival,
				}

				err = warc.NewArchive(archivalRequest, archivePath)
				if err != nil {
					cError.Printf("Failed to create archive: %v\n", err)
					return
				}
			}
		}()
	}

	// Make sure title is not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Save bookmark to database
	_, err = DB.SaveBookmarks(book)
	if err != nil {
		cError.Printf("Failed to save bookmark: %v\n", err)
		return
	}

	// Save article image to local disk
	imgPath := fp.Join(DataDir, "thumb", fmt.Sprintf("%d", book.ID))
	for _, imageURL := range imageURLs {
		err = downloadBookImage(imageURL, imgPath, time.Minute)
		if err == nil {
			break
		} else {
			cError.Printf("Failed to download image: %v\n", err)
			continue
		}
	}

	// Print added bookmark
	fmt.Println()
	printBookmarks(book)
}
