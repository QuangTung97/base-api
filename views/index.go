package views

import (
	_ "embed"
	"html/template"
)

//go:embed index.html
var indexTmplStr string
var indexTmpl = Load(indexTmplStr)

type IndexData struct {
}

func Index(data IndexData) (template.HTML, error) {
	return indexTmpl.Render(data)
}
