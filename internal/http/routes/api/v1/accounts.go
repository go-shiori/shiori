package api_v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type AccountsAPIRoutes struct {
	logger *logrus.Logger
	deps   *dependencies.Dependencies
}

func (r *AccountsAPIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	g.Use(middleware.AdminRequired())
	g.GET("/", r.listHandler)
	g.POST("/", r.createHandler)
	g.DELETE("/:id", r.deleteHandler)
	g.PATCH("/:id", r.updateHandler)

	return r
}

func NewAccountsAPIRoutes(logger *logrus.Logger, deps *dependencies.Dependencies) *AccountsAPIRoutes {
	return &AccountsAPIRoutes{
		logger: logger,
		deps:   deps,
	}
}

// listHandler godoc
//
//	@Summary		List accounts
//	@Description	List accounts
//	@Tags			accounts
//	@Produce		json
//	@Success		200	{array}		model.AccountDTO
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/api/v1/accounts [get]
func (r *AccountsAPIRoutes) listHandler(c *gin.Context) {
	accounts, err := r.deps.Domains.Accounts.ListAccounts(c.Request.Context())
	if err != nil {
		r.logger.WithError(err).Error("error getting accounts")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response.Send(c, http.StatusOK, accounts)
}

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

// createHandler godoc
//
//	@Summary	Create an account
//	@Tags		accounts
//	@Produce	json
//	@Success	201	{array}		model.AccountDTO
//	@Failure	400	{string}	string	"Bad Request"
//	@Failure	500	{string}	string	"Internal Server Error"
//	@Router		/api/v1/accounts [post]
func (r *AccountsAPIRoutes) createHandler(c *gin.Context) {
	var payload createAccountPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		r.logger.WithError(err).Error("error parsing json")
		response.SendError(c, http.StatusBadRequest, "invalid json")
		return
	}

	account, err := r.deps.Domains.Accounts.CreateAccount(c.Request.Context(), payload.ToAccountDTO())
	if err, isValidationErr := err.(model.ValidationError); isValidationErr {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, model.ErrAlreadyExists) {
		response.SendError(c, http.StatusConflict, "account already exists")
		return
	}

	if err != nil {
		r.logger.WithError(err).Error("error creating account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response.Send(c, http.StatusCreated, account)
}

// deleteHandler godoc
//
//	@Summary	Delete an account
//	@Tags		accounts
//	@Produce	json
//	@Success	204	{string}	string	"No content"
//	@Failure	500	{string}	string	"Internal Server Error"
//	@Router		/api/v1/accounts/{id} [delete]
func (r *AccountsAPIRoutes) deleteHandler(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		r.logger.WithError(err).Error("error parsing id")
		response.SendError(c, http.StatusBadRequest, "invalid id")
		return
	}

	err = r.deps.Domains.Accounts.DeleteAccount(c.Request.Context(), id)
	if errors.Is(err, model.ErrNotFound) {
		response.SendError(c, http.StatusNotFound, "account not found")
		return
	}

	if err != nil {
		r.logger.WithError(err).Error("error deleting account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response.Send(c, http.StatusNoContent, nil)
}

// updateHandler godoc
//
//	@Summary	Update an account
//	@Tags		accounts
//	@Produce	json
//	@Success	200	{array}		updateAccountPayload
//	@Failure	400	{string}	string	"Bad Request"
//	@Failure	500	{string}	string	"Internal Server Error"
//	@Router		/api/v1/accounts/{id} [patch]
func (r *AccountsAPIRoutes) updateHandler(c *gin.Context) {
	accountID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		r.logger.WithError(err).Error("error parsing id")
		response.SendError(c, http.StatusBadRequest, "invalid id")
		return
	}

	var payload updateAccountPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		r.logger.WithError(err).Error("error binding json")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Not checking the old password since admins/owners can update any account
	updatedAccount := payload.ToAccountDTO()
	updatedAccount.ID = model.DBID(accountID)

	account, err := r.deps.Domains.Accounts.UpdateAccount(c.Request.Context(), updatedAccount)
	if errors.Is(err, model.ErrNotFound) {
		response.SendError(c, http.StatusNotFound, "account not found")
		return
	}
	if errors.Is(err, model.ErrAlreadyExists) {
		response.SendError(c, http.StatusConflict, "account already exists")
		return
	}

	if err, isValidationErr := err.(model.ValidationError); isValidationErr {
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err != nil {
		r.logger.WithError(err).Error("error updating account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response.Send(c, http.StatusOK, account)
}
