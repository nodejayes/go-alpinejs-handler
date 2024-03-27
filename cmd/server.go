package main

import (
	"github.com/nodejayes/go-alpinejs-handler/cmd/apps/counter"
	"net/http"

	goalpinejshandler "github.com/nodejayes/go-alpinejs-handler"
)

func main() {
	router := http.NewServeMux()

	config := goalpinejshandler.Config{
		ActionUrl:         "/action",
		EventUrl:          "/events",
		ClientIDHeaderKey: "clientId",
		Pages: []goalpinejshandler.Page{
			counter.NewPage(),
		},
	}
	goalpinejshandler.Register(router, &config)

	http.ListenAndServe(":40000", router)
}
