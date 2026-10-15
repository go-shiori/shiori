package api_v1

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandleOIDCLogin(t *testing.T) {
	logger := logrus.New()

	t.Run("OIDC disabled", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = false

		w := testutil.PerformRequest(deps, HandleOIDCLogin, "GET", "/api/v1/auth/oidc/login")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("OIDC enabled but provider fails", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = "http://invalid-issuer"

		w := testutil.PerformRequest(deps, HandleOIDCLogin, "GET", "/api/v1/auth/oidc/login")
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Successful redirect", func(t *testing.T) {
		var ts *httptest.Server
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/.well-known/openid-configuration" {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{
					"issuer": "%s",
					"authorization_endpoint": "%s/auth",
					"token_endpoint": "%s/token",
					"jwks_uri": "%s/keys",
					"id_token_signing_alg_values_supported": ["RS256"]
				}`, ts.URL, ts.URL, ts.URL, ts.URL)
				return
			}
		}))
		defer ts.Close()

		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = ts.URL
		deps.Config().Http.OIDCClientID = "test-client"
		deps.Config().Http.OIDCScopes = "profile,email"

		w := testutil.PerformRequest(deps, HandleOIDCLogin, "GET", "/api/v1/auth/oidc/login")
		require.Equal(t, http.StatusFound, w.Code)
		require.Contains(t, w.Header().Get("Location"), "response_type=code")
		require.Contains(t, w.Header().Get("Location"), "client_id=test-client")
		require.Contains(t, w.Header().Get("Set-Cookie"), "oidc_state=")
	})
}

func TestHandleOIDCCallback(t *testing.T) {
	logger := logrus.New()

	t.Run("OIDC disabled", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = false

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback")
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Missing state cookie", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=test")
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid state", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=wrong", func(c model.WebContext) {
			c.Request().AddCookie(&http.Cookie{Name: "oidc_state", Value: "correct"})
		})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("OIDC enabled but provider fails", func(t *testing.T) {
		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = "http://invalid-issuer"

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=test", func(c model.WebContext) {
			c.Request().AddCookie(&http.Cookie{Name: "oidc_state", Value: "test"})
		})
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Full flow success", func(t *testing.T) {
		// Generate RSA key for signing JWT
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		var ts *httptest.Server
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/.well-known/openid-configuration":
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{
					"issuer": "%s",
					"authorization_endpoint": "%s/auth",
					"token_endpoint": "%s/token",
					"jwks_uri": "%s/keys",
					"id_token_signing_alg_values_supported": ["RS256"]
				}`, ts.URL, ts.URL, ts.URL, ts.URL)
			case "/keys":
				w.Header().Set("Content-Type", "application/json")
				key := jose.JSONWebKey{
					Key:       &privateKey.PublicKey,
					KeyID:     "test-key-id",
					Algorithm: "RS256",
					Use:       "sig",
				}
				jwks := jose.JSONWebKeySet{
					Keys: []jose.JSONWebKey{key},
				}
				json.NewEncoder(w).Encode(jwks)
			case "/token":
				w.Header().Set("Content-Type", "application/json")
				// Create ID Token
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"iss":                ts.URL,
					"sub":                "test-user",
					"aud":                "test-client",
					"exp":                time.Now().Add(time.Hour).Unix(),
					"iat":                time.Now().Unix(),
					"preferred_username": "test-user",
				})
				token.Header["kid"] = "test-key-id"
				tokenString, _ := token.SignedString(privateKey)

				fmt.Fprintf(w, `{
					"access_token": "access-token",
					"id_token": "%s",
					"token_type": "Bearer",
					"expires_in": 3600
				}`, tokenString)
			}
		}))
		defer ts.Close()

		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = ts.URL
		deps.Config().Http.OIDCClientID = "test-client"
		deps.Config().Http.OIDCUsernameClaim = "preferred_username"
		deps.Config().Http.OIDCAutoRegister = true

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=test&code=test-code", func(c model.WebContext) {
			c.Request().AddCookie(&http.Cookie{Name: "oidc_state", Value: "test"})
		})

		require.Equal(t, http.StatusFound, w.Code)
		require.Contains(t, w.Header().Get("Location"), "/login?token=")

		// Verify account was created
		account, err := deps.Domains().Accounts().GetAccountByUsername(ctx, "test-user")
		require.NoError(t, err)
		require.Equal(t, "test-user", account.Username)
	})

	t.Run("Full flow - account not found and no auto-register", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		var ts *httptest.Server
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/.well-known/openid-configuration":
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{
					"issuer": "%s",
					"authorization_endpoint": "%s/auth",
					"token_endpoint": "%s/token",
					"jwks_uri": "%s/keys",
					"id_token_signing_alg_values_supported": ["RS256"]
				}`, ts.URL, ts.URL, ts.URL, ts.URL)
			case "/keys":
				w.Header().Set("Content-Type", "application/json")
				key := jose.JSONWebKey{Key: &privateKey.PublicKey, KeyID: "test-key-id", Algorithm: "RS256", Use: "sig"}
				jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{key}}
				json.NewEncoder(w).Encode(jwks)
			case "/token":
				w.Header().Set("Content-Type", "application/json")
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"iss": ts.URL, "sub": "test-user-no-reg", "aud": "test-client",
					"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
					"preferred_username": "test-user-no-reg",
				})
				token.Header["kid"] = "test-key-id"
				tokenString, _ := token.SignedString(privateKey)
				fmt.Fprintf(w, `{"access_token": "access-token", "id_token": "%s", "token_type": "Bearer"}`, tokenString)
			}
		}))
		defer ts.Close()

		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = ts.URL
		deps.Config().Http.OIDCClientID = "test-client"
		deps.Config().Http.OIDCUsernameClaim = "preferred_username"
		deps.Config().Http.OIDCAutoRegister = false

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=test&code=test-code", func(c model.WebContext) {
			c.Request().AddCookie(&http.Cookie{Name: "oidc_state", Value: "test"})
		})

		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Full flow - token exchange failure", func(t *testing.T) {
		var ts *httptest.Server
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/.well-known/openid-configuration":
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"issuer": "%s", "token_endpoint": "%s/token"}`, ts.URL, ts.URL)
			case "/token":
				w.WriteHeader(http.StatusBadRequest)
			}
		}))
		defer ts.Close()

		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = ts.URL

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=test&code=test-code", func(c model.WebContext) {
			c.Request().AddCookie(&http.Cookie{Name: "oidc_state", Value: "test"})
		})

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Full flow - missing username claim", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		var ts *httptest.Server
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/.well-known/openid-configuration":
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"issuer": "%s", "token_endpoint": "%s/token", "jwks_uri": "%s/keys"}`, ts.URL, ts.URL, ts.URL)
			case "/keys":
				w.Header().Set("Content-Type", "application/json")
				key := jose.JSONWebKey{Key: &privateKey.PublicKey, KeyID: "test-key-id", Algorithm: "RS256", Use: "sig"}
				json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{key}})
			case "/token":
				w.Header().Set("Content-Type", "application/json")
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"iss": ts.URL, "sub": "test-user", "aud": "test-client",
					"exp": time.Now().Add(time.Hour).Unix(),
				})
				token.Header["kid"] = "test-key-id"
				tokenString, _ := token.SignedString(privateKey)
				fmt.Fprintf(w, `{"access_token": "foo", "id_token": "%s"}`, tokenString)
			}
		}))
		defer ts.Close()

		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = ts.URL
		deps.Config().Http.OIDCClientID = "test-client"
		deps.Config().Http.OIDCUsernameClaim = "missing"

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=test&code=test-code", func(c model.WebContext) {
			c.Request().AddCookie(&http.Cookie{Name: "oidc_state", Value: "test"})
		})

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Full flow - id token verification failure", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		var ts *httptest.Server
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/.well-known/openid-configuration":
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"issuer": "%s", "token_endpoint": "%s/token", "jwks_uri": "%s/keys"}`, ts.URL, ts.URL, ts.URL)
			case "/keys":
				w.Header().Set("Content-Type", "application/json")
				key := jose.JSONWebKey{Key: &privateKey.PublicKey, KeyID: "test-key-id", Algorithm: "RS256", Use: "sig"}
				json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{key}})
			case "/token":
				w.Header().Set("Content-Type", "application/json")
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"iss": ts.URL, "sub": "test-user", "aud": "test-client",
					"exp": time.Now().Add(time.Hour).Unix(),
				})
				token.Header["kid"] = "test-key-id"
				// Sign with WRONG key
				tokenString, _ := token.SignedString(otherKey)
				fmt.Fprintf(w, `{"access_token": "foo", "id_token": "%s"}`, tokenString)
			}
		}))
		defer ts.Close()

		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = ts.URL
		deps.Config().Http.OIDCClientID = "test-client"

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=test&code=test-code", func(c model.WebContext) {
			c.Request().AddCookie(&http.Cookie{Name: "oidc_state", Value: "test"})
		})

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Full flow - no id_token in token response", func(t *testing.T) {
		var ts *httptest.Server
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/.well-known/openid-configuration":
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"issuer": "%s", "token_endpoint": "%s/token"}`, ts.URL, ts.URL)
			case "/token":
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"access_token": "foo"}`)
			}
		}))
		defer ts.Close()

		ctx := context.Background()
		_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
		deps.Config().Http.OIDCEnabled = true
		deps.Config().Http.OIDCIssuer = ts.URL

		w := testutil.PerformRequest(deps, HandleOIDCCallback, "GET", "/api/v1/auth/oidc/callback?state=test&code=test-code", func(c model.WebContext) {
			c.Request().AddCookie(&http.Cookie{Name: "oidc_state", Value: "test"})
		})

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
