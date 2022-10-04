package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

func downloadBookmarkContent(book *model.Bookmark, dataDir string, request *http.Request) (*model.Bookmark, error) {
	content, contentType, err := core.DownloadBookmark(book.URL)
	if err != nil {
		return nil, fmt.Errorf("error downloading bookmark: %s", err)
	}

	processRequest := core.ProcessRequest{
		DataDir:     dataDir,
		Bookmark:    *book,
		Content:     content,
		ContentType: contentType,
	}

	result, isFatalErr, err := core.ProcessBookmark(processRequest)
	content.Close()

	if err != nil && isFatalErr {
		panic(fmt.Errorf("failed to process bookmark: %v", err))
	}

	return &result, err
}

// apiLogin is handler for POST /api/login
func (h *handler) apiLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Decode request
	request := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Remember bool   `json:"remember"`
		Owner    bool   `json:"owner"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Prepare function to generate session
	genSession := func(account model.Account, expTime time.Duration) {
		// Create session ID
		sessionID, err := uuid.NewV4()
		checkError(err)

		// Save session ID to cache
		strSessionID := sessionID.String()
		h.SessionCache.Set(strSessionID, account, expTime)

		// Save user's session IDs to cache as well
		// useful for mass logout
		sessionIDs := []string{strSessionID}
		if val, found := h.UserCache.Get(request.Username); found {
			sessionIDs = val.([]string)
			sessionIDs = append(sessionIDs, strSessionID)
		}
		h.UserCache.Set(request.Username, sessionIDs, -1)

		// Send login result
		account.Password = ""
		loginResult := struct {
			Session string        `json:"session"`
			Account model.Account `json:"account"`
			Expires string        `json:"expires"`
		}{strSessionID, account, time.Now().Add(expTime).Format(time.RFC1123)}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(&loginResult)
		checkError(err)
	}

	// Check if user's database is empty or there are no owner.
	// If yes, and user uses default account, let him in.
	searchOptions := database.GetAccountsOptions{
		Owner: true,
	}

	accounts, err := h.DB.GetAccounts(ctx, searchOptions)
	checkError(err)

	if len(accounts) == 0 && request.Username == "shiori" && request.Password == "gopher" {
		genSession(model.Account{
			Username: "shiori",
			Owner:    true,
		}, time.Hour)
		return
	}

	// Get account data from database
	account, exist, err := h.DB.GetAccount(ctx, request.Username)
	checkError(err)

	if !exist {
		panic(fmt.Errorf("username doesn't exist"))
	}

	// Compare password with database
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(request.Password))
	if err != nil {
		panic(fmt.Errorf("username and password don't match"))
	}

	// If login request is as owner, make sure this account is owner
	if request.Owner && !account.Owner {
		panic(fmt.Errorf("account level is not sufficient as owner"))
	}

	// Calculate expiration time
	expTime := time.Hour
	if request.Remember {
		expTime = time.Hour * 24 * 30
	}

	// Create session
	genSession(account, expTime)
}

// apiLogout is handler for POST /api/logout
func (h *handler) apiLogout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get session ID
	sessionID := h.getSessionID(r)
	if sessionID != "" {
		h.SessionCache.Delete(sessionID)
	}

	fmt.Fprint(w, 1)
}

// apiGetBookmarks is handler for GET /api/bookmarks
func (h *handler) apiGetBookmarks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Get URL queries
	keyword := r.URL.Query().Get("keyword")
	strPage := r.URL.Query().Get("page")
	strTags := r.URL.Query().Get("tags")
	strExcludedTags := r.URL.Query().Get("exclude")

	tags := strings.Split(strTags, ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{}
	}

	excludedTags := strings.Split(strExcludedTags, ",")
	if len(excludedTags) == 1 && excludedTags[0] == "" {
		excludedTags = []string{}
	}

	page, _ := strconv.Atoi(strPage)
	if page < 1 {
		page = 1
	}

	// Prepare filter for database
	searchOptions := database.GetBookmarksOptions{
		Tags:         tags,
		ExcludedTags: excludedTags,
		Keyword:      keyword,
		Limit:        30,
		Offset:       (page - 1) * 30,
		OrderMethod:  database.ByLastAdded,
	}

	// Calculate max page
	nBookmarks, err := h.DB.GetBookmarksCount(ctx, searchOptions)
	checkError(err)
	maxPage := int(math.Ceil(float64(nBookmarks) / 30))

	// Fetch all matching bookmarks
	bookmarks, err := h.DB.GetBookmarks(ctx, searchOptions)
	checkError(err)

	// Get image URL for each bookmark, and check if it has archive
	for i := range bookmarks {
		strID := strconv.Itoa(bookmarks[i].ID)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		archivePath := fp.Join(h.DataDir, "archive", strID)

		if fileExists(imgPath) {
			bookmarks[i].ImageURL = path.Join(h.RootPath, "bookmark", strID, "thumb")
		}

		if fileExists(archivePath) {
			bookmarks[i].HasArchive = true
		}
	}

	// Return JSON response
	resp := map[string]interface{}{
		"page":      page,
		"maxPage":   maxPage,
		"bookmarks": bookmarks,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&resp)
	checkError(err)
}

// apiGetTags is handler for GET /api/tags
func (h *handler) apiGetTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Fetch all tags
	tags, err := h.DB.GetTags(ctx)
	checkError(err)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&tags)
	checkError(err)
}

// apiRenameTag is handler for PUT /api/tag
func (h *handler) apiRenameTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	tag := model.Tag{}
	err = json.NewDecoder(r.Body).Decode(&tag)
	checkError(err)

	// Update name
	err = h.DB.RenameTag(ctx, tag.ID, tag.Name)
	checkError(err)

	fmt.Fprint(w, 1)
}

// Bookmark is the record for an URL.
type apiInsertBookmarkPayload struct {
	URL           string      `json:"url"`
	Title         string      `json:"title"`
	Excerpt       string      `json:"excerpt"`
	Tags          []model.Tag `json:"tags"`
	CreateArchive bool        `json:"createArchive"`
	MakePublic    int         `json:"public"`
	Async         bool        `json:"async"`
}

// newApiInsertBookmarkPayload
// Returns the payload struct with its defaults
func newAPIInsertBookmarkPayload() *apiInsertBookmarkPayload {
	return &apiInsertBookmarkPayload{
		CreateArchive: false,
		Async:         true,
	}
}

// apiInsertBookmark is handler for POST /api/bookmark
func (h *handler) apiInsertBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	payload := newAPIInsertBookmarkPayload()
	err = json.NewDecoder(r.Body).Decode(&payload)
	checkError(err)

	book := &model.Bookmark{
		URL:           payload.URL,
		Title:         payload.Title,
		Excerpt:       payload.Excerpt,
		Tags:          payload.Tags,
		Public:        payload.MakePublic,
		CreateArchive: payload.CreateArchive,
	}

	// Clean up bookmark URL
	book.URL, err = core.RemoveUTMParams(book.URL)
	if err != nil {
		panic(fmt.Errorf("failed to clean URL: %v", err))
	}

	// Make sure bookmark's title not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Save bookmark to database
	results, err := h.DB.SaveBookmarks(ctx, *book)
	if err != nil || len(results) == 0 {
		panic(fmt.Errorf("failed to save bookmark: %v", err))
	}

	if payload.Async {
		go func() {
			bookmark, err := downloadBookmarkContent(&results[0], h.DataDir, r)
			if err != nil {
				log.Printf("error downloading boorkmark: %s", err)
			}
			if _, err := h.DB.SaveBookmarks(ctx, *bookmark); err != nil {
				log.Printf("failed to save bookmark: %s", err)
			}
		}()
	} else {
		// Workaround. Download content after saving the bookmark so we have the proper database
		// id already set in the object regardless of the database engine.
		book, err = downloadBookmarkContent(book, h.DataDir, r)
		if err != nil {
			log.Printf("error downloading boorkmark: %s", err)
		}
		if _, err := h.DB.SaveBookmarks(ctx, *book); err != nil {
			log.Printf("failed to save bookmark: %s", err)
		}
	}

	// Return the new bookmark
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(results[0])
	checkError(err)
}

// apiDeleteBookmarks is handler for DELETE /api/bookmark
func (h *handler) apiDeleteBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	ids := []int{}
	err = json.NewDecoder(r.Body).Decode(&ids)
	checkError(err)

	// Delete bookmarks
	err = h.DB.DeleteBookmarks(ctx, ids...)
	checkError(err)

	// Delete thumbnail image and archives from local disk
	for _, id := range ids {
		strID := strconv.Itoa(id)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		archivePath := fp.Join(h.DataDir, "archive", strID)

		os.Remove(imgPath)
		os.Remove(archivePath)
	}

	fmt.Fprint(w, 1)
}

// apiUpdateBookmark is handler for PUT /api/bookmarks
func (h *handler) apiUpdateBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := model.Bookmark{}
	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Validate input
	if request.Title == "" {
		panic(fmt.Errorf("title must not empty"))
	}

	// Get existing bookmark from database
	filter := database.GetBookmarksOptions{
		IDs:         []int{request.ID},
		WithContent: true,
	}

	bookmarks, err := h.DB.GetBookmarks(ctx, filter)
	checkError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("no bookmark with matching ids"))
	}

	// Set new bookmark data
	book := bookmarks[0]
	book.URL = request.URL
	book.Title = request.Title
	book.Excerpt = request.Excerpt
	book.Public = request.Public

	// Clean up bookmark URL
	book.URL, err = core.RemoveUTMParams(book.URL)
	if err != nil {
		panic(fmt.Errorf("failed to clean URL: %v", err))
	}

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
	res, err := h.DB.SaveBookmarks(ctx, book)
	checkError(err)

	// Add thumbnail image to the saved bookmarks again
	newBook := res[0]
	newBook.ImageURL = request.ImageURL
	newBook.HasArchive = request.HasArchive

	// Return new saved result
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&newBook)
	checkError(err)
}

// apiUpdateCache is handler for PUT /api/cache
func (h *handler) apiUpdateCache(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := struct {
		IDs           []int `json:"ids"`
		KeepMetadata  bool  `json:"keepMetadata"`
		CreateArchive bool  `json:"createArchive"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Get existing bookmark from database
	filter := database.GetBookmarksOptions{
		IDs:         request.IDs,
		WithContent: true,
	}

	bookmarks, err := h.DB.GetBookmarks(ctx, filter)
	checkError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("no bookmark with matching ids"))
	}

	// For web interface, let's limit to max 20 IDs to update, and 5 for archival.
	// This is done to prevent the REST request from client took too long to finish.
	if len(bookmarks) > 20 {
		panic(fmt.Errorf("max 20 bookmarks to update"))
	} else if len(bookmarks) > 5 && request.CreateArchive {
		panic(fmt.Errorf("max 5 bookmarks to update with archival"))
	}

	// Fetch data from internet
	mx := sync.RWMutex{}
	wg := sync.WaitGroup{}
	chDone := make(chan struct{})
	chProblem := make(chan int, 10)
	semaphore := make(chan struct{}, 10)

	for i, book := range bookmarks {
		wg.Add(1)

		// Mark whether book will be archived
		book.CreateArchive = request.CreateArchive

		go func(i int, book model.Bookmark, keepMetadata bool) {
			// Make sure to finish the WG
			defer wg.Done()

			// Register goroutine to semaphore
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			// Download data from internet
			content, contentType, err := core.DownloadBookmark(book.URL)
			if err != nil {
				chProblem <- book.ID
				return
			}

			request := core.ProcessRequest{
				DataDir:     h.DataDir,
				Bookmark:    book,
				Content:     content,
				ContentType: contentType,
				KeepTitle:   keepMetadata,
				KeepExcerpt: keepMetadata,
			}

			book, _, err = core.ProcessBookmark(request)
			content.Close()

			if err != nil {
				chProblem <- book.ID
				return
			}

			// Update list of bookmarks
			mx.Lock()
			bookmarks[i] = book
			mx.Unlock()
		}(i, book, request.KeepMetadata)
	}

	// Receive all problematic bookmarks
	idWithProblems := []int{}
	go func() {
		for {
			select {
			case <-chDone:
				return
			case id := <-chProblem:
				idWithProblems = append(idWithProblems, id)
			}
		}
	}()

	// Wait until all download finished
	wg.Wait()
	close(chDone)

	// Update database
	_, err = h.DB.SaveBookmarks(ctx, bookmarks...)
	checkError(err)

	// Return new saved result
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&bookmarks)
	checkError(err)
}

// apiUpdateBookmarkTags is handler for PUT /api/bookmarks/tags
func (h *handler) apiUpdateBookmarkTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
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
	filter := database.GetBookmarksOptions{
		IDs:         request.IDs,
		WithContent: true,
	}

	bookmarks, err := h.DB.GetBookmarks(ctx, filter)
	checkError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("no bookmark with matching ids"))
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
	bookmarks, err = h.DB.SaveBookmarks(ctx, bookmarks...)
	checkError(err)

	// Get image URL for each bookmark
	for i := range bookmarks {
		strID := strconv.Itoa(bookmarks[i].ID)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		imgURL := path.Join(h.RootPath, "bookmark", strID, "thumb")

		if fileExists(imgPath) {
			bookmarks[i].ImageURL = imgURL
		}
	}

	// Return new saved result
	err = json.NewEncoder(w).Encode(&bookmarks)
	checkError(err)
}

// apiGetAccounts is handler for GET /api/accounts
func (h *handler) apiGetAccounts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Get list of usernames from database
	accounts, err := h.DB.GetAccounts(ctx, database.GetAccountsOptions{})
	checkError(err)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&accounts)
	checkError(err)
}

// apiInsertAccount is handler for POST /api/accounts
func (h *handler) apiInsertAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	var account model.Account
	err = json.NewDecoder(r.Body).Decode(&account)
	checkError(err)

	// Save account to database
	err = h.DB.SaveAccount(ctx, account)
	checkError(err)

	fmt.Fprint(w, 1)
}

// apiUpdateAccount is handler for PUT /api/accounts
func (h *handler) apiUpdateAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := struct {
		Username    string `json:"username"`
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
		Owner       bool   `json:"owner"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Get existing account data from database
	account, exist, err := h.DB.GetAccount(ctx, request.Username)
	checkError(err)

	if !exist {
		panic(fmt.Errorf("username doesn't exist"))
	}

	// Compare old password with database
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(request.OldPassword))
	if err != nil {
		panic(fmt.Errorf("old password doesn't match"))
	}

	// Save new password to database
	account.Password = request.NewPassword
	account.Owner = request.Owner
	err = h.DB.SaveAccount(ctx, account)
	checkError(err)

	// Delete user's sessions
	if val, found := h.UserCache.Get(request.Username); found {
		userSessions := val.([]string)
		for _, session := range userSessions {
			h.SessionCache.Delete(session)
		}

		h.UserCache.Delete(request.Username)
	}

	fmt.Fprint(w, 1)
}

// apiDeleteAccount is handler for DELETE /api/accounts
func (h *handler) apiDeleteAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	usernames := []string{}
	err = json.NewDecoder(r.Body).Decode(&usernames)
	checkError(err)

	// Delete accounts
	err = h.DB.DeleteAccounts(ctx, usernames...)
	checkError(err)

	// Delete user's sessions
	var userSessions []string
	for _, username := range usernames {
		if val, found := h.UserCache.Get(username); found {
			userSessions = val.([]string)
			for _, session := range userSessions {
				h.SessionCache.Delete(session)
			}

			h.UserCache.Delete(username)
		}
	}

	fmt.Fprint(w, 1)
}
