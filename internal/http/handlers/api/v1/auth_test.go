package api_v1

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandleLogin(t *testing.T) {
	logger := logrus.New()
	// _, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("invalid json payload", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{"username":}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing username", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{"password": "test"}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing password", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{"username": "test"}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{"username": "test", "password": "wrong"}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("successful login", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		account := testutil.GetValidAccount().ToDTO()
		account.Password = "test"
		_, err := deps.Domains().Accounts().CreateAccount(context.Background(), account)
		require.NoError(t, err)

		body := `{
			"username": "test",
			"password": "test",
			"remember_me": true
		}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "token", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
		response.AssertMessageJSONKeyValue(t, "expires", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
	})
}

func TestHandleRefreshToken(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("requires authentication", func(t *testing.T) {
		w := testutil.PerformRequest(deps, HandleRefreshToken, "POST", "/refresh")
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("successful refresh", func(t *testing.T) {
		account := testutil.GetValidAccount().ToDTO()
		account.Password = "test"
		_, err := deps.Domains().Accounts().CreateAccount(context.Background(), account)
		require.NoError(t, err)

		w := testutil.PerformRequest(deps, HandleRefreshToken, "POST", "/refresh", testutil.WithAccount(&account))
		require.Equal(t, http.StatusAccepted, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "token", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
		response.AssertMessageJSONKeyValue(t, "expires", func(t *testing.T, value any) {
			require.NotZero(t, value)
		})
	})
}

func TestHandleGetMe(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("requires authentication", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		HandleGetMe(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns user info", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		HandleGetMe(deps, c)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "username", func(t *testing.T, value any) {
			require.Equal(t, "user", value)
		})
		response.AssertMessageJSONKeyValue(t, "owner", func(t *testing.T, value any) {
			require.False(t, value.(bool))
		})
	})

	t.Run("returns admin info", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeAdmin(c)
		HandleGetMe(deps, c)
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "username", func(t *testing.T, value any) {
			require.Equal(t, "user", value)
		})
		response.AssertMessageJSONKeyValue(t, "owner", func(t *testing.T, value any) {
			require.True(t, value.(bool))
		})
	})
}

func TestHandleUpdateLoggedAccount(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	account, err := deps.Domains().Accounts().CreateAccount(context.Background(), model.AccountDTO{
		Username: "shiori",
		Password: "gopher",
		Owner:    model.Ptr(true),
		Config: model.Ptr(model.UserConfig{
			ShowId:        true,
			ListMode:      true,
			HideThumbnail: true,
			HideExcerpt:   true,
			KeepMetadata:  true,
			UseArchive:    true,
			CreateEbook:   true,
			MakePublic:    true,
		}),
	})
	require.NoError(t, err)

	t.Run("requires authentication", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		HandleUpdateLoggedAccount(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json payload", func(t *testing.T) {
		body := `invalid json`
		w := testutil.PerformRequest(deps, HandleUpdateLoggedAccount, "PATCH", "/account", testutil.WithBody(body), testutil.WithAccount(account))
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("missing old password", func(t *testing.T) {
		body := `{"new_password": "newpass"}`
		w := testutil.PerformRequest(deps, HandleUpdateLoggedAccount, "PATCH", "/account", testutil.WithBody(body), testutil.WithAccount(account))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("incorrect old password", func(t *testing.T) {
		body := `{
			"old_password": "wrong",
			"new_password": "newpass"
		}`
		w := testutil.PerformRequest(deps, HandleUpdateLoggedAccount, "PATCH", "/account", testutil.WithBody(body), testutil.WithAccount(account))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("successful update", func(t *testing.T) {
		body := `{
			"old_password": "gopher",
			"new_password": "newpass",
			"config": {
				"ShowId": true,
				"ListMode": true
			}
		}`
		w := testutil.PerformRequest(deps, HandleUpdateLoggedAccount, "PATCH", "/account", testutil.WithBody(body), testutil.WithAccount(account))
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "username", func(t *testing.T, value any) {
			require.Equal(t, "shiori", value)
		})
		response.AssertMessageJSONKeyValue(t, "config", func(t *testing.T, value any) {
			config := value.(map[string]any)
			require.True(t, config["ShowId"].(bool))
			require.True(t, config["ListMode"].(bool))
		})
	})
}

func TestHandleLogout(t *testing.T) {
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, context.Background(), logger)

	t.Run("requires authentication", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		HandleLogout(deps, c)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("successful logout", func(t *testing.T) {
		c, w := testutil.NewTestWebContext()
		testutil.SetFakeUser(c)
		HandleLogout(deps, c)
		require.Equal(t, http.StatusOK, w.Code)
	})
}
