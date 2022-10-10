package response

import "github.com/gofiber/fiber/v2"

const internalServerErrorMessage = "Internal server error, please contact an administrator"

// New provides a shortcut to a successful response object
func New(ok bool, statusCode int, data interface{}) *Response {
	return NewResponse(ok, data, nil, statusCode)
}

// Send provides a shortcut to send a (potentially) successful response
func Send(ctx *fiber.Ctx, statusCode int, data interface{}) error {
	return New(true, statusCode, data).Send(ctx)
}

// SendError provides a shortcut to send an unsuccessful response
func SendError(ctx *fiber.Ctx, statusCode int, data interface{}) error {
	return New(false, statusCode, data).Send(ctx)
}

// SendErrorWithParams the same as above but for errors that require error parameters
func SendErrorWithParams(ctx *fiber.Ctx, statusCode int, data interface{}, errorParams map[string]string) error {
	return NewResponse(false, data, errorParams, statusCode).Send(ctx)
}

// SendInternalServerError directly sends an internal server error response
func SendInternalServerError(ctx *fiber.Ctx) error {
	return SendError(ctx, fiber.StatusInternalServerError, internalServerErrorMessage)
}
