package flanet

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"fleta/flanetinterface"
	"fleta/formulator"
	"fleta/minninggroup"
	"fleta/mock"
	"fleta/mock/mockblock"
	"fleta/mock/mocknet"
	"fleta/peerlist"
	util "fleta/samutil"
	"fleta/samutil/concentrator"
)

//flanet error list
var (
	ErrUnregisteredCaster = errors.New("Unregistered Caster")
)

//Flanet is gateway of fleta
type Flanet struct {
	sync.Mutex
	flanetID   int
	flanetType string
	concentrator.Router
}

//NewFlanet is instance of flanet
func NewFlanet(i int, nodeType string) *Flanet {
	f := &Flanet{
		flanetID:   i,
		flanetType: nodeType,
	}
	f.Router.Init(f, nodeType)
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

//TODO MOCK DATA
func (f *Flanet) GetObserverNodeAddr() string {
	from := simulationdata.ObserverNodeStartIndex
	to := simulationdata.ObserverNodeStartIndex + simulationdata.ObserverNodeCount
	targetID := rand.Intn(to-from) + from
	return util.Sha256HexInt(targetID)
}

//IFormulator is requested to implement a list of Formulator functions
type IFormulator interface {
	concentrator.ConnStore
	DetectFormulatorNode(node flanetinterface.Node)
	NewPeer(address string)
	FormulatorList() ([]*flanetinterface.Node, error)
	SelfNode() *flanetinterface.Node
	CheckFormulator(addr string) error
}

func (f *Flanet) formulator() (IFormulator, error) {
	var i IFormulator
	i = &formulator.Formulator{}
	connStore := f.LocalRouter(i.Location())
	if cs, ok := connStore.(IFormulator); ok {
		return cs, nil
	}
	return nil, ErrUnregisteredCaster
}

/*
Formulator Impls
*/
func (f *Flanet) DetectFormulatorNode(node flanetinterface.Node) error {
	fm, err := f.formulator()
	if err != nil {
		return err
	}
	fm.DetectFormulatorNode(node)
	return nil
}

func (f *Flanet) NewPeer(peer string) error {
	fm, err := f.formulator()
	if err != nil {
		return err
	}
	fm.NewPeer(peer)
	return nil
}

func (f *Flanet) FormulatorList() ([]*flanetinterface.Node, error) {
	fm, err := f.formulator()
	if err != nil {
		return nil, err
	}
	return fm.FormulatorList()
}

func (f *Flanet) CheckFormulator(addr string) error {
	fm, err := f.formulator()
	if err != nil {
		return err
	}
	return fm.CheckFormulator(addr)
}

//IPeerList is requested to implement a list of Peerlist functions
type IPeerList interface {
	concentrator.ConnStore
	BroadCastToFormulator(fp util.FletaPacket) error
	PeerList() []flanetinterface.Node
	PeerSend(address string, fp util.FletaPacket) error
}

func (f *Flanet) peerlist() (IPeerList, error) {
	var i IPeerList
	i = &peerlist.PeerList{}
	connStore := f.LocalRouter(i.Location())
	if cs, ok := connStore.(IPeerList); ok {
		return cs, nil
	}
	return nil, ErrUnregisteredCaster

}

/*
PeerList Impls
*/
func (f *Flanet) BroadCastToFormulator(fp util.FletaPacket) error {
	p, err := f.peerlist()
	if err != nil {
		return err
	}
	return p.BroadCastToFormulator(fp)
}

func (f *Flanet) PeerList() ([]flanetinterface.Node, error) {
	p, err := f.peerlist()
	if err != nil {
		return nil, err
	}

	return p.PeerList(), nil
}

func (f *Flanet) PeerSend(addr string, val util.FletaPacket) error {
	p, err := f.peerlist()
	if err != nil {
		return err
	}
	return p.PeerSend(addr, val)
}

//IMinninggroup is requested to implement a list of Block functions
type IMinningGroup interface {
	concentrator.ConnStore
	NewBlock(block *mockblock.Block) error
	NewFormulator(node *flanetinterface.Node)
}

func (f *Flanet) minninggroup() (IMinningGroup, error) {
	var i IMinningGroup
	i = &minninggroup.MinningGroup{}
	connStore := f.LocalRouter(i.Location())
	if cs, ok := connStore.(IMinningGroup); ok {
		return cs, nil
	}
	return nil, ErrUnregisteredCaster
}

func (f *Flanet) NewFormulator(node *flanetinterface.Node) error {
	m, err := f.minninggroup()
	if err != nil {
		return err
	}
	m.NewFormulator(node)
	return nil
}

func (f *Flanet) NewBlock(block *mockblock.Block) error {
	m, err := f.minninggroup()
	if err != nil {
		return err
	}

	return m.NewBlock(block)
}

/*
Block Impls
*/

//IBlock is requested to implement a list of Block functions
type ISync interface {
	concentrator.ConnStore
	MakeBlock(node *flanetinterface.Node) error
	GetMakeBlockTime(addr string) (time.Time, error)
	SeedNodeAddr() (string, error)
	GetBlockHeight() int
}

func (f *Flanet) sync() (ISync, error) {
	var i ISync
	i = &mockblock.Sync{}
	connStore := f.LocalRouter(i.Location())
	if cs, ok := connStore.(ISync); ok {
		return cs, nil
	}
	return nil, ErrUnregisteredCaster
}

/*
Block Impls
*/
func (f *Flanet) MakeBlock() error {
	b, err := f.sync()
	if err != nil {
		return err
	}
	fm, err := f.formulator()
	if err != nil {
		return err
	}

	node := fm.SelfNode()
	b.MakeBlock(node)

	return nil
}

func (f *Flanet) GetMakeBlockTime(addr string) (time.Time, error) {
	b, err := f.sync()
	if err != nil {
		return time.Time{}, err
	}

	return b.GetMakeBlockTime(addr)
}

func (f *Flanet) SeedNodeAddr() (string, error) {
	b, err := f.sync()
	if err != nil {
		return "", err
	}

	return b.SeedNodeAddr()
}
func (f *Flanet) GetBlockHeight() int {
	b, err := f.sync()
	if err != nil {
		return -1
	}

	return b.GetBlockHeight()
}
