package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type AccountAPIRoutes struct {
	logger *logrus.Logger
	deps   *config.Dependencies
}

func (r *AccountAPIRoutes) Setup(group *gin.RouterGroup) model.Routes {
	group.GET("/me", r.meHandler)
	group.POST("/login", r.loginHandler)
	group.POST("/refresh", r.refreshHandler)
	group.POST("/logout", r.logoutHandler)
	return r
}

func (r *AccountAPIRoutes) setCookie(c *gin.Context, token string, expiration time.Time) {
	c.SetCookie("auth", token, int(expiration.Unix()), "/", "", !r.deps.Config.Development, false)
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
	Token string `json:"token"`
}

// loginHandler godoc
// @Summary      Login to an account
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        payload   body    loginRequestPayload    false  "Login data"
// @Router       /api/v1/account/login [post]
func (r *AccountAPIRoutes) loginHandler(c *gin.Context) {
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

	expiration := time.Now().Add(time.Hour)
	if payload.RememberMe {
		expiration = time.Now().Add(time.Hour * 24 * 30)
	}

	token, err := r.deps.Domains.Auth.CreateTokenForAccount(account, expiration)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	responseMessage := loginResponseMessage{
		Token: token,
	}

	r.setCookie(c, token, expiration)

	response.Send(c, http.StatusOK, responseMessage)
}

func (r *AccountAPIRoutes) refreshHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if ctx.UserIsLogged() {
		response.SendError(c, http.StatusForbidden, nil)
		return
	}

	expiration := time.Now().Add(time.Hour * 72)
	account, _ := c.Get("account")
	token, err := r.deps.Domains.Auth.CreateTokenForAccount(account.(*model.Account), expiration)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	responseMessage := loginResponseMessage{
		Token: token,
	}

	r.setCookie(c, token, expiration)

	response.Send(c, http.StatusAccepted, responseMessage)
}

func (r *AccountAPIRoutes) meHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if ctx.UserIsLogged() {
		response.SendError(c, http.StatusUnauthorized, nil)
	}

	account, _ := c.Get("account")
	response.Send(c, http.StatusOK, account.(*model.Account))
}

func (r *AccountAPIRoutes) logoutHandler(c *gin.Context) {
	ctx := context.NewContextFromGin(c)
	if ctx.UserIsLogged() {
		response.SendError(c, http.StatusUnauthorized, nil)
		return
	}

	c.SetCookie("auth", "", 0, "/", "", !r.deps.Config.Development, false)

	// no-op server side, at least for now
	response.Send(c, http.StatusOK, "logged out")
}

func NewAccountAPIRoutes(logger *logrus.Logger, deps *config.Dependencies) *AccountAPIRoutes {
	return &AccountAPIRoutes{
		logger: logger,
		deps:   deps,
	}
}
