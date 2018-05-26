package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	nurl "net/url"
	"os"
	fp "path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	dt "github.com/RadhiFadlillah/shiori/database"
	"github.com/RadhiFadlillah/shiori/model"
	"github.com/RadhiFadlillah/shiori/readability"
	valid "github.com/asaskevich/govalidator"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

// cmdHandler is handler for all action in AccountCmd
type cmdHandler struct {
	db      dt.Database
	dataDir string
}

// addBookmark is handler for adding new bookmark
func (h *cmdHandler) addBookmark(cmd *cobra.Command, args []string) {
	// Read flag and arguments
	url := args[0]
	title, _ := cmd.Flags().GetString("title")
	excerpt, _ := cmd.Flags().GetString("excerpt")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	offline, _ := cmd.Flags().GetBool("offline")

	// Make sure URL valid
	parsedURL, err := nurl.Parse(url)
	if err != nil || !valid.IsRequestURL(url) {
		cError.Println("URL is not valid")
		return
	}

	// Clear UTM parameters from URL
	clearUTMParams(parsedURL)

	// Create bookmark item
	book := model.Bookmark{
		URL:     parsedURL.String(),
		Title:   normalizeSpace(title),
		Excerpt: normalizeSpace(excerpt),
	}

	// Get new bookmark id
	book.ID, err = h.db.GetNewID("bookmark")
	if err != nil {
		cError.Println(err)
		return
	}

	// Set bookmark tags
	book.Tags = make([]model.Tag, len(tags))
	for i, tag := range tags {
		book.Tags[i].Name = strings.TrimSpace(tag)
	}

	// If it's not offline mode, fetch data from internet
	if !offline {
		article, _ := readability.FromURL(parsedURL, 20*time.Second)

		book.Author = article.Meta.Author
		book.MinReadTime = article.Meta.MinReadTime
		book.MaxReadTime = article.Meta.MaxReadTime
		book.Content = article.Content
		book.HTML = article.RawContent

		// If title and excerpt doesnt have submitted value, use from article
		if book.Title == "" {
			book.Title = article.Meta.Title
		}

		if book.Excerpt == "" {
			book.Excerpt = article.Meta.Excerpt
		}

		// Save bookmark image to local disk
		imgPath := fp.Join(h.dataDir, "thumb", fmt.Sprintf("%d", book.ID))
		err = downloadFile(article.Meta.Image, imgPath, 20*time.Second)
		if err == nil {
			book.ImageURL = fmt.Sprintf("/thumb/%d", book.ID)
		}
	}

	// Make sure title is not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Save bookmark to database
	book.ID, err = h.db.CreateBookmark(book)
	if err != nil {
		cError.Println(err)
		return
	}

	printBookmarks(book)
}

// printBookmarks is handler for printing list of saved bookmarks
func (h *cmdHandler) printBookmarks(cmd *cobra.Command, args []string) {
	// Read flags
	useJSON, _ := cmd.Flags().GetBool("json")
	indexOnly, _ := cmd.Flags().GetBool("index-only")

	// Convert args to ids
	ids, err := parseIndexList(args)
	if err != nil {
		cError.Println(err)
		return
	}

	// Read bookmarks from database
	bookmarks, err := h.db.GetBookmarks(false, ids...)
	if err != nil {
		cError.Println(err)
		return
	}

	if len(bookmarks) == 0 {
		if len(args) > 0 {
			cError.Println("No matching index found")
		} else {
			cError.Println("No bookmarks saved yet")
		}
		return
	}

	// Print data
	if useJSON {
		bt, err := json.MarshalIndent(&bookmarks, "", "    ")
		if err != nil {
			cError.Println(err)
			return
		}

		fmt.Println(string(bt))
		return
	}

	if indexOnly {
		for _, bookmark := range bookmarks {
			fmt.Printf("%d ", bookmark.ID)
		}
		fmt.Println()
		return
	}

	printBookmarks(bookmarks...)
}

// searchBookmarks is handler for searching bookmarks with matching keyword or tags
func (h *cmdHandler) searchBookmarks(cmd *cobra.Command, args []string) {
	// Read flags
	tags, _ := cmd.Flags().GetStringSlice("tags")
	useJSON, _ := cmd.Flags().GetBool("json")
	indexOnly, _ := cmd.Flags().GetBool("index-only")

	// Fetch keyword
	keyword := ""
	if len(args) > 0 {
		keyword = args[0]
	}

	// Read bookmarks from database
	bookmarks, err := h.db.SearchBookmarks(false, keyword, tags...)
	if err != nil {
		cError.Println(err)
		return
	}

	if len(bookmarks) == 0 {
		cError.Println("No matching bookmarks found")
		return
	}

	// Print data
	if useJSON {
		bt, err := json.MarshalIndent(&bookmarks, "", "    ")
		if err != nil {
			cError.Println(err)
			return
		}

		fmt.Println(string(bt))
		return
	}

	if indexOnly {
		for _, bookmark := range bookmarks {
			fmt.Printf("%d ", bookmark.ID)
		}
		fmt.Println()
		return
	}

	printBookmarks(bookmarks...)
}

// updateBookmarks is handler for updating bookmarks
func (h *cmdHandler) updateBookmarks(cmd *cobra.Command, args []string) {
	// Parse flags
	url, _ := cmd.Flags().GetString("url")
	title, _ := cmd.Flags().GetString("title")
	excerpt, _ := cmd.Flags().GetString("excerpt")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	offline, _ := cmd.Flags().GetBool("offline")
	skipConfirm, _ := cmd.Flags().GetBool("yes")
	dontOverwrite := cmd.Flags().Changed("dont-overwrite")

	title = normalizeSpace(title)
	excerpt = normalizeSpace(excerpt)

	// Convert args to ids
	ids, err := parseIndexList(args)
	if err != nil {
		cError.Println(err)
		return
	}

	// Check if --url flag is used
	if cmd.Flags().Changed("url") {
		// Make sure URL is valid
		parsedURL, err := nurl.Parse(url)
		if err != nil || !valid.IsRequestURL(url) {
			cError.Println("URL is not valid")
			return
		}

		// Clear UTM parameters from URL
		clearUTMParams(parsedURL)
		url = parsedURL.String()

		// Make sure there is only one arguments
		if len(ids) != 1 {
			cError.Println("Update only accepts one index while using --url flag")
			return
		}
	}

	// If no arguments (i.e all bookmarks will be updated),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmUpdate := ""
		fmt.Print("Update ALL bookmarks? (y/n): ")
		fmt.Scanln(&confirmUpdate)

		if confirmUpdate != "y" {
			fmt.Println("No bookmarks updated")
			return
		}
	}

	// Prepare wait group
	wg := sync.WaitGroup{}

	// Fetch bookmarks from database
	bookmarks, err := h.db.GetBookmarks(true, ids...)
	if err != nil {
		cError.Println(err)
		return
	}

	if len(bookmarks) == 0 {
		cError.Println("No matching index found")
		return
	}

	// If not offline, fetch articles from internet
	if !offline {
		fmt.Println("Fetching new bookmarks data")

		// Start progress bar
		uiprogress.Start()
		bar := uiprogress.AddBar(len(bookmarks)).AppendCompleted().PrependElapsed()

		for i, book := range bookmarks {
			wg.Add(1)

			go func(pos int, book model.Bookmark) {
				defer func() {
					bar.Incr()
					wg.Done()
				}()

				// If used, use submitted URL
				if url != "" {
					book.URL = url
				}

				// Parse URL
				parsedURL, err := nurl.Parse(book.URL)
				if err != nil || !valid.IsRequestURL(book.URL) {
					return
				}

				// Fetch data from internet
				article, err := readability.FromURL(parsedURL, 20*time.Second)
				if err != nil {
					return
				}

				book.Author = article.Meta.Author
				book.MinReadTime = article.Meta.MinReadTime
				book.MaxReadTime = article.Meta.MaxReadTime
				book.Content = article.Content
				book.HTML = article.RawContent

				if !dontOverwrite {
					book.Title = article.Meta.Title
					book.Excerpt = article.Meta.Excerpt
				}

				// Update bookmark image in local disk
				imgPath := fp.Join(h.dataDir, "thumb", fmt.Sprintf("%d", book.ID))
				err = downloadFile(article.Meta.Image, imgPath, 20*time.Second)
				if err == nil {
					book.ImageURL = fmt.Sprintf("/thumb/%d", book.ID)
				}

				bookmarks[pos] = book
			}(i, book)
		}

		wg.Wait()
		uiprogress.Stop()
		fmt.Println("\nSaving new data")
	}

	// Map the tags to be added or deleted from flag --tags
	addedTags := make(map[string]struct{})
	deletedTags := make(map[string]struct{})
	for _, tag := range tags {
		tagName := strings.ToLower(tag)
		tagName = strings.TrimSpace(tagName)

		if strings.HasPrefix(tagName, "-") {
			tagName = strings.TrimPrefix(tagName, "-")
			deletedTags[tagName] = struct{}{}
		} else {
			addedTags[tagName] = struct{}{}
		}
	}

	// Set title, excerpt and tags from user submitted value
	for i, book := range bookmarks {
		// Check if user submit his own title or excerpt
		if title != "" {
			book.Title = title
		}

		if excerpt != "" {
			book.Excerpt = excerpt
		}

		// Make sure title is not empty
		if book.Title == "" {
			book.Title = book.URL
		}

		// Generate new tags
		tempAddedTags := make(map[string]struct{})
		for key, value := range addedTags {
			tempAddedTags[key] = value
		}

		newTags := []model.Tag{}
		for _, tag := range book.Tags {
			if _, isDeleted := deletedTags[tag.Name]; isDeleted {
				tag.Deleted = true
			}

			if _, alreadyExist := addedTags[tag.Name]; alreadyExist {
				delete(tempAddedTags, tag.Name)
			}

			newTags = append(newTags, tag)
		}

		for tag := range tempAddedTags {
			newTags = append(newTags, model.Tag{Name: tag})
		}

		book.Tags = newTags

		// Set bookmark new data
		bookmarks[i] = book
	}

	// Update database
	result, err := h.db.UpdateBookmarks(bookmarks...)
	if err != nil {
		cError.Println(err)
		return
	}

	// Print update result
	printBookmarks(result...)
}

// deleteBookmarks is handler for deleting bookmarks
func (h *cmdHandler) deleteBookmarks(cmd *cobra.Command, args []string) {
	// Parse flags
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	// If no arguments (i.e all bookmarks going to be deleted),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmDelete := ""
		fmt.Print("Remove ALL bookmarks? (y/n): ")
		fmt.Scanln(&confirmDelete)

		if confirmDelete != "y" {
			fmt.Println("No bookmarks deleted")
			return
		}
	}

	// Convert args to ids
	ids, err := parseIndexList(args)
	if err != nil {
		cError.Println(err)
		return
	}

	// Delete bookmarks from database
	err = h.db.DeleteBookmarks(ids...)
	if err != nil {
		cError.Println(err)
	}

	// Delete thumbnail image from local disk
	for _, id := range ids {
		imgPath := fp.Join(h.dataDir, "thumb", fmt.Sprintf("%d", id))
		os.Remove(imgPath)
	}

	fmt.Println("Bookmark(s) have been deleted")
}

// openBookmarks is handler for opening bookmarks
func (h *cmdHandler) openBookmarks(cmd *cobra.Command, args []string) {
	// Parse flags
	cacheMode, _ := cmd.Flags().GetBool("cache")
	trimSpace, _ := cmd.Flags().GetBool("trim-space")
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	// If no arguments (i.e all bookmarks will be opened),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmOpen := ""
		fmt.Print("Open ALL bookmarks? (y/n): ")
		fmt.Scanln(&confirmOpen)

		if confirmOpen != "y" {
			return
		}
	}

	// Convert args to ids
	ids, err := parseIndexList(args)
	if err != nil {
		cError.Println(err)
		return
	}

	// Fetch bookmarks from database
	bookmarks, err := h.db.GetBookmarks(true, ids...)
	if err != nil {
		cError.Println(err)
		return
	}

	if len(bookmarks) == 0 {
		if len(args) > 0 {
			cError.Println("No matching index found")
		} else {
			cError.Println("No saved bookmarks yet")
		}
		return
	}

	// If not cache mode, open bookmarks in browser
	if !cacheMode {
		for _, book := range bookmarks {
			err = openBrowser(book.URL)
			if err != nil {
				cError.Printf("Failed to open %s: %v\n", book.URL, err)
			}
		}
		return
	}

	// Show bookmarks content in terminal
	termWidth := getTerminalWidth()
	if termWidth < 50 {
		termWidth = 50
	}

	for _, book := range bookmarks {
		if trimSpace {
			words := strings.Fields(book.Content)
			book.Content = strings.Join(words, " ")
		}

		cIndex.Printf("%d. ", book.ID)
		cTitle.Println(book.Title)
		fmt.Println()

		if book.Content == "" {
			cError.Println("This bookmark doesn't have any cached content")
		} else {
			fmt.Println(book.Content)
		}

		fmt.Println()
		cSymbol.Println(strings.Repeat("-", termWidth))
		fmt.Println()
	}
}

// importBookmarks is handler for importing bookmarks.
// Accept exactly one argument, the file to be imported.
func (h *cmdHandler) importBookmarks(cmd *cobra.Command, args []string) {
	// Parse flags
	generateTag := cmd.Flags().Changed("generate-tag")

	// If user doesn't specify, ask if tag need to be generated
	if !generateTag {
		var submit string
		fmt.Print("Add parents folder as tag? (y/n): ")
		fmt.Scanln(&submit)

		generateTag = submit == "y"
	}

	// Open bookmark's file
	srcFile, err := os.Open(args[0])
	if err != nil {
		cError.Println(err)
		return
	}
	defer srcFile.Close()

	// Parse bookmark's file
	doc, err := goquery.NewDocumentFromReader(srcFile)
	if err != nil {
		cError.Println(err)
		return
	}

	bookmarks := []model.Bookmark{}
	doc.Find("dt>a").Each(func(_ int, a *goquery.Selection) {
		// Get related elements
		dt := a.Parent()
		dl := dt.Parent()

		// Get metadata
		title := a.Text()
		url, _ := a.Attr("href")
		strTags, _ := a.Attr("tags")
		strModified, _ := a.Attr("last_modified")
		intModified, _ := strconv.ParseInt(strModified, 10, 64)
		modified := time.Unix(intModified, 0)

		// Get bookmark tags
		tags := []model.Tag{}
		for _, strTag := range strings.Split(strTags, ",") {
			if strTag != "" {
				tags = append(tags, model.Tag{Name: strTag})
			}
		}

		// Get bookmark excerpt
		excerpt := ""
		if dd := dt.Next(); dd.Is("dd") {
			excerpt = dd.Text()
		}

		// Get category name for this bookmark
		// and add it as tags (if necessary)
		category := ""
		if dtCategory := dl.Prev(); dtCategory.Is("h3") {
			category = dtCategory.Text()
			category = normalizeSpace(category)
			category = strings.ToLower(category)
			category = strings.Replace(category, " ", "-", -1)
		}

		if category != "" && generateTag {
			tags = append(tags, model.Tag{Name: category})
		}

		// Add item to list
		bookmark := model.Bookmark{
			URL:      url,
			Title:    normalizeSpace(title),
			Excerpt:  normalizeSpace(excerpt),
			Modified: modified.Format("2006-01-02 15:04:05"),
			Tags:     tags,
		}

		bookmarks = append(bookmarks, bookmark)
	})

	// Save bookmarks to database
	for _, book := range bookmarks {
		// Make sure URL valid
		parsedURL, err := nurl.Parse(book.URL)
		if err != nil || !valid.IsRequestURL(book.URL) {
			cError.Println("URL is not valid")
			continue
		}

		// Clear UTM parameters from URL
		clearUTMParams(parsedURL)
		book.URL = parsedURL.String()

		// Save book to database
		book.ID, err = h.db.CreateBookmark(book)
		if err != nil {
			cError.Println(err)
			continue
		}

		printBookmarks(book)
	}
}

// exportBookmarks is handler for exporting bookmarks.
// Accept exactly one argument, the file to be exported.
func (h *cmdHandler) exportBookmarks(cmd *cobra.Command, args []string) {
	// Fetch bookmarks from database
	bookmarks, err := h.db.GetBookmarks(false)
	if err != nil {
		cError.Println(err)
		return
	}

	if len(bookmarks) == 0 {
		cError.Println("No saved bookmarks yet")
		return
	}

	// Make sure destination directory exist
	dstDir := fp.Dir(args[0])
	os.MkdirAll(dstDir, os.ModePerm)

	// Open destination file
	dstFile, err := os.Create(args[0])
	if err != nil {
		cError.Println(err)
		return
	}
	defer dstFile.Close()

	// Create template
	funcMap := template.FuncMap{
		"unix": func(str string) int64 {
			t, err := time.Parse("2006-01-02 15:04:05", str)
			if err != nil {
				return time.Now().Unix()
			}

			return t.Unix()
		},
		"combine": func(tags []model.Tag) string {
			strTags := make([]string, len(tags))
			for i, tag := range tags {
				strTags[i] = tag.Name
			}

			return strings.Join(strTags, ",")
		},
	}

	tplContent := `<!DOCTYPE NETSCAPE-Bookmark-file-1>` +
		`<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">` +
		`<TITLE>Bookmarks</TITLE>` +
		`<H1>Bookmarks</H1>` +
		`<DL><p>` +
		`{{range $book := .}}` +
		`<DT><A HREF="{{$book.URL}}" ADD_DATE="{{unix $book.Modified}}" TAGS="{{combine $book.Tags}}">{{$book.Title}}</A>` +
		`{{if gt (len $book.Excerpt) 0}}<DD>{{$book.Excerpt}}{{end}}{{end}}` +
		`</DL><p>`

	tpl, err := template.New("export").Funcs(funcMap).Parse(tplContent)
	if err != nil {
		cError.Println(err)
		return
	}

	// Execute template
	err = tpl.Execute(dstFile, &bookmarks)
	if err != nil {
		cError.Println(err)
		return
	}

	fmt.Println("Export finished")
}

func printBookmarks(bookmarks ...model.Bookmark) {
	for _, bookmark := range bookmarks {
		// Create bookmark index
		strBookmarkIndex := fmt.Sprintf("%d. ", bookmark.ID)
		strSpace := strings.Repeat(" ", len(strBookmarkIndex))

		// Print bookmark title
		cIndex.Print(strBookmarkIndex)
		cTitle.Print(bookmark.Title)

		// Print read time
		if bookmark.MinReadTime > 0 {
			readTime := fmt.Sprintf(" (%d-%d minutes)", bookmark.MinReadTime, bookmark.MaxReadTime)
			if bookmark.MinReadTime == bookmark.MaxReadTime {
				readTime = fmt.Sprintf(" (%d minutes)", bookmark.MinReadTime)
			}
			cReadTime.Println(readTime)
		} else {
			fmt.Println()
		}

		// Print bookmark URL
		cSymbol.Print(strSpace + "> ")
		cURL.Println(bookmark.URL)

		// Print bookmark excerpt
		if bookmark.Excerpt != "" {
			cSymbol.Print(strSpace + "+ ")
			cExcerpt.Println(bookmark.Excerpt)
		}

		// Print bookmark tags
		if len(bookmark.Tags) > 0 {
			cSymbol.Print(strSpace + "# ")
			for i, tag := range bookmark.Tags {
				if i == len(bookmark.Tags)-1 {
					cTag.Println(tag.Name)
				} else {
					cTag.Print(tag.Name + ", ")
				}
			}
		}

		// Append new line
		fmt.Println()
	}
}
