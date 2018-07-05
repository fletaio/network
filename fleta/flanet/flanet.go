package flanet

import (
	"fleta/flanetinterface"
	"fleta/formulator"
	"fleta/minninggroup"
	"fleta/mock/mockblock"
	"fleta/mock/mocknet"
	"fleta/peerlist"
	"fleta/util"
	"fleta/util/netcaster"
)

//Flanet is gateway of fleta
type Flanet struct {
	flanetID   int
	flanetType string
	*netcaster.CastRouter
}

//NewFlanet is instance of flanet
func NewFlanet(i int, nodeType string) *Flanet {
	f := &Flanet{
		flanetID:   i,
		flanetType: nodeType,
	}
	f.CastRouter = &netcaster.CastRouter{
		f,
		make([]*netcaster.NetCaster, 0),
	}
	return f
}

//GetNodeType return node type
func (f *Flanet) GetNodeType() string {
	return f.flanetType
}

//GetFlanetID  return ID
func (f *Flanet) getFlanetID() int {
	return f.flanetID
}

//Localhost return localhost
func (f *Flanet) Localhost() string {
	return util.Sha256HexInt(f.flanetID)
}

//OpenFlanet is wait of all accept
func (f *Flanet) OpenFlanet() error {
	listen, err := mocknet.Listen("tcp", ":3000")
	if err != nil {
		return err
	}
	ls := listen

	for {
		conn, err := ls.Accept()
		if err != nil {
			continue
		}
		go f.CommandRouter(conn)
	}

}

/*
Formulator Impls
*/
func (f *Flanet) DetectMasterNode(node flanetinterface.Node) {
	connStore := f.LocalRouter(formulator.Location())
	if fm, ok := connStore.(*formulator.Formulator); ok {
		fm.DetectMasterNode(node)
	}

}

func (f *Flanet) NewPeer(peer string) {
	connStore := f.LocalRouter(formulator.Location())
	if fm, ok := connStore.(*formulator.Formulator); ok {
		fm.NewPeer(peer)
	}

}

/*
PeerList Impls
*/
func (f *Flanet) PeerList() []flanetinterface.Node {
	connStore := f.LocalRouter(peerlist.Location())
	if cs, ok := connStore.(*peerlist.PeerList); ok {
		return cs.PeerList()
	}

	return nil
}
func (f *Flanet) PeerSend(addr string, val util.FletaPacket) error {
	connStore := f.LocalRouter(peerlist.Location())
	if cs, ok := connStore.(*peerlist.PeerList); ok {
		return cs.PeerSend(addr, val)
	}

	return nil
}

func (f *Flanet) ImMinningGroup(group []*mockblock.BlockGen) {
	connStore := f.LocalRouter(minninggroup.Location())
	if cs, ok := connStore.(*minninggroup.MinningGroup); ok {
		cs.GroupList = group
		go cs.ImMinningGroup()
	}
}

func (f *Flanet) CalculateScore() []*mockblock.BlockGen {
	connStore := f.LocalRouter(formulator.Location())
	if cs, ok := connStore.(*formulator.Formulator); ok {
		return cs.CalculateScore()
	}
	return nil
}
