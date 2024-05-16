package views

import (
	"bytes"
	"html/template"
)

type Template struct {
	tmpl *template.Template
}

func Load(templateStr string) *Template {
	t, err := template.New("empty").Parse(templateStr)
	if err != nil {
		panic(err)
	}
	return &Template{
		tmpl: t,
	}
}

func (t *Template) Render(data any) (template.HTML, error) {
	var buf bytes.Buffer
	err := t.tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
