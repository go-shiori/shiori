package webserver

import (
	"fmt"
	"net/http"

	"github.com/go-shiori/shiori/internal/database"
	cch "github.com/patrickmn/go-cache"
)

var developmentMode = false

// Handler is handler for serving the web interface.
type handler struct {
	DB           database.DB
	DataDir      string
	UserCache    *cch.Cache
	SessionCache *cch.Cache
	ArchiveCache *cch.Cache
}

// prepareLoginCache prepares login cache for future use
func (h *handler) prepareLoginCache() {
	h.SessionCache.OnEvicted(func(key string, val interface{}) {
		username := val.(string)
		arr, found := h.UserCache.Get(username)
		if !found {
			return
		}

		sessionIDs := arr.([]string)
		for i := 0; i < len(sessionIDs); i++ {
			if sessionIDs[i] == key {
				sessionIDs = append(sessionIDs[:i], sessionIDs[i+1:]...)
				break
			}
		}

		h.UserCache.Set(username, sessionIDs, -1)
	})
}

func (h *handler) getSessionID(r *http.Request) string {
	// Get session-id from header and cookie
	headerSessionID := r.Header.Get("X-Session-Id")
	cookieSessionID := func() string {
		cookie, err := r.Cookie("session-id")
		if err != nil {
			return ""
		}

		return cookie.Value
	}()

	// Session ID in cookie is more priority than in header
	sessionID := headerSessionID
	if cookieSessionID != "" {
		sessionID = cookieSessionID
	}

	return sessionID
}

// validateSession checks whether user session is still valid or not
func (h *handler) validateSession(r *http.Request) error {
	sessionID := h.getSessionID(r)
	if sessionID == "" {
		return fmt.Errorf("session is not exist")
	}

	// Make sure session is not expired yet
	val, found := h.SessionCache.Get(sessionID)
	if !found {
		return fmt.Errorf("session has been expired")
	}

	// If this is not get request, make sure it's owner
	if r.Method != "" && r.Method != "GET" {
		if isOwner := val.(bool); !isOwner {
			return fmt.Errorf("account level is not sufficient")
		}
	}

	return nil
}
