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
	reporter    *TestReporter
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
		reporter:    NewTestReporter(),
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
	th.reporter.AddResult(th.name, false, screenshotPath, errMsg)

	if th.runningInCI {
		th.addToGithubStepSummary(fmt.Sprintf("### ❌ %s", th.name))
	}
}

// addToGithubStepSummary adds a message to the $GITHUB_STEP_SUMMARY
func (th *TestHelper) addToGithubStepSummary(content string) error {
	// Retrieve the path to $GITHUB_STEP_SUMMARY from the environment variable
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		return fmt.Errorf("GITHUB_STEP_SUMMARY environment variable is not set")
	}

	// Open the file in append mode, create if it doesn't exist
	file, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the content to the file
	_, err = file.WriteString(content)
	return err
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
func (pr *PlaywrightRequire) True(value bool, msgAndArgs ...interface{}) {
	if !value {
		screenshotPath := pr.helper.captureScreenshot(pr.helper.name)
		pr.helper.HandleError(screenshotPath, msgAndArgs...)
	} else {
		pr.helper.reporter.AddResult(pr.helper.name, true, "", "")
	}
	pr.Assertions.True(value, msgAndArgs...)
}

// False asserts that the specified value is false and takes a screenshot on failure
func (pr *PlaywrightRequire) False(value bool, msgAndArgs ...interface{}) {
	if value {
		pr.helper.captureScreenshot(pr.helper.name)
	}
	pr.Assertions.False(value, msgAndArgs...)
}

// Equal asserts that two objects are equal and takes a screenshot on failure
func (pr *PlaywrightRequire) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected != actual {
		pr.helper.captureScreenshot(pr.helper.name)
	}
	pr.Assertions.Equal(expected, actual, msgAndArgs...)
}

// NoError asserts that a function returned no error and takes a screenshot on failure
func (pr *PlaywrightRequire) NoError(err error, msgAndArgs ...interface{}) {
	if err != nil {
		pr.helper.captureScreenshot(pr.helper.name)
	}
	pr.Assertions.NoError(err, msgAndArgs...)
}

// Error asserts that a function returned an error and takes a screenshot on failure
func (pr *PlaywrightRequire) Error(err error, msgAndArgs ...interface{}) {
	if err == nil {
		pr.helper.captureScreenshot(pr.helper.name)
	}
	pr.Assertions.Error(err, msgAndArgs...)
}

// Close cleans up resources and generates the report
func (th *TestHelper) Close() {
	if err := th.reporter.GenerateHTML(); err != nil {
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
