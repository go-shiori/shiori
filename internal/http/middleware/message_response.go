package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/go-shiori/shiori/internal/model"
)

type responseMiddlewareBody struct {
	Ok      bool `json:"ok"`
	Message any  `json:"message"`
}

type MessageResponseMiddleware struct {
	deps                   model.Dependencies
	originalResponseWriter http.ResponseWriter
}

func (m *MessageResponseMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	// Create a response recorder and wrap the original ResponseWriter
	m.originalResponseWriter = c.ResponseWriter()
	recorder := newResponseRecorder(m.originalResponseWriter)
	c.SetResponseWriter(recorder)
	return nil
}

func (m *MessageResponseMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	writer := c.ResponseWriter()

	// Get the response recorder
	recorder, ok := writer.(*responseRecorder)
	if !ok {
		return nil
	}

	// Copy all headers to the original response writer
	for k, v := range recorder.header {
		m.originalResponseWriter.Header()[k] = v
	}

	// If it's not a JSON response, write the original response and return
	if ct := recorder.header.Get("Content-Type"); ct != "application/json" {
		m.originalResponseWriter.WriteHeader(recorder.statusCode)
		_, err := m.originalResponseWriter.Write(recorder.body.Bytes())
		return err
	}

	// For JSON responses, wrap them in our format
	wrappedResponse := responseMiddlewareBody{
		Ok:      recorder.statusCode < 400,
		Message: nil,
	}

	// If there's a response body, parse it
	if recorder.body.Len() > 0 {
		var originalBody any
		if err := json.NewDecoder(&recorder.body).Decode(&originalBody); err != nil {
			return err
		}
		wrappedResponse.Message = originalBody
	}

	// Write the status code and content type
	m.originalResponseWriter.Header().Set("Content-Type", "application/json")
	m.originalResponseWriter.WriteHeader(recorder.statusCode)

	// Write the wrapped response
	return json.NewEncoder(m.originalResponseWriter).Encode(wrappedResponse)
}

func NewMessageResponseMiddleware(deps model.Dependencies) *MessageResponseMiddleware {
	return &MessageResponseMiddleware{deps: deps}
}

// responseRecorder is a custom ResponseWriter that captures the response
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
	header     http.Header
}

func newResponseRecorder(original http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: original,
		statusCode:     http.StatusOK,
		header:         make(http.Header),
	}
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	// Only write to the buffer, we'll write to the actual ResponseWriter in OnResponse
	return r.body.Write(b)
}
