package api_v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

type createAccountPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Owner    bool   `json:"owner"`
}

func (p *createAccountPayload) ToAccountDTO() model.AccountDTO {
	return model.AccountDTO{
		Username: p.Username,
		Password: p.Password,
		Owner:    &p.Owner,
	}
}

// @Summary		List accounts
// @Description	List accounts
// @Tags			accounts
// @Produce		json
// @Success		200	{array}		model.AccountDTO
// @Failure		500	{string}	string	"Internal Server Error"
// @Router			/api/v1/accounts [get]
func HandleListAccounts(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		return
	}

	accounts, err := deps.Domains().Accounts().ListAccounts(c.Request().Context())
	if err != nil {
		deps.Logger().WithError(err).Error("error getting accounts")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusOK, accounts)
}

// @Summary	Create an account
// @Tags		accounts
// @Accept		json
// @Produce	json
// @Success	201	{object}	model.AccountDTO
// @Failure	400	{object}	nil	"Bad Request"
// @Failure	409	{object}	nil	"Account already exists"
// @Failure	500	{object}	nil	"Internal Server Error"
// @Router		/api/v1/accounts [post]
func HandleCreateAccount(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		return
	}

	var payload createAccountPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "invalid json", nil)
		return
	}

	account, err := deps.Domains().Accounts().CreateAccount(c.Request().Context(), payload.ToAccountDTO())
	if err, isValidationErr := err.(model.ValidationError); isValidationErr {
		response.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if errors.Is(err, model.ErrAlreadyExists) {
		response.SendError(c, http.StatusConflict, "account already exists", nil)
		return
	}

	if err != nil {
		deps.Logger().WithError(err).Error("error creating account")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusCreated, account)
}

// @Summary	Delete an account
// @Tags		accounts
// @Produce	json
// @Param		id	path		int	true	"Account ID"
// @Success	204	{object}	nil	"No content"
// @Failure	400	{object}	nil	"Invalid ID"
// @Failure	404	{object}	nil	"Account not found"
// @Failure	500	{object}	nil	"Internal Server Error"
// @Router		/api/v1/accounts/{id} [delete]
func HandleDeleteAccount(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		return
	}

	id, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "invalid id", nil)
		return
	}

	err = deps.Domains().Accounts().DeleteAccount(c.Request().Context(), id)
	if errors.Is(err, model.ErrNotFound) {
		response.SendError(c, http.StatusNotFound, "account not found", nil)
		return
	}

	if err != nil {
		deps.Logger().WithError(err).Error("error deleting account")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusNoContent, nil)
}

// @Summary	Update an account
// @Tags		accounts
// @Accept		json
// @Produce	json
// @Param		id		path		int						true	"Account ID"
// @Param		account	body		updateAccountPayload	true	"Account data"
// @Success	200		{object}	model.AccountDTO
// @Failure	400		{object}	nil	"Invalid ID/data"
// @Failure	404		{object}	nil	"Account not found"
// @Failure	409		{object}	nil	"Account already exists"
// @Failure	500		{object}	nil	"Internal Server Error"
// @Router		/api/v1/accounts/{id} [patch]
func HandleUpdateAccount(deps model.Dependencies, c model.WebContext) {
	if err := middleware.RequireLoggedInAdmin(deps, c); err != nil {
		return
	}

	accountID, err := strconv.Atoi(c.Request().PathValue("id"))
	if err != nil {
		response.SendError(c, http.StatusBadRequest, "invalid id", nil)
		return
	}

	var payload updateAccountPayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		response.SendError(c, http.StatusBadRequest, "invalid json", nil)
		return
	}

	updatedAccount := payload.ToAccountDTO()
	updatedAccount.ID = model.DBID(accountID)

	account, err := deps.Domains().Accounts().UpdateAccount(c.Request().Context(), updatedAccount)
	if errors.Is(err, model.ErrNotFound) {
		response.SendError(c, http.StatusNotFound, "account not found", nil)
		return
	}
	if errors.Is(err, model.ErrAlreadyExists) {
		response.SendError(c, http.StatusConflict, "account already exists", nil)
		return
	}
	if err, isValidationErr := err.(model.ValidationError); isValidationErr {
		response.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err != nil {
		deps.Logger().WithError(err).Error("error updating account")
		response.SendInternalServerError(c)
		return
	}

	response.Send(c, http.StatusOK, account)
}
