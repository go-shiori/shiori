package middleware

import (
	"errors"
	"net"

	"github.com/go-shiori/shiori/internal/model"
)

// AuthMiddleware handles authentication for incoming request by checking the token
// from the Authorization header or the token cookie and setting the account in the
// request context.
type AuthSSOProxyMiddleware struct {
	deps model.Dependencies

	trustedIPs []*net.IPNet
}

func NewAuthSSOProxyMiddleware(deps model.Dependencies) *AuthSSOProxyMiddleware {
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

	return &AuthSSOProxyMiddleware{
		deps:       deps,
		trustedIPs: trustedIPs,
	}
}

func (m *AuthSSOProxyMiddleware) OnRequest(deps model.Dependencies, c model.WebContext) error {
	if c.UserIsLogged() {
		return nil
	}

	account, err := m.ssoAccount(deps, c)
	if err != nil {
		deps.Logger().
			WithError(err).
			WithField("remote_addr", c.Request().RemoteAddr).
			WithField("request_id", c.GetRequestID()).
			Error("getting sso account")
		return nil
	}
	if account != nil {
		c.SetAccount(account)
		return nil
	}

	return nil
}

func (m *AuthSSOProxyMiddleware) ssoAccount(deps model.Dependencies, c model.WebContext) (*model.AccountDTO, error) {
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
			return nil, err
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
func (m *AuthSSOProxyMiddleware) isTrustedIP(ip net.IP) bool {
	for _, net := range m.trustedIPs {
		if ok := net.Contains(ip); ok {
			return true
		}
	}
	return false
}

func (m *AuthSSOProxyMiddleware) OnResponse(deps model.Dependencies, c model.WebContext) error {
	return nil
}
