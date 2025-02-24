package model

import "net/http"

const (
	// ContextAccountKey is the key used to store the account model in the gin context.
	ContextAccountKey = "account"

	// AuthorizationHeader is the name of the header used to send the token.
	AuthorizationHeader = "Authorization"
	// AuthorizationTokenType is the type of token used in the Authorization header.
	AuthorizationTokenType = "Bearer"
)

// WebContext represents the context of an HTTP request
type WebContext interface {
	Request() *http.Request
	ResponseWriter() http.ResponseWriter
	GetAccount() *AccountDTO
	SetAccount(*AccountDTO)
	UserIsLogged() bool
	GetRequestID() string
	SetRequestID(id string)
}

// Handler is a custom handler function that receives dependencies and web context
type HttpHandler func(deps Dependencies, c WebContext)

// Middleware defines the interface for request/response customization
type HttpMiddleware interface {
	OnRequest(deps Dependencies, c WebContext) error
	OnResponse(deps Dependencies, c WebContext) error
}
