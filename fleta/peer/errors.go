package peer

import (
	"errors"
)

var (
	// ErrMismatchMagicword TODO
	ErrPeerNotExist = errors.New("peer not exist")
	ErrPeerNotAlive = errors.New("peer not alive")
	ErrAddExistPeer = errors.New("try add exist peer")
)
