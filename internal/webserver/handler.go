package webserver

import (
	"errors"
	"fmt"
	"net"
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
	trustedIPs   []*net.IPNet
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
	var account *model.AccountDTO
	var err error

	if h.dependencies.Config().Http.SSOProxyAuth {
		account, err = h.ssoAccount(r)
		if err != nil {
			h.dependencies.Logger().WithError(err).Error("getting sso account")
		}
	}

	if account == nil {
		account, err = h.tokenAccount(r)
		if err != nil {
			return err
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

func (h *Handler) tokenAccount(r *http.Request) (*model.AccountDTO, error) {
	authorization := r.Header.Get(model.AuthorizationHeader)
	if authorization == "" {
		// Get token from cookie
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			return nil, fmt.Errorf("session is not exist")
		}

		authorization = tokenCookie.Value
	}

	if authorization != "" {
		authParts := strings.SplitN(authorization, " ", 2)
		if len(authParts) != 2 || authParts[0] != model.AuthorizationTokenType {
			return nil, fmt.Errorf("session has been expired")
		}

		account, err := h.dependencies.Domains().Auth().CheckToken(r.Context(), authParts[1])
		if err != nil {
			return nil, fmt.Errorf("session has been expired")
		}

		return account, nil
	}

	return nil, errors.New("session has been expired")
}

func (h *Handler) ssoAccount(r *http.Request) (*model.AccountDTO, error) {
	remoteAddr := r.RemoteAddr
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		var addrErr *net.AddrError
		if errors.As(err, &addrErr) && addrErr.Err == "missing port in address" {
			ip = remoteAddr
		} else {
			return nil, err
		}
	}
	requestIP := net.ParseIP(ip)
	if !h.isTrustedIP(requestIP) {
		return nil, fmt.Errorf("'%s' is not a trusted ip", r.RemoteAddr)
	}

	headerName := h.dependencies.Config().Http.SSOProxyAuthHeaderName
	userName := r.Header.Get(headerName)
	if userName == "" {
		return nil, nil
	}

	account, err := h.dependencies.Domains().Accounts().GetAccountByUsername(r.Context(), userName)
	if err != nil {
		return nil, err
	}

	return account, nil
}
func (h *Handler) isTrustedIP(ip net.IP) bool {
	for _, net := range h.trustedIPs {
		if ok := net.Contains(ip); ok {
			return true
		}
	}
	return false
}
