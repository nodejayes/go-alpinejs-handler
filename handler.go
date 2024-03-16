package goalpinejshandler

import (
	"fmt"
	"net/http"
)

type processor struct {
	handlers map[string]func(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool)
	messagePool *MessagePool
}

func (ctx *processor) dispatch(message Message, res http.ResponseWriter, req *http.Request) error {
	handler := ctx.handlers[message.Type]
	if handler != nil {
		handler(message, res, req, ctx.messagePool)
		return nil
	}
	return fmt.Errorf("handler %s not found", message.Type)
}

func (ctx *processor) registerHandlers(handlers []ActionHandler, messagePool *MessagePool) {
	ctx.messagePool = messagePool
	for _, handler := range handlers {
		ctx.handlers[handler.GetActionType()] = handler.Handle
	}
}

var actionProcessor = &processor{
	handlers: make(map[string]func(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool)),
}