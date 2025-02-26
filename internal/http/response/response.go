package response

import (
	"encoding/json"

	"github.com/go-shiori/shiori/internal/model"
)

type Response struct {
	// Ok if the response was successful or not
	Ok bool `json:"ok"`

	// Message the payload of the response, depending on the endpoint/response status
	Message any `json:"message"`

	// ErrorParams parameters defined if the response is not successful to help client's debugging
	ErrorParams map[string]string `json:"error_params,omitempty"`

	// statusCode used for the http response status code
	statusCode int
}

func (r *Response) IsError() bool {
	return !r.Ok
}

func (r *Response) GetMessage() any {
	return r.Message
}

func (r *Response) Send(c model.WebContext) error {
	c.ResponseWriter().Header().Set("Content-Type", "application/json")
	c.ResponseWriter().WriteHeader(r.statusCode)
	return json.NewEncoder(c.ResponseWriter()).Encode(r)
}

func NewResponse(ok bool, message any, errorParams map[string]string, statusCode int) *Response {
	return &Response{
		Ok:          ok,
		Message:     message,
		ErrorParams: errorParams,
		statusCode:  statusCode,
	}
}
