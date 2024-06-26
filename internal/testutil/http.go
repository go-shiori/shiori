package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/model"
)

// NewGin returns a new gin engine with test mode enabled.
func NewGin() *gin.Engine {
	engine := gin.New()
	gin.SetMode(gin.TestMode)
	return engine
}

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

func PerformRequest(handler http.Handler, method, path string, options ...Option) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	return PerformRequestWithRecorder(recorder, handler, method, path, options...)
}

func PerformRequestWithRecorder(recorder *httptest.ResponseRecorder, r http.Handler, method, path string, options ...Option) *httptest.ResponseRecorder {
	request, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	for _, opt := range options {
		opt(request)
	}
	r.ServeHTTP(recorder, request)
	return recorder
}

// FakeUserLoggedInMiddlewware is a middleware that sets a fake user account to context.
// Keep in mind that this users is not saved in database so any tests that use this middleware
// should not rely on database.
func FakeUserLoggedInMiddlewware(ctx *gin.Context) {
	ctx.Set("account", &model.Account{
		ID:       1,
		Username: "user",
		Owner:    false,
	})
}

// FakeAdminLoggedInMiddlewware is a middleware that sets a fake admin account to context.
// Keep in mind that this users is not saved in database so any tests that use this middleware
// should not rely on database.
func FakeAdminLoggedInMiddlewware(ctx *gin.Context) {
	ctx.Set("account", &model.Account{
		ID:       1,
		Username: "admin",
		Owner:    true,
	})
}

// AuthUserMiddleware is a middleware that manually sets an user as authenticated in the context
// to be used in tests.
func AuthUserMiddleware(user *model.AccountDTO) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("account", user)
	}
}
