package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

// AuthMiddleware provides basic authentication capabilities to all routes underneath
// its usage, only allowing authenticated users access and set a custom local context
// `account` with the account model for the logged in user.
func AuthMiddleware(deps *dependencies.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getTokenFromHeader(c)
		if token == "" {
			token = getTokenFromCookie(c)
		}

		account, err := deps.Domains.Auth.CheckToken(c.Request.Context(), token)
		if err != nil {
			deps.Log.WithError(err).Error("Failed to check token")
			return
		}

		c.Set(model.ContextAccountKey, account)
	}
}

// AuthenticationRequired provides a middleware that checks if the user is logged in, returning
// a 401 error if not.
func AuthenticationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.NewContextFromGin(c)
		if !ctx.UserIsLogged() {
			response.SendError(c, http.StatusUnauthorized, nil)
			return
		}
	}
}

// AdminRequired provides a middleware that checks if the user is logged in and is an admin, returning
// a 403 error if not.
func AdminRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := context.NewContextFromGin(ctx)
		account := c.GetAccount()
		if account == nil || !account.IsOwner() {
			response.SendError(ctx, http.StatusForbidden, nil)
			return
		}
	}
}

// getTokenFromHeader returns the token from the Authorization header, if any.
func getTokenFromHeader(c *gin.Context) string {
	authorization := c.GetHeader(model.AuthorizationHeader)
	if authorization == "" {
		return ""
	}

	authParts := strings.SplitN(authorization, " ", 2)
	if len(authParts) != 2 && authParts[0] != model.AuthorizationTokenType {
		return ""
	}

	return authParts[1]
}

// getTokenFromCookie returns the token from the token cookie, if any.
func getTokenFromCookie(c *gin.Context) string {
	cookie, err := c.Cookie("token")
	if err != nil {
		return ""
	}

	return cookie
}
