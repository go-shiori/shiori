package response

import "github.com/gofiber/fiber/v2"

type Response struct {
	Ok          bool              `json:"ok"`
	Message     interface{}       `json:"message"`
	ErrorParams map[string]string `json:"error_params,omitempty"`
	statusCode  int
}

func (m *Response) IsError() bool {
	return m.Ok
}

func (m *Response) Send(c *fiber.Ctx) error {
	return c.Status(m.statusCode).JSON(m)
}

func NewResponse(ok bool, message interface{}, errorParams map[string]string, statusCode int) *Response {
	return &Response{
		Ok:          ok,
		Message:     message,
		ErrorParams: errorParams,
		statusCode:  statusCode,
	}
}
