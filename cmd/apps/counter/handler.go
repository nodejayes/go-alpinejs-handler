package counter

import (
	"encoding/json"
	"fmt"
	goalpinejshandler "github.com/nodejayes/go-alpinejs-handler"
	"net/http"
)

type (
	handler struct {
		Value   int       `json:"value"`
		History []history `json:"history"`
	}
	handlerArguments struct {
		Operation string `json:"operation"`
		Value     int    `json:"value"`
	}
)

func (ctx *handler) GetName() string {
	return "counter"
}

func (ctx *handler) GetActionType() string {
	return fmt.Sprintf("[%s] operation", ctx.GetName())
}

func (ctx *handler) Authorized(msg goalpinejshandler.Message, res http.ResponseWriter, req *http.Request, messagePool *goalpinejshandler.MessagePool, tools *goalpinejshandler.Tools) error {
	return nil
}

func (ctx *handler) OnDestroy(clientID string) {
	println(fmt.Sprintf("clientID: %s disconnected clear up something here", clientID))
	ctx.Value = 0
	ctx.History = make([]history, 0)
}

func (ctx *handler) GetDefaultState() any {
	return ctx
}

func (ctx *handler) Handle(msg goalpinejshandler.Message, res http.ResponseWriter, req *http.Request, messagePool *goalpinejshandler.MessagePool, tools *goalpinejshandler.Tools) {
	content, err := json.Marshal(msg.Payload)
	if err != nil {
		return
	}
	var args handlerArguments
	err = json.Unmarshal(content, &args)
	if err != nil {
		return
	}

	switch args.Operation {
	case "add":
		ctx.History = append(ctx.History, history{
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
