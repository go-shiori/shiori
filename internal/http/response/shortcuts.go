package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const internalServerErrorMessage = "Internal server error, please contact an administrator"

// New provides a shortcut to a successful response object
func New(ok bool, statusCode int, data interface{}) *Response {
	return NewResponse(ok, data, nil, statusCode)
}

// Send provides a shortcut to send a (potentially) successful response
func Send(ctx *gin.Context, statusCode int, data interface{}) {
	New(true, statusCode, data).Send(ctx)
}

// SendError provides a shortcut to send an unsuccessful response
func SendError(ctx *gin.Context, statusCode int, data interface{}) {
	New(false, statusCode, data).Send(ctx)
}

// SendErrorWithParams the same as above but for errors that require error parameters
func SendErrorWithParams(ctx *gin.Context, statusCode int, data interface{}, errorParams map[string]string) {
	NewResponse(false, data, errorParams, statusCode).Send(ctx)
}

// SendInternalServerError directly sends an internal server error response
func SendInternalServerError(ctx *gin.Context) {
	SendError(ctx, http.StatusInternalServerError, internalServerErrorMessage)
}
