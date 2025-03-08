package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

// AuthMiddleware handles authentication for incoming request by checking the token
// from the Authorization header or the token cookie and setting the account in the
// request context.
type AuthMiddleware struct {
	deps model.Dependencies
}

func NewAuthMiddleware(deps model.Dependencies) *AuthMiddleware {
	return &AuthMiddleware{deps: deps}
}

func (m *AuthMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	token := getTokenFromHeader(c.Request())
	if token == "" {
		token = getTokenFromCookie(c.Request())
	}

	if token == "" {
		return nil
	}

	account, err := deps.Domains().Auth().CheckToken(c.Request().Context(), token)
	if err != nil {
		// If we fail to check token, remove the token cookie and redirect to login
		deps.Logger().WithError(err).WithField("request_id", c.GetRequestID()).Error("Failed to check token")
		http.SetCookie(c.ResponseWriter(), &http.Cookie{
			Name:   "token",
			Value:  "",
			MaxAge: -1,
		})
		return nil
	}

	c.SetAccount(account)
	return nil
}

func (m *AuthMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	return nil
}

// RequireLoggedInUser ensures a user is authenticated
func RequireLoggedInUser(deps model.Dependencies, c model.WebContext) error {
	if !c.UserIsLogged() {
		response.SendError(c, http.StatusUnauthorized, "Authentication required", nil)
		return fmt.Errorf("authentication required")
	}
	return nil
}

// RequireLoggedInAdmin ensures a user is authenticated and is an admin
func RequireLoggedInAdmin(deps model.Dependencies, c model.WebContext) error {
	account := c.GetAccount()
	if err := RequireLoggedInUser(deps, c); err != nil {
		return err
	}

	if !account.IsOwner() {
		response.SendError(c, http.StatusForbidden, "Admin access required", nil)
		return fmt.Errorf("admin access required")
	}

	return nil
}

// getTokenFromHeader returns the token from the Authorization header
func getTokenFromHeader(r *http.Request) string {
	authorization := r.Header.Get(model.AuthorizationHeader)
	if authorization == "" {
		return ""
	}

	authParts := strings.SplitN(authorization, " ", 2)
	if len(authParts) != 2 || authParts[0] != model.AuthorizationTokenType {
		return ""
	}

	return authParts[1]
}

// getTokenFromCookie returns the token from the token cookie
func getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err != nil {
		return ""
	}
	return cookie.Value
}
