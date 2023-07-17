package model

import "github.com/gin-gonic/gin"

const (
	// ContextAccountKey is the key used to store the account model in the gin context.
	ContextAccountKey = "account"

	// AuthorizationHeader is the name of the header used to send the token.
	AuthorizationHeader = "Authorization"
	// AuthorizationTokenType is the type of token used in the Authorization header.
	AuthorizationTokenType = "Bearer"
)

type Routes interface {
	Setup(group *gin.RouterGroup) Routes
}
