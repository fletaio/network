package message

import (
	"errors"
)

var (
	// ErrMismatchMagicword TODO
	ErrUnvalidMessageType    = errors.New("unvalid message type")
	ErrUnknownMessageType    = errors.New("unknown payload type")
	ErrInvalidMsgConsumeSize = errors.New("invalid message consume size")
)
