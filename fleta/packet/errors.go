package packet

import (
	"errors"
)

var (
	// ErrMismatchMagicword TODO
	ErrMismatchMagicword  = errors.New("mismatch magicword in header")
	ErrMismatchPacketSize = errors.New("mismatch packet size")
	ErrInvalidIntegrity   = errors.New("invalid integrity")
	ErrMessageUnresolved  = errors.New("unresolved message")
)
