package goalpinejshandler

import (
	"fmt"
	"net/http"
)

type processor struct {
	tools       *Tools
	protectors  map[string]func(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool, tools *Tools) error
	handlers    map[string]func(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool, tools *Tools)
	messagePool *MessagePool
}

func (ctx *processor) registerTools(tools *Tools) {
	ctx.tools = tools
}

func (ctx *processor) dispatch(message Message, res http.ResponseWriter, req *http.Request) error {
	handler := ctx.handlers[message.Type]
	protector := ctx.protectors[message.Type]
	if handler != nil {
		if protector != nil {
			err := protector(message, res, req, ctx.messagePool, ctx.tools)
			if err != nil {
				return err
			}
		}
		handler(message, res, req, ctx.messagePool, ctx.tools)
		return nil
	}
	return fmt.Errorf("handler %s not found", message.Type)
}

func (ctx *processor) registerHandlers(handlers []ActionHandler, messagePool *MessagePool) {
	ctx.messagePool = messagePool
	for _, handler := range handlers {
		ctx.handlers[handler.GetActionType()] = handler.Handle
		protectedHandler, ok := handler.(protectedActionHandler)
		if ok {
			ctx.protectors[protectedHandler.GetActionType()] = protectedHandler.Authorized
		}
	}
}

var actionProcessor = &processor{
	protectors: make(map[string]func(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool, tools *Tools) error),
	handlers:   make(map[string]func(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool, tools *Tools)),
}
