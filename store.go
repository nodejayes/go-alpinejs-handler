package goalpinejshandler

import (
	"net/http"

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
		Clients map[string]Client
	}
)

func newClientStore() *clientStore {
	return &clientStore{
		Clients: make(map[string]Client),
	}
}

func (ctx *clientStore) Add(client Client) {
	ctx.Clients[client.ID] = client
}

func (ctx *clientStore) Remove(client Client) {
	delete(ctx.Clients, client.ID)
}

func (ctx *clientStore) Get(filter func(client Client) bool) []Client {
	result := make([]Client, 0)
	for _, cl := range ctx.Clients {
		if filter(cl) {
			result = append(result, cl)
		}
	}
	return result
}
