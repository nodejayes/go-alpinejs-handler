package components

import goalpinejshandler "github.com/nodejayes/go-alpinejs-handler"

type Button struct {
	goalpinejshandler.ViewTools
	Label goalpinejshandler.Template
}

func NewButton(label string) *Button {
	return &Button{
		Label: NewText(label),
	}
}

func (ctx *Button) Name() string {
	return "button"
}

func (ctx *Button) Render() string {
	return `<button>{{ .Paint .Label }}</button>`
}
