package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
)

// AuthMiddleware provides basic authentication capabilities to all routes underneath
// its usage, only allowing authenticated users access and set a custom local context
// `account` with the account model for the logged in user.
func AuthMiddleware(deps *config.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			log.Println("no header")
			return
		}

		authParts := strings.SplitN(authorization, " ", 2)
		if len(authParts) != 2 && authParts[0] != "Bearer" {
			log.Println("no correct header")
			return
		}

		account, err := deps.Domains.Auth.CheckToken(c, authParts[1])
		if err != nil {
			log.Println("no correct token: ", err.Error())
			return
		}

		c.Set("account", account)
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
