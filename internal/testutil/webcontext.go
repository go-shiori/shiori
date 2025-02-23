package testutil

import (
	"net/http"
	"net/http/httptest"

	"github.com/go-shiori/shiori/internal/model"
)

// NewTestWebContext creates a new WebContext with test recorder and request
func NewTestWebContext() (model.WebContext, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	return NewWebContext(w, r), w
}

// NewTestWebContextWithMethod creates a new WebContext with specified method
func NewTestWebContextWithMethod(method, path string, opts ...Option) (model.WebContext, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	for _, opt := range opts {
		opt(r)
	}
	return NewWebContext(w, r), w
}

type testWebContext struct {
	req     *http.Request
	resp    http.ResponseWriter
	account *model.AccountDTO
}

func NewWebContext(w http.ResponseWriter, r *http.Request) model.WebContext {
	return &testWebContext{
		req:  r,
		resp: w,
	}
}

func (c *testWebContext) GetAccount() *model.AccountDTO       { return c.account }
func (c *testWebContext) Request() *http.Request              { return c.req }
func (c *testWebContext) ResponseWriter() http.ResponseWriter { return c.resp }
func (c *testWebContext) UserIsLogged() bool                  { return c.account != nil }
func (c *testWebContext) SetAccount(a *model.AccountDTO) {
	c.account = a
}
