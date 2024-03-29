package goalpinejshandler

import (
	"bytes"
	"fmt"
	di "github.com/nodejayes/generic-di"
	"html/template"
)

type ViewTools struct{}

func (ctx *ViewTools) Paint(tmpl Component) template.HTML {
	buf := bytes.NewBuffer([]byte{})
	t := template.Must(template.New(tmpl.Name()).Parse(tmpl.Render()))
	err := t.Execute(buf, tmpl)
	if err != nil {
		return template.HTML(fmt.Sprintf("<p>Error on Render Component: %s</p>", err.Error()))
	}
	return template.HTML(buf.String())
}

func (ctx *ViewTools) Style(names ...string) template.HTML {
	if len(names) < 1 {
		return template.HTML(di.Inject[styleRegistry]("global").Build())
	}
	buf := bytes.NewBuffer([]byte{})
	for _, name := range names {
		buf.WriteString(di.Inject[styleRegistry](name).Build())
	}
	return template.HTML(buf.String())
}
