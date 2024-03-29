package goalpinejshandler

import (
	"bytes"
	di "github.com/nodejayes/generic-di"
	"slices"
	"sync"
)

func init() {
	di.Injectable(newStyleRegistry)
}

type styleRegistry struct {
	m      *sync.Mutex
	styles []string
}

func newStyleRegistry() *styleRegistry {
	return &styleRegistry{
		m:      &sync.Mutex{},
		styles: make([]string, 0),
	}
}

func (ctx *styleRegistry) Register(style string) {
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

func (ctx *styleRegistry) Build() string {
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
