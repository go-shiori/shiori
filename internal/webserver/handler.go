package webserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-shiori/shiori/internal/model"
	cch "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// Handler is Handler for serving the web interface.
type Handler struct {
	DB        model.DB
	DataDir   string
	RootPath  string
	UserCache *cch.Cache
	// SessionCache *cch.Cache
	ArchiveCache *cch.Cache
	Log          bool

	dependencies model.Dependencies
}

func (h *Handler) PrepareSessionCache() {
	// h.SessionCache.OnEvicted(func(key string, val interface{}) {
	// 	account := val.(*model.AccountDTO)
	// 	arr, found := h.UserCache.Get(account.Username)
	// 	if !found {
	// 		return
	// 	}

	// 	sessionIDs := arr.([]string)
	// 	for i := 0; i < len(sessionIDs); i++ {
	// 		if sessionIDs[i] == key {
	// 			sessionIDs = append(sessionIDs[:i], sessionIDs[i+1:]...)
	// 			break
	// 		}
	// 	}

	// 	h.UserCache.Set(account.Username, sessionIDs, -1)
	// })
}

// validateSession checks whether user session is still valid or not
func (h *Handler) validateSession(r *http.Request) error {
	authorization := r.Header.Get(model.AuthorizationHeader)
	if authorization == "" {
		// Get token from cookie
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			return fmt.Errorf("session is not exist")
		}

		authorization = tokenCookie.Value
	}

	var account *model.AccountDTO

	if authorization != "" {
		var err error

		authParts := strings.SplitN(authorization, " ", 2)
		if len(authParts) != 2 && authParts[0] != model.AuthorizationTokenType {
			return fmt.Errorf("session has been expired")
		}

		account, err = h.dependencies.Domains().Auth().CheckToken(r.Context(), authParts[1])
		if err != nil {
			return fmt.Errorf("session has been expired")
		}
	}

	if r.Method != "" && r.Method != "GET" && account.Owner != nil && !*account.Owner {
		return fmt.Errorf("account level is not sufficient")
	}

	h.dependencies.Logger().WithFields(logrus.Fields{
		"username": account.Username,
		"method":   r.Method,
		"path":     r.URL.Path,
	}).Info("allowing legacy api access using JWT token")

	return nil

}
