package api_v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

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
	Token      string `json:"token"`
	Expiration int64  `json:"expires"`
}

// @Summary	Login to an account using username and password
// @Tags		Auth
// @Accept		json
// @Produce	json
// @Param		payload	body		loginRequestPayload		false	"Login data"
// @Success	200		{object}	loginResponseMessage	"Login successful"
// @Failure	400		{object}	nil						"Invalid login data"
// @Router		/api/v1/auth/login [post]
func HandleLogin(deps model.Dependencies, c model.WebContext) {
	var payload loginRequestPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	account, err := deps.Domains().Auth().GetAccountFromCredentials(c.Request().Context(), payload.Username, payload.Password)
	if err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	expiration := time.Hour
	if payload.RememberMe {
		expiration = time.Hour * 24 * 30
	}

	expirationTime := time.Now().Add(expiration)

	token, err := deps.Domains().Auth().CreateTokenForAccount(account, expirationTime)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusOK, loginResponseMessage{
		Token:      token,
		Expiration: expirationTime.Unix(),
	})
}

// @Summary					Refresh a token for an account
// @Tags						Auth
// @securityDefinitions.apikey	ApiKeyAuth
// @Produce					json
// @Success					200	{object}	loginResponseMessage	"Refresh successful"
// @Failure					403	{object}	nil						"Token not provided/invalid"
// @Router						/api/v1/auth/refresh [post]
func HandleRefreshToken(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	expiration := time.Now().Add(time.Hour * 72)
	account := c.GetAccount()
	token, err := deps.Domains().Auth().CreateTokenForAccount(account, expiration)
	if err != nil {
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusAccepted, loginResponseMessage{
		Token: token,
	})
}

// @Summary					Get information for the current logged in user
// @Tags						Auth
// @securityDefinitions.apikey	ApiKeyAuth
// @Produce					json
// @Success					200	{object}	model.Account
// @Failure					403	{object}	nil	"Token not provided/invalid"
// @Router						/api/v1/auth/me [get]
func HandleGetMe(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}
	response.Send(c, http.StatusOK, c.GetAccount())
}

type updateAccountPayload struct {
	OldPassword string            `json:"old_password"`
	NewPassword string            `json:"new_password"`
	Username    string            `json:"username"`
	Owner       *bool             `json:"owner"`
	Config      *model.UserConfig `json:"config"`
}

func (p *updateAccountPayload) IsValid() error {
	if p.NewPassword != "" && p.OldPassword == "" {
		return fmt.Errorf("To update the password the old one must be provided")
	}
	return nil
}

func (p *updateAccountPayload) ToAccountDTO() model.AccountDTO {
	account := model.AccountDTO{}
	if p.NewPassword != "" {
		account.Password = p.NewPassword
	}
	if p.Owner != nil {
		account.Owner = p.Owner
	}
	if p.Config != nil {
		account.Config = p.Config
	}
	if p.Username != "" {
		account.Username = p.Username
	}
	return account
}

// @Summary					Update account information
// @Tags						Auth
// @securityDefinitions.apikey	ApiKeyAuth
// @Param						payload	body	updateAccountPayload	false	"Account data"
// @Produce					json
// @Success					200	{object}	model.Account
// @Failure					403	{object}	nil	"Token not provided/invalid"
// @Router						/api/v1/auth/account [patch]
func HandleUpdateLoggedAccount(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	var payload updateAccountPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendInternalServerError(c)
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	account := c.GetAccount()

	if payload.NewPassword != "" {
		_, err := deps.Domains().Auth().GetAccountFromCredentials(c.Request().Context(), account.Username, payload.OldPassword)
		if err != nil {
			response.SendError(c, http.StatusBadRequest, "Old password is incorrect", nil)
			return
		}
	}

	updatedAccount := payload.ToAccountDTO()
	updatedAccount.ID = account.ID

	account, err := deps.Domains().Accounts().UpdateAccount(c.Request().Context(), updatedAccount)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to update account")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusOK, account)
}

// @Summary					Logout from the current session
// @Tags						Auth
// @securityDefinitions.apikey	ApiKeyAuth
// @Produce					json
// @Success					200	{object}	nil	"Logout successful"
// @Failure					403	{object}	nil	"Token not provided/invalid"
// @Router						/api/v1/auth/logout [post]
func HandleLogout(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInUser(deps, c); err != nil {
		return
	}

	// Remove token cookie
	c.Request().AddCookie(&http.Cookie{
		Name:  "token",
		Value: "",
	})

	response.Send(c, http.StatusOK, nil)
}
