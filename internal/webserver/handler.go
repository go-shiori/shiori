package webserver

import (
	"fmt"
	"github.com/gofrs/uuid"
	"html/template"
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/warc"
	cch "github.com/patrickmn/go-cache"
)

var developmentMode = false

// Handler is handler for serving the web interface.
type handler struct {
	DB           database.DB
	DataDir      string
	RootPath     string
	UserCache    *cch.Cache
	SessionCache *cch.Cache
	ArchiveCache *cch.Cache
	Log          bool

	templates map[string]*template.Template
}

func (h *handler) prepareSessionCache() {
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

func (h *handler) prepareArchiveCache() {
	h.ArchiveCache.OnEvicted(func(key string, data interface{}) {
		archive := data.(*warc.Archive)
		archive.Close()
	})
}

func (h *handler) prepareTemplates() error {
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

func (h *handler) getSessionID(r *http.Request) string {
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
func (h *handler) getSessionAccount(r *http.Request) (*model.Account, error) {
	sessionID := h.getSessionID(r)
	if sessionID == "" {
		return nil, fmt.Errorf("session is not exist")
	}

	// Make sure session is not expired yet
	val, found := h.SessionCache.Get(sessionID)
	if !found {
		return nil, fmt.Errorf("session has been expired")
	}

	account := val.(model.Account)
	return &account, nil
}

// validateSession checks whether user session is still valid or not
func (h *handler) validateSession(r *http.Request) error {
	account, err := h.getSessionAccount(r)
	if err != nil {
		return err
	}

	// If this is not get request, make sure it's owner
	if r.Method != "" && r.Method != "GET" {
		if !account.Owner {
			return fmt.Errorf("account level is not sufficient")
		}
	}

	return nil
}

func (h *handler) createSession(account model.Account, expTime time.Duration) (string, error) {
	// Create session ID
	sessionID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	// Save session ID to cache
	strSessionID := sessionID.String()
	h.SessionCache.Set(strSessionID, account, expTime)

	// Save user's session IDs to cache as well
	// useful for mass logout
	sessionIDs := []string{strSessionID}
	if val, found := h.UserCache.Get(account.Username); found {
		sessionIDs = val.([]string)
		sessionIDs = append(sessionIDs, strSessionID)
	}
	h.UserCache.Set(account.Username, sessionIDs, -1)

	return strSessionID, nil
}
