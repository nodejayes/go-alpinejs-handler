package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	goalpinejshandler "github.com/nodejayes/go-alpinejs-handler"
)

type History struct {
	ID      int    `json:"id"`
	Counter string `json:"counter"`
}

type CounterHandler struct {
	Value   int       `json:"value"`
	History []History `json:"history"`
}

type CounterHandlerArguments struct {
	Operation string `json:"operation"`
	Value     int    `json:"value"`
}

func (ctx *CounterHandler) GetName() string {
	return "counter"
}

func (ctx *CounterHandler) GetActionType() string {
	return fmt.Sprintf("[%s] operation", ctx.GetName())
}

func (ctx *CounterHandler) Authorized(msg goalpinejshandler.Message, res http.ResponseWriter, req *http.Request, messagePool *goalpinejshandler.MessagePool, tools *goalpinejshandler.Tools) error {
	return nil
}

func (ctx *CounterHandler) GetDefaultState() string {
	stream, err := json.Marshal(ctx)
	if err != nil {
		return ""
	}
	return string(stream)
}

func (ctx *CounterHandler) Handle(msg goalpinejshandler.Message, res http.ResponseWriter, req *http.Request, messagePool *goalpinejshandler.MessagePool, tools *goalpinejshandler.Tools) {
	content, err := json.Marshal(msg.Payload)
	if err != nil {
		return
	}
	var args CounterHandlerArguments
	err = json.Unmarshal(content, &args)
	if err != nil {
		return
	}

	switch args.Operation {
	case "add":
		ctx.History = append(ctx.History, History{
			ID:      len(ctx.History) + 1,
			Counter: fmt.Sprintf("Counter %v", ctx.Value),
		})
		ctx.Value += args.Value
	case "sub":
		if ctx.Value > 0 {
			ctx.History = ctx.History[:len(ctx.History)-1]
			ctx.Value -= args.Value
		}
	}

	messagePool.Add(goalpinejshandler.ChannelMessage{
		ClientFilter: func(client goalpinejshandler.Client) bool {
			return true
		},
		Message: goalpinejshandler.Message{
			Type:    fmt.Sprintf("[%s] update", ctx.GetName()),
			Payload: ctx,
		},
	})
}

func main() {
	router := http.NewServeMux()

	config := goalpinejshandler.Config{
		ActionUrl:         "/action",
		EventUrl:          "/events",
		ClientIDHeaderKey: "clientId",
		Handlers: []goalpinejshandler.ActionHandler{
			&CounterHandler{
				Value:   0,
				History: make([]History, 0),
			},
		},
	}
	goalpinejshandler.Register(router, &config)
	router.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
				<title>Test alpinestorehandler</title>
				%s
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
			</body>
		</html>
		`, goalpinejshandler.HeadScripts())))
	})

	http.ListenAndServe(":40000", router)
}
