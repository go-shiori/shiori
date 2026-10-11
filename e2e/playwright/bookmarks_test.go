package playwright

import (
	"fmt"
	"testing"

	"github.com/go-shiori/shiori/e2e/e2eutil"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBookmarks(t *testing.T) {
	testServerURL := e2eutil.StartTestServer(t)

	container := e2eutil.NewShioriContainer(t, "")
	baseURL := fmt.Sprintf("http://localhost:%s", container.GetPort())

	mainTestHelper, err := NewTestHelper(t, "main")
	require.NoError(t, err)
	defer mainTestHelper.Close()

	// Login as admin
	t.Run("001 login", func(t *testing.T) {
		_, err = mainTestHelper.page.Goto(baseURL)
		mainTestHelper.Require().NoError(t, err, "Navigate to base URL")

		usernameLocator := mainTestHelper.page.Locator("#username")
		mainTestHelper.Require().NoError(t, usernameLocator.WaitFor(), "Wait for username field")
		mainTestHelper.Require().NoError(t, usernameLocator.Fill("shiori"), "Fill username")
		mainTestHelper.Require().NoError(t, mainTestHelper.page.Locator("#password").Fill("gopher"), "Fill password")
		mainTestHelper.Require().NoError(t, mainTestHelper.page.Locator(".button").Click(), "Click login")
		mainTestHelper.Require().NoError(t, mainTestHelper.page.Locator("#bookmarks-grid").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(1000),
		}), "Wait for bookmarks grid")
	})

	t.Run("002 create bookmark with URL only", func(t *testing.T) {
		// Click add bookmark button
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Add new bookmark"]`).Click(),
			"Click add new bookmark")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for add bookmark dialog")

		// Fill URL only
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="url"]`).Fill(testServerURL+"/index.html"),
			"Fill URL field")

		// Submit
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click OK button")

		// Wait for dialog to close (longer timeout since bookmark processing fetches the URL)
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(10000),
			}),
			"Wait for dialog to close")

		// Verify bookmark appears in grid
		bookmarkCount, err := mainTestHelper.page.Locator("#bookmarks-grid .bookmark").Count()
		mainTestHelper.Require().NoError(t, err, "Count bookmarks")
		mainTestHelper.Require().Equal(t, 1, bookmarkCount, "One bookmark should exist")
	})

	t.Run("003 create bookmark with custom title and excerpt", func(t *testing.T) {
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Add new bookmark"]`).Click(),
			"Click add new bookmark")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for add bookmark dialog")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="url"]`).Fill(testServerURL+"/custom.html"),
			"Fill URL")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="title"]`).Fill("My Custom Title"),
			"Fill title")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="excerpt"]`).Fill("My custom excerpt text"),
			"Fill excerpt")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click OK button")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(10000),
			}),
			"Wait for dialog to close")

		// Verify bookmark with custom title appears (first in list since newest)
		titleText, err := mainTestHelper.page.Locator("#bookmarks-grid .bookmark").First().Locator(".title").TextContent()
		mainTestHelper.Require().NoError(t, err, "Get bookmark title")
		mainTestHelper.Require().Contains(t, titleText, "My Custom Title", "Bookmark should have custom title")

		bookmarkCount, err := mainTestHelper.page.Locator("#bookmarks-grid .bookmark").Count()
		mainTestHelper.Require().NoError(t, err, "Count bookmarks")
		mainTestHelper.Require().Equal(t, 2, bookmarkCount, "Two bookmarks should exist")
	})

	t.Run("004 create bookmark with archive option", func(t *testing.T) {
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Add new bookmark"]`).Click(),
			"Click add new bookmark")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for add bookmark dialog")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="url"]`).Fill(testServerURL+"/archive.html"),
			"Fill URL")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="create_archive"]`).Check(),
			"Check create archive")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click OK button")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(10000),
			}),
			"Wait for dialog to close")

		bookmarkCount, err := mainTestHelper.page.Locator("#bookmarks-grid .bookmark").Count()
		mainTestHelper.Require().NoError(t, err, "Count bookmarks")
		mainTestHelper.Require().Equal(t, 3, bookmarkCount, "Three bookmarks should exist")
	})

	t.Run("005 create bookmark with ebook option", func(t *testing.T) {
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Add new bookmark"]`).Click(),
			"Click add new bookmark")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for add bookmark dialog")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="url"]`).Fill(testServerURL+"/ebook.html"),
			"Fill URL")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="create_ebook"]`).Check(),
			"Check create ebook")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click OK button")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(10000),
			}),
			"Wait for dialog to close")

		bookmarkCount, err := mainTestHelper.page.Locator("#bookmarks-grid .bookmark").Count()
		mainTestHelper.Require().NoError(t, err, "Count bookmarks")
		mainTestHelper.Require().Equal(t, 4, bookmarkCount, "Four bookmarks should exist")
	})

	t.Run("006 create bookmark with tags", func(t *testing.T) {
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Add new bookmark"]`).Click(),
			"Click add new bookmark")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for add bookmark dialog")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="url"]`).Fill(testServerURL+"/tagged.html"),
			"Fill URL")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="tags"]`).Fill("test-tag"),
			"Fill tags")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click OK button")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(10000),
			}),
			"Wait for dialog to close")

		// Verify the newest bookmark (first in list) has the tag
		tagText, err := mainTestHelper.page.Locator("#bookmarks-grid .bookmark").First().Locator(".bookmark-tags").TextContent()
		mainTestHelper.Require().NoError(t, err, "Get bookmark tags text")
		mainTestHelper.Require().Contains(t, tagText, "test-tag", "Bookmark should have test-tag")
	})

	t.Run("007 create bookmark as public", func(t *testing.T) {
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Add new bookmark"]`).Click(),
			"Click add new bookmark")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for add bookmark dialog")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="url"]`).Fill(testServerURL+"/public.html"),
			"Fill URL")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="makePublic"]`).Check(),
			"Check make public")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click OK button")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(10000),
			}),
			"Wait for dialog to close")

		// Verify the newest bookmark has the public icon (fa-eye)
		publicIcon := mainTestHelper.page.Locator("#bookmarks-grid .bookmark").First().Locator(".fa-eye")
		visible, err := publicIcon.IsVisible()
		mainTestHelper.Require().NoError(t, err, "Check public icon visibility")
		mainTestHelper.Require().True(t, visible, "Public bookmark should show eye icon")
	})

	t.Run("008 empty URL shows error", func(t *testing.T) {
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Add new bookmark"]`).Click(),
			"Click add new bookmark")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for add bookmark dialog")

		// Leave URL empty, click OK
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click OK with empty URL")

		// Should show error dialog
		dialogContent := mainTestHelper.page.Locator(".custom-dialog-content")
		mainTestHelper.Require().NoError(t,
			dialogContent.WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(2000),
			}),
			"Wait for error dialog")

		errorText, err := dialogContent.TextContent()
		mainTestHelper.Require().NoError(t, err, "Get error dialog text")
		mainTestHelper.Require().Equal(t, "URL must not empty", errorText, "Verify empty URL error message")

		// Dismiss error dialog
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Dismiss error dialog")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(1000),
			}),
			"Wait for error dialog to close")
	})

	t.Run("009 batch add tags keeps existing tags", func(t *testing.T) {
		// The bookmark created with "test-tag" is second, after the public one
		taggedBookmark := mainTestHelper.page.Locator("#bookmarks-grid .bookmark").Nth(1)
		tagText, err := taggedBookmark.Locator(".bookmark-tags").TextContent()
		mainTestHelper.Require().NoError(t, err, "Get bookmark tags text")
		mainTestHelper.Require().Contains(t, tagText, "test-tag", "Bookmark should start with test-tag")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Batch edit"]`).Click(),
			"Click batch edit")
		mainTestHelper.Require().NoError(t,
			taggedBookmark.Locator(".bookmark-selector").Click(),
			"Select tagged bookmark")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[title="Add tags"]`).Click(),
			"Click add tags")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(1000),
			}),
			"Wait for add tags dialog")

		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`[name="tags"]`).Fill("bulk-tag"),
			"Fill tags")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(`.custom-dialog-button.main`).Click(),
			"Click OK button")
		mainTestHelper.Require().NoError(t,
			mainTestHelper.page.Locator(".custom-dialog").WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateHidden,
				Timeout: playwright.Float(5000),
			}),
			"Wait for dialog to close")

		tagText, err = taggedBookmark.Locator(".bookmark-tags").TextContent()
		mainTestHelper.Require().NoError(t, err, "Get bookmark tags text")
		mainTestHelper.Require().Contains(t, tagText, "test-tag", "Bookmark should keep test-tag")
		mainTestHelper.Require().Contains(t, tagText, "bulk-tag", "Bookmark should have bulk-tag")
	})
}
