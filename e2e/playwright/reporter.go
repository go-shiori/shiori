package playwright

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AssertionResult struct {
	Message    string
	Status     string
	Error      string
	Screenshot string // Base64 screenshot, only for failures
}

type TestResult struct {
	Name       string
	Status     string
	Timestamp  time.Time
	Assertions []AssertionResult
}

type TestReporter struct {
	Results map[string]*TestResult
}

var globalReporter = &TestReporter{
	Results: make(map[string]*TestResult),
}

func GetReporter() *TestReporter {
	return globalReporter
}

func (r *TestReporter) AddResult(testName string, passed bool, screenshotPath string, message, errorMessage string) {
	status := "Passed"
	if !passed {
		status = "Failed"
	}

	var screenshot string
	if !passed && screenshotPath != "" {
		imageFile, err := os.Open(screenshotPath)
		if err == nil {
			defer imageFile.Close()
			if data, err := io.ReadAll(imageFile); err == nil {
				screenshot = "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
			} else {
				fmt.Printf("Failed to read screenshot %s: %v\n", screenshotPath, err)
			}
		} else {
			fmt.Printf("Failed to open screenshot %s: %v\n", screenshotPath, err)
		}
	}

	// Get or create test result
	testResult, exists := r.Results[testName]
	if !exists {
		testResult = &TestResult{
			Name:       testName,
			Status:     "Passed",
			Timestamp:  time.Now(),
			Assertions: make([]AssertionResult, 0),
		}
		r.Results[testName] = testResult
	}

	// Add assertion result
	testResult.Assertions = append(testResult.Assertions, AssertionResult{
		Message:    message,
		Error:      errorMessage,
		Status:     status,
		Screenshot: screenshot,
	})

	// Update test status if any assertion failed
	if !passed {
		testResult.Status = "Failed"
	}
}

func (r *TestReporter) GenerateHTML() error {
	const tmpl = `
<!DOCTYPE html>
<html>
<head>
    <title>Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .test { margin: 20px 0; padding: 15px; border: 1px solid #ddd; }
        .test.passed { background-color: #e8f5e9; }
        .test.failed { background-color: #ffebee; }
        .assertions { margin: 10px 0; }
        .assertion.failed { font-weight: bold; }
        img { max-width: 800px; margin: 10px 0; }
        .assertion-msg { font-weight: bold; }
        .error-details { color: #d32f2f; margin: 5px 0; }
    </style>
</head>
<body>
    <h1>Test Results</h1>
    {{range .Results}}
    <div class="test {{.Status | toLowerCase}}">
        <h3>{{.Name}}</h3>
        <p><b>Status:</b> {{.Status}}</p>
		{{if eq .Status "Failed"}}
        <ul class="assertions">
            {{range .Assertions}}
                <li class="assertion {{.Status | toLowerCase}}">
                    <p>{{if eq .Status "Passed"}}✓ {{end}}{{.Message}}</p>
                    <p class="error-details">{{.Error}}</p>
                    {{if .Screenshot}}<p><img src="{{.Screenshot | safeURL}}" alt="Failure screenshot"></p>{{end}}
                </li>
            {{end}}
        </ul>
  		{{end}}
    </div>
    {{end}}
</body>
</html>`

	t := template.New("report")
	t = t.Funcs(template.FuncMap{
		"toLowerCase": strings.ToLower,
		"safeHTML":    func(s string) template.HTML { return template.HTML(s) },
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
	})

	t, err := t.Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	if err := os.MkdirAll("test-results", 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %v", err)
	}

	basePath := os.Getenv("CONTEXT_PATH")
	if basePath == "" {
		basePath = "."
	}

	f, err := os.Create(filepath.Join(basePath, "e2e-report.html"))
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer f.Close()

	return t.Execute(f, r)
}
