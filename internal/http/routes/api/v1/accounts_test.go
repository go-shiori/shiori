package api_v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/http/middleware"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAccountRouteAuthorization(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("require authentication", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "GET", "/")
		require.Equal(t, http.StatusForbidden, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
	})

	t.Run("require admin user", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		token, err := deps.Domains.Auth.CreateTokenForAccount(account, time.Now().Add(time.Hour))
		require.NoError(t, err)

		w := testutil.PerformRequest(g, "GET", "/", testutil.WithAuthToken(token))
		require.Equal(t, http.StatusForbidden, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
	})
}

func TestAccountList(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("database error", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		// Force DB error by clearing the deps
		deps.Database.ReaderDB().Close()

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "GET", "/", testutil.WithAuthToken(token))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
	t.Run("return account", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "GET", "/", testutil.WithAuthToken(token))
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertOk(t)
		require.Len(t, response.Response.Message, 1)
	})

	t.Run("return accounts", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		_, err = deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "GET", "/", testutil.WithAuthToken(token))
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertOk(t)
		require.Len(t, response.Response.Message, 2)
	})

}

func TestAccountCreate(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("database error", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		// Force DB error by clearing the deps
		deps.Database.WriterDB().Close()

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "gopher",
			"password": "shiori"
		}`), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("duplicate username", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		// Create first account
		_, err = deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		// Try to create account with same username
		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "gopher",
			"password": "shiori"
		}`), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusConflict, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertNotOk(t)
	})

	t.Run("create owner account", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "gopher",
			"password": "shiori",
			"owner": true
		}`), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusCreated, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)

		require.NoError(t, err)
		require.True(t, response.Response.Message.(map[string]interface{})["owner"].(bool))
	})

	t.Run("invalid payload", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`invalid`), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusBadRequest, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
	})

	t.Run("create account ok", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "gopher",
			"password": "shiori"
		}`), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusCreated, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertOk(t)

	})

	t.Run("empty username", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "",
			"password": "shiori"
		}`), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusBadRequest, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
	})

	t.Run("empty password", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "POST", "/", testutil.WithBody(`{
			"username": "gopher",
			"password": ""
		}`), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusBadRequest, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)

		response.AssertNotOk(t)
	})
}

func TestAccountDelete(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("database error", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		// Force DB error by clearing the deps
		deps.Database.WriterDB().Close()

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "DELETE", "/"+strconv.Itoa(int(account.ID)), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("delete owner account", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		owner := true
		account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
			Owner:    &owner,
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "DELETE", "/"+strconv.Itoa(int(account.ID)), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "DELETE", "/"+strconv.Itoa(int(account.ID)), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("account not found", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))
		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "DELETE", "/99", testutil.WithAuthToken(token))
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))
		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "DELETE", "/invalid", testutil.WithAuthToken(token))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAccountUpdate(t *testing.T) {
	logger := logrus.New()
	ctx := context.TODO()

	t.Run("database error", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		// Close dataase connection to force error
		deps.Database.ReaderDB().Close()

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "PATCH", "/"+strconv.Itoa(int(account.ID)),
			testutil.WithBody(`{"username":"newname"}`),
			testutil.WithAuthToken(token))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update to existing username", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		// Create first account
		_, err = deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher1",
			Password: "shiori",
		})
		require.NoError(t, err)

		// Create second account
		account2, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher2",
			Password: "shiori",
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		// Try to update second account to first account's username
		w := testutil.PerformRequest(g, "PATCH", "/"+strconv.Itoa(int(account2.ID)),
			testutil.WithBody(`{"username":"gopher1"}`),
			testutil.WithAuthToken(token))
		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("update with empty changes", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "PATCH", "/"+strconv.Itoa(int(account.ID)),
			testutil.WithBody(`{}`),
			testutil.WithAuthToken(token))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
	for _, tc := range []struct {
		name    string
		payload updateAccountPayload
		code    int
		cmp     func(t *testing.T, initial *model.AccountDTO, payload updateAccountPayload, storedAccount model.Account)
	}{
		{
			name: "success change username",
			payload: updateAccountPayload{
				Username: "gopher2",
			},
			code: http.StatusOK,
			cmp: func(t *testing.T, initial *model.AccountDTO, payload updateAccountPayload, storedAccount model.Account) {
				require.Equal(t, payload.Username, storedAccount.Username)
			},
		},
		{
			name: "success change password",
			payload: updateAccountPayload{
				OldPassword: "gopher",
				NewPassword: "gopher2",
			},
			code: http.StatusOK,
			cmp: func(t *testing.T, initial *model.AccountDTO, payload updateAccountPayload, storedAccount model.Account) {
				require.NotEqual(t, initial.Password, storedAccount.Password)
			},
		},
		{
			name: "success change owner",
			payload: updateAccountPayload{
				Owner: model.Ptr(true),
			},
			code: http.StatusOK,
			cmp: func(t *testing.T, initial *model.AccountDTO, payload updateAccountPayload, storedAccount model.Account) {
				require.Equal(t, *payload.Owner, storedAccount.Owner)
			},
		},
		{
			name: "change entire account",
			payload: updateAccountPayload{
				Username:    "gopher2",
				NewPassword: "gopher2",
				Owner:       model.Ptr(true),
			},
			code: http.StatusOK,
			cmp: func(t *testing.T, initial *model.AccountDTO, payload updateAccountPayload, storedAccount model.Account) {
				require.Equal(t, payload.Username, storedAccount.Username)
				require.NotEqual(t, initial.Password, storedAccount.Password)
				require.Equal(t, *payload.Owner, storedAccount.Owner)
			},
		},
		{
			name:    "invalid update",
			payload: updateAccountPayload{},
			code:    http.StatusBadRequest,
			cmp: func(t *testing.T, initial *model.AccountDTO, payload updateAccountPayload, storedAccount model.Account) {
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gin.New()
			_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
			g.Use(middleware.AuthMiddleware(deps))

			_, token, err := testutil.NewAdminUser(deps)
			require.NoError(t, err)

			account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
				Username: "gopher",
				Password: "shiori",
			})
			require.NoError(t, err)

			router := NewAccountsAPIRoutes(logger, deps)
			router.Setup(g.Group("/"))

			body, err := json.Marshal(tc.payload)
			require.NoError(t, err)

			w := testutil.PerformRequest(g, "PATCH", "/"+strconv.Itoa(int(account.ID)), testutil.WithBody(string(body)), testutil.WithAuthToken(token))
			require.Equal(t, tc.code, w.Code)

			storedAccount, _, err := deps.Database.GetAccount(ctx, account.ID)
			require.NoError(t, err)

			tc.cmp(t, account, tc.payload, *storedAccount)
		})
	}

	t.Run("invalid payload", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))

		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)

		account, err := deps.Domains.Accounts.CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))

		w := testutil.PerformRequest(g, "PATCH", "/"+strconv.Itoa(int(account.ID)), testutil.WithBody(`invalid`), testutil.WithAuthToken(token))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("account not found", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))
		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "PATCH", "/99", testutil.WithAuthToken(token), testutil.WithBody(`{"username":"gopher"}`))
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		g := gin.New()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		g.Use(middleware.AuthMiddleware(deps))
		_, token, err := testutil.NewAdminUser(deps)
		require.NoError(t, err)
		router := NewAccountsAPIRoutes(logger, deps)
		router.Setup(g.Group("/"))
		w := testutil.PerformRequest(g, "PATCH", "/invalid", testutil.WithAuthToken(token))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}
