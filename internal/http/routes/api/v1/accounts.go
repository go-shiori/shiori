package api_v1

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type AccountsAPIRoutes struct {
	logger *logrus.Logger
	deps   *dependencies.Dependencies
}

func (r *AccountsAPIRoutes) Setup(g *gin.RouterGroup) model.Routes {
	g.GET("/", r.listHandler)
	g.POST("/", r.createHandler)
	g.DELETE("/:id", r.deleteHandler)
	// g.PUT("/:id", r.updateHandler)

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
// @Summary List accounts
// @Description List accounts
// @Tags accounts
// @Produce json
// @Success 200 {array} model.AccountDTO
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/accounts [get]
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
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	IsVisitor bool   `json:"is_visitor"`
}

func (p *createAccountPayload) IsValid() error {
	if p.Username == "" {
		return fmt.Errorf("username should not be empty")
	}
	if p.Password == "" {
		return fmt.Errorf("password should not be empty")
	}
	return nil
}

func (p *createAccountPayload) ToAccountDTO() model.AccountDTO {
	return model.AccountDTO{
		Username: p.Username,
		Password: p.Password,
		Owner:    !p.IsVisitor,
	}
}

// createHandler godoc
//
// @Summary Create an account
// @Tags accounts
// @Produce json
// @Success 201 {array} model.AccountDTO
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/accounts [post]
func (r *AccountsAPIRoutes) createHandler(c *gin.Context) {
	var payload createAccountPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		r.logger.WithError(err).Error("error binding json")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := payload.IsValid(); err != nil {
		r.logger.WithError(err).Error("error validating payload")
		response.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	account, err := r.deps.Domains.Accounts.CreateAccount(c.Request.Context(), payload.ToAccountDTO())
	if err != nil {
		r.logger.WithError(err).Error("error creating account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response.Send(c, http.StatusCreated, account)
}

// deleteHandler godoc
//
// @Summary Delete an account
// @Tags accounts
// @Produce json
// @Success 204 {string} string "No content"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/accounts/{id} [delete]
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

// func (r *AccountsAPIRoutes) updateHandler(c *gin.Context) {
// 	id := c.Param("id")

// 	var payload model.AccountDTO
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		r.logger.WithError(err).Error("error binding json")
// 		c.AbortWithStatus(http.StatusBadRequest)
// 		return
// 	}

// 	account, err := r.deps.Domains.Accounts.UpdateAccount(c.Request.Context(), id, payload)
// 	if err != nil {
// 		r.logger.WithError(err).Error("error updating account")
// 		c.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}

// 	response.Send(c, http.StatusOK, account)
// }
