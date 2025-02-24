package webcontext

import (
	"context"
	"net/http"
)

// WebContext wraps the standard request and response writer
type WebContext struct {
	request        *http.Request
	responseWriter http.ResponseWriter
}

// NewWebContext creates a new WebContext from http.ResponseWriter and *http.Request
func NewWebContext(w http.ResponseWriter, r *http.Request) *WebContext {
	return &WebContext{
		request:        r,
		responseWriter: w,
	}
}

// Context returns the request's context
func (c *WebContext) Context() context.Context {
	return c.request.Context()
}

// WithContext returns a shallow copy of c with its context changed to ctx
func (c *WebContext) WithContext(ctx context.Context) *WebContext {
	c2 := new(WebContext)
	*c2 = *c
	c2.request = c2.request.WithContext(ctx)
	return c2
}

func (c *WebContext) ResponseWriter() http.ResponseWriter {
	return c.responseWriter
}

func (c *WebContext) Request() *http.Request {
	return c.request
}
