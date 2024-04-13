package response

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	// Response payload
	// Ok if the response was successful or not
	Ok bool `json:"ok"`

	// Message the payload of the response, depending on the endpoint/response status
	Message any `json:"message"`

	// ErrorParams parameters defined if the response is not successful to help client's debugging
	ErrorParams map[string]string `json:"error_params,omitempty"`

	// statusCode used for the http response status code
	statusCode int
}

func (m *Response) IsError() bool {
	return !m.Ok
}

func (m *Response) GetMessage() any {
	return m.Message
}

func (m *Response) Send(c *gin.Context) {
	c.Status(m.statusCode)
	c.JSON(m.statusCode, m)
}

func NewResponse(ok bool, message any, errorParams map[string]string, statusCode int) *Response {
	return &Response{
		Ok:          ok,
		Message:     message,
		ErrorParams: errorParams,
		statusCode:  statusCode,
	}
}
