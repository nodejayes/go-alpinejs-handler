package goalpinejshandler

type (
	MessagePool struct {
		messages chan ChannelMessage
	}
	ChannelMessage struct {
		ClientFilter func(client Client) bool
		Message      Message
	}
)

func newMessagePool() *MessagePool {
	return &MessagePool{
		messages: make(chan ChannelMessage),
	}
}

func (ctx *MessagePool) Add(msg ChannelMessage) {
	ctx.messages <- msg
}

func (ctx *MessagePool) Pull() chan ChannelMessage {
	return ctx.messages
}
