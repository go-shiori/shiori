package webserver

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path"
	fp "path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
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
		OrderLatest: true,
	}

	// Calculate max page
	nBookmarks, err := h.DB.GetBookmarksCount(searchOptions)
	checkError(err)
	maxPage := int(math.Ceil(float64(nBookmarks) / 30))

	// Fetch all matching bookmarks
	bookmarks, err := h.DB.GetBookmarks(searchOptions)
	checkError(err)

	// Get image URL for each bookmark
	for i := range bookmarks {
		strID := strconv.Itoa(bookmarks[i].ID)
		imgPath := fp.Join(h.DataDir, "thumb", strID)
		imgURL := path.Join("/", "thumb", strID)

		if fileExists(imgPath) {
			bookmarks[i].ImageURL = imgURL
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
