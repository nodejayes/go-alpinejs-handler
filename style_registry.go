package goalpinejshandler

import (
	"bytes"
	di "github.com/nodejayes/generic-di"
	"slices"
	"sync"
)

func init() {
	di.Injectable(NewStyleRegistry)
}

type StyleRegistry struct {
	m      *sync.Mutex
	styles []string
}

func NewStyleRegistry() *StyleRegistry {
	return &StyleRegistry{
		m:      &sync.Mutex{},
		styles: make([]string, 0),
	}
}

func (ctx *StyleRegistry) Register(style string) {
	ctx.m.Lock()
	defer ctx.m.Unlock()

	idx := slices.IndexFunc(ctx.styles, func(s string) bool {
		return s == style
	})
	if idx >= 0 {
		return
	}
	ctx.styles = append(ctx.styles, style)
}

func (ctx *StyleRegistry) Build() string {
	ctx.m.Lock()
	defer ctx.m.Unlock()

	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(`<style type="text/css">`)
	for _, style := range ctx.styles {
		buf.WriteString(style)
	}
	buf.WriteString(`</style>`)
	return buf.String()
}
