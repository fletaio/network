package concentrator

import (
	util "fleta/samutil"
	"net"
)

//Caster TODO
type Caster struct {
	*Router
	EmbeddedCaster
	Commands map[string]func(net.Conn, util.FletaPacket) (exit bool, err error)
}

//Init TODO
func (nc *Caster) Init(cs EmbeddedCaster) {
	nc.EmbeddedCaster = cs
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
func (nc *Caster) ConsignmentCast(caster string, fp util.FletaPacket) ([]string, error) {
	for _, tnc := range nc.NcList {
		if tnc.Location() == caster {
			return tnc.BroadCast(fp)
		}
	}
	return nil, ErrNotFoundCaster
}

//BroadCast TODO
func (nc *Caster) BroadCast(fp util.FletaPacket) ([]string, error) {
	cl := nc.GetConnList()
	addrArr := make([]string, len(cl))
	for i, conn := range cl {
		addrArr[i] = conn.RemoteAddr().String()
		p, err := fp.Packet()
		if err != nil {
			return nil, err
		}
		conn.Write(p)
	}
	return addrArr, nil
}
