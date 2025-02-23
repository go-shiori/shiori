package response

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-shiori/shiori/internal/http/templates"
	"github.com/go-shiori/shiori/internal/model"
)

const internalServerErrorMessage = "Internal server error, please contact an administrator"

// New provides a shortcut to a successful response object
func New(ok bool, statusCode int, data any) *Response {
	return NewResponse(ok, data, nil, statusCode)
}

// Send provides a shortcut to send a (potentially) successful response
func Send(c model.WebContext, statusCode int, message any) error {
	resp := NewResponse(statusCode < 400, message, nil, statusCode)
	return resp.Send(c)
}

// SendError provides a shortcut to send an unsuccessful response
func SendError(c model.WebContext, statusCode int, message any, errorParams map[string]string) error {
	resp := NewResponse(false, message, errorParams, statusCode)
	return resp.Send(c)
}

// SendErrorWithParams the same as above but for errors that require error parameters
func SendErrorWithParams(c model.WebContext, statusCode int, data any, errorParams map[string]string) error {
	return NewResponse(false, data, errorParams, statusCode).Send(c)
}

// SendInternalServerError directly sends an internal server error response
func SendInternalServerError(c model.WebContext) error {
	return SendError(c, http.StatusInternalServerError, internalServerErrorMessage, nil)
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
	c.ResponseWriter().Header().Set("Content-Type", "application/json")
	c.ResponseWriter().WriteHeader(statusCode)
	return json.NewEncoder(c.ResponseWriter()).Encode(data)
}

// SendErrorJSON is a helper function to send error JSON responses
func SendErrorJSON(c model.WebContext, statusCode int, message string) error {
	return SendError(c, statusCode, message, nil)
}

// SendTemplate renders and sends an HTML template
func SendTemplate(c model.WebContext, name string, data any) error {
	c.ResponseWriter().Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.RenderTemplate(c.ResponseWriter(), name, data); err != nil {
		return SendInternalServerError(c)
	}
	return nil
}
