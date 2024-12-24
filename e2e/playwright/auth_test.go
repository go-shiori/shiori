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
	require.NoError(t, err)
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	defer browser.Close()

	// Create a new browser context for each test to ensure clean state
	context, err := browser.NewContext()
	require.NoError(t, err)
	defer context.Close()

	t.Run("successful login with default credentials", func(t *testing.T) {
		page, err := context.NewPage()
		require.NoError(t, err)
		defer page.Close()

		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err)

		// Fill in the login form with default credentials
		require.NoError(t, page.Fill("#username", "shiori"))
		require.NoError(t, page.Fill("#password", "gopher"))

		// Click the login button
		require.NoError(t, page.Click(".button"))

		// Wait for navigation and verify we're logged in by checking for bookmarks page element
		_, err = page.WaitForSelector("#page-content")
		require.NoError(t, err)
	})

	t.Run("failed login with wrong username", func(t *testing.T) {
		page, err := context.NewPage()
		require.NoError(t, err)
		defer page.Close()

		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err)

		// Fill in the login form with wrong username
		require.NoError(t, page.Fill("#username", "wrong_user"))
		require.NoError(t, page.Fill("#password", "gopher"))

		// Click the login button
		require.NoError(t, page.Click(".button"))

		// Verify error message appears
		errorText, err := page.TextContent(".error-message")
		require.NoError(t, err)
		require.Contains(t, errorText, "username or password invalid")
	})

	t.Run("failed login with wrong password", func(t *testing.T) {
		page, err := context.NewPage()
		require.NoError(t, err)
		defer page.Close()

		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err)

		// Fill in the login form with wrong password
		require.NoError(t, page.Fill("#username", "shiori"))
		require.NoError(t, page.Fill("#password", "wrong_password"))

		// Click the login button
		require.NoError(t, page.Click(".button"))

		// Verify error message appears
		errorText, err := page.TextContent(".error-message")
		require.NoError(t, err)
		require.Contains(t, errorText, "username or password invalid")
	})

	t.Run("empty username validation", func(t *testing.T) {
		page, err := context.NewPage()
		require.NoError(t, err)
		defer page.Close()

		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err)

		// Fill in only password
		require.NoError(t, page.Fill("#password", "gopher"))

		// Click the login button
		require.NoError(t, page.Click(".button"))

		// Verify error message appears
		errorText, err := page.TextContent(".error-message")
		require.NoError(t, err)
		require.Contains(t, errorText, "Username must not empty")
	})
}
