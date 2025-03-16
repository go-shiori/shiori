package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-shiori/shiori/internal/model"
)

type responseMiddlewareBody struct {
	Ok      bool `json:"ok"`
	Message any  `json:"message"`
}

type MessageResponseMiddleware struct {
	deps model.Dependencies
}

func (m *MessageResponseMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	if c.Request().Header.Get("X-Shiori-Response-Format") == "new" {
		return nil
	}

	// Create a response recorder and wrap the original ResponseWriter
	recorder := newResponseRecorder(c.ResponseWriter())
	c.SetResponseWriter(recorder)
	return nil
}

func (m *MessageResponseMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	if c.Request().Header.Get("X-Shiori-Response-Format") == "new" {
		return nil
	}

	writer := c.ResponseWriter()

	// Get the response recorder
	recorder, ok := writer.(*responseRecorder)
	if !ok {
		return nil
	}

	// Copy all headers to the original response writer
	for k, v := range recorder.header {
		if k != "Content-Length" {
			recorder.ResponseWriter.Header().Set(k, strings.Join(v, ""))
		}
	}

	// Write the status code
	recorder.ResponseWriter.WriteHeader(recorder.statusCode)

	// If it's not a JSON response, write the original response and return
	if ct := recorder.header.Get("Content-Type"); ct != "application/json" {
		_, err := recorder.ResponseWriter.Write(recorder.body.Bytes())
		return err
	}

	// For JSON responses, wrap them in our format
	wrappedResponse := responseMiddlewareBody{
		Ok:      recorder.statusCode < 400,
		Message: nil,
	}

	// If there's a response body and status code allows body, parse it
	if recorder.body.Len() > 0 && recorder.statusCode != http.StatusNoContent {
		var originalBody any
		if err := json.NewDecoder(&recorder.body).Decode(&originalBody); err != nil {
			return err
		}
		wrappedResponse.Message = originalBody
		// Write the status code and wrapped response
		if err := json.NewEncoder(recorder.ResponseWriter).Encode(wrappedResponse); err != nil {
			return err
		}
	}

	return nil
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
		body:           bytes.Buffer{},
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
