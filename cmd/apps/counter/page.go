package counter

import (
	"fmt"
	goalpinejshandler "github.com/nodejayes/go-alpinejs-handler"
	"github.com/nodejayes/go-alpinejs-handler/cmd/components"
)

const pageId = "counter"

type Page struct {
	goalpinejshandler.ViewTools
	CustomButton1 goalpinejshandler.Component
	CustomButton2 goalpinejshandler.Component
}

func style() string {
	return `
	* {
	  font-family: system-ui;
	  font-size: 15px;
	  margin: 0;
	  padding: 0;
	}
	html, body {
	  width: 100vw;
	  height: 100vh;
	}`
}

func NewPage() *Page {
	goalpinejshandler.RegisterStyle(pageId, style())
	return &Page{
		CustomButton1: components.NewButton("Custom Button 1"),
		CustomButton2: components.NewButton("Custom Button 2"),
	}
}

func (ctx *Page) Name() string {
	return pageId
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
				{{ .Style }}
				{{ .Style "counter" }}
			</head>
			<body>
				<h1>Überflieger</h1>
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
				{{ .Paint .CustomButton1 }}
				{{ .Paint .CustomButton2 }}
			</body>
		</html>
`
}
