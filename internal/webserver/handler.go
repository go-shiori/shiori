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

// validateSession checks whether user session is still valid or not
func (h *handler) validateSession(r *http.Request) error {
	// Get session-id from cookie
	sessionID, err := r.Cookie("session-id")
	if err != nil {
		if err == http.ErrNoCookie {
			return fmt.Errorf("session is not exist")
		}
		return err
	}

	// Make sure session is not expired yet
	if _, found := h.SessionCache.Get(sessionID.Value); !found {
		return fmt.Errorf("session has been expired")
	}

	return nil
}
