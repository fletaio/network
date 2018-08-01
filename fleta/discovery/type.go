package discovery

import (
	"fleta/message"
)

var (
	// PingMessageType TODO
	// PingMessageType = message.MessageType(15)
	PingMessageType = message.DefineType("PING")
	// PongMessageType TODO
	// PongMessageType = message.MessageType(iota)
	PongMessageType = message.DefineType("PONG")
	// FindNodeMessageType TODO
	// FindNodeMessageType = message.MessageType(iota)
	FindNodeMessageType = message.DefineType("FINDNODE")
	// NodeInfoMessageType TODO
	// NodeInfoMessageType = message.MessageType(iota)
	NodeInfoMessageType = message.DefineType("NODEINFO")
)

// TypeOfMessage TODO
func TypeOfMessage(m message.Message) (message.MessageType, error) {
	switch m.(type) {
	case *Ping:
		return PingMessageType, nil
	case *Pong:
		return PongMessageType, nil
	default:
		return 0, message.ErrUnknownMessageType
	}
}

// NewMessageByType TODO
func NewMessageByType(t message.MessageType) (message.Message, error) {
	switch t {
	case PingMessageType:
		return new(Ping), nil
	case PongMessageType:
		return new(Pong), nil
	default:
		return nil, message.ErrUnknownMessageType
	}
}
