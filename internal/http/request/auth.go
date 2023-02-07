package request

import (
	"github.com/gofiber/fiber/v2"
)

func IsLogged(c *fiber.Ctx) bool {
	return c.Locals("account") != nil
}
