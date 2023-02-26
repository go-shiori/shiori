package api

import (
	"context"
	"fmt"
	"time"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/request"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AccountAPIRoutes struct {
	logger *logrus.Logger
	router *fiber.App
	deps   *config.Dependencies
}

func (r *AccountAPIRoutes) Setup() *AccountAPIRoutes {
	r.router.Get("/me", r.meHandler)
	r.router.Post("/login", r.loginHandler)
	r.router.Post("/refresh", r.refreshHandler)
	r.router.Post("/logout", r.logoutHandler)
	return r
}

func (r *AccountAPIRoutes) setCookie(c *fiber.Ctx, token string, expiration time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:    "auth",
		Value:   token,
		Path:    "/",
		Secure:  !r.deps.Config.Development,
		Expires: expiration,
	})
}

func (r *AccountAPIRoutes) Router() *fiber.App {
	return r.router
}

type loginRequestPayload struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

func (p *loginRequestPayload) IsValid() error {
	if p.Username == "" {
		return fmt.Errorf("username should not be empty")
	}
	if p.Password == "" {
		return fmt.Errorf("password should not be empty")
	}
	return nil
}

type loginResponseMessage struct {
	Token string `json:"token"`
}

func (r *AccountAPIRoutes) loginHandler(c *fiber.Ctx) error {
	ctx := context.Background()

	var payload loginRequestPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.SendInternalServerError(c)
	}

	if err := payload.IsValid(); err != nil {
		return response.SendError(c, 400, err.Error())
	}

	account, err := r.deps.Domains.Auth.GetAccountFromCredentials(ctx, payload.Username, payload.Password)
	if err != nil {
		return response.SendError(c, 400, err.Error())
	}

	expiration := time.Now().Add(time.Hour)
	if payload.RememberMe {
		expiration = time.Now().Add(time.Hour * 24 * 30)
	}

	token, err := r.deps.Domains.Auth.CreateTokenForAccount(account, expiration)
	if err != nil {
		return response.SendInternalServerError(c)
	}

	responseMessage := loginResponseMessage{
		Token: token,
	}

	r.setCookie(c, token, expiration)

	return response.Send(c, 200, responseMessage)
}

func (r *AccountAPIRoutes) refreshHandler(c *fiber.Ctx) error {
	if !request.IsLogged(c) {
		return response.SendError(c, 403, nil)
	}

	expiration := time.Now().Add(time.Hour * 72)
	account := c.Locals("account").(model.Account)
	token, err := r.deps.Domains.Auth.CreateTokenForAccount(&account, expiration)
	if err != nil {
		return response.SendInternalServerError(c)
	}

	responseMessage := loginResponseMessage{
		Token: token,
	}

	r.setCookie(c, token, expiration)

	return response.Send(c, 202, responseMessage)
}

func (r *AccountAPIRoutes) meHandler(c *fiber.Ctx) error {
	if !request.IsLogged(c) {
		return response.SendError(c, 403, nil)
	}

	account := c.Locals("account").(model.Account)
	return response.Send(c, 200, account)
}

func (r *AccountAPIRoutes) logoutHandler(c *fiber.Ctx) error {
	if !request.IsLogged(c) {
		return response.SendError(c, 403, nil)
	}

	c.ClearCookie("auth")

	// no-op server side, at least for now
	return response.Send(c, 200, "logged out")
}

func NewAccountAPIRoutes(logger *logrus.Logger, deps *config.Dependencies) *AccountAPIRoutes {
	return &AccountAPIRoutes{
		logger: logger,
		router: fiber.New(),
		deps:   deps,
	}
}
