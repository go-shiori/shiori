package api_v1

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandleListAccounts(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		HandleListAccounts(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("requires admin access", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		HandleListAccounts(deps, c)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("database error", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)

		// Force DB error by closing connection
		deps.Database().ReaderDB().Close()

		HandleListAccounts(deps, c)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("returns accounts list", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create test account
		_, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		HandleListAccounts(deps, c)
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)
		require.Len(t, response.Response.Message, 1) // Admin + created account
	})
}

func TestHandleCreateAccount(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		HandleCreateAccount(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("requires admin access", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		HandleCreateAccount(deps, c)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `invalid json`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetFakeAdmin(c)
			HandleCreateAccount(deps, c)
		}, "POST", "/api/v1/accounts", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("database error", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Force DB error
		deps.Database().WriterDB().Close()

		body := `{
			"username": "gopher",
			"password": "shiori"
		}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetFakeAdmin(c)
			HandleCreateAccount(deps, c)
		}, "POST", "/api/v1/accounts", testutil.WithBody(body))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("account already exists", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create first account
		_, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "gopher",
			Password: "shiori",
		})
		require.NoError(t, err)

		// Try to create duplicate account
		body := `{
			"username": "gopher",
			"password": "shiori"
		}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetFakeAdmin(c)
			HandleCreateAccount(deps, c)
		}, "POST", "/api/v1/accounts", testutil.WithBody(body))
		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("successful creation", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{
			"username": "newuser",
			"password": "password",
			"owner": false
		}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetFakeAdmin(c)
			HandleCreateAccount(deps, c)
		}, "POST", "/api/v1/accounts", testutil.WithBody(body))
		require.Equal(t, http.StatusCreated, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)
		response.AssertMessageContains(t, "id")
		require.NotZero(t, response.Response.Message.(map[string]interface{})["id"])
	})
}

func TestHandleDeleteAccount(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		HandleDeleteAccount(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("requires admin access", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		HandleDeleteAccount(deps, c)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		testutil.SetRequestPathValue(c, "id", "invalid")
		HandleDeleteAccount(deps, c)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("account not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		testutil.SetRequestPathValue(c, "id", "999")
		HandleDeleteAccount(deps, c)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successful deletion", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create account to delete
		account, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "todelete",
			Password: "password",
		})
		require.NoError(t, err)

		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		testutil.SetRequestPathValue(c, "id", strconv.Itoa(int(account.ID)))
		HandleDeleteAccount(deps, c)
		require.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestHandleUpdateAccount(t *testing.T) {
	logger := logrus.New()
	ctx := context.Background()

	t.Run("requires authentication", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		HandleUpdateAccount(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("requires admin access", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		HandleUpdateAccount(deps, c)
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		testutil.SetRequestPathValue(c, "id", "invalid")
		HandleUpdateAccount(deps, c)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `invalid json`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetFakeAdmin(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/1", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("account not found", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{"username": "newname"}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetRequestPathValue(c, "id", "999")
			testutil.SetFakeAdmin(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/999", testutil.WithBody(body))
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successful update", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Create account to update
		account, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "shiori",
			Password: "gopher",
		})
		require.NoError(t, err)

		body := `{
			"username": "updated",
			"owner": true
		}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetRequestPathValue(c, "id", strconv.Itoa(int(account.ID)))
			testutil.SetFakeAdmin(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/"+strconv.Itoa(int(account.ID)), testutil.WithBody(body))
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)
		response.AssertMessageContains(t, "owner")
		require.True(t, response.Response.Message.(map[string]any)["owner"].(bool))
	})

	t.Run("update with empty payload", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		account, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "shiori",
			Password: "gopher",
			Owner:    model.Ptr(false),
			Config: model.Ptr(model.UserConfig{
				ShowId:        true,
				ListMode:      true,
				HideThumbnail: true,
			}),
		})
		require.NoError(t, err)

		body := `{}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetRequestPathValue(c, "id", strconv.Itoa(int(account.ID)))
			testutil.SetFakeAdmin(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/"+strconv.Itoa(int(account.ID)), testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)

		// Verify no changes were made
		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertNotOk(t)
	})

	t.Run("update username only", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		account, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "shiori",
			Password: "gopher",
		})
		require.NoError(t, err)

		body := `{"username": "newname"}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetRequestPathValue(c, "id", strconv.Itoa(int(account.ID)))
			testutil.SetFakeAdmin(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/"+strconv.Itoa(int(account.ID)), testutil.WithBody(body))
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)
		require.Equal(t, "newname", response.Response.Message.(map[string]any)["username"])
	})

	t.Run("update password only", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		account, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "shiori",
			Password: "gopher",
		})
		require.NoError(t, err)

		body := `{"new_password": "newpass"}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetRequestPathValue(c, "id", strconv.Itoa(int(account.ID)))
			testutil.SetFakeAdmin(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/"+strconv.Itoa(int(account.ID)), testutil.WithBody(body))
		require.Equal(t, http.StatusOK, w.Code)

		// Verify we can login with new password
		loginBody := `{"username": "shiori", "password": "newpass"}`
		w = testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(loginBody))
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("only admin can update other's passwords", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		account, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "shiori",
			Password: "gopher",
		})
		require.NoError(t, err)

		body := `{"new_password": "newpass"}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetRequestPathValue(c, "id", strconv.Itoa(int(account.ID)))
			testutil.SetFakeUser(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/"+strconv.Itoa(int(account.ID)), testutil.WithBody(body))
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("update config only", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		account, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "shiori",
			Password: "gopher",
			Config: model.Ptr(model.UserConfig{
				ShowId:   false,
				ListMode: false,
			}),
		})
		require.NoError(t, err)

		body := `{
			"config": {
				"ShowId": true,
				"ListMode": true,
				"HideThumbnail": true,
				"HideExcerpt": true,
				"Theme": "dark",
				"KeepMetadata": true,
				"UseArchive": true,
				"CreateEbook": true,
				"MakePublic": true
			}
		}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetRequestPathValue(c, "id", strconv.Itoa(int(account.ID)))
			testutil.SetFakeAdmin(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/"+strconv.Itoa(int(account.ID)), testutil.WithBody(body))
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)

		config := response.Response.Message.(map[string]any)["config"].(map[string]any)
		require.True(t, config["ShowId"].(bool))
		require.True(t, config["ListMode"].(bool))
		require.True(t, config["HideThumbnail"].(bool))
		require.True(t, config["HideExcerpt"].(bool))
		require.Equal(t, "dark", config["Theme"])
		require.True(t, config["KeepMetadata"].(bool))
		require.True(t, config["UseArchive"].(bool))
		require.True(t, config["CreateEbook"].(bool))
		require.True(t, config["MakePublic"].(bool))
	})

	t.Run("update all fields", func(t *testing.T) {
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		account, err := deps.Domains().Accounts().CreateAccount(ctx, model.AccountDTO{
			Username: "shiori",
			Password: "gopher",
			Owner:    model.Ptr(false),
			Config: model.Ptr(model.UserConfig{
				ShowId:   false,
				ListMode: false,
			}),
		})
		require.NoError(t, err)

		body := `{
			"username": "updated",
			"new_password": "newpass",
			"owner": true,
			"config": {
				"ShowId": true,
				"ListMode": true,
				"HideThumbnail": true,
				"HideExcerpt": true,
				"Theme": "dark"
			}
		}`
		w := testutil.PerformRequest(deps, func(deps model.Dependencies, c model.WebContext) {
			testutil.SetRequestPathValue(c, "id", strconv.Itoa(int(account.ID)))
			testutil.SetFakeAdmin(c)
			HandleUpdateAccount(deps, c)
		}, "PATCH", "/api/v1/accounts/"+strconv.Itoa(int(account.ID)), testutil.WithBody(body))
		require.Equal(t, http.StatusOK, w.Code)

		response, err := testutil.NewTestResponseFromReader(w.Body)
		require.NoError(t, err)
		response.AssertOk(t)

		msg := response.Response.Message.(map[string]any)
		require.Equal(t, "updated", msg["username"])
		require.True(t, msg["owner"].(bool))

		config := msg["config"].(map[string]any)
		require.True(t, config["ShowId"].(bool))
		require.True(t, config["ListMode"].(bool))
		require.True(t, config["HideThumbnail"].(bool))
		require.True(t, config["HideExcerpt"].(bool))
		require.Equal(t, "dark", config["Theme"])

		// Verify password change
		loginBody := `{"username": "updated", "password": "newpass"}`
		w = testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(loginBody))
		require.Equal(t, http.StatusOK, w.Code)
	})
}
