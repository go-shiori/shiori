package middleware

import (
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware provides basic authentication capabilities to all routes underneath
// its usage, only allowing authenticated users access and set a custom local context
// `account` with the account model for the logged in user.
func AuthMiddleware(secretKey string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: []byte(secretKey),
		SuccessHandler: func(c *fiber.Ctx) error {
			user := c.Locals("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)
			account := claims["account"].(map[string]interface{})
			c.Locals("account", model.Account{
				Username: account["username"].(string),
				ID:       int(account["id"].(float64)),
				Owner:    account["owner"].(bool),
			})
			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.SendError(c, fiber.StatusUnauthorized, err.Error())
		},
	})
}
