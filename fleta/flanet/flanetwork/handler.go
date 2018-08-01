package flanetwork

import (
	"fleta/message"
	"fleta/packet"
)

type MessageHandler struct {
	Next message.Resolver
	// init consumer channel
	Consumer chan message.Message
}

func NewMessageHandler(next message.Resolver, consumer chan message.Message) *MessageHandler {
	return &MessageHandler{
		Next:     next,
		Consumer: consumer,
	}
}

func (mp *MessageHandler) Resolve(msgType message.MessageType, payload packet.Payload) (message.Message, message.Processor, int64, error) {
	msg, err := NewMessageByType(msgType)
	if err == message.ErrUnknownMessageType {
		return mp.Next.Resolve(msgType, payload)
	} else if err != nil {
		return nil, nil, 0, err
	}

	n, err := msg.ReadFrom(payload)
	if err != nil {
		return nil, nil, n, err
	}

	return msg, mp, n, nil
}

func (mp *MessageHandler) Process(msg message.Message) error {

	if mp.Consumer == nil {
		return ErrConsumerNotExist
	}

	//send to consumer channel
	mp.Consumer <- msg

	return nil
}

// import (
// 	"flenet/message"
// 	"flenet/packet"
// )

// type MessageHandler struct {
// 	Next message.Resolver
// }

// func NewMessageHandler(next message.Resolver) *MessageHandler {
// 	return &MessageHandler{
// 		Next: next,
// 	}
// }

// func (mp *MessageHandler) Resolve(msgType message.MessageType, payload packet.Payload) (message.Message, message.Processor, int64, error) {
// 	msg, err := NewMessageByType(msgType)
// 	if err == message.ErrUnknownMessageType {
// 		return mp.Next.Resolve(msgType, payload)
// 	} else if err != nil {
// 		return nil, nil, 0, err
// 	}

// 	n, err := msg.ReadFrom(payload)
// 	if err != nil {
// 		return nil, nil, n, err
// 	}

// 	return msg, mp, n, nil
// }

// func (mp *MessageHandler) Process(msg message.Message) error {
// 	mType, err := TypeOfMessage(msg)
// 	if err != nil {
// 		return err
// 	}
// 	switch mType {
// 	case FormulatorListMessageType:

// 	}
// 	return nil
// }
