package renderer

import (
	"bytes"
	"html/template"
	"io"
)

type Renderer struct {
	templates *template.Template
}

func New(dir string) Renderer {
	return Renderer{
		templates: parseHTMLTemplates(dir),
	}
}

func (r Renderer) RenderTemplate(w io.Writer, templateName TemplateName, data interface{}) error {
	// buffer the template execution so that partial writes don't occur in case of errors
	buf := &bytes.Buffer{}
	err := r.templates.ExecuteTemplate(buf, string(templateName), data)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	return err
}
