package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/http/context"
	"github.com/go-shiori/shiori/internal/http/response"
	"github.com/go-shiori/shiori/internal/model"
)

// AuthMiddleware provides basic authentication capabilities to all routes underneath
// its usage, only allowing authenticated users access and set a custom local context
// `account` with the account model for the logged in user.
func AuthMiddleware(deps *dependencies.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getTokenFromHeader(c)
		if token == "" {
			token = getTokenFromCookie(c)
		}
		if token == "" {
			token = getTokenFromAuthHeader(c, deps)
		}

		account, err := deps.Domains.Auth.CheckToken(c, token)
		if err != nil {
			return
		}

		c.Set(model.ContextAccountKey, account)
	}
}

// AuthenticationRequired provides a middleware that checks if the user is logged in, returning
// a 401 error if not.
func AuthenticationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.NewContextFromGin(c)
		if !ctx.UserIsLogged() {
			response.SendError(c, http.StatusUnauthorized, nil)
			return
		}
	}
}

// getAuth user from oauth proxy, if any
func getTokenFromAuthHeader(c *gin.Context, deps *dependencies.Dependencies) string {
	if deps.Config.Http.ReverseProxyAuthUser == "" {
		deps.Log.Debugf("auth proxy: reverse-proxy-auth-user not set")
		return ""
	}
	authUser := c.GetHeader(deps.Config.Http.ReverseProxyAuthUser)
	if authUser == "" {
		deps.Log.Debugf("auth proxy: can not get user header from proxy")
		return ""
	}
	remoteAddr := c.ClientIP()
	deps.Log.Debugf("auth proxy: got auth user (%s), client ip (%s)", authUser, remoteAddr)
	cidrs, err := newCIDRs(deps.Config.Http.TrustedProxies)
	if err != nil {
		deps.Log.Errorf("auth proxy: trusted proxy config error (%v)", err)
		return ""
	}
	canTrustProxy := false
	if len(deps.Config.Http.TrustedProxies) == 0 || cidrs.ContainStringIP(remoteAddr) {
		canTrustProxy = true
	}
	if canTrustProxy {
		account, exit, err := deps.Database.GetAccount(c, authUser)
		if err == nil && exit {
			token, err := deps.Domains.Auth.CreateTokenForAccount(&account, time.Now().Add(time.Hour*24*365))
			if err != nil {
				deps.Log.Errorf("auth proxy: create token error %v", err)
				return ""
			}
			sessionId, err := deps.Domains.LegacyLogin(account, time.Hour*24*30)
			if err != nil {
				deps.Log.Errorf("auth proxy: create session error %v", err)
				return ""
			}
			deps.Log.Debugf("auth proxy: write session %s token %s", sessionId, token)
			sessionCookie := &http.Cookie{Name: "session-id", Value: sessionId}
			http.SetCookie(c.Writer, sessionCookie)
			c.Request.AddCookie(sessionCookie)
			tokenCookie := &http.Cookie{Name: "token", Value: token}
			http.SetCookie(c.Writer, tokenCookie)
			c.Request.AddCookie(tokenCookie)

			return token
		} else {
			deps.Log.Warnf("auth proxy: no such user (%s) or error %v", authUser, err)
		}
	} else if authUser != "" {
		deps.Log.Warnf("auth proxy: invalid auth request from %s", remoteAddr)
	}
	return ""
}

// getTokenFromHeader returns the token from the Authorization header, if any.
func getTokenFromHeader(c *gin.Context) string {
	authorization := c.GetHeader(model.AuthorizationHeader)
	if authorization == "" {
		return ""
	}

	authParts := strings.SplitN(authorization, " ", 2)
	if len(authParts) != 2 && authParts[0] != model.AuthorizationTokenType {
		return ""
	}

	return authParts[1]
}

// getTokenFromCookie returns the token from the token cookie, if any.
func getTokenFromCookie(c *gin.Context) string {
	cookie, err := c.Cookie("token")
	if err != nil {
		return ""
	}

	return cookie
}
