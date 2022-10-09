package middleware

import (
	"net/http"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

func JSONMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if string(c.Request().Header.ContentType()) != "application/json" {
			return response.SendError(c, http.StatusNotAcceptable, "")
		}

		c.Response().Header.Add("Content-Type", "application/json")

		return c.Next()
	}
}
