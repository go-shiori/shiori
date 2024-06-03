package api_v1

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAccountList(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("empty account list", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "GET", "/")
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertMessageIsEmptyList(t)
	})

	t.Run("return account", func(t *testing.T) {
		ctx := context.TODO()

		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		_, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "GET", "/")
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertOk(t)
		require.Len(t, response.Response.Message, 1)
	})
}

func TestAccountCreate(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("create account ok", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "gopher",
			"password": "shiori"
		}`))
		require.Equal(t, http.StatusCreated, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertOk(t)

	})

	t.Run("empty username", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "",
			"password": "shiori"
		}`))
		require.Equal(t, http.StatusBadRequest, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
	})

	t.Run("empty password", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "gopher",
			"password": ""
		}`))
		require.Equal(t, http.StatusBadRequest, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
	})
}

func TestAccountDelete(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("success", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "DELETE", "/"+strconv.Itoa(int(account.ID)))
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("account not found", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "DELETE", "/99")
		require.Equal(t, http.StatusNotFound, w.Code)
	})
}
