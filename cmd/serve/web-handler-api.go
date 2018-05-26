package serve

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"os"
	fp "path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/RadhiFadlillah/shiori/model"
	"github.com/RadhiFadlillah/shiori/readability"
	valid "github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

// login is handler for POST /api/login
func (h *webHandler) apiLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Decode request
	var request model.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Get account data from database
	account, err := h.db.GetAccount(request.Username)
	checkError(err)

	// Compare password with database
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(request.Password))
	if err != nil {
		panic(fmt.Errorf("Username and password don't match"))
	}

	// Calculate expiration time
	nbf := time.Now()
	exp := time.Now().Add(12 * time.Hour)
	if request.Remember {
		exp = time.Now().Add(7 * 24 * time.Hour)
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf": nbf.Unix(),
		"exp": exp.Unix(),
		"sub": account.ID,
	})

	tokenString, err := token.SignedString(h.jwtKey)
	checkError(err)

	// Return token
	fmt.Fprint(w, tokenString)
}

// apiGetBookmarks is handler for GET /api/bookmarks
func (h *webHandler) apiGetBookmarks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Get URL queries
	keyword := r.URL.Query().Get("keyword")
	strTags := r.URL.Query().Get("tags")
	tags := strings.Split(strTags, ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{}
	}

	// Fetch all matching bookmarks
	bookmarks, err := h.db.SearchBookmarks(true, keyword, tags...)
	checkError(err)

	err = json.NewEncoder(w).Encode(&bookmarks)
	checkError(err)
}

// apiGetTags is handler for GET /api/tags
func (h *webHandler) apiGetTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Fetch all tags
	tags, err := h.db.GetTags()
	checkError(err)

	err = json.NewEncoder(w).Encode(&tags)
	checkError(err)
}

// apiInsertBookmark is handler for POST /api/bookmark
func (h *webHandler) apiInsertBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Decode request
	book := model.Bookmark{}
	err = json.NewDecoder(r.Body).Decode(&book)
	checkError(err)

	// Make sure URL valid
	parsedURL, err := nurl.Parse(book.URL)
	if err != nil || !valid.IsRequestURL(book.URL) {
		panic(fmt.Errorf("URL is not valid"))
	}

	// Clear UTM parameter from URL
	clearUTMParams(parsedURL)
	book.URL = parsedURL.String()

	// Get new bookmark id
	book.ID, err = h.db.GetNewID("bookmark")
	checkError(err)

	// Fetch data from internet
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

	// Make sure title is not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Check if book has content
	if book.Content != "" {
		book.HasContent = true
	}

	// Save bookmark image to local disk
	imgPath := fp.Join(h.dataDir, "thumb", fmt.Sprintf("%d", book.ID))
	err = downloadFile(article.Meta.Image, imgPath, 20*time.Second)
	if err == nil {
		book.ImageURL = fmt.Sprintf("/thumb/%d", book.ID)
	}

	// Save bookmark to database
	_, err = h.db.CreateBookmark(book)
	checkError(err)

	// Return new saved result
	err = json.NewEncoder(w).Encode(&book)
	checkError(err)
}

// apiDeleteBookmarks is handler for DELETE /api/bookmark
func (h *webHandler) apiDeleteBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Decode request
	ids := []int{}
	err = json.NewDecoder(r.Body).Decode(&ids)
	checkError(err)

	// Delete bookmarks
	err = h.db.DeleteBookmarks(ids...)
	checkError(err)

	// Delete thumbnail image from local disk
	for _, id := range ids {
		imgPath := fp.Join(h.dataDir, "thumb", fmt.Sprintf("%d", id))
		os.Remove(imgPath)
	}

	fmt.Fprint(w, 1)
}

// apiUpdateBookmark is handler for PUT /api/bookmarks
func (h *webHandler) apiUpdateBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Decode request
	request := model.Bookmark{}
	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Validate input
	if request.Title == "" {
		panic(fmt.Errorf("Title must not empty"))
	}

	// Get existing bookmark from database
	bookmarks, err := h.db.GetBookmarks(true, request.ID)
	checkError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("No bookmark with matching index"))
	}

	// Set new bookmark data
	book := bookmarks[0]
	book.Title = request.Title
	book.Excerpt = request.Excerpt

	// Set new tags
	for i := range book.Tags {
		book.Tags[i].Deleted = true
	}

	for _, newTag := range request.Tags {
		for i, oldTag := range book.Tags {
			if newTag.Name == oldTag.Name {
				newTag.ID = oldTag.ID
				book.Tags[i].Deleted = false
				break
			}
		}

		if newTag.ID == 0 {
			book.Tags = append(book.Tags, newTag)
		}
	}

	// Update database
	res, err := h.db.UpdateBookmarks(book)
	checkError(err)

	// Return new saved result
	err = json.NewEncoder(w).Encode(&res[0])
	checkError(err)
}

// apiUpdateBookmarkTags is handler for PUT /api/bookmarks/tags
func (h *webHandler) apiUpdateBookmarkTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Decode request
	request := struct {
		IDs  []int       `json:"ids"`
		Tags []model.Tag `json:"tags"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Validate input
	if len(request.IDs) == 0 || len(request.Tags) == 0 {
		panic(fmt.Errorf("IDs and tags must not empty"))
	}

	// Get existing bookmark from database
	bookmarks, err := h.db.GetBookmarks(true, request.IDs...)
	checkError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("No bookmark with matching index"))
	}

	// Set new tags
	for i, book := range bookmarks {
		for _, newTag := range request.Tags {
			for _, oldTag := range book.Tags {
				if newTag.Name == oldTag.Name {
					newTag.ID = oldTag.ID
					break
				}
			}

			if newTag.ID == 0 {
				book.Tags = append(book.Tags, newTag)
			}
		}

		bookmarks[i] = book
	}

	// Update database
	res, err := h.db.UpdateBookmarks(bookmarks...)
	checkError(err)

	// Return new saved result
	err = json.NewEncoder(w).Encode(&res)
	checkError(err)
}

// apiUpdateCache is handler for PUT /api/cache
func (h *webHandler) apiUpdateCache(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Decode request
	ids := []int{}
	err = json.NewDecoder(r.Body).Decode(&ids)
	checkError(err)

	// Prepare wait group
	wg := sync.WaitGroup{}

	// Fetch bookmarks from database
	books, err := h.db.GetBookmarks(false, ids...)
	checkError(err)

	// Download new cache data
	for i, book := range books {
		wg.Add(1)

		go func(pos int, book model.Bookmark) {
			defer wg.Done()

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

			book.Excerpt = article.Meta.Excerpt
			book.Author = article.Meta.Author
			book.MinReadTime = article.Meta.MinReadTime
			book.MaxReadTime = article.Meta.MaxReadTime
			book.Content = article.Content
			book.HTML = article.RawContent

			// Make sure title is not empty
			if article.Meta.Title != "" {
				book.Title = article.Meta.Title
			}

			// Check if book has content
			if book.Content != "" {
				book.HasContent = true
			}

			// Update bookmark image in local disk
			imgPath := fp.Join(h.dataDir, "thumb", fmt.Sprintf("%d", book.ID))
			err = downloadFile(article.Meta.Image, imgPath, 20*time.Second)
			if err == nil {
				book.ImageURL = fmt.Sprintf("/thumb/%d", book.ID)
			}

			books[pos] = book
		}(i, book)
	}

	// Wait until all finished
	wg.Wait()

	// Update database
	res, err := h.db.UpdateBookmarks(books...)
	checkError(err)

	// Return new saved result
	err = json.NewEncoder(w).Encode(&res)
	checkError(err)
}

func downloadFile(url, dstPath string, timeout time.Duration) error {
	// Fetch data from URL
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure destination directory exist
	err = os.MkdirAll(fp.Dir(dstPath), os.ModePerm)
	if err != nil {
		return err
	}

	// Create destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Write response body to the file
	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func clearUTMParams(url *nurl.URL) {
	newQuery := nurl.Values{}
	for key, value := range url.Query() {
		if !strings.HasPrefix(key, "utm_") {
			newQuery[key] = value
		}
	}

	url.RawQuery = newQuery.Encode()
}
