package api

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type DebugAPIRoutes struct {
	logger *logrus.Logger
	router *fiber.App
	deps   *config.Dependencies
}

func (r *DebugAPIRoutes) Setup() *DebugAPIRoutes {
	r.router.Get("/create_user", r.createUserHandler)
	return r
}

func (r *DebugAPIRoutes) Router() *fiber.App {
	return r.router
}

func (r *DebugAPIRoutes) createUserHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	account := model.Account{
		Username: "shiori",
		Password: "gopher",
		Owner:    true,
	}

	if err := r.deps.Database.SaveAccount(ctx, account); err != nil {
		return response.SendError(c, 500, err.Error())
	}
	return response.Send(c, 201, account)
}

func NewDebugPIRoutes(logger *logrus.Logger, deps *config.Dependencies) *DebugAPIRoutes {
	return &DebugAPIRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
	}
}
