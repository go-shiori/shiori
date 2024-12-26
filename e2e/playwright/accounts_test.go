package playwright

import (
	"fmt"
	"testing"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2EAccounts(t *testing.T) {
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

	context, err := browser.NewContext()
	require.NoError(t, err)

	t.Cleanup(func() {
		context.Close()
	})

	page, err := context.NewPage()
	require.NoError(t, err)
	defer page.Close()

	t.Run("login as admin", func(t *testing.T) {
		// Navigate to the login page
		_, err = page.Goto(baseURL)
		require.NoError(t, err)

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
		require.NoError(t, page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}))
	})

	t.Run("create new admin account", func(t *testing.T) {
		// Navigate to settings page
		page.Locator(`[title="Settings"]`).Click()
		require.NoError(t, page.Locator(".setting-container").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}))

		// Click on "Add new account" <a> element
		page.Locator(`[title="Add new account"]`).Click()
		page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		})

		// Fill modal
		page.Locator(`[name="username"]`).Fill("admin2")
		page.Locator(`[name="password"]`).Fill("admin2")
		page.Locator(`[name="repeat_password"]`).Fill("admin2")
		page.Locator(`[name="admin"]`).Check()

		// Click on "Ok" button
		page.Locator(`.custom-dialog-button.main`).Click()

		// Wait for modal to disappear
		page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(1000),
		})

		// Check if new account is created
		accountsCount, err := page.Locator(".accounts-list li").Count()
		require.NoError(t, err)
		require.Equal(t, 2, accountsCount)
	})

	t.Run("create new user account", func(t *testing.T) {
		// Click on "Add new account" <a> element
		page.Locator(`[title="Add new account"]`).Click()
		page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		})

		// Fill modal
		page.Locator(`[name="username"]`).Fill("user1")
		page.Locator(`[name="password"]`).Fill("user1")
		page.Locator(`[name="repeat_password"]`).Fill("user1")

		// Click on "Ok" button
		page.Locator(`.custom-dialog-button.main`).Click()

		// Wait for modal to disappear
		page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(1000),
		})

		// Check if new account is created
		accountsCount, err := page.Locator(".accounts-list li").Count()
		require.NoError(t, err)
		require.Equal(t, 3, accountsCount)
	})
}
