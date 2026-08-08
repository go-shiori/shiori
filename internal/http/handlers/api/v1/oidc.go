package api_v1

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
	"golang.org/x/oauth2"
)

// @Summary	Redirect to OIDC provider for login
// @Tags		Auth
// @Router		/api/v1/auth/oidc/login [get]
func HandleOIDCLogin(deps model.Dependencies, c model.WebContext) {
	cfg := deps.Config().Http
	if !cfg.OIDCEnabled {
		response.SendError(c, http.StatusNotFound, "OIDC is not enabled")
		return
	}

	provider, err := oidc.NewProvider(c.Request().Context(), cfg.OIDCIssuer)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to get OIDC provider")
		response.SendInternalServerError(c)
		return
	}

	scopes := []string{oidc.ScopeOpenID}
	if cfg.OIDCScopes != "" {
		scopes = append(scopes, strings.Split(cfg.OIDCScopes, ",")...)
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		RedirectURL:  cfg.OIDCRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	state := generateState()
	// Store state in a cookie
	http.SetCookie(c.ResponseWriter(), &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
		MaxAge:   int(time.Minute * 5 / time.Second),
	})

	http.Redirect(c.ResponseWriter(), c.Request(), oauth2Config.AuthCodeURL(state), http.StatusFound)
}

// @Summary	OIDC callback URL
// @Tags		Auth
// @Router		/api/v1/auth/oidc/callback [get]
func HandleOIDCCallback(deps model.Dependencies, c model.WebContext) {
	cfg := deps.Config().Http
	if !cfg.OIDCEnabled {
		response.SendError(c, http.StatusNotFound, "OIDC is not enabled")
		return
	}

	// Verify state
	stateCookie, err := c.Request().Cookie("oidc_state")
	if err != nil || stateCookie.Value != c.Request().URL.Query().Get("state") {
		response.SendError(c, http.StatusBadRequest, "Invalid state")
		return
	}

	// Delete state cookie
	http.SetCookie(c.ResponseWriter(), &http.Cookie{
		Name:   "oidc_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	provider, err := oidc.NewProvider(c.Request().Context(), cfg.OIDCIssuer)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to get OIDC provider")
		response.SendInternalServerError(c)
		return
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		RedirectURL:  cfg.OIDCRedirectURL,
		Endpoint:     provider.Endpoint(),
	}

	token, err := oauth2Config.Exchange(c.Request().Context(), c.Request().URL.Query().Get("code"))
	if err != nil {
		deps.Logger().WithError(err).Error("failed to exchange token")
		response.SendInternalServerError(c)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		deps.Logger().Error("no id_token in token response")
		response.SendInternalServerError(c)
		return
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID})
	idToken, err := verifier.Verify(c.Request().Context(), rawIDToken)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to verify id token")
		response.SendInternalServerError(c)
		return
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		deps.Logger().WithError(err).Error("failed to parse claims")
		response.SendInternalServerError(c)
		return
	}

	username, _ := claims[cfg.OIDCUsernameClaim].(string)
	if username == "" {
		deps.Logger().Errorf("username claim %s not found in id token", cfg.OIDCUsernameClaim)
		response.SendError(c, http.StatusBadRequest, "Username not found in OIDC claims")
		return
	}

	// Get or create account
	account, err := deps.Domains().Accounts().GetAccountByUsername(c.Request().Context(), username)
	if err != nil {
		if cfg.OIDCAutoRegister {
			// Create account
			newAccount := model.AccountDTO{
				Username: username,
				Password: generateState(), // Random password, won't be used
			}
			account, err = deps.Domains().Accounts().CreateAccount(c.Request().Context(), newAccount)
			if err != nil {
				deps.Logger().WithError(err).Error("failed to create account")
				response.SendInternalServerError(c)
				return
			}
		} else {
			deps.Logger().WithError(err).Errorf("account %s not found and auto-registration is disabled", username)
			response.SendError(c, http.StatusForbidden, "Account not found")
			return
		}
	}

	// Create session token
	expirationTime := time.Now().Add(time.Hour * 24 * 30) // 30 days
	shioriToken, err := deps.Domains().Auth().CreateTokenForAccount(account, expirationTime)
	if err != nil {
		deps.Logger().WithError(err).Error("failed to create shiori token")
		response.SendInternalServerError(c)
		return
	}

	// Redirect back to UI with token
	redirectURL := fmt.Sprintf("/login?token=%s&expires=%d", shioriToken, expirationTime.Unix())
	http.Redirect(c.ResponseWriter(), c.Request(), redirectURL, http.StatusFound)
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
