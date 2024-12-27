package playwright

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestHelper wraps common test functionality
type TestHelper struct {
	name    string
	page    playwright.Page
	browser playwright.Browser
	context playwright.BrowserContext
	t       require.TestingT
}

// NewTestHelper creates a new test helper instance
func NewTestHelper(t require.TestingT, name string) (*TestHelper, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	context, err := browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("could not create context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	return &TestHelper{
		name:    name,
		page:    page,
		browser: browser,
		context: context,
		t:       t,
	}, nil
}

// Require returns a custom assertion object that takes screenshots on failure
func (th *TestHelper) Require() *PlaywrightRequire {
	return &PlaywrightRequire{
		Assertions: require.New(th.t),
		helper:     th,
	}
}

func (th *TestHelper) HandleError(t *testing.T, screenshotPath string, msg, err string) {
	GetReporter().AddResult(t.Name(), false, screenshotPath, msg, err)
	t.Error(msg) // Also log the error to the test output
}

func (th *TestHelper) HandleSuccess(t *testing.T, message string) {
	GetReporter().AddResult(t.Name(), true, "", message, "")
}

// PlaywrightRequire wraps require.Assertions to add screenshot capability
type PlaywrightRequire struct {
	*require.Assertions
	helper *TestHelper
}

// captureScreenshot saves a screenshot to the screenshots directory
func (th *TestHelper) captureScreenshot(testName string) string {
	timestamp := time.Now().Format("20060102-150405")
	tmpDir, err := os.MkdirTemp("", "playwright-screenshots")
	if err != nil {
		th.t.Errorf("Failed to create temporary directory: %v\n", err)
		return ""
	}
	filePath := filepath.Join(tmpDir, fmt.Sprintf("%s-%s.png", testName, timestamp))

	// Get the full path without the filename from `filename` and create the directories
	if err := os.MkdirAll(path.Dir(filePath), 0755); err != nil {
		th.t.Errorf("Failed to create screenshots directory: %v\n", err)
		return ""
	}

	// Create screenshots directory if it doesn't exist
	if err := os.MkdirAll("screenshots", 0755); err != nil {
		th.t.Errorf("Failed to create screenshots directory: %v\n", err)
		return ""
	}

	// Take screenshot
	if _, err := th.page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(filePath),
		FullPage: playwright.Bool(true),
	}); err != nil {
		th.t.Errorf("Failed to capture screenshot: %v\n", err)
		return ""
	}

	fmt.Printf("Screenshot saved: %s\n", filePath)

	return filePath
}

func (pr *PlaywrightRequire) Assert(t *testing.T, assertFn func() error, msgAndArgs ...interface{}) {
	err := assertFn()
	var msg string
	if len(msgAndArgs) > 0 {
		if format, ok := msgAndArgs[0].(string); ok && len(msgAndArgs) > 1 {
			msg = fmt.Sprintf(format, msgAndArgs[1:]...)
		} else {
			msg = fmt.Sprint(msgAndArgs...)
		}
	}
	if err == nil {
		pr.helper.HandleSuccess(t, msg)
	} else {
		screenshotPath := pr.helper.captureScreenshot(t.Name())
		pr.helper.HandleError(t, screenshotPath, msg, err.Error())
	}
}

// True asserts that the specified value is true and takes a screenshot on failure
func (pr *PlaywrightRequire) True(t *testing.T, value bool, msgAndArgs ...interface{}) {
	pr.Assert(t, func() error {
		var err error
		if !value {
			err = fmt.Errorf("Expected value to be true but got false in test '%s'", t.Name())
		}
		return err
	}, msgAndArgs...)
	pr.Assertions.True(value, msgAndArgs...)
}

// False asserts that the specified value is false and takes a screenshot on failure
func (pr *PlaywrightRequire) False(t *testing.T, value bool, msgAndArgs ...interface{}) {
	pr.Assert(t, func() error {
		var err error
		if value {
			err = fmt.Errorf("Expected value to be false but got true in test '%s'", t.Name())
		}
		return err
	}, msgAndArgs...)
	pr.Assertions.False(value, msgAndArgs...)
}

// Equal asserts that two objects are equal and takes a screenshot on failure
func (pr *PlaywrightRequire) Equal(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	pr.Assert(t, func() error {
		var err error
		if expected != actual {
			err = fmt.Errorf("Expected values to be equal in test '%s':\nexpected: %v\nactual: %v", t.Name(), expected, actual)
		}
		return err
	}, msgAndArgs...)
	pr.Assertions.Equal(expected, actual, msgAndArgs...)
}

// NoError asserts that a function returned no error and takes a screenshot on failure
func (pr *PlaywrightRequire) NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	pr.Assert(t, func() error {
		var assertErr error
		if err != nil {
			assertErr = fmt.Errorf("Expected no error but got error in test '%s': %v", t.Name(), err)
		}
		return assertErr
	}, msgAndArgs...)
	pr.Assertions.NoError(err, msgAndArgs...)
}

// Error asserts that a function returned an error and takes a screenshot on failure
func (pr *PlaywrightRequire) Error(t *testing.T, err error, msgAndArgs ...interface{}) {
	pr.Assert(t, func() error {
		var assertErr error
		if err == nil {
			assertErr = fmt.Errorf("Expected error but got none in test '%s'", t.Name())
		}
		return assertErr
	}, msgAndArgs...)
	pr.Assertions.Error(err, msgAndArgs...)
}

// Close cleans up resources and generates the report
func (th *TestHelper) Close() {
	if err := GetReporter().GenerateHTML(); err != nil {
		fmt.Printf("Failed to generate HTML report: %v\n", err)
	}
	if th.page != nil {
		th.page.Close()
	}
	if th.context != nil {
		th.context.Close()
	}
	if th.browser != nil {
		th.browser.Close()
	}
}
