package goalpinejshandler

type (
	Message struct {
		Type    string `json:"type"`
		Payload any    `json:"payload"`
	}

	Response struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}
)
