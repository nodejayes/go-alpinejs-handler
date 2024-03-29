package components

import (
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
	Label goalpinejshandler.Component
}

func NewButton(label string) *Button {
	goalpinejshandler.RegisterGlobalStyle(style())
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
