package middleware

import (
	"fmt"
	"net"
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

	trustedIPs []*net.IPNet
}

func NewAuthMiddleware(deps model.Dependencies) *AuthMiddleware {
	plainIPs := deps.Config().Http.SSOTrustedProxy
	trustedIPs := make([]*net.IPNet, len(plainIPs))
	for i, ip := range plainIPs {
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			deps.Logger().WithError(err).WithField("ip", ip).Error("Failed to parse trusted ip cidr")
			continue
		}

		trustedIPs[i] = ipNet
	}

	return &AuthMiddleware{
		deps:       deps,
		trustedIPs: trustedIPs,
	}
}

func (m *AuthMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	if account := m.ssoAccount(deps, c); account != nil {
		c.SetAccount(account)
		return nil
	}

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

func (m *AuthMiddleware) ssoAccount(deps model.Dependencies, c model.WebContext) *model.AccountDTO {
	if !deps.Config().Http.SSOEnable {
		return nil
	}

	requestIP := net.ParseIP(c.Request().RemoteAddr)
	if !m.isTrustedIP(requestIP) {
		return nil
	}

	headerName := deps.Config().Http.SSOHeaderName
	userName := c.Request().Header.Get(headerName)
	if userName == "" {
		return nil
	}

	account, err := deps.Domains().Accounts().GetAccountByUsername(c.Request().Context(), userName)
	if err != nil {
		deps.Logger().WithError(err).WithField("request_id", c.GetRequestID()).Error("Failed to get account from sso header")
		return nil
	}

	return account
}
func (m *AuthMiddleware) isTrustedIP(ip net.IP) bool {
	for _, net := range m.trustedIPs {
		if ok := net.Contains(ip); ok {
			return true
		}
	}
	return false
}

func (m *AuthMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	return nil
}

// RequireLoggedInUser ensures a user is authenticated
func RequireLoggedInUser(deps model.Dependencies, c model.WebContext) error {
	if !c.UserIsLogged() {
		response.SendError(c, http.StatusUnauthorized, "Authentication required")
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
		response.SendError(c, http.StatusForbidden, "Admin access required")
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
