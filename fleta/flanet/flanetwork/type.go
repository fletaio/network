package flanetwork

import "fleta/message"

//hss commands
const (
	FormulatorListMessageType   = message.MessageType(50) // + iota
	AskFormulatorMessageType    = message.MessageType(51)
	AnswerFormulatorMessageType = message.MessageType(52)
)

// TypeOfMessage TODO
func TypeOfMessage(m message.Message) (message.MessageType, error) {
	switch m.(type) {
	case *FormulatorList:
		return FormulatorListMessageType, nil
	case *AskFormulator:
		return AskFormulatorMessageType, nil
	case *AnswerFormulator:
		return AnswerFormulatorMessageType, nil
	default:
		return 0, message.ErrUnknownMessageType
	}
}

// NewMessageByType TODO
func NewMessageByType(t message.MessageType) (message.Message, error) {
	switch t {
	case FormulatorListMessageType:
		return new(FormulatorList), nil
	case AskFormulatorMessageType:
		return new(AskFormulator), nil
	case AnswerFormulatorMessageType:
		return new(AnswerFormulator), nil
	default:
		return nil, message.ErrUnknownMessageType
	}
}
