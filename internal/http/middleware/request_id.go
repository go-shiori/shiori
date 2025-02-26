package middleware

import (
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofrs/uuid/v5"
)

const (
	// RequestIDHeader is the header key for the request ID
	RequestIDHeader = "X-Request-ID"
)

// RequestIDMiddleware adds a unique request ID to each request
type RequestIDMiddleware struct {
	deps model.Dependencies
}

// NewRequestIDMiddleware creates a new RequestIDMiddleware
func NewRequestIDMiddleware(deps model.Dependencies) *RequestIDMiddleware {
	return &RequestIDMiddleware{deps: deps}
}

// OnRequest adds a request ID to the request context and response headers
func (m *RequestIDMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	// Generate request ID
	requestID, err := uuid.NewV4()
	if err != nil {
		deps.Logger().WithError(err).Error("Failed to generate request ID")
		return err
	}

	// Add request ID to response headers
	c.ResponseWriter().Header().Set(RequestIDHeader, requestID.String())

	// Add request ID to context
	c.SetRequestID(requestID.String())

	return nil
}

// OnResponse is a no-op for this middleware
func (m *RequestIDMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	return nil
}
