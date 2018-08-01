package message

import "fleta/packet"

type Handler interface {
	Resolver
	Processor
}

type Resolver interface {
	Resolve(t MessageType, payload packet.Payload) (Message, Processor, int64, error)
}

type Processor interface {
	Process(msg Message) error
}

type BaseHandler struct {
}
