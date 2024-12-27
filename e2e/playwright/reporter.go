package playwright

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"
	"time"
)

type TestResult struct {
	Name          string
	Status        string
	ScreenshotPath string
	Timestamp     time.Time
	Error         string
}

type TestReporter struct {
	Results []TestResult
}

func NewTestReporter() *TestReporter {
	return &TestReporter{
		Results: make([]TestResult, 0),
	}
}

func (r *TestReporter) AddResult(name string, passed bool, screenshotPath string, err string) {
	status := "Passed"
	if !passed {
		status = "Failed"
	}
	
	r.Results = append(r.Results, TestResult{
		Name:          name,
		Status:        status,
		ScreenshotPath: screenshotPath,
		Timestamp:     time.Now(),
		Error:         err,
	})
}

func (r *TestReporter) GenerateHTML() error {
	const tmpl = `
<!DOCTYPE html>
<html>
<head>
    <title>Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .test { margin: 10px 0; padding: 10px; border: 1px solid #ddd; }
        .passed { background-color: #e8f5e9; }
        .failed { background-color: #ffebee; }
        img { max-width: 800px; margin-top: 10px; }
    </style>
</head>
<body>
    <h1>Test Results</h1>
    {{range .Results}}
    <div class="test {{.Status | toLowerCase}}">
        <h3>{{.Name}}</h3>
        <p>Status: {{.Status}}</p>
        <p>Time: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
        {{if .Error}}
        <p>Error: {{.Error}}</p>
        {{end}}
        {{if .ScreenshotPath}}
        <img src="{{.ScreenshotPath}}" alt="Test Screenshot">
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
