package webserver

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	cch "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

var developmentMode = false

// Handler is Handler for serving the web interface.
type Handler struct {
	DB           database.DB
	DataDir      string
	RootPath     string
	UserCache    *cch.Cache
	SessionCache *cch.Cache
	ArchiveCache *cch.Cache
	Log          bool

	depenencies *config.Dependencies

	templates map[string]*template.Template
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

func (h *Handler) PrepareTemplates() error {
	// Prepare variables
	var err error
	h.templates = make(map[string]*template.Template)

	// Prepare func map
	funcMap := template.FuncMap{
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	// Create template for login, index and content
	for _, name := range []string{"login", "index", "content"} {
		h.templates[name], err = createTemplate(name+".html", funcMap)
		if err != nil {
			return err
		}
	}

	// Create template for archive overlay
	h.templates["archive"], err = template.New("archive").Delims("$$", "$$").Parse(
		`<div id="shiori-archive-header">
		<p id="shiori-logo"><span>栞</span>shiori</p>
		<div class="spacer"></div>
		<a href="$$.URL$$" target="_blank">View Original</a>
		$$if .HasContent$$
		<a href="/bookmark/$$.ID$$/content">View Readable</a>
		$$end$$
		</div>`)
	if err != nil {
		return err
	}

	return nil
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

		account, err := h.depenencies.Domains.Auth.CheckToken(r.Context(), authParts[1])
		if err != nil {
			return err
		}

		if r.Method != "" && r.Method != "GET" && !account.Owner {
			return fmt.Errorf("account level is not sufficient")
		}

		h.depenencies.Log.WithFields(logrus.Fields{
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
