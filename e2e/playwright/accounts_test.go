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
		mainTestHelper.Require().NoError(err, "Failed to navigate to base URL")

		// Get locators for form elements
		usernameLocator := mainTestHelper.page.Locator("#username")
		passwordLocator := mainTestHelper.page.Locator("#password")
		buttonLocator := mainTestHelper.page.Locator(".button")

		// Wait for and fill the login form
		mainTestHelper.Require().NoError(usernameLocator.WaitFor(), "Username field not found")
		mainTestHelper.Require().NoError(usernameLocator.Fill("shiori"), "Failed to fill username")
		mainTestHelper.Require().NoError(passwordLocator.Fill("gopher"), "Failed to fill password")

		// Click login and wait for success
		mainTestHelper.Require().NoError(buttonLocator.Click(), "Failed to click login button")
		mainTestHelper.Require().NoError(mainTestHelper.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}))
	})

	t.Run("002 create new admin account", func(t *testing.T) {
		// Navigate to settings page
		mainTestHelper.page.Locator(`[title="Settings"]`).Click()
		mainTestHelper.Require().NoError(mainTestHelper.page.Locator(".setting-container").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}))

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
		mainTestHelper.Require().NoError(err, "Failed to count accounts in list")
		mainTestHelper.Require().Equal(2, accountsCount, "Expected 2 accounts after deleting user account")
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
		mainTestHelper.Require().NoError(err, "Failed to count accounts in list")
		mainTestHelper.Require().Equal(3, accountsCount, "Expected 3 accounts after creating user account")
	})

	t.Run("004 check admin account created successfully", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err, "Failed to create test helper")
		defer th.Close()

		// Navigate to the login page
		_, err = th.page.Goto(baseURL)
		th.Require().NoError(err, "Failed to navigate to base URL")

		// Get locators for form elements
		usernameLocator := th.page.Locator("#username")
		passwordLocator := th.page.Locator("#password")
		buttonLocator := th.page.Locator(".button")

		// Wait for and fill the login form
		th.Require().NoError(usernameLocator.WaitFor(), "Username field not found")
		th.Require().NoError(usernameLocator.Fill("admin2"), "Failed to fill username")
		th.Require().NoError(passwordLocator.Fill("admin2"), "Failed to fill password")

		// Click login and wait for success
		th.Require().NoError(buttonLocator.Click(), "Failed to click login button")
		th.Require().NoError(th.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}))

		// Navigate to settings
		th.page.Locator(`[title="Settings"]`).Click()
		th.Require().NoError(th.page.Locator(".setting-container").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}))

		// Check if can see system info (admin only)
		visible, err := th.page.Locator(`#setting-system-info`).IsVisible()
		th.Require().NoError(err, "Failed to check visibility of system info section")
		th.Require().True(visible, "System info section should be visible for admin user")
	})

	t.Run("005 check user account created successfully", func(t *testing.T) {
		th, err := NewTestHelper(t, t.Name())
		require.NoError(t, err)

		defer th.Close()

		// Navigate to the login page
		_, err = th.page.Goto(baseURL)
		th.Require().NoError(err, "Failed to navigate to base URL")

		// Get locators for form elements
		usernameLocator := th.page.Locator("#username")
		passwordLocator := th.page.Locator("#password")
		buttonLocator := th.page.Locator(".button")

		// Wait for and fill the login form
		th.Require().NoError(usernameLocator.WaitFor())
		th.Require().NoError(usernameLocator.Fill("user1"))
		th.Require().NoError(passwordLocator.Fill("user1"))

		// Click login and wait for success
		th.Require().NoError(buttonLocator.Click())
		th.Require().NoError(th.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}))

		// Navigate to settings
		th.page.Locator(`[title="Settings"]`).Click()
		th.Require().NoError(th.page.Locator(".setting-container").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}))

		// Check if can see system info (admin only)
		visible, err := th.page.Locator(`#setting-system-info`).IsVisible()
		th.Require().NoError(err, "Failed to check visibility of system info section")
		th.Require().False(visible, "System info section should not be visible for regular user")

		// My account settings is visible
		visible, err = th.page.Locator(`#setting-my-account`).IsVisible()
		th.Require().NoError(err, "Failed to check visibility of account settings")
		th.Require().True(visible, "Account settings should be visible for user")

		// Check change password requires current password
		th.page.Locator(`li[shiori-username="user1"] a[title="Change password"]`).Click()
		th.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		})
		visible, err = th.page.Locator(`[name="old_password"]`).IsVisible()
		th.Require().NoError(err, "Failed to check visibility of old password field")
		th.Require().True(visible, "Old password field should be visible when changing password")

		// Fill modal
		th.page.Locator(`[name="old_password"]`).Fill("user1")
		th.page.Locator(`[name="new_password"]`).Fill("new_user1")
		th.page.Locator(`[name="repeat_password"]`).Fill("new_user1")

		// Click on "Ok" button
		th.page.Locator(`.custom-dialog-button.main`).Click()

		// Wait for modal to display text: "Password has been changed."
		dialogContent := th.page.Locator(".custom-dialog-content")
		dialogContent.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		})

		contentText, err := dialogContent.TextContent()
		th.Require().NoError(err, "Failed to get dialog content text")
		th.Require().Equal("Password has been changed.", contentText, "Password change confirmation message not shown")
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
		mainTestHelper.Require().NoError(err, "Failed to count accounts in list")
		mainTestHelper.Require().Equal(2, accountsCount, "Expected 2 accounts after creating admin account")

		time.Sleep(5 * time.Second)
	})

	t.Run("007 change password for admin account", func(t *testing.T) {
		// Click on "Change password" button
		mainTestHelper.page.Locator(`li[shiori-username="admin2"] a[title="Change password"]`).Click()
		mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		})

		// Fill modal
		mainTestHelper.page.Locator(`[name="password"]`).Fill("admin3")
		mainTestHelper.page.Locator(`[name="repeat_password"]`).Fill("admin3")

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

		t.Run("0071 login with new password", func(t *testing.T) {
			th, err := NewTestHelper(t, t.Name())
			require.NoError(t, err)
			defer th.Close()

			// Go to login page
			th.page.Goto(baseURL)

			// Wait for login page
			th.page.Locator("#username").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			})
			th.Require().NoError(th.page.Locator("#username").Fill("admin2"), "Failed to fill username")
			th.Require().NoError(th.page.Locator("#password").Fill("admin3"), "Failed to fill password")
			th.Require().NoError(th.page.Locator(".button").Click(), "Failed to click login button")
			th.Require().NoError(th.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}))
		})
	})

	t.Run("008 logout", func(t *testing.T) {
		// Click on "Logout" button
		mainTestHelper.page.Locator(`a[title="Logout"]`).Click()

		// Wait for login page
		mainTestHelper.page.Locator("#login-scene").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		})
	})
}
