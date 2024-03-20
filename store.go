package goalpinejshandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sync"

	di "github.com/nodejayes/generic-di"
)

func init() {
	di.Injectable(newClientStore)
}

type (
	Client struct {
		ID       string
		Response http.ResponseWriter
		Request  *http.Request
	}
	clientStore struct {
		m       *sync.Mutex
		Clients map[string][]Client
	}
)

func newClientStore() *clientStore {
	return &clientStore{
		m:       &sync.Mutex{},
		Clients: make(map[string][]Client),
	}
}

func (ctx *clientStore) Add(client Client) {
	ctx.m.Lock()
	defer ctx.m.Unlock()

	if ctx.Clients[client.ID] == nil {
		ctx.Clients[client.ID] = make([]Client, 0)
	}
	ctx.Clients[client.ID] = append(ctx.Clients[client.ID], client)
}

func (ctx *clientStore) Remove(client Client) {
	ctx.m.Lock()
	defer ctx.m.Unlock()

	if slices.Contains(ctx.Clients[client.ID], client) {
		ctx.Clients[client.ID] = slices.Delete(ctx.Clients[client.ID], slices.Index(ctx.Clients[client.ID], client), 1)
	}

	if len(ctx.Clients[client.ID]) < 1 {
		delete(ctx.Clients, client.ID)
	}
}

func (ctx *clientStore) Get(filter func(client Client) bool) []Client {
	result := make([]Client, 0)
	for _, cl := range ctx.Clients {
		for _, c := range cl {
			if filter(c) {
				result = append(result, c)
			}
		}
	}
	return result
}

func (ctx *Client) SendMessage(msg ChannelMessage) {
	message, err := json.Marshal(msg.Message)
	if err != nil {
		return
	}
	data, err := formatMessage(string(message))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ctx.Response.Write([]byte(data))
	rc := http.NewResponseController(ctx.Response)
	err = rc.Flush()
	if err != nil {
		println(err.Error())
		flusher, ok := ctx.Response.(http.Flusher)
		if !ok {
			println("flusher not supported")
			return
		}
		flusher.Flush()
	}
}
