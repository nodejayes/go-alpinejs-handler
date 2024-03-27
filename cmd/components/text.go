package components

type Text struct {
	Content string
}

func NewText(content string) *Text {
	return &Text{
		Content: content,
	}
}

func (ctx *Text) Name() string {
	return "text"
}

func (ctx *Text) Render() string {
	return `<span>{{ .Content }}</span>`
}
