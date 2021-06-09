package renderer

import (
	"html/template"
	"path"
)

type TemplateName string

const (
	TemplateNameIndex   TemplateName = "index"
	TemplateNameSuccess TemplateName = "success"
)

func parseHTMLTemplates(dir string) *template.Template {
	tpl := template.Must(template.ParseGlob(path.Join(dir, "*.gohtml")))
	return tpl
}
