package concentrator

import (
	"errors"
	"net"

	util "fleta/samutil"
)

//concentrator error list
var (
	ErrNotFoundCaster  = errors.New("ErrNotFoundCaster")
	ErrNotFoundCommand = errors.New("ErrNotFoundCommand")
	ErrEmptyPacket     = errors.New("ErrEmpyPacket")
)

//IRouter is impl list of Router
type IRouter interface {
	Log(format string, msg ...interface{})
	Debug(format string, msg ...interface{})
	Error(format string, msg ...interface{})
	Localhost() string
	// GetHint() string
}

//Router TODO
type Router struct {
	// FlanetID int
	IRouter
	Hint   string
	NcList []ICaster
}

//CommandRouter TODO
func (c *Router) Init(ir IRouter, hint string) {
	c.IRouter = ir
	c.Hint = hint
}

//CommandRouter TODO
func (c *Router) CommandRouter(conn net.Conn) {
	readyToReadChan := make(chan bool)
	exitGo := make(chan bool)
	fChan := make(chan util.FletaPacket, 1)
	go util.ReadLoopFletaPacket(fChan, conn, readyToReadChan, exitGo)
	<-readyToReadChan
	close(readyToReadChan)
	for {
		fp := <-fChan
		if fp.Command == "" {
			break
		}

		exit, err := c.RunCommand(conn, fp)
		if err != nil {
			c.Error("%s %s", err, fp)
		}
		exitGo <- exit

	}

}

//RunCommand TODO
func (c *Router) RunCommand(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
	if fp.Command == "" {
		return false, ErrEmptyPacket
	}
	for _, nc := range c.NcList {
		location := nc.Location()
		if fp.Command[:2] == location {
			if function := nc.GetCommands(fp.Command); function != nil {
				return function(conn, fp)
			} else {
				return false, ErrNotFoundCommand
			}
		}
	}
	c.Log("cLocalhost : %s", fp.Command[:2])
	return false, ErrNotFoundCaster
}

//LocalRouter TODO
func (c *Router) LocalRouter(caster string) ConnStore {
	for _, tnc := range c.NcList {
		if tnc.Location() == caster {
			return tnc
		}
	}
	return nil
}

//LocalRouter TODO
func (c *Router) GetHint() string {
	return c.Hint
}

//RegistCaster TODO
func (c *Router) RegistCaster(nc ICaster) {
	nc.setRouter(c)
	c.NcList = append(c.NcList, nc)
	// c.Log("RegistCaster %d", len(c.NcList))
}

func (c *Router) VisualizationData() map[string][]string {
	m := make(map[string][]string)
	for _, tnc := range c.NcList {
		m[tnc.Location()] = tnc.VisualizationData()
	}
	return m
}

//ConnStore TODO
type ConnStore interface {
	GetConnList() []net.Conn
	Location() string
	VisualizationData() []string
}

//ICaster TODO
type ICaster interface {
	ConnStore

	setRouter(c *Router)
	GetCommands(string) func(net.Conn, util.FletaPacket) (exit bool, err error)
	BroadCast(fp util.FletaPacket) error
}

//Caster TODO
type Caster struct {
	*Router
	ConnStore
	Commands map[string]func(net.Conn, util.FletaPacket) (exit bool, err error)
}

func (nc *Caster) Init(cs ConnStore) {
	nc.ConnStore = cs
}

func (nc *Caster) setRouter(c *Router) {
	nc.Router = c
}

//GetHint TODO
func (nc *Caster) GetHint() string {
	return nc.Router.GetHint()
}

//GetCommands TODO
func (nc *Caster) GetCommands(key string) func(net.Conn, util.FletaPacket) (exit bool, err error) {
	return nc.Commands[key]
}

//AddCommand TODO
func (nc *Caster) AddCommand(command string, f func(net.Conn, util.FletaPacket) (exit bool, err error)) {
	if nc.Commands == nil {
		nc.Commands = make(map[string]func(net.Conn, util.FletaPacket) (exit bool, err error))
	}
	nc.Commands[command] = f
}

//ConsignmentCast TODO
func (nc *Caster) ConsignmentCast(caster string, fp util.FletaPacket) error {
	for _, tnc := range nc.NcList {
		if tnc.Location() == caster {
			tnc.BroadCast(fp)
			return nil
		}
	}
	return ErrNotFoundCaster
}

//BroadCast TODO
func (nc *Caster) BroadCast(fp util.FletaPacket) error {
	cl := nc.GetConnList()
	for _, conn := range cl {
		p, err := fp.Packet()
		if err != nil {
			return err
		}
		conn.Write(p)
	}
	return nil
}
