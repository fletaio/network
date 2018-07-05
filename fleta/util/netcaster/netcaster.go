package netcaster

import (
	"fleta/util"
	"net"
)

//ICastRouter TODO
type ICastRouter interface {
	Log(format string, msg ...interface{})
	Localhost() string
}

//CastRouter TODO
type CastRouter struct {
	// FlanetID int
	ICastRouter
	NcList []*NetCaster
}

//CommandRouter TODO
func (c *CastRouter) CommandRouter(conn net.Conn) {
	readyToReadChan := make(chan bool)
	fChan := make(chan util.FletaPacket, 1)
	go util.ReadLoopFletaPacket(fChan, conn, readyToReadChan)
	<-readyToReadChan
	close(readyToReadChan)
	for {
		fp := <-fChan
		if fp.Command == "" {
			break
		}
		for _, nc := range c.NcList {
			if fp.Command[:2] == nc.CS.Location() {
				nc.CS.ProcessPacket(conn, fp)
				break
			}
		}
		if fp.Command == "MGEXLOOP" {
			return
		}
	}

}

//LocalRouter TODO
func (c *CastRouter) LocalRouter(caster string) ConnStore {
	for _, tnc := range c.NcList {
		if tnc.CS.Location() == caster {
			return tnc.CS
		}
	}
	return nil
}

//PutNetCaster TODO
func (c *CastRouter) PutNetCaster(nc *NetCaster) {
	c.NcList = append(c.NcList, nc)
	// c.Log("PutNetCaster %d", len(c.NcList))
}

//ConnStore TODO
type ConnStore interface {
	GetConnList() []net.Conn
	ProcessPacket(net.Conn, util.FletaPacket) error
	Location() string
}

//NetCaster TODO
type NetCaster struct {
	CS ConnStore
	*CastRouter
	Hint string
}

//Router TODO
func (nc *NetCaster) Router(caster string, fp util.FletaPacket) {
	for _, tnc := range nc.NcList {
		if tnc.CS.Location() == caster {
			tnc.BroadCast(fp)
			return
		}
	}
}

//BroadCast TODO
func (nc NetCaster) BroadCast(fp util.FletaPacket) error {
	cl := nc.CS.GetConnList()
	for _, conn := range cl {
		p, err := fp.Packet()
		if err != nil {
			return err
		}
		conn.Write(p)
	}
	return nil
}
