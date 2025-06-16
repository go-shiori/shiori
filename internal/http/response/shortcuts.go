package response

import (
	"net/http"
	"net/url"

	"github.com/go-shiori/shiori/internal/http/templates"
	"github.com/go-shiori/shiori/internal/model"
)

const internalServerErrorMessage = "Internal server error, please contact an administrator"

// New provides a shortcut to a successful response object
func New(statusCode int, data any) *Response {
	return NewResponse(data, statusCode)
}

// Send provides a shortcut to send a (potentially) successful response
func Send(c model.WebContext, statusCode int, message any, contentType string) error {
	return NewResponse(message, statusCode).Send(c, contentType)
}

// SendError provides a shortcut to send an unsuccessful response
func SendError(c model.WebContext, statusCode int, message any) error {
	resp := NewResponse(struct {
		Error string `json:"error"`
	}{Error: message.(string)}, statusCode)
	return resp.SendJSON(c)
}

// SendErrorWithParams the same as above but for errors that require error parameters
func SendErrorWithParams(c model.WebContext, statusCode int, data any, errorParams map[string]string) error {
	return NewResponse(data, statusCode).SendJSON(c)
}

// SendInternalServerError directly sends an internal server error response
func SendInternalServerError(c model.WebContext) error {
	return SendError(c, http.StatusInternalServerError, internalServerErrorMessage)
}

// RedirectToLogin redirects to the login page with an optional destination
func RedirectToLogin(c model.WebContext, webroot, dst string) {
	redirectURL := url.URL{
		Path: webroot,
		RawQuery: url.Values{
			"dst": []string{dst},
		}.Encode(),
	}
	http.Redirect(c.ResponseWriter(), c.Request(), redirectURL.String(), http.StatusFound)
}

// NotFound sends a not found response
func NotFound(c model.WebContext) {
	http.NotFound(c.ResponseWriter(), c.Request())
}

// SendJSON is a helper function to send JSON responses
func SendJSON(c model.WebContext, statusCode int, data any) error {
	response := NewResponse(data, statusCode)
	return response.SendJSON(c)
}

// SendTemplate renders and sends an HTML template
func SendTemplate(c model.WebContext, name string, data any) error {
	c.ResponseWriter().Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.RenderTemplate(c.ResponseWriter(), name, data); err != nil {
		return SendInternalServerError(c)
	}
	return nil
}
