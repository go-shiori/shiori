package templates

import (
	"fmt"
	"html/template"
	"io"

	views "github.com/go-shiori/shiori/internal/view"
)

const (
	leftTemplateDelim  = "$$"
	rightTemplateDelim = "$$"
)

var templates *template.Template

// SetupTemplates initializes the templates for the webserver
func SetupTemplates() error {
	var err error
	templates, err = template.New("html").
		Delims(leftTemplateDelim, rightTemplateDelim).
		ParseFS(views.Templates, "*.html")
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}
	return nil
}

// RenderTemplate renders a template with the given data
func RenderTemplate(w io.Writer, name string, data any) error {
	if templates == nil {
		if err := SetupTemplates(); err != nil {
			return fmt.Errorf("failed to setup templates: %w", err)
		}
	}
	return templates.ExecuteTemplate(w, name, data)
}
