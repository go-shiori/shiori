package playwright

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestHelper wraps common test functionality
type TestHelper struct {
	name        string
	page        playwright.Page
	browser     playwright.Browser
	context     playwright.BrowserContext
	t           require.TestingT
	runningInCI bool
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
		name:        name,
		page:        page,
		browser:     browser,
		context:     context,
		t:           t,
		runningInCI: os.Getenv("GITHUB_STEP_SUMMARY") != "",
	}, nil
}

// Require returns a custom assertion object that takes screenshots on failure
func (th *TestHelper) Require() *PlaywrightRequire {
	return &PlaywrightRequire{
		Assertions: require.New(th.t),
		helper:     th,
	}
}

func (th *TestHelper) HandleError(screenshotPath string, msgAndArgs ...interface{}) {
	errMsg := fmt.Sprint(msgAndArgs...)
	GetReporter().AddResult(th.name, false, screenshotPath, errMsg)
}

func (th *TestHelper) HandleSuccess() {
	GetReporter().AddResult(th.name, true, "", "")
}

// PlaywrightRequire wraps require.Assertions to add screenshot capability
type PlaywrightRequire struct {
	*require.Assertions
	helper *TestHelper
}

// captureScreenshot saves a screenshot to the screenshots directory
func (th *TestHelper) captureScreenshot(testName string) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("screenshots/%s-%s.png", testName, timestamp)

	// Get the full path without the filename from `filename` and create the directories
	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		fmt.Printf("Failed to create screenshots directory: %v\n", err)
		panic(err)
	}

	// Create screenshots directory if it doesn't exist
	if err := os.MkdirAll("screenshots", 0755); err != nil {
		fmt.Printf("Failed to create screenshots directory: %v\n", err)
		return ""
	}

	// Take screenshot
	if _, err := th.page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(filename),
		FullPage: playwright.Bool(true),
	}); err != nil {
		fmt.Printf("Failed to capture screenshot: %v\n", err)
		return ""
	}

	fmt.Printf("Screenshot saved: %s\n", filename)

	return filename
}

// True asserts that the specified value is true and takes a screenshot on failure
func (pr *PlaywrightRequire) True(t *testing.T, value bool, msgAndArgs ...interface{}) {
	if !value {
		screenshotPath := pr.helper.captureScreenshot(t.Name())
		msg := fmt.Sprintf("Expected value to be true but got false in test '%s'", t.Name())
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...)
		}
		pr.helper.HandleError(screenshotPath, msg)
	} else {
		pr.helper.HandleSuccess()
	}
	pr.Assertions.True(value, msgAndArgs...)
}

// False asserts that the specified value is false and takes a screenshot on failure
func (pr *PlaywrightRequire) False(t *testing.T, value bool, msgAndArgs ...interface{}) {
	if value {
		screenshotPath := pr.helper.captureScreenshot(t.Name())
		msg := fmt.Sprintf("Expected value to be false but got true in test '%s'", t.Name())
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...)
		}
		pr.helper.HandleError(screenshotPath, msg)
	} else {
		pr.helper.HandleSuccess()
	}
	pr.Assertions.False(value, msgAndArgs...)
}

// Equal asserts that two objects are equal and takes a screenshot on failure
func (pr *PlaywrightRequire) Equal(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected != actual {
		screenshotPath := pr.helper.captureScreenshot(t.Name())
		msg := fmt.Sprintf("Expected values to be equal in test '%s':\nexpected: %v\nactual: %v", t.Name(), expected, actual)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...)
		}
		pr.helper.HandleError(screenshotPath, msg)
	} else {
		pr.helper.HandleSuccess()
	}
	pr.Assertions.Equal(expected, actual, msgAndArgs...)
}

// NoError asserts that a function returned no error and takes a screenshot on failure
func (pr *PlaywrightRequire) NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	if err != nil {
		screenshotPath := pr.helper.captureScreenshot(t.Name())
		msg := fmt.Sprintf("Expected no error but got error in test '%s': %v", t.Name(), err)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...)
		}
		pr.helper.HandleError(screenshotPath, msg)
	} else {
		pr.helper.HandleSuccess()
	}
	pr.Assertions.NoError(err, msgAndArgs...)
}

// Error asserts that a function returned an error and takes a screenshot on failure
func (pr *PlaywrightRequire) Error(t *testing.T, err error, msgAndArgs ...interface{}) {
	if err == nil {
		screenshotPath := pr.helper.captureScreenshot(t.Name())
		msg := fmt.Sprintf("Expected error but got none in test '%s'", t.Name())
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprint(msgAndArgs...)
		}
		pr.helper.HandleError(screenshotPath, msg)
	} else {
		pr.helper.HandleSuccess()
	}
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
