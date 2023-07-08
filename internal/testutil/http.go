package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
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
