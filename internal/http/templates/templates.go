package templates

import (
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
	views "github.com/go-shiori/shiori/internal/view"
)

const (
	leftTemplateDelim  = "$$"
	rightTemplateDelim = "$$"
)

// SetupTemplates sets up the templates for the webserver.
func SetupTemplates(engine *gin.Engine) error {
	engine.Delims(leftTemplateDelim, rightTemplateDelim)
	tmpl, err := template.New("html").Delims(leftTemplateDelim, rightTemplateDelim).ParseFS(views.Templates, "*.html")
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}
	engine.SetHTMLTemplate(tmpl)
	return nil
}
