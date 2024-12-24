package playwright_test

import (
	"fmt"
	"testing"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/playwright-community/playwright-go"
	expect "github.com/playwright-community/playwright-go/expect"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	// Start a new Shiori container
	container := e2eutil.NewShioriContainer(t, "")
	baseURL := fmt.Sprintf("http://localhost:%s", container.GetPort())

	// Initialize the browser
	pw, err := playwright.Run()
	require.NoError(t, err)
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	require.NoError(t, err)
	defer browser.Close()

	// Create a new browser context for each test to ensure clean state
	context, err := browser.NewContext()
	require.NoError(t, err)
	defer context.Close()

	t.Run("successful login with default credentials", func(t *testing.T) {
		_, err = browser.NewContext()
		require.NoError(t, err)

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

		// Wait for and fill the login form
		require.NoError(t, expect.Expect(usernameLocator).ToBeVisible())
		require.NoError(t, usernameLocator.Fill("shiori"))
		require.NoError(t, passwordLocator.Fill("gopher"))

		// Click login and wait for success
		require.NoError(t, buttonLocator.Click())
		require.NoError(t, expect.Expect(page.Locator("#bookmarks-grid")).ToBeVisible())
	})

	t.Run("failed login with wrong username", func(t *testing.T) {
		_, err = browser.NewContext()
		require.NoError(t, err)

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
		require.NoError(t, expect.Expect(usernameLocator).ToBeVisible())
		require.NoError(t, usernameLocator.Fill("wrong_user"))
		require.NoError(t, passwordLocator.Fill("gopher"))

		// Click login and verify error
		require.NoError(t, buttonLocator.Click())
		errorText, err := errorLocator.TextContent()
		require.NoError(t, err)
		require.Contains(t, errorText, "username or password do not match")
	})

	t.Run("failed login with wrong password", func(t *testing.T) {
		_, err = browser.NewContext()
		require.NoError(t, err)

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
		require.NoError(t, expect.Expect(usernameLocator).ToBeVisible())
		require.NoError(t, usernameLocator.Fill("shiori"))
		require.NoError(t, passwordLocator.Fill("wrong_password"))

		// Click login and verify error
		require.NoError(t, buttonLocator.Click())
		errorText, err := errorLocator.TextContent()
		require.NoError(t, err)
		require.Contains(t, errorText, "username or password invalid")
	})

	t.Run("empty username validation", func(t *testing.T) {
		_, err = browser.NewContext()
		require.NoError(t, err)

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
		require.NoError(t, expect.Expect(usernameLocator).ToBeVisible())
		require.NoError(t, passwordLocator.Fill("gopher"))

		// Click login and verify error
		require.NoError(t, buttonLocator.Click())
		errorText, err := errorLocator.TextContent()
		require.NoError(t, err)
		require.Contains(t, errorText, "Username must not empty")
	})
}
