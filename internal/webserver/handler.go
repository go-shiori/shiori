package webserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
	cch "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Handler is Handler for serving the web interface.
type Handler struct {
	DB           database.DB
	DataDir      string
	RootPath     string
	UserCache    *cch.Cache
	SessionCache *cch.Cache
	ArchiveCache *cch.Cache
	Log          bool

	dependencies *dependencies.Dependencies
}

func (h *Handler) PrepareSessionCache() {
	h.SessionCache.OnEvicted(func(key string, val interface{}) {
		account := val.(model.Account)
		arr, found := h.UserCache.Get(account.Username)
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

		h.UserCache.Set(account.Username, sessionIDs, -1)
	})
}

func (h *Handler) GetSessionID(r *http.Request) string {
	// Try to get session ID from the header
	sessionID := r.Header.Get("X-Session-Id")

	// If not, try it from the cookie
	if sessionID == "" {
		cookie, err := r.Cookie("session-id")
		if err != nil {
			return ""
		}

		sessionID = cookie.Value
	}

	return sessionID
}

// validateSession checks whether user session is still valid or not
func (h *Handler) validateSession(r *http.Request) error {
	authorization := r.Header.Get(model.AuthorizationHeader)
	if authorization != "" {
		authParts := strings.SplitN(authorization, " ", 2)
		if len(authParts) != 2 && authParts[0] != model.AuthorizationTokenType {
			return fmt.Errorf("session has been expired")
		}

		account, err := h.dependencies.Domains.Auth.CheckToken(r.Context(), authParts[1])
		if err != nil {
			return fmt.Errorf("session has been expired")
		}

		if r.Method != "" && r.Method != "GET" && account.Owner != nil && !*account.Owner {
			return fmt.Errorf("account level is not sufficient")
		}

		h.dependencies.Log.WithFields(logrus.Fields{
			"username": account.Username,
			"method":   r.Method,
			"path":     r.URL.Path,
		}).Info("allowing legacy api access using JWT token")

		return nil
	}

	sessionID := h.GetSessionID(r)
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
		if account := val.(model.Account); !account.Owner {
			return fmt.Errorf("account level is not sufficient")
		}
	}

	return nil
}
