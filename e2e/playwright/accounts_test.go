package playwright

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestE2EAccounts(t *testing.T) {
	// Start a new Shiori container
	container := e2eutil.NewShioriContainer(t, "")
	baseURL := fmt.Sprintf("http://localhost:%s", container.GetPort())

	mainTestHelper, err := NewTestHelper(t, "main")
	require.NoError(t, err)
	defer mainTestHelper.Close()

	t.Run("001 login as admin", func(t *testing.T) {
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

	t.Run("002 create new admin account", func(t *testing.T) {
		// Navigate to settings page
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Settings"]`).Click(),
			"Click on settings button")
		mainTestHelper.Require().NoError(t, mainTestHelper.page.Locator(".setting-container").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}), "Wait for settings page to show up")

		// Click on "Add new account" <a> element
		mainTestHelper.page.Locator(`[title="Add new account"]`).Click()
		mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		})

		// Fill modal
		mainTestHelper.page.Locator(`[name="username"]`).Fill("admin2")
		mainTestHelper.page.Locator(`[name="password"]`).Fill("admin2")
		mainTestHelper.page.Locator(`[name="repeat_password"]`).Fill("admin2")
		mainTestHelper.page.Locator(`[name="admin"]`).Check()

		// Click on "Ok" button
		mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click()

		// Wait for modal to disappear
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(1000),
			}),
			"Wait for modal to disappear")

		// Refresh account list
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`a[title="Refresh accounts"]`).Click(),
			"Click on refresh accounts button")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".loading-overlay").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(1000),
			}),
			"Wait for loading overlay to disappear")

		// Check if new account is created
		accountsCount, err := mainTestHelper.page.Locator(".accounts-list li").Count()
		mainTestHelper.Require().NoError(t, err, "Count accounts in list")
		mainTestHelper.Require().Equal(t, 2, accountsCount, "Verify 2 accounts present after creating new admin account")
	})

	t.Run("003 create new user account", func(t *testing.T) {
		// Click on "Add new account" <a> element
		mainTestHelper.page.Locator(`[title="Add new account"]`).Click()
		mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		})

		// Fill modal
		mainTestHelper.page.Locator(`[name="username"]`).Fill("user1")
		mainTestHelper.page.Locator(`[name="password"]`).Fill("user1")
		mainTestHelper.page.Locator(`[name="repeat_password"]`).Fill("user1")

		// Click on "Ok" button
		mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click()

		// Wait for modal to disappear
		mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(1000),
		})

		// Refresh account list
		mainTestHelper.page.Locator(`a[title="Refresh accounts"]`).Click()
		mainTestHelper.page.Locator(".loading-overlay").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(1000),
		})

		// Check if new account is created
		accountsCount, err := mainTestHelper.page.Locator(".accounts-list li").Count()
		mainTestHelper.Require().NoError(t, err, "Failed to count accounts in list")
		mainTestHelper.Require().Equal(t, 3, accountsCount, "Expected 3 accounts after creating user account")
	})

	t.Run("004 check admin account created successfully", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err, "Create test helper")
		defer th.Close()

		// Navigate to the login page
		_, err = th.page.Goto(baseURL)
		th.Require().NoError(t, err, "Navigate to base URL")

		// Get locators for form elements
		usernameLocator := th.page.Locator("#username")
		passwordLocator := th.page.Locator("#password")
		buttonLocator := th.page.Locator(".button")

		// Wait for and fill the login form
		th.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
		th.Require().NoError(t, usernameLocator.Fill("admin2"), "Fill username field")
		th.Require().NoError(t, passwordLocator.Fill("admin2"), "Fill password field")

		// Click login and wait for success
		th.Require().NoError(t, buttonLocator.Click(), "Click login button")
		th.Require().NoError(t, th.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}), "Wait for bookmarks section to show up")

		// Navigate to settings
		th.Require().NoError(t,
			th.page.Locator(`[title="Settings"]`).Click(),
			"Click on settings button")
		th.Require().NoError(t,
			th.page.Locator(".setting-container").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for settings page to show up")

		// Check if can see system info (admin only)
		visible, err := th.page.Locator(`#setting-system-info`).IsVisible()
		th.Require().NoError(t, err, "Check visibility of system info section")
		th.Require().True(t, visible, "Verify system info section visibility for admin user")
	})

	t.Run("005 check user account created successfully", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err, "Create test helper")

		defer th.Close()

		// Navigate to the login page
		_, err = th.page.Goto(baseURL)
		th.Require().NoError(t, err, "Navigate to base URL")

		// Get locators for form elements
		usernameLocator := th.page.Locator("#username")
		passwordLocator := th.page.Locator("#password")
		buttonLocator := th.page.Locator(".button")

		// Wait for and fill the login form
		th.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
		th.Require().NoError(t, usernameLocator.Fill("user1"), "Fill username field")
		th.Require().NoError(t, passwordLocator.Fill("user1"), "Fill password field")

		// Click login and wait for success
		th.Require().NoError(t, buttonLocator.Click(), "Click login button")
		th.Require().NoError(t, th.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}), "Wait for bookmarks section to show up")

		// Navigate to settings
		th.Require().NoError(t,
			th.page.Locator(`[title="Settings"]`).Click(),
			"Click on settings button")
		th.Require().NoError(t, th.page.Locator(".setting-container").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}), "Wait for settings page to show up")

		// Check if can see system info (admin only)
		visible, err := th.page.Locator(`#setting-system-info`).IsVisible()
		th.Require().NoError(t, err, "Check visibility of system info section")
		th.Require().False(t, visible, "Verify system info section not visible for regular user")

		// My account settings is visible
		visible, err = th.page.Locator(`#setting-my-account`).IsVisible()
		th.Require().NoError(t, err, "Check visibility of account settings")
		th.Require().True(t, visible, "Verify account settings visibility for user")

		// Check change password requires current password
		th.Require().NoError(t,
			th.page.Locator(`li[shiori-username="user1"] a[title="Change password"]`).Click(),
			"Click on change password button")
		th.Require().NoError(t,
			th.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for change password modal to show up")
		visible, err = th.page.Locator(`[name="old_password"]`).IsVisible()
		th.Require().NoError(t, err, "Check visibility of old password field")
		th.Require().True(t, visible, "Verify old password field visibility when changing password")

		// Fill modal
		th.Require().NoError(t,
			th.page.Locator(`[name="old_password"]`).Fill("user1"),
			"Fill old password field")
		th.Require().NoError(t,
			th.page.Locator(`[name="new_password"]`).Fill("new_user1"),
			"Fill new password field")
		th.Require().NoError(t,
			th.page.Locator(`[name="repeat_password"]`).Fill("new_user1"),
			"Fill repeat password field")

		// Click on "Ok" button
		th.Require().NoError(t,
			th.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click on ok button")

		// Wait for modal to display text: "Password has been changed."
		dialogContent := th.page.Locator(".custom-dialog-content")
		th.Require().NoError(t,
			dialogContent.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for dialog content to show up")

		contentText, err := dialogContent.TextContent()
		th.Require().NoError(t, err, "Get dialog content text")
		th.Require().Equal(t, "Password has been changed.", contentText, "Verify password change confirmation message")
	})

	t.Run("006 delete user account", func(t *testing.T) {
		// Click on "Delete" button
		mainTestHelper.page.Locator(`li[shiori-username="user1"] a[title="Delete account"]`).Click()
		mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		})

		// Click on "Ok" button
		mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click()

		// Wait for modal to disappear
		mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(1000),
		})

		// Refresh account list
		mainTestHelper.page.Locator(`a[title="Refresh accounts"]`).Click()
		mainTestHelper.page.Locator(".loading-overlay").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateHidden,
			Timeout: playwright.Float(1000),
		})

		// Check if account is deleted
		accountsCount, err := mainTestHelper.page.Locator(".accounts-list li").Count()
		mainTestHelper.Require().NoError(t, err, "Count accounts in list")
		mainTestHelper.Require().Equal(t, 2, accountsCount, "Verify 2 accounts present after creating admin account")

		time.Sleep(5 * time.Second)
	})

	t.Run("007 change password for admin account", func(t *testing.T) {
		// Click on "Change password" button
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`li[shiori-username="admin2"] a[title="Change password"]`).Click(),
			"Click change password button")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for password dialog to appear")

		// Fill modal
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="new_password"]`).Fill("admin3"),
			"Fill new password")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="repeat_password"]`).Fill("admin3"),
			"Fill repeat password")

		// Click on "Ok" button
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click ok button")

		// Wait for modal to display text: "Password has been changed."
		dialogContent := mainTestHelper.page.Locator(".custom-dialog-content")
		mainTestHelper.Require().NoError(t,
			dialogContent.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for dialog content to show up")

		contentText, err := dialogContent.TextContent()
		mainTestHelper.Require().NoError(t, err, "Get dialog content text")
		mainTestHelper.Require().Equal(t, "Password has been changed.", contentText, "Verify password change confirmation message")

		// Click on "Ok" button
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click ok button")

		// Wait for modal to disappear
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(2000),
			}),
			"Wait for dialog to close")

		// Refresh account list
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`a[title="Refresh accounts"]`).Click(),
			"Click refresh accounts")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".loading-overlay").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(1000),
			}),
			"Wait for refresh to complete")

		t.Run("0071 login with new password", func(t *testing.T) {
			th, err := NewTestHelper(t, t.Name())
			require.NoError(t, err, "Failed to create test helper")
			defer th.Close()

			// Navigate to the login page
			_, err = th.page.Goto(baseURL)
			th.Require().NoError(t, err, "Navigate to base URL")

			// Wait for login page
			th.Require().NoError(t,
				th.page.Locator("#username").WaitFor(playwright.LocatorWaitForOptions{
					State:   playwright.WaitForSelectorStateVisible,
					Timeout: playwright.Float(1000),
				}),
				"Wait for login page")
			th.Require().NoError(t, th.page.Locator("#username").Fill("admin2"), "Fill username field")
			th.Require().NoError(t, th.page.Locator("#password").Fill("admin3"), "Fill password field")
			th.Require().NoError(t, th.page.Locator(".button").Click(), "Click login button")
			th.Require().NoError(t, th.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}), "Wait for bookmarks section to show up")
		})
	})

	t.Run("008 logout", func(t *testing.T) {
		// Click on "Logout" button
		mainTestHelper.Require().NoError(t, mainTestHelper.page.Locator(`a[title="Logout"]`).Click(), "Click on logout button")

		// Wait for modal to display text
		dialogContent := mainTestHelper.page.Locator(".custom-dialog-content")
		mainTestHelper.Require().NoError(t,
			dialogContent.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for dialog content to show up")

		contentText, err := dialogContent.TextContent()
		mainTestHelper.Require().NoError(t, err, "Get dialog content text")
		mainTestHelper.Require().Equal(t, "Are you sure you want to log out ?", contentText, "Verify logout confirmation message")

		// Click on "Yes" button
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click Yes button")

		// Wait for login page
		err = mainTestHelper.page.Locator("#login-scene").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		})
		mainTestHelper.Require().NoError(t, err, "Wait for login page")
	})
}
