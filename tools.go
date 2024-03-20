package goalpinejshandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Tools struct {
	config *Config
}

func NewTools(config *Config) *Tools {
	return &Tools{
		config: config,
	}
}

func (ctx *Tools) GetClientId(req *http.Request) string {
	return req.Header.Get(ctx.config.ClientIDHeaderKey)
}

func jsonResponse(res http.ResponseWriter, statusCode int, data any) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	str, _ := json.Marshal(data)
	res.Write(str)
}

func formatMessage(data string) (string, error) {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("data: %v\n", data))
	sb.WriteString("\n")

	return sb.String(), nil
}
