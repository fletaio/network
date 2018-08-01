package concentrator

import (
	"errors"
	util "fleta/samutil"
	"net"
)

//concentrator error list
var (
	ErrNotFoundCaster  = errors.New("ErrNotFoundCaster")
	ErrNotFoundCommand = errors.New("ErrNotFoundCommand")
	ErrEmptyPacket     = errors.New("ErrEmpyPacket")
)

//ConnStore TODO
type EmbeddedCaster interface {
	GetConnList() []net.Conn
	Location() string
	VisualizationData() []string
	RegisteredRouter() error
	Close()
}

//IRouter is impl list of Router
type IRouter interface {
	Log(format string, msg ...interface{})
	Debug(format string, msg ...interface{})
	Error(format string, msg ...interface{})
	Localhost() string
}

//ICaster TODO
type ICaster interface {
	EmbeddedCaster

	setRouter(c *Router)
	GetCommands(string) func(net.Conn, util.FletaPacket) (exit bool, err error)
	BroadCast(fp util.FletaPacket) ([]string, error)
}
