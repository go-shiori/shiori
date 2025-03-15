package response

import (
	"encoding/json"
	"net/http"

	"github.com/go-shiori/shiori/internal/model"
)

type Response struct {
	// Data the payload of the response, depending on the endpoint/response status
	Data any `json:"message"`

	// statusCode used for the http response status code
	statusCode int
}

// GetData returns the data of the response
func (r *Response) GetData() any {
	return r.Data
}

// IsError returns true if the response is an error
func (r *Response) IsError() bool {
	return r.statusCode >= http.StatusBadRequest
}

// Send sends the response to the client
func (r *Response) Send(c model.WebContext, contentType string) error {
	c.ResponseWriter().Header().Set("Content-Type", contentType)
	c.ResponseWriter().WriteHeader(r.statusCode)
	_, err := c.ResponseWriter().Write([]byte(r.GetData().(string)))
	return err
}

// SendJSON sends the response to the client
func (r *Response) SendJSON(c model.WebContext) error {
	c.ResponseWriter().Header().Set("Content-Type", "application/json")
	c.ResponseWriter().WriteHeader(r.statusCode)
	return json.NewEncoder(c.ResponseWriter()).Encode(r.GetData())
}

// NewResponse creates a new response
func NewResponse(message any, statusCode int) *Response {
	return &Response{
		Data:       message,
		statusCode: statusCode,
	}
}
