package goalpinejshandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/google/uuid"
	di "github.com/nodejayes/generic-di"
)

const ContentTypeKey = "Content-Type"

var messagesPool = newMessagePool()

type (
	Component interface {
		Name() string
		Render() string
	}
	Page interface {
		Component
		Route() string
		Handlers() []ActionHandler
	}
	ActionHandler interface {
		GetName() string
		GetActionType() string
		GetDefaultState() any
		Handle(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool, tools *Tools)
	}
	protectedActionHandler interface {
		ActionHandler
		Authorized(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool, tools *Tools) error
	}
	destroyableHandler interface {
		OnDestroy(clientId string)
	}
	Config struct {
		EventUrl                string
		ActionUrl               string
		ClientIDHeaderKey       string
		SocketReconnectInterval int
		Pages                   []Page
	}
)

func Register(router *http.ServeMux, config *Config) {
	if config.SocketReconnectInterval < 1 {
		config.SocketReconnectInterval = 5000
	}
	setupOutgoing(router, config)
	setupIncoming(router, config)
	actionProcessor.registerTools(newTools(config))
	for _, page := range config.Pages {
		actionProcessor.registerHandlers(page.Handlers(), messagesPool)
		router.HandleFunc(page.Route(), usePage(page, config))
	}
}

func RegisterGlobalStyle(style string) {
	di.Inject[styleRegistry]("global").Register(style)
}

func RegisterStyle(name, style string) {
	di.Inject[styleRegistry](name).Register(style)
}

func usePage(page Page, config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer([]byte{})
		tmpl, err := template.New(page.Name()).Parse(page.Render())
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte{})
			return
		}
		tmpl = addScriptTemplates(tmpl, config, page.Handlers())
		err = tmpl.ExecuteTemplate(buf, page.Name(), page)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte{})
			return
		}
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf.Bytes())
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte{})
			return
		}
	}
}

func setupOutgoing(router *http.ServeMux, config *Config) {
	router.HandleFunc(fmt.Sprintf("GET %s", config.EventUrl), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set(ContentTypeKey, "text/event-stream")
		res.Header().Set("Cache-Control", "no-cache")
		res.Header().Set("Connection", "keep-alive")
		connectionID := uuid.NewString()

		clientStore, clientID, failRegisterInClient := registerInClientStore(req, config, res, connectionID)
		if failRegisterInClient {
			return
		}

		requestClosed := false
		closeLocker := &sync.Mutex{}
		go func() {
			<-req.Context().Done()
			closeLocker.Lock()
			requestClosed = true
			clients := clientStore.Get(func(client Client) bool { return client.ConnectionID == connectionID })
			if len(clients) > 0 {
				for _, client := range clients {
					clientStore.Remove(client)
				}
			}
			clientIDClients := clientStore.Get(func(client Client) bool {
				return client.ID == clientID
			})
			if len(clientIDClients) < 1 {
				for _, page := range config.Pages {
					for _, handler := range page.Handlers() {
						destroyableHandler, ok := handler.(destroyableHandler)
						if ok {
							destroyableHandler.OnDestroy(clientID)
						}
					}
				}
			}
			closeLocker.Unlock()
		}()

		sendConnectedInfo(clientID)

		for msg := range messagesPool.Pull() {
			if requestClosed {
				return
			}
			client := clientStore.Get(msg.ClientFilter)
			if len(client) < 1 {
				continue
			}
			for _, c := range client {
				c.SendMessage(msg)
			}
		}
	})
}

func sendConnectedInfo(clientID string) {
	go func() {
		messagesPool.Add(ChannelMessage{
			Message: Message{
				Type:    "connected",
				Payload: clientID,
			},
			ClientFilter: func(client Client) bool {
				return client.ID == clientID
			},
		})
	}()
}

func registerInClientStore(req *http.Request, config *Config, res http.ResponseWriter, connectionID string) (*clientStore, string, bool) {
	cls := di.Inject[clientStore]()
	clientID := req.URL.Query().Get(config.ClientIDHeaderKey)
	_, err := uuid.Parse(clientID)
	if err != nil {
		jsonResponse(res, http.StatusBadRequest, Response{
			Code:  http.StatusBadRequest,
			Error: "clientId not found in header",
		})
		return nil, "", true
	}
	cls.Add(Client{
		ID:           clientID,
		ConnectionID: connectionID,
		Response:     res,
		Request:      req,
	})
	return cls, clientID, false
}

func setupIncoming(router *http.ServeMux, config *Config) {
	router.HandleFunc(fmt.Sprintf("POST %s", config.ActionUrl), func(res http.ResponseWriter, req *http.Request) {
		var msg Message
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			jsonResponse(res, http.StatusInternalServerError, Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			})
			return
		}
		err = actionProcessor.dispatch(msg, res, req)
		if err != nil {
			jsonResponse(res, http.StatusInternalServerError, Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			})
			return
		}
		jsonResponse(res, http.StatusOK, Response{
			Code:  http.StatusOK,
			Error: "",
		})
	})
}
