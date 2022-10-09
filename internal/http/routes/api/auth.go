package api

import (
	"context"
	"fmt"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthAPIRoutes struct {
	logger *zap.Logger
	router *fiber.App
	deps   *config.Dependencies
}

func (r *AuthAPIRoutes) Setup() *AuthAPIRoutes {
	r.router.Post("/login", r.loginHandler)
	return r
}

func (r *AuthAPIRoutes) Router() *fiber.App {
	return r.router
}

type loginPayload struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

func (p *loginPayload) IsValid() error {
	if p.Username == "" {
		return fmt.Errorf("username should not be empty")
	}
	if p.Password == "" {
		return fmt.Errorf("password should not be empty")
	}
	return nil
}

func (r *AuthAPIRoutes) loginHandler(c *fiber.Ctx) error {
	ctx := context.Background()

	var payload loginPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.SendError(c, 500, err)
	}

	if err := payload.IsValid(); err != nil {
		return response.SendError(c, 400, err.Error())
	}

	account, err := r.deps.Domains.Auth.Login(ctx, payload.Username, payload.Password)
	if err != nil {
		return response.SendError(c, 400, err.Error())
	}

	return response.Send(c, 200, account)
}

func NewAuthAPIRoutes(logger *zap.Logger, deps *config.Dependencies) *AuthAPIRoutes {
	routes := AuthAPIRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
	}
	routes.Setup()
	return &routes
}
