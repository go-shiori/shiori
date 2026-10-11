package api_v1

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

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
		body := `{"Username":}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing username", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{"Password": "test"}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing password", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{"Username": "test"}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		body := `{"Username": "test", "Password": "wrong"}`
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
			"Username": "test",
			"Password": "test",
			"RememberMe": true
		}`
		w := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body))
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "Token", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
		response.AssertMessageJSONKeyValue(t, "Expiration", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
	})

	t.Run("remember_me sets correct expiration", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)

		// Test with remember_me=false (should be 1 hour)
		account1 := testutil.GetValidAccount().ToDTO()
		account1.Username = "test1"
		account1.Password = "test"
		createdAccount1, err := deps.Domains().Accounts().CreateAccount(context.Background(), account1)
		require.NoError(t, err)

		body1 := `{
			"Username": "test1",
			"Password": "test",
			"RememberMe": false
		}`
		w1 := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body1))
		require.Equal(t, http.StatusOK, w1.Code)

		response1 := testutil.NewTestResponseFromRecorder(w1)
		var loginResp1 struct {
			Token      string
			Expiration int64
		}
		err = json.Unmarshal(response1.Response.GetData().([]byte), &loginResp1)
		require.NoError(t, err)

		now := time.Now()
		expectedExpiration1 := now.Add(time.Hour)
		actualExpiration1 := time.Unix(loginResp1.Expiration, 0)

		// Allow 5 seconds tolerance for test execution time
		diff1 := actualExpiration1.Sub(expectedExpiration1)
		require.True(t, diff1 < 5*time.Second && diff1 > -5*time.Second, "Expected expiration around 1 hour from now, got %v", diff1)

		// Verify token is valid and can be checked
		accountFromToken1, err := deps.Domains().Auth().CheckToken(ctx, loginResp1.Token)
		require.NoError(t, err)
		require.NotNil(t, accountFromToken1)
		require.Equal(t, createdAccount1.ID, accountFromToken1.ID)

		// Test with remember_me=true (should be 30 days)
		account2 := testutil.GetValidAccount().ToDTO()
		account2.Username = "test2"
		account2.Password = "test"
		createdAccount2, err := deps.Domains().Accounts().CreateAccount(context.Background(), account2)
		require.NoError(t, err)

		body2 := `{
			"Username": "test2",
			"Password": "test",
			"RememberMe": true
		}`
		w2 := testutil.PerformRequest(deps, HandleLogin, "POST", "/login", testutil.WithBody(body2))
		require.Equal(t, http.StatusOK, w2.Code)

		response2 := testutil.NewTestResponseFromRecorder(w2)
		var loginResp2 struct {
			Token      string
			Expiration int64
		}
		err = json.Unmarshal(response2.Response.GetData().([]byte), &loginResp2)
		require.NoError(t, err)

		expectedExpiration2 := now.Add(time.Hour * 24 * 30)
		actualExpiration2 := time.Unix(loginResp2.Expiration, 0)

		// Allow 5 seconds tolerance for test execution time
		diff2 := actualExpiration2.Sub(expectedExpiration2)
		require.True(t, diff2 < 5*time.Second && diff2 > -5*time.Second, "Expected expiration around 30 days from now, got %v", diff2)

		// Verify token is valid and can be checked
		accountFromToken2, err := deps.Domains().Auth().CheckToken(ctx, loginResp2.Token)
		require.NoError(t, err)
		require.NotNil(t, accountFromToken2)
		require.Equal(t, createdAccount2.ID, accountFromToken2.ID)
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
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "Token", func(t *testing.T, value any) {
			require.NotEmpty(t, value)
		})
		response.AssertMessageJSONKeyValue(t, "Expiration", func(t *testing.T, value any) {
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
		response.AssertMessageJSONKeyValue(t, "Username", func(t *testing.T, value any) {
			require.Equal(t, "user", value)
		})
		response.AssertMessageJSONKeyValue(t, "Owner", func(t *testing.T, value any) {
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
		response.AssertMessageJSONKeyValue(t, "Username", func(t *testing.T, value any) {
			require.Equal(t, "user", value)
		})
		response.AssertMessageJSONKeyValue(t, "Owner", func(t *testing.T, value any) {
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
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing old password", func(t *testing.T) {
		body := `{"NewPassword": "newpass"}`
		w := testutil.PerformRequest(deps, HandleUpdateLoggedAccount, "PATCH", "/account", testutil.WithBody(body), testutil.WithAccount(account))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("incorrect old password", func(t *testing.T) {
		body := `{
			"OldPassword": "wrong",
			"NewPassword": "newpass"
		}`
		w := testutil.PerformRequest(deps, HandleUpdateLoggedAccount, "PATCH", "/account", testutil.WithBody(body), testutil.WithAccount(account))
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("deleted account returns 404", func(t *testing.T) {
		body := `{"Config": {"ShowId": true}}`
		deleted := &model.AccountDTO{ID: 9999, Username: "deleted", Owner: model.Ptr(false)}
		w := testutil.PerformRequest(deps, HandleUpdateLoggedAccount, "PATCH", "/account", testutil.WithBody(body), testutil.WithAccount(deleted))
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successful update", func(t *testing.T) {
		body := `{
			"OldPassword": "gopher",
			"NewPassword": "newpass",
			"Config": {
				"ShowId": true,
				"ListMode": true
			}
		}`
		w := testutil.PerformRequest(deps, HandleUpdateLoggedAccount, "PATCH", "/account", testutil.WithBody(body), testutil.WithAccount(account))
		require.Equal(t, http.StatusOK, w.Code)

		response := testutil.NewTestResponseFromRecorder(w)
		response.AssertOk(t)
		response.AssertMessageJSONKeyValue(t, "Username", func(t *testing.T, value any) {
			require.Equal(t, "shiori", value)
		})
		response.AssertMessageJSONKeyValue(t, "Config", func(t *testing.T, value any) {
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
