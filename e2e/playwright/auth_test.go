package playwright

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	// Start a new Shiori container
	container := e2eutil.NewShioriContainer(t, "")
	baseURL := fmt.Sprintf("http://localhost:%s", container.GetPort())

	mainTestHelper, err := NewTestHelper(t, "main")
	require.NoError(t, err)
	defer mainTestHelper.Close()

	t.Run("successful login with default credentials", func(t *testing.T) {
		// Navigate to the login page
		_, err = mainTestHelper.page.Goto(baseURL)
		mainTestHelper.Require().NoError(t, err, "Navigate to base URL")

		// Get locators for form elements
		usernameLocator := mainTestHelper.page.Locator("#username")
		passwordLocator := mainTestHelper.page.Locator("#password")
		buttonLocator := mainTestHelper.page.Locator(".button")

		// Wait for and fill the login form
		mainTestHelper.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
		mainTestHelper.Require().NoError(t, usernameLocator.Fill("shiori"), "Fill username field")
		mainTestHelper.Require().NoError(t, passwordLocator.Fill("gopher"), "Fill password field")

		// Click login and wait for success
		mainTestHelper.Require().NoError(t, buttonLocator.Click(), "Click login button")
		mainTestHelper.Require().NoError(t, mainTestHelper.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}), "Wait for bookmarks section to show up")
	})

	t.Run("failed login with wrong username", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err)
		defer th.Close()

		// Navigate to the login page
		_, err = th.page.Goto(baseURL)
		th.Require().NoError(t, err, "Navigate to base URL")

		// Get locators for form elements
		usernameLocator := th.page.Locator("#username")
		passwordLocator := th.page.Locator("#password")
		buttonLocator := th.page.Locator(".button")
		errorLocator := th.page.Locator(".error-message")

		// Wait for and fill the login form
		th.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
		th.Require().NoError(t, usernameLocator.Fill("wrong_user"), "Fill username field")
		th.Require().NoError(t, passwordLocator.Fill("gopher"), "Fill password field")

		// Click login and verify error
		th.Require().NoError(t, buttonLocator.Click(), "Click login button")
		errorText, err := errorLocator.TextContent()
		th.Require().NoError(t, err, "Get error message text")
		th.Require().Contains(t, errorText, "Invalid username or password")
	})

	t.Run("failed login with wrong password", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err)
		defer th.Close()

		// Navigate to the login page
		_, err = th.page.Goto(baseURL)
		th.Require().NoError(t, err, "Navigate to base URL")

		// Get locators for form elements
		usernameLocator := th.page.Locator("#username")
		passwordLocator := th.page.Locator("#password")
		buttonLocator := th.page.Locator(".button")
		errorLocator := th.page.Locator(".error-message")

		// Wait for and fill the login form
		th.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
		th.Require().NoError(t, usernameLocator.Fill("shiori"), "Fill username field")
		th.Require().NoError(t, passwordLocator.Fill("wrong_password"), "Fill password field")

		// Click login and verify error
		th.Require().NoError(t, buttonLocator.Click(), "Click login button")
		errorText, err := errorLocator.TextContent()
		th.Require().NoError(t, err, "Get error message text")
		th.Require().Contains(t, errorText, "Invalid username or password")
	})

	t.Run("empty username validation", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err)
		defer th.Close()

		// Navigate to the login page
		_, err = th.page.Goto(baseURL)
		th.Require().NoError(t, err, "Navigate to base URL")

		// Get locators for form elements
		usernameLocator := th.page.Locator("#username")
		passwordLocator := th.page.Locator("#password")
		buttonLocator := th.page.Locator(".button")
		errorLocator := th.page.Locator(".error-message")

		// Wait for form and fill only password
		th.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
		th.Require().NoError(t, passwordLocator.Fill("gopher"), "Fill password field")

		// Click login and verify error
		th.Require().NoError(t, buttonLocator.Click(), "Click login button")
		errorText, err := errorLocator.TextContent()
		th.Require().NoError(t, err, "Get error message text")
		th.Require().Contains(t, errorText, "Username must not empty")
	})

	t.Run("remember me extends session expiration", func(t *testing.T) {
		// loginResponse is used to parse the API response.
		// SendJSON encodes the data directly (no wrapper), so the response is:
		// {"Token": "...", "Expiration": 1234567890}
		type loginResponse struct {
			Token      string `json:"Token"`
			Expiration int64  `json:"Expiration"`
		}

		t.Run("without remember me token expires in about 1 hour", func(t *testing.T) {
			th, err := NewTestHelper(t, t.Name())
			require.NoError(t, err)
			defer th.Close()

			_, err = th.page.Goto(baseURL)
			th.Require().NoError(t, err, "Navigate to base URL")

			usernameLocator := th.page.Locator("#username")
			th.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
			th.Require().NoError(t, usernameLocator.Fill("shiori"), "Fill username field")
			th.Require().NoError(t, th.page.Locator("#password").Fill("gopher"), "Fill password field")

			// Do NOT check remember-me, intercept the login API response
			resp, err := th.page.ExpectResponse("**/api/v1/auth/login", func() error {
				return th.page.Locator(".button").Click()
			})
			th.Require().NoError(t, err, "Intercept login response")

			body, err := resp.Body()
			th.Require().NoError(t, err, "Read response body")

			var loginResp loginResponse
			th.Require().NoError(t, json.Unmarshal(body, &loginResp), "Parse login response JSON")

			// Token expiration should be ~1 hour from now (less than 2 hours)
			expiresAt := time.Unix(loginResp.Expiration, 0)
			untilExpiry := time.Until(expiresAt)
			th.Require().True(t, untilExpiry < 2*time.Hour,
				fmt.Sprintf("Without remember-me, token should expire within 2 hours, got %v", untilExpiry))
			th.Require().True(t, untilExpiry > 30*time.Minute,
				fmt.Sprintf("Token should expire in at least 30 minutes, got %v", untilExpiry))
		})

		t.Run("with remember me token expires in about 30 days", func(t *testing.T) {
			th, err := NewTestHelper(t, t.Name())
			require.NoError(t, err)
			defer th.Close()

			_, err = th.page.Goto(baseURL)
			th.Require().NoError(t, err, "Navigate to base URL")

			usernameLocator := th.page.Locator("#username")
			th.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
			th.Require().NoError(t, usernameLocator.Fill("shiori"), "Fill username field")
			th.Require().NoError(t, th.page.Locator("#password").Fill("gopher"), "Fill password field")

			// Check the remember-me checkbox (legacy UI uses name="remember")
			th.Require().NoError(t, th.page.Locator("input[name='remember']").Check(), "Check remember-me checkbox")

			// Intercept the login API response
			resp, err := th.page.ExpectResponse("**/api/v1/auth/login", func() error {
				return th.page.Locator(".button").Click()
			})
			th.Require().NoError(t, err, "Intercept login response")

			body, err := resp.Body()
			th.Require().NoError(t, err, "Read response body")

			var loginResp loginResponse
			th.Require().NoError(t, json.Unmarshal(body, &loginResp), "Parse login response JSON")

			// Token expiration should be ~30 days from now (more than 29 days)
			expiresAt := time.Unix(loginResp.Expiration, 0)
			untilExpiry := time.Until(expiresAt)
			th.Require().True(t, untilExpiry > 29*24*time.Hour,
				fmt.Sprintf("With remember-me, token should expire after 29 days, got %v", untilExpiry))
		})
	})

	t.Run("unauthenticated user sees login form", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err)
		defer th.Close()

		// Navigate to the app without any session
		_, err = th.page.Goto(baseURL)
		th.Require().NoError(t, err, "Navigate to base URL")

		// The login form should be displayed
		th.Require().NoError(t, th.page.Locator("#login-scene").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		}), "Login form should be visible for unauthenticated user")

		// The main app content should NOT be visible
		visible, err := th.page.Locator("#main-scene").IsVisible()
		th.Require().NoError(t, err, "Check main scene visibility")
		th.Require().False(t, visible, "Main app content should not be visible when unauthenticated")
	})

	t.Run("deep-link redirect after login", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err)
		defer th.Close()

		// Navigate to the login page with a dst parameter pointing to the settings page
		_, err = th.page.Goto(baseURL + "?dst=%2F%23setting")
		th.Require().NoError(t, err, "Navigate to login with dst param")

		// Wait for login form
		th.Require().NoError(t, th.page.Locator("#username").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3000),
		}), "Wait for login form")

		// Log in with valid credentials
		th.Require().NoError(t, th.page.Locator("#username").Fill("shiori"), "Fill username field")
		th.Require().NoError(t, th.page.Locator("#password").Fill("gopher"), "Fill password field")
		th.Require().NoError(t, th.page.Locator(".button").Click(), "Click login button")

		// After login, the app should redirect to /#setting (the dst destination)
		th.Require().NoError(t, th.page.WaitForURL("**/#setting", playwright.PageWaitForURLOptions{
			Timeout: playwright.Float(3000),
		}), "Wait for redirect to settings hash")
	})
}
