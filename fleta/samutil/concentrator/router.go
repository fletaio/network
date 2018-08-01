package concentrator

import (
	util "fleta/samutil"
	"net"
)

//Router TODO
type Router struct {
	// FlanetID int
	IRouter
	Hint   string
	NcList []ICaster
}

//Init TODO
func (c *Router) Init(ir IRouter, hint string) {
	c.IRouter = ir
	c.Hint = hint
}

//Close TODO
func (c *Router) Close() {
	for _, nc := range c.NcList {
		nc.Close()
	}

}

//CommandRouter TODO
func (c *Router) CommandRouter(conn net.Conn) {
	exitGo := make(chan bool)
	fChan := make(chan util.FletaPacket, 1)
	go util.ReadLoopFletaPacket(fChan, conn, exitGo)
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
func (c *Router) LocalRouter(caster string) EmbeddedCaster {
	for _, tnc := range c.NcList {
		if tnc.Location() == caster {
			return tnc
		}
	}
	return nil
}

//GetHint TODO
func (c *Router) GetHint() string {
	return c.Hint
}

//RegistCaster TODO
func (c *Router) RegistCaster(nc ICaster) error {
	nc.setRouter(c)
	c.NcList = append(c.NcList, nc)
	return nc.RegisteredRouter()
	// c.Log("RegistCaster %d", len(c.NcList))
}

//VisualizationData TODO
func (c *Router) VisualizationData() map[string][]string {
	m := make(map[string][]string)
	for _, tnc := range c.NcList {
		m[tnc.Location()] = tnc.VisualizationData()
	}
	return m
}
