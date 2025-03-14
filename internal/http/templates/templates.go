package templates

import (
	"fmt"
	"html/template"
	"io"

	"github.com/go-shiori/shiori/internal/config"
	views "github.com/go-shiori/shiori/internal/view"
	webapp "github.com/go-shiori/shiori/webapp"
)

const (
	leftTemplateDelim  = "$$"
	rightTemplateDelim = "$$"
)

var templates *template.Template

// SetupTemplates initializes the templates for the webserver
func SetupTemplates(config *config.Config) error {
	var err error
	fs := views.Templates

	globs := []string{"*.html"}

	if config.Http.ServeWebUIV2 {
		fs = webapp.Templates
		globs = []string{"**/*.html"}
	}

	templates, err = template.New("html").
		Delims(leftTemplateDelim, rightTemplateDelim).
		ParseFS(fs, globs...)

	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}
	return nil
}

// RenderTemplate renders a template with the given data
func RenderTemplate(w io.Writer, name string, data any) error {
	if templates == nil {
		return fmt.Errorf("templates not initialized")
	}
	return templates.ExecuteTemplate(w, name, data)
}
