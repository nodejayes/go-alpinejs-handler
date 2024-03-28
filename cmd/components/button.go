package components

import (
	di "github.com/nodejayes/generic-di"
	goalpinejshandler "github.com/nodejayes/go-alpinejs-handler"
)

func style() string {
	return `
	button.primary {
		background-color: red;
	}`
}

type Button struct {
	goalpinejshandler.ViewTools
	Label goalpinejshandler.Template
}

func NewButton(label string) *Button {
	di.Inject[goalpinejshandler.StyleRegistry]().Register(style())
	return &Button{
		Label: NewText(label),
	}
}

func (ctx *Button) Name() string {
	return "button"
}

func (ctx *Button) Render() string {
	return `<button class="primary">{{ .Paint .Label }}</button>`
}
