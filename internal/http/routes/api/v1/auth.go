package api_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type AuthAPIRoutes struct {
	logger             *logrus.Logger
	deps               *dependencies.Dependencies
	legacyLoginHandler model.LegacyLoginHandler
}

func (r *AuthAPIRoutes) Setup(group *gin.RouterGroup) model.Routes {
	group.GET("/me", r.meHandler)
	group.POST("/login", r.loginHandler)
	group.POST("/refresh", r.refreshHandler)
	group.PATCH("/account", r.updateHandler)
	return r
}

type loginRequestPayload struct {
	Username   string `json:"username"    validate:"required"`
	Password   string `json:"password"    validate:"required"`
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
	Token      string `json:"token"`
	SessionID  string `json:"session"` // Deprecated, used only for legacy APIs
	Expiration int64  `json:"expires"` // Deprecated, used only for legacy APIs
}

type settingRequestPayload struct {
	Config model.UserConfig `json:"config"`
}

// loginHandler godoc
//
//	@Summary	Login to an account using username and password
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		loginRequestPayload		false	"Login data"
//	@Success	200		{object}	loginResponseMessage	"Login successful"
//	@Failure	400		{object}	nil						"Invalid login data"
//	@Router		/api/v1/auth/login [post]
func (r *AuthAPIRoutes) loginHandler(c *gin.Context) {
	var payload loginRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.SendInternalServerError(c)
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	account, err := r.deps.Domains.Auth.GetAccountFromCredentials(c, payload.Username, payload.Password)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	expiration := time.Hour
	if payload.RememberMe {
		expiration = time.Hour * 24 * 30
	}

	expirationTime := time.Now().Add(expiration)

	token, err := r.deps.Domains.Auth.CreateTokenForAccount(account, expirationTime)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	sessionID, err := r.legacyLoginHandler(*account, expiration)
	if err != nil {
		r.logger.WithError(err).Error("failed execute legacy login handler")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusOK, loginResponseMessage{
		Token:      token,
		SessionID:  sessionID,
		Expiration: expirationTime.Unix(),
	})
}

// refreshHandler godoc
//
//	@Summary					Refresh a token for an account
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Produce					json
//	@Success					200	{object}	loginResponseMessage	"Refresh successful"
//	@Failure					403	{object}	nil						"Token not provided/invalid"
//	@Router						/api/v1/auth/refresh [post]
func (r *AuthAPIRoutes) refreshHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if !ctx.UserIsLogged() {
		response.SendError(c, http.StatusForbidden, nil)
		return
	}

	expiration := time.Now().Add(time.Hour * 72)
	account, _ := c.Get(model.ContextAccountKey)
	token, err := r.deps.Domains.Auth.CreateTokenForAccount(account.(*model.Account), expiration)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusAccepted, loginResponseMessage{
		Token: token,
	})
}

// meHandler godoc
//
//	@Summary					Get information for the current logged in user
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Produce					json
//	@Success					200	{object}	model.Account
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/auth/me [get]
func (r *AuthAPIRoutes) meHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if !ctx.UserIsLogged() {
		response.SendError(c, http.StatusForbidden, nil)
		return
	}

	response.Send(c, http.StatusOK, ctx.GetAccount())
}

// updateHandler godoc
//
//	@Summary					Perform actions on the currently logged-in user.
//	@Tags						Auth
//	@securityDefinitions.apikey	ApiKeyAuth
//	@Param						payload	body	settingRequestPayload	false	"Config data"
//	@Produce					json
//	@Success					200	{object}	model.Account
//	@Failure					403	{object}	nil	"Token not provided/invalid"
//	@Router						/api/v1/auth/account [patch]
func (r *AuthAPIRoutes) updateHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if !ctx.UserIsLogged() {
		response.SendError(c, http.StatusForbidden, nil)
	}
	var payload settingRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.SendInternalServerError(c)
	}

	account := ctx.GetAccount()
	if account == nil {
		response.SendError(c, http.StatusUnauthorized, nil)
		return
	}
	account.Config = payload.Config

	err := r.deps.Database.SaveAccountSettings(c, *account)
	if err != nil {
		response.SendInternalServerError(c)
	}

	response.Send(c, http.StatusOK, ctx.GetAccount())
}

func NewAuthAPIRoutes(logger *logrus.Logger, deps *dependencies.Dependencies, loginHandler model.LegacyLoginHandler) *AuthAPIRoutes {
	return &AuthAPIRoutes{
		logger:             logger,
		deps:               deps,
		legacyLoginHandler: loginHandler,
	}
}
