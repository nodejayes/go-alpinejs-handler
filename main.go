package goalpinejshandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	di "github.com/nodejayes/generic-di"
)

var messagesPool = NewMessagePool()

type (
	ActionHandler interface {
		GetName() string
		GetActionType() string
		GetDefaultState() string
		Handle(msg Message, res http.ResponseWriter, req *http.Request, messagePool *MessagePool)
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
		w.Header().Add("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(getJsScript()))
	})
	router.HandleFunc("/alpinestorehandler_app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(getAppScript(*config)))
	})
}

func setupOutgoing(router *http.ServeMux, config *Config) {
	router.HandleFunc(fmt.Sprintf("GET %s", config.EventUrl), func(res http.ResponseWriter, req *http.Request) {
		clientStore := di.Inject[clientStore]()
		clientID := req.URL.Query().Get(config.ClientIDHeaderKey)
		_, err := uuid.Parse(clientID)
		if err != nil {
			jsonResponse(res, http.StatusBadRequest, Response{
				Code:  http.StatusBadRequest,
				Error: "clientId not found in header",
			})
			return
		}
		clientStore.Add(Client{
			ID:       clientID,
			Response: res,
			Request:  req,
		})
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

		res.Header().Set("Content-Type", "text/event-stream")
		res.Header().Set("Cache-Control", "no-cache")
		res.Header().Set("Connection", "keep-alive")

		for msg := range messagesPool.Pull() {
			client := clientStore.Get(msg.ClientFilter)
			if len(client) < 1 {
				continue
			}
			message, err := json.Marshal(msg.Message)
			if err != nil {
				continue
			}
			data, err := formatMessage(string(message))
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			res.Write([]byte(data))
			rc := http.NewResponseController(res)
			err = rc.Flush()
			if err != nil {
				println(err.Error())
				flusher, ok := res.(http.Flusher)
				if !ok {
					println("flusher not supported")
					continue
				}
				flusher.Flush()
				continue
			}
		}
	})
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
