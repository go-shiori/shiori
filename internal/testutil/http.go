package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/model"
)

type Option = func(*http.Request)

func WithBody(body string) Option {
	return func(request *http.Request) {
		request.Body = io.NopCloser(strings.NewReader(body))
	}
}

func WithHeader(name, value string) Option {
	return func(request *http.Request) {
		request.Header.Add(name, value)
	}
}

func WithAuthToken(token string) Option {
	return func(request *http.Request) {
		request.Header.Add(model.AuthorizationHeader, model.AuthorizationTokenType+" "+token)
	}
}

// PerformRequest executes a request against a handler
func PerformRequest(deps model.Dependencies, handler model.HttpHandler, method, path string, options ...Option) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	for _, opt := range options {
		opt(r)
	}

	c := NewWebContext(w, r)
	handler(deps, c)

	return w
}

// FakeAccount creates a fake account for testing
func FakeAccount(isAdmin bool) *model.AccountDTO {
	return &model.AccountDTO{
		ID:       1,
		Username: "user",
		Owner:    model.Ptr(isAdmin),
	}
}

// WithFakeAccount adds a fake account to the request context
func WithFakeAccount(isAdmin bool) Option {
	return func(r *http.Request) {
		c := NewWebContext(nil, r)
		c.SetAccount(FakeAccount(isAdmin))
	}
}

// FakeUserLoggedInMiddlewware is a middleware that sets a fake user account to context.
// Keep in mind that this users is not saved in database so any tests that use this middleware
// should not rely on database.
func FakeUserLoggedInMiddlewware(ctx *gin.Context) {
	ctx.Set("account", &model.AccountDTO{
		ID:       1,
		Username: "user",
		Owner:    model.Ptr(false),
	})
}

// FakeAdminLoggedInMiddlewware is a middleware that sets a fake admin account to context.
// Keep in mind that this users is not saved in database so any tests that use this middleware
// should not rely on database.
func FakeAdminLoggedInMiddlewware(ctx *gin.Context) {
	ctx.Set("account", &model.AccountDTO{
		ID:       1,
		Username: "user",
		Owner:    model.Ptr(true),
	})
}

// AuthUserMiddleware is a middleware that manually sets an user as authenticated in the context
// to be used in tests.
func AuthUserMiddleware(user *model.AccountDTO) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("account", user)
	}
}
