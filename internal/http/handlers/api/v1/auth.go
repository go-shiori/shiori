package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

type loginRequestPayload struct {
	Username   string
	Password   string
	RememberMe bool
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
	Token      string
	Expiration int64
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
		response.SendError(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	account, err := deps.Domains().Auth().GetAccountFromCredentials(c.Request().Context(), payload.Username, payload.Password)
	if errors.Is(err, model.ErrInvalidCredentials) {
		response.SendError(c, http.StatusBadRequest, "Invalid username or password")
		return
	}
	if err != nil {
		deps.Logger().WithError(err).Error("unexpected error during authentication")
		response.SendInternalServerError(c)
		return
	}

	expiration := time.Hour
	if payload.RememberMe {
		expiration = time.Hour * 24 * 30
	}

	expirationTime := time.Now().Add(expiration)

	token, err := deps.Domains().Auth().CreateTokenForAccount(account, expirationTime)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to create auth token")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, loginResponseMessage{
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

	expiration := time.Now().UTC().Add(time.Hour * 24 * 30)
	account := c.GetAccount()
	token, err := deps.Domains().Auth().CreateTokenForAccount(account, expiration)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to create refresh token")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, loginResponseMessage{
		Token:      token,
		Expiration: expiration.Unix(),
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
	response.SendJSON(c, http.StatusOK, c.GetAccount())
}

type updateAccountPayload struct {
	OldPassword string
	NewPassword string
	Username    string
	Owner       *bool
	Config      *model.UserConfig
}

func (p *updateAccountPayload) IsValid() error {
	if p.NewPassword != "" && p.OldPassword == "" {
		return fmt.Errorf("to update the password the old one must be provided")
	}
	return nil
}

func (p *updateAccountPayload) ToAccountDTO() model.AccountDTO {
	account := model.AccountDTO{
		Config: p.Config,
	}
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
		response.SendError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := payload.IsValid(); err != nil {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	account := c.GetAccount()

	// Non-admin users cannot change the Owner flag
	if payload.Owner != nil && !account.IsOwner() {
		response.SendError(c, http.StatusForbidden, "Only admins can change owner status")
		return
	}

	if payload.NewPassword != "" {
		_, err := deps.Domains().Auth().GetAccountFromCredentials(c.Request().Context(), account.Username, payload.OldPassword)
		if errors.Is(err, model.ErrInvalidCredentials) {
			response.SendError(c, http.StatusBadRequest, "Old password is incorrect")
			return
		}
		if err != nil {
			deps.Logger().WithError(err).Error("failed to verify old password")
			response.SendInternalServerError(c)
			return
		}
	}

	// TODO: Use a method in the AccountDTO to apply the updates directly:
	// account := domains.Accounts().GetAccount(...)
	// account.ApplyUpdates(payload)
	updatedAccount := payload.ToAccountDTO()
	updatedAccount.ID = account.ID

	account, err := deps.Domains().Accounts().UpdateAccount(c.Request().Context(), updatedAccount)
	if errors.Is(err, model.ErrNotFound) {
		response.SendError(c, http.StatusNotFound, "account not found")
		return
	}
	if errors.Is(err, model.ErrAlreadyExists) {
		response.SendError(c, http.StatusConflict, "account already exists")
		return
	}
	if validationErr, ok := err.(model.ValidationError); ok {
		response.SendError(c, http.StatusBadRequest, validationErr.Error())
		return
	}
	if err != nil {
		deps.Logger().WithError(err).Error("failed to update account")
		response.SendInternalServerError(c)
		return
	}

	response.SendJSON(c, http.StatusOK, account)
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
	http.SetCookie(c.ResponseWriter(), &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	response.SendJSON(c, http.StatusOK, nil)
}
