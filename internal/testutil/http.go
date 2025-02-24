package testutil

import (
	"io"
	"net/http/httptest"
	"strings"

	"github.com/go-shiori/shiori/internal/http/webcontext"
	"github.com/go-shiori/shiori/internal/model"
)

type Option = func(c model.WebContext)

// NewTestWebContext creates a new WebContext with test recorder and request
func NewTestWebContext() (model.WebContext, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	return webcontext.NewWebContext(w, r), w
}

// NewTestWebContextWithMethod creates a new WebContext with specified method
func NewTestWebContextWithMethod(method, path string, opts ...Option) (model.WebContext, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	c := webcontext.NewWebContext(w, r)
	for _, opt := range opts {
		opt(c)
	}
	return c, w
}

func WithBody(body string) Option {
	return func(c model.WebContext) {
		c.Request().Body = io.NopCloser(strings.NewReader(body))
	}
}

func WithHeader(name, value string) Option {
	return func(c model.WebContext) {
		c.Request().Header.Add(name, value)
	}
}

// WithAuthToken adds an authorization token to the request
func WithAuthToken(token string) Option {
	return func(c model.WebContext) {
		c.Request().Header.Add(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token)
	}
}

func WithAccount(account *model.AccountDTO) Option {
	return func(c model.WebContext) {
		c.SetAccount(account)
	}
}

// WithFakeAccount adds a fake account to the request context
func WithFakeAccount(isAdmin bool) Option {
	return func(c model.WebContext) {
		c.SetAccount(FakeAccount(isAdmin))
	}
}

func WithRequestPathValue(key, value string) Option {
	return func(c model.WebContext) {
		c.Request().SetPathValue(key, value)
	}
}

// PerformRequest executes a request against a handler
func PerformRequest(deps model.Dependencies, handler model.HttpHandler, method, path string, options ...Option) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	c := webcontext.NewWebContext(w, r)
	for _, opt := range options {
		opt(c)
	}

	handler(deps, c)

	return w
}

// PerformRequestOnRecorder executes a request against a handler and returns the response recorder
func PerformRequestOnRecorder(deps model.Dependencies, w *httptest.ResponseRecorder, handler model.HttpHandler, method, path string, options ...Option) {
	r := httptest.NewRequest(method, path, nil)
	c := webcontext.NewWebContext(w, r)
	for _, opt := range options {
		opt(c)
	}
	handler(deps, c)
}

// FakeAccount creates a fake account for testing
func FakeAccount(isAdmin bool) *model.AccountDTO {
	return &model.AccountDTO{
		ID:       1,
		Username: "user",
		Owner:    model.Ptr(isAdmin),
	}
}

// SetFakeUser sets a fake user account in the WebContext
func SetFakeUser(c model.WebContext) {
	c.SetAccount(&model.AccountDTO{
		ID:       1,
		Username: "user",
		Owner:    model.Ptr(false),
	})
}

// SetFakeAdmin sets a fake admin account in the WebContext
func SetFakeAdmin(c model.WebContext) {
	c.SetAccount(&model.AccountDTO{
		ID:       1,
		Username: "user",
		Owner:    model.Ptr(true),
	})
}

// WithFakeUser returns an Option that sets a fake user account
func WithFakeUser() Option {
	return WithFakeAccount(false)
}

// WithFakeAdmin returns an Option that sets a fake admin account
func WithFakeAdmin() Option {
	return WithFakeAccount(true)
}

// SetRequestPathValue sets a path value for the request
func SetRequestPathValue(c model.WebContext, key, value string) {
	c.Request().SetPathValue(key, value)
}
