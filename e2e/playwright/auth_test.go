package playwright_test

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

	// Initialize the browser
	pw, err := playwright.Run()
	require.NoError(t, err, "Initialize Playwright")
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	require.NoError(t, err, "Launch browser")
	defer browser.Close()

	t.Run("successful login with default credentials", func(t *testing.T) {
		context, err := browser.NewContext()
		require.NoError(t, err, "Create browser context")

		t.Cleanup(func() {
			context.Close()
		})

		page, err := context.NewPage()
		require.NoError(t, err, "Create new page")
		defer page.Close()

		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err, "Navigate to login page")

		// Get locators for form elements
		usernameLocator := page.Locator("#username")
		passwordLocator := page.Locator("#password")
		buttonLocator := page.Locator(".button")

		// Wait for and fill the login form
		require.NoError(t, usernameLocator.WaitFor())
		require.NoError(t, usernameLocator.Fill("shiori"))
		require.NoError(t, passwordLocator.Fill("gopher"))

		// Click login and wait for success
		require.NoError(t, buttonLocator.Click())
		require.NoError(t, page.Locator("#bookmarks-grid").WaitFor())
	})

	t.Run("failed login with wrong username", func(t *testing.T) {
		context, err := browser.NewContext()
		require.NoError(t, err)

		t.Cleanup(func() {
			context.Close()
		})

		page, err := context.NewPage()
		require.NoError(t, err)
		defer page.Close()

		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err)

		// Get locators for form elements
		usernameLocator := page.Locator("#username")
		passwordLocator := page.Locator("#password")
		buttonLocator := page.Locator(".button")
		errorLocator := page.Locator(".error-message")

		// Wait for and fill the login form
		require.NoError(t, usernameLocator.WaitFor())
		require.NoError(t, usernameLocator.Fill("wrong_user"))
		require.NoError(t, passwordLocator.Fill("gopher"))

		// Click login and verify error
		require.NoError(t, buttonLocator.Click())
		errorText, err := errorLocator.TextContent()
		require.NoError(t, err, "Get error message text")
		require.Contains(t, errorText, "username or password do not match", "Verify error message for wrong username")
	})

	t.Run("failed login with wrong password", func(t *testing.T) {
		context, err := browser.NewContext()
		require.NoError(t, err)

		t.Cleanup(func() {
			context.Close()
		})

		page, err := context.NewPage()
		require.NoError(t, err)
		defer page.Close()

		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err)

		// Get locators for form elements
		usernameLocator := page.Locator("#username")
		passwordLocator := page.Locator("#password")
		buttonLocator := page.Locator(".button")
		errorLocator := page.Locator(".error-message")

		// Wait for and fill the login form
		require.NoError(t, usernameLocator.WaitFor())
		require.NoError(t, usernameLocator.Fill("shiori"))
		require.NoError(t, passwordLocator.Fill("wrong_password"))

		// Click login and verify error
		require.NoError(t, buttonLocator.Click())
		errorText, err := errorLocator.TextContent()
		require.NoError(t, err)
		require.Contains(t, errorText, "username or password do not match")
	})

	t.Run("empty username validation", func(t *testing.T) {
		context, err := browser.NewContext()
		require.NoError(t, err)

		t.Cleanup(func() {
			context.Close()
		})

		page, err := context.NewPage()
		require.NoError(t, err)
		defer page.Close()

		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err)

		// Get locators for form elements
		usernameLocator := page.Locator("#username")
		passwordLocator := page.Locator("#password")
		buttonLocator := page.Locator(".button")
		errorLocator := page.Locator(".error-message")

		// Wait for form and fill only password
		require.NoError(t, usernameLocator.WaitFor())
		require.NoError(t, passwordLocator.Fill("gopher"))

		// Click login and verify error
		require.NoError(t, buttonLocator.Click())
		errorText, err := errorLocator.TextContent()
		require.NoError(t, err)
		require.Contains(t, errorText, "Username must not empty")
	})
}
