package router

import "errors"

//router error list
var (
	ErrNotConnected              = errors.New("not connected")
	ErrMismatchHashSize          = errors.New("mismatch hash size")
	ErrNotFoundLogicalConnection = errors.New("not found logical connection")
	ErrListenFirst               = errors.New("not found listen address")
	ErrDuplicateAccept           = errors.New("cannot accept ")
)
