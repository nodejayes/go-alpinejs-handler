package goalpinejshandler

import (
	"bytes"
	"fmt"
	"html/template"
)

type ViewTools struct{}

func (ctx *ViewTools) Paint(tmpl Template) template.HTML {
	buf := bytes.NewBuffer([]byte{})
	t := template.Must(template.New(tmpl.Name()).Parse(tmpl.Render()))
	err := t.Execute(buf, tmpl)
	if err != nil {
		return template.HTML(fmt.Sprintf("<p>Error on Render Template: %s</p>", err.Error()))
	}
	return template.HTML(buf.String())
}
