package playwright

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
)

type AssertionResult struct {
	Message        string
	Status         string
	ScreenshotPath string
	ScreenshotB64  string
	Timestamp      time.Time
	Error          string
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

func (r *TestReporter) AddResult(testName string, assertionMsg string, passed bool, screenshotPath string, err string) {
	status := "Passed"
	if !passed {
		status = "Failed"
	}

	var b64Screenshot string
	if screenshotPath != "" {
		if data, err := os.ReadFile(screenshotPath); err == nil {
			b64Screenshot = "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
		} else {
			fmt.Printf("Failed to read screenshot %s: %v\n", screenshotPath, err)
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
		Message:        assertionMsg,
		Status:         status,
		ScreenshotPath: screenshotPath,
		ScreenshotB64:  b64Screenshot,
		Timestamp:      time.Now(),
		Error:          err,
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
        .assertion { margin: 10px 0; padding: 10px; border: 1px solid #eee; }
        .passed { background-color: #e8f5e9; }
        .failed { background-color: #ffebee; }
        img { max-width: 800px; margin-top: 10px; }
        h4 { margin: 5px 0; }
    </style>
</head>
<body>
    <h1>Test Results</h1>
    {{range .Results}}
    <div class="test {{.Status | toLowerCase}}">
        <h3>{{.Name}}</h3>
        <p>Status: {{.Status}}</p>
        <p>Time: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
        
        {{range .Assertions}}
        <div class="assertion {{.Status | toLowerCase}}">
            <h4>{{.Message}}</h4>
            <p>Status: {{.Status}}</p>
            <p>Time: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
            {{if .Error}}
            <p>Error: {{.Error}}</p>
            {{end}}
            {{if .ScreenshotB64}}
            <img src="{{.ScreenshotB64}}" alt="Test Screenshot">
            {{end}}
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>`

	t := template.New("report")
	t = t.Funcs(template.FuncMap{
		"toLowerCase": strings.ToLower,
	})

	t, err := t.Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	if err := os.MkdirAll("test-results", 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %v", err)
	}

	f, err := os.Create("test-results/report.html")
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer f.Close()

	return t.Execute(f, r)
}
