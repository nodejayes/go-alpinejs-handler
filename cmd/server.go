package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	goalpinejshandler "github.com/nodejayes/go-alpinejs-handler"
)

type CounterHandler struct {
	Value int `json:"value"`
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

func (ctx *CounterHandler) GetDefaultState() string {
	stream, err := json.Marshal(ctx)
	if err != nil {
		return ""
	}
	return string(stream)
}

func (ctx *CounterHandler) Handle(msg goalpinejshandler.Message, res http.ResponseWriter, req *http.Request, messagePool *goalpinejshandler.MessagePool) {
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
		ctx.Value += args.Value
	case "sub":
		ctx.Value -= args.Value
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
		ActionUrl:            "/action",
		EventUrl:             "/events",
		ClientIDHeaderKey:    "clientId",
		SendConnectedAfterMs: 100,
		Handlers: []goalpinejshandler.ActionHandler{
			&CounterHandler{},
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
				<div x-data="$store.counter.state">
					<span x-text="value"></span>
				</div>
				<button x-data @click="$store.counter.emit({operation:'add',value:1})">+</button>
				<button x-data @click="$store.counter.emit({operation:'sub',value:1})">-</button>
			</body>
		</html>
		`, goalpinejshandler.HeadScripts())))
	})

	http.ListenAndServe(":40000", router)
}
