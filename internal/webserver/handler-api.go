package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	nurl "net/url"
	"os"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-shiori/go-readability"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/pkg/warc"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

// apiLogin is handler for POST /api/login
func (h *handler) apiLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Decode request
	var request model.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Prepare function to generate session
	genSession := func(expTime time.Duration) {
		// Create session ID
		sessionID, err := uuid.NewV4()
		checkError(err)

		// Save session ID to cache
		strSessionID := sessionID.String()
		h.SessionCache.Set(strSessionID, request.Username, expTime)

		// Save user's session IDs to cache as well
		// useful for mass logout
		sessionIDs := []string{strSessionID}
		if val, found := h.UserCache.Get(request.Username); found {
			sessionIDs = val.([]string)
			sessionIDs = append(sessionIDs, strSessionID)
		}
		h.UserCache.Set(request.Username, sessionIDs, -1)

		// Return session ID to user in cookies
		http.SetCookie(w, &http.Cookie{
			Name:    "session-id",
			Value:   strSessionID,
			Path:    "/",
			Expires: time.Now().Add(expTime),
		})

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, strSessionID)
	}

	// Check if user's database is empty.
	// If database still empty, and user uses default account, let him in.
	accounts, err := h.DB.GetAccounts("")
	checkError(err)

	if len(accounts) == 0 && request.Username == "shiori" && request.Password == "gopher" {
		genSession(time.Hour)
		return
	}

	// Get account data from database
	account, exist := h.DB.GetAccount(request.Username)
	if !exist {
		panic(fmt.Errorf("username doesn't exist"))
	}

	// Compare password with database
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(request.Password))
	if err != nil {
		panic(fmt.Errorf("username and password don't match"))
	}

	// Calculate expiration time
	expTime := time.Hour
	if request.Remember > 0 {
		expTime = time.Duration(request.Remember) * time.Hour
	} else {
		expTime = -1
	}

	// Create session
	genSession(expTime)
}

// apiLogout is handler for POST /api/logout
func (h *handler) apiLogout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get session ID
	sessionID, err := r.Cookie("session-id")
	if err != nil {
		if err == http.ErrNoCookie {
			panic(fmt.Errorf("session is expired"))
		} else {
			panic(err)
		}
	}

	h.SessionCache.Delete(sessionID.Value)
	fmt.Fprint(w, 1)
}

// apiGetBookmarks is handler for GET /api/bookmarks
func (h *handler) apiGetBookmarks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Get URL queries
	keyword := r.URL.Query().Get("keyword")
	strTags := r.URL.Query().Get("tags")
	strPage := r.URL.Query().Get("page")

	tags := strings.Split(strTags, ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{}
	}

	page, _ := strconv.Atoi(strPage)
	if page < 1 {
		page = 1
	}

	// Prepare filter for database
	searchOptions := database.GetBookmarksOptions{
		Tags:        tags,
		Keyword:     keyword,
		Limit:       30,
		Offset:      (page - 1) * 30,
		OrderMethod: database.ByLastAdded,
	}

	// Calculate max page
	nBookmarks, err := h.DB.GetBookmarksCount(searchOptions)
	checkError(err)
	maxPage := int(math.Ceil(float64(nBookmarks) / 30))

	// Fetch all matching bookmarks
	bookmarks, err := h.DB.GetBookmarks(searchOptions)
	checkError(err)

	// Get image URL for each bookmark, and check if it has archive
	for i := range bookmarks {
		strID := strconv.Itoa(bookmarks[i].ID)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		archivePath := fp.Join(h.DataDir, "archive", strID)

		if fileExists(imgPath) {
			bookmarks[i].ImageURL = path.Join("/", "bookmark", strID, "thumb")
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
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Fetch all tags
	tags, err := h.DB.GetTags()
	checkError(err)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&tags)
	checkError(err)
}

// apiInsertBookmark is handler for POST /api/bookmark
func (h *handler) apiInsertBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	book := model.Bookmark{}
	err = json.NewDecoder(r.Body).Decode(&book)
	checkError(err)

	// Clean up URL by removing its fragment and UTM parameters
	tmp, err := nurl.Parse(book.URL)
	if err != nil || tmp.Scheme == "" || tmp.Hostname() == "" {
		panic(fmt.Errorf("URL is not valid"))
	}

	tmp.Fragment = ""
	clearUTMParams(tmp)
	book.URL = tmp.String()

	// Create bookmark ID
	book.ID, err = h.DB.CreateNewID("bookmark")
	if err != nil {
		panic(fmt.Errorf("failed to create ID: %v", err))
	}

	// Fetch data from internet
	var imageURLs []string
	func() {
		// Prepare download request
		req, err := http.NewRequest("GET", book.URL, nil)
		if err != nil {
			return
		}

		// Send download request
		req.Header.Set("User-Agent", "Shiori/2.0.0 (+https://github.com/go-shiori/shiori)")
		resp, err := httpClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// Split response body so it can be processed twice
		archivalInput := bytes.NewBuffer(nil)
		readabilityInput := bytes.NewBuffer(nil)
		readabilityCheckInput := bytes.NewBuffer(nil)
		multiWriter := io.MultiWriter(archivalInput, readabilityInput, readabilityCheckInput)

		_, err = io.Copy(multiWriter, resp.Body)
		if err != nil {
			return
		}

		// If this is HTML, parse for readable content
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			isReadable := readability.IsReadable(readabilityCheckInput)

			article, err := readability.FromReader(readabilityInput, book.URL)
			if err != nil {
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

			if !isReadable {
				book.Content = ""
			}
		}

		// If needed, create offline archive as well
		if book.CreateArchive {
			archivePath := fp.Join(h.DataDir, "archive", fmt.Sprintf("%d", book.ID))
			archivalRequest := warc.ArchivalRequest{
				URL:         book.URL,
				Reader:      archivalInput,
				ContentType: contentType,
			}

			err = warc.NewArchive(archivalRequest, archivePath)
			if err != nil {
				return
			}
		}
	}()

	// Make sure title is not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Save bookmark to database
	results, err := h.DB.SaveBookmarks(book)
	if err != nil || len(results) == 0 {
		panic(fmt.Errorf("failed to save bookmark: %v", err))
	}
	book = results[0]

	// Save article image to local disk
	strID := strconv.Itoa(book.ID)
	imgPath := fp.Join(h.DataDir, "thumb", strID)
	for _, imageURL := range imageURLs {
		err = downloadBookImage(imageURL, imgPath, time.Minute)
		if err == nil {
			book.ImageURL = path.Join("/", "bookmark", strID, "thumb")
			break
		}
	}

	// Return the new bookmark
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&book)
	checkError(err)
}

// apiDeleteBookmarks is handler for DELETE /api/bookmark
func (h *handler) apiDeleteBookmark(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	ids := []int{}
	err = json.NewDecoder(r.Body).Decode(&ids)
	checkError(err)

	// Delete bookmarks
	err = h.DB.DeleteBookmarks(ids...)
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
	// Make sure session still valid
	err := h.validateSession(r)
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
	filter := database.GetBookmarksOptions{
		IDs:         []int{request.ID},
		WithContent: true,
	}

	bookmarks, err := h.DB.GetBookmarks(filter)
	checkError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("no bookmark with matching ids"))
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
	res, err := h.DB.SaveBookmarks(book)
	checkError(err)

	// Add thumbnail image to the saved bookmarks again
	newBook := res[0]
	newBook.ImageURL = request.ImageURL

	// Return new saved result
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&newBook)
	checkError(err)
}

// apiUpdateCache is handler for PUT /api/cache
func (h *handler) apiUpdateCache(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := struct {
		IDs           []int `json:"ids"`
		CreateArchive bool  `json:"createArchive"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Get existing bookmark from database
	filter := database.GetBookmarksOptions{
		IDs:         request.IDs,
		WithContent: true,
	}

	bookmarks, err := h.DB.GetBookmarks(filter)
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

		go func(i int, book model.Bookmark) {
			// Make sure to finish the WG
			defer wg.Done()

			// Register goroutine to semaphore
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			// Prepare download request
			req, err := http.NewRequest("GET", book.URL, nil)
			if err != nil {
				chProblem <- book.ID
				return
			}

			// Send download request
			req.Header.Set("User-Agent", "Shiori/2.0.0 (+https://github.com/go-shiori/shiori)")
			resp, err := httpClient.Do(req)
			if err != nil {
				chProblem <- book.ID
				return
			}
			defer resp.Body.Close()

			// Split response body so it can be processed twice
			archivalInput := bytes.NewBuffer(nil)
			readabilityInput := bytes.NewBuffer(nil)
			readabilityCheckInput := bytes.NewBuffer(nil)
			multiWriter := io.MultiWriter(archivalInput, readabilityInput, readabilityCheckInput)

			_, err = io.Copy(multiWriter, resp.Body)
			if err != nil {
				chProblem <- book.ID
				return
			}

			// If this is HTML, parse for readable content
			strID := strconv.Itoa(book.ID)
			contentType := resp.Header.Get("Content-Type")

			if strings.Contains(contentType, "text/html") {
				isReadable := readability.IsReadable(readabilityCheckInput)

				article, err := readability.FromReader(readabilityInput, book.URL)
				if err != nil {
					chProblem <- book.ID
					return
				}

				book.Author = article.Byline
				book.Content = article.TextContent
				book.HTML = article.Content

				if article.Title != "" {
					book.Title = article.Title
				}

				if article.Excerpt != "" {
					book.Excerpt = article.Excerpt
				}

				if !isReadable {
					book.Content = ""
				}

				// Get image for thumbnail and save it to local disk
				var imageURLs []string
				if article.Image != "" {
					imageURLs = append(imageURLs, article.Image)
				}

				if article.Favicon != "" {
					imageURLs = append(imageURLs, article.Favicon)
				}

				// Save article image to local disk
				imgPath := fp.Join(h.DataDir, "thumb", strID)
				for _, imageURL := range imageURLs {
					err = downloadBookImage(imageURL, imgPath, time.Minute)
					if err == nil {
						book.ImageURL = path.Join("/", "bookmark", strID, "thumb")
						break
					}
				}
			}

			// If needed, update offline archive as well.
			// Make sure to delete the old one first.
			if request.CreateArchive {
				archivePath := fp.Join(h.DataDir, "archive", strID)
				os.Remove(archivePath)

				archivalRequest := warc.ArchivalRequest{
					URL:         book.URL,
					Reader:      archivalInput,
					ContentType: contentType,
				}

				err = warc.NewArchive(archivalRequest, archivePath)
				if err != nil {
					chProblem <- book.ID
					return
				}
			}

			// Update list of bookmarks
			mx.Lock()
			bookmarks[i] = book
			mx.Unlock()
		}(i, book)
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
	_, err = h.DB.SaveBookmarks(bookmarks...)
	checkError(err)

	// Return new saved result
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&bookmarks)
	checkError(err)
}

// apiUpdateBookmarkTags is handler for PUT /api/bookmarks/tags
func (h *handler) apiUpdateBookmarkTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	bookmarks, err := h.DB.GetBookmarks(filter)
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
	bookmarks, err = h.DB.SaveBookmarks(bookmarks...)
	checkError(err)

	// Get image URL for each bookmark
	for i := range bookmarks {
		strID := strconv.Itoa(bookmarks[i].ID)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		imgURL := path.Join("/", "bookmark", strID, "thumb")

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
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Get list of usernames from database
	accounts, err := h.DB.GetAccounts("")
	checkError(err)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&accounts)
	checkError(err)
}

// apiInsertAccount is handler for POST /api/accounts
func (h *handler) apiInsertAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	var account model.Account
	err = json.NewDecoder(r.Body).Decode(&account)
	checkError(err)

	// Save account to database
	err = h.DB.SaveAccount(account.Username, account.Password)
	checkError(err)

	fmt.Fprint(w, 1)
}

// apiUpdateAccount is handler for PUT /api/accounts
func (h *handler) apiUpdateAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	request := struct {
		Username    string `json:"username"`
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&request)
	checkError(err)

	// Get existing account data from database
	account, exist := h.DB.GetAccount(request.Username)
	if !exist {
		panic(fmt.Errorf("username doesn't exist"))
	}

	// Compare old password with database
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(request.OldPassword))
	if err != nil {
		panic(fmt.Errorf("old password doesn't match"))
	}

	// Save new password to database
	err = h.DB.SaveAccount(request.Username, request.NewPassword)
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
	// Make sure session still valid
	err := h.validateSession(r)
	checkError(err)

	// Decode request
	usernames := []string{}
	err = json.NewDecoder(r.Body).Decode(&usernames)
	checkError(err)

	// Delete accounts
	err = h.DB.DeleteAccounts(usernames...)
	checkError(err)

	// Delete user's sessions
	userSessions := []string{}
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
