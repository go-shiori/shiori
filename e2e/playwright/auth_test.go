package playwright

import (
	"fmt"
	"testing"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/playwright-community/playwright-go"
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
		th.Require().Contains(t, errorText, "username or password do not match")
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
		th.Require().Contains(t, errorText, "username or password do not match")
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
}
