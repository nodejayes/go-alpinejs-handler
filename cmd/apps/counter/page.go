package counter

import (
	"fmt"
	goalpinejshandler "github.com/nodejayes/go-alpinejs-handler"
	"github.com/nodejayes/go-alpinejs-handler/cmd/components"
)

type Page struct {
	goalpinejshandler.ViewTools
	CustomButton goalpinejshandler.Template
}

func NewPage() *Page {
	return &Page{
		CustomButton: components.NewButton("Custom Button"),
	}
}

func (ctx *Page) Name() string {
	return "counter"
}

func (ctx *Page) Route() string {
	return fmt.Sprintf("/%s", ctx.Name())
}

func (ctx *Page) Handlers() []goalpinejshandler.ActionHandler {
	return []goalpinejshandler.ActionHandler{
		&handler{
			Value:   0,
			History: make([]history, 0),
		},
	}
}

func (ctx *Page) Render() string {
	return `
		<!DOCTYPE html>
		<html>
			<head>
				<title>Test alpinestorehandler</title>
				{{ template "alpinejs" }}
				{{ template "alpinejs_handler_lib" }}
				{{ template "alpinejs_handler_stores" }}
			</head>
			<body>
				<div x-data="$store.counter.state" x-init="$store.counter.emit({operation:'get'})">
					<span x-text="value"></span>
					<ul>
						<template x-for="hist in history">
							<li>
								<span x-text="hist.id"></span>
								<span x-text="hist.counter"></span>
							</li>
						</template>
					</ul>
				</div>
				<button x-data @click="$store.counter.emit({operation:'add',value:1})">+</button>
				<button x-data @click="$store.counter.emit({operation:'sub',value:1})">-</button>
				{{ .Paint .CustomButton }}
			</body>
		</html>
`
}
