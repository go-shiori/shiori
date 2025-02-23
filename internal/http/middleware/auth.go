package middleware

import (
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
		deps.Logger().WithError(err).Error("Failed to check token")
		return err
	}

	c.SetAccount(account)
	return nil
}

func (m *AuthMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	return nil
}

// RequireAuthMiddleware ensures a user is authenticated
type RequireAuthMiddleware struct{}

func NewRequireAuthMiddleware() *RequireAuthMiddleware {
	return &RequireAuthMiddleware{}
}

func (m *RequireAuthMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	if !c.UserIsLogged() {
		response.SendError(c, http.StatusUnauthorized, "Authentication required", nil)
		return nil
	}
	return nil
}

func (m *RequireAuthMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	return nil
}

// RequireAdminMiddleware ensures a user is authenticated and is an admin
type RequireAdminMiddleware struct{}

func NewRequireAdminMiddleware() *RequireAdminMiddleware {
	return &RequireAdminMiddleware{}
}

func (m *RequireAdminMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	account := c.GetAccount()
	if account == nil || !account.IsOwner() {
		response.SendError(c, http.StatusForbidden, "Admin access required", nil)
		return nil
	}
	return nil
}

func (m *RequireAdminMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	return nil
}

// Helper functions
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

func getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err != nil {
		return ""
	}
	return cookie.Value
}
