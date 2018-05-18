package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	nurl "net/url"
	"strings"
	"time"

	"github.com/RadhiFadlillah/shiori/model"
	"github.com/RadhiFadlillah/shiori/readability"
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
	tags := strings.Fields(strTags)

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
	parsedURL, err := nurl.ParseRequestURI(book.URL)
	if err != nil || parsedURL.Host == "" {
		panic(fmt.Errorf("URL is not valid"))
	}

	// Clear UTM parameters from URL
	book.URL = clearUTMParams(parsedURL)

	// Fetch data from internet
	article, _ := readability.Parse(book.URL, 10*time.Second)

	book.ImageURL = article.Meta.Image
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
		book.Title = "Untitled"
	}

	// Save to database
	book.ID, err = h.db.CreateBookmark(book)
	checkError(err)

	// Return new saved result
	err = json.NewEncoder(w).Encode(&book)
	checkError(err)
}

// apiUpdateBookmark is handler for PUT /api/bookmark
func (h *webHandler) apiUpdateBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Get url queries
	_, dontOverwrite := r.URL.Query()["dont-overwrite"]

	// Decode request
	request := model.Bookmark{}
	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Make sure URL valid
	parsedURL, err := nurl.ParseRequestURI(request.URL)
	if err != nil || parsedURL.Host == "" {
		panic(fmt.Errorf("URL is not valid"))
	}

	// Clear UTM parameters from URL
	request.URL = clearUTMParams(parsedURL)

	// Get existing bookmark from database
	bookmarks, err := h.db.GetBookmarks(true, fmt.Sprintf("%d", request.ID))
	checkError(err)

	if len(bookmarks) == 0 {
		panic(fmt.Errorf("No bookmark with matching index"))
	}

	book := bookmarks[0]
	book.URL = request.URL

	// Fetch data from internet
	article, err := readability.Parse(book.URL, 10*time.Second)
	checkError(err)

	book.ImageURL = article.Meta.Image
	book.Author = article.Meta.Author
	book.MinReadTime = article.Meta.MinReadTime
	book.MaxReadTime = article.Meta.MaxReadTime
	book.Content = article.Content
	book.HTML = article.RawContent

	if !dontOverwrite {
		book.Title = article.Meta.Title
		book.Excerpt = article.Meta.Excerpt
	}

	// Check if user submit his own title or excerpt
	if request.Title != "" {
		book.Title = request.Title
	}

	if request.Excerpt != "" {
		book.Excerpt = request.Excerpt
	}

	// Make sure title is not empty
	if book.Title == "" {
		book.Title = "Untitled"
	}

	// Create new tags from request
	addedTags := make(map[string]struct{})
	deletedTags := make(map[string]struct{})
	for _, tag := range request.Tags {
		tagName := strings.ToLower(tag.Name)
		tagName = strings.TrimSpace(tagName)

		if strings.HasPrefix(tagName, "-") {
			tagName = strings.TrimPrefix(tagName, "-")
			deletedTags[tagName] = struct{}{}
		} else {
			addedTags[tagName] = struct{}{}
		}
	}

	newTags := []model.Tag{}
	for _, tag := range book.Tags {
		if _, isDeleted := deletedTags[tag.Name]; isDeleted {
			tag.Deleted = true
		}

		if _, alreadyExist := addedTags[tag.Name]; alreadyExist {
			delete(addedTags, tag.Name)
		}

		newTags = append(newTags, tag)
	}

	for tag := range addedTags {
		newTags = append(newTags, model.Tag{Name: tag})
	}

	book.Tags = newTags

	// Update database
	book.Modified = time.Now().UTC().Format("2006-01-02 15:04:05")
	res, err := h.db.UpdateBookmarks(book)
	checkError(err)

	// Return new saved result
	err = json.NewEncoder(w).Encode(&res[0])
	checkError(err)
}

// apiDeleteBookmarks is handler for DELETE /api/bookmark
func (h *webHandler) apiDeleteBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check token
	err := h.checkAPIToken(r)
	checkError(err)

	// Decode request
	indices := []string{}
	err = json.NewDecoder(r.Body).Decode(&indices)
	checkError(err)

	// Delete bookmarks
	err = h.db.DeleteBookmarks(indices...)
	checkError(err)

	fmt.Fprint(w, 1)
}

func clearUTMParams(url *nurl.URL) string {
	newQuery := nurl.Values{}
	for key, value := range url.Query() {
		if strings.HasPrefix(key, "utm_") {
			continue
		}

		newQuery[key] = value
	}

	url.RawQuery = newQuery.Encode()
	return url.String()
}
