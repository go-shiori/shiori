package middleware

import (
	"errors"
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
	plainIPs := deps.Config().Http.SSOProxyAuthTrusted
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
	account, err := m.ssoAccount(deps, c)
	if err != nil {
		deps.Logger().
			WithError(err).
			WithField("remote_addr", c.Request().RemoteAddr).
			WithField("request_id", c.GetRequestID()).
			Error("getting sso account")
	}
	if account != nil {
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

	account, err = deps.Domains().Auth().CheckToken(c.Request().Context(), token)
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

func (m *AuthMiddleware) ssoAccount(deps model.Dependencies, c model.WebContext) (*model.AccountDTO,error) {
	if !deps.Config().Http.SSOProxyAuth {
		return nil, nil
	}

	remoteAddr := c.Request().RemoteAddr
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		var addrErr *net.AddrError
		if errors.As(err, &addrErr) && addrErr.Err == "missing port in address" {
			ip = remoteAddr
		} else {
			return nil,err
		}
	}
	requestIP := net.ParseIP(ip)
	if !m.isTrustedIP(requestIP) {
		return nil, errors.New("remoteAddr is not a trusted ip") 
	}

	headerName := deps.Config().Http.SSOProxyAuthHeaderName
	userName := c.Request().Header.Get(headerName)
	if userName == "" {
		return nil, nil
	}

	account, err := deps.Domains().Accounts().GetAccountByUsername(c.Request().Context(), userName)
	if err != nil {
		return nil, err
	}

	return account, nil
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
