package goalpinejshandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	di "github.com/nodejayes/generic-di"
)

const ContentTypeKey = "Content-Type"

var messagesPool = NewMessagePool()

type (
	ActionHandler interface {
		GetName() string
		GetActionType() string
		GetDefaultState() string
		Handle(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool)
	}
	ProtectedActionHandler interface {
		ActionHandler
		Authorized(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool) error
	}
	Config struct {
		EventUrl                string
		ActionUrl               string
		ClientIDHeaderKey       string
		SendConnectedAfterMs    int
		SocketReconnectInterval int
		Handlers                []ActionHandler
	}
)

func Register(router *http.ServeMux, config *Config) {
	if config.SendConnectedAfterMs < 1 {
		config.SendConnectedAfterMs = 500
	}
	if config.SocketReconnectInterval < 1 {
		config.SocketReconnectInterval = 5000
	}
	setupOutgoing(router, config)
	setupIncoming(router, config)
	setupScripts(router, config)
	actionProcessor.registerHandlers(config.Handlers, messagesPool)
}

func setupScripts(router *http.ServeMux, config *Config) {
	router.HandleFunc("/alpinestorehandler_lib.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(ContentTypeKey, "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(getJsScript()))
	})
	router.HandleFunc("/alpinestorehandler_app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(ContentTypeKey, "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(getAppScript(*config)))
	})
}

func setupOutgoing(router *http.ServeMux, config *Config) {
	router.HandleFunc(fmt.Sprintf("GET %s", config.EventUrl), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set(ContentTypeKey, "text/event-stream")
		res.Header().Set("Cache-Control", "no-cache")
		res.Header().Set("Connection", "keep-alive")

		clientStore, clientID, failRegisterInClient := registerInClientStore(req, config, res)
		if failRegisterInClient {
			return
		}

		requestClosed := false
		closeLocker := &sync.Mutex{}
		go func() {
			<-req.Context().Done()
			closeLocker.Lock()
			requestClosed = true
			clients := clientStore.Get(func(client Client) bool { return client.ID == clientID })
			if len(clients) > 0 {
				for _, client := range clients {
					clientStore.Remove(client)
				}
			}
			closeLocker.Unlock()
		}()

		sendConnectedInfo(config, clientID)

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

func sendConnectedInfo(config *Config, clientID string) {
	go func() {
		time.Sleep(time.Duration(config.SendConnectedAfterMs) * time.Millisecond)
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

func registerInClientStore(req *http.Request, config *Config, res http.ResponseWriter) (*clientStore, string, bool) {
	clientStore := di.Inject[clientStore]()
	clientID := req.URL.Query().Get(config.ClientIDHeaderKey)
	_, err := uuid.Parse(clientID)
	if err != nil {
		jsonResponse(res, http.StatusBadRequest, Response{
			Code:  http.StatusBadRequest,
			Error: "clientId not found in header",
		})
		return nil, "", true
	}
	clientStore.Add(Client{
		ID:       clientID,
		Response: res,
		Request:  req,
	})
	return clientStore, clientID, false
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
