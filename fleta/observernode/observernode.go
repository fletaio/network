package observernode

import (
	"errors"
	"fleta/mock/mocknet"
	"net"

	util "fleta/samutil"
	"fleta/samutil/concentrator"
)

//observernode error list
var (
	ErrInvalidIndex = errors.New("Invalid Index")
)

type ObserverNode struct {
	fi FlanetImpl
	concentrator.Caster

	observerList []net.Conn
}

func (o ObserverNode) Location() string {
	return "OB"
}
func (o *ObserverNode) GetConnList() []net.Conn {
	return o.observerList
}
func (o *ObserverNode) VisualizationData() []string {
	list := []string{}
	for _, conn := range o.observerList {
		if conn != nil {
			list = append(list, conn.RemoteAddr().String())
		}
	}
	return list
}

//RegisteredRouter TODO
func (o *ObserverNode) RegisteredRouter() error {
	return nil
}

//Close TODO
func (o *ObserverNode) Close() {
}

//TODO define observernode address
func New(fi FlanetImpl) *ObserverNode {
	o := &ObserverNode{
		observerList: make([]net.Conn, 4),
	}
	o.fi = fi
	o.Caster.Init(o)

	return o
}

//TODO define observernode address
func (o *ObserverNode) ConnectObserver(nodeIndex int) {
	add1 := util.Sha256HexInt((nodeIndex + 1) % 5)
	add2 := util.Sha256HexInt((nodeIndex + 3) % 5)

	conn1, err := mocknet.Dial("tcp", add1, o.Localhost())
	conn2, err := mocknet.Dial("tcp", add2, o.Localhost())
	o.observerList[0] = conn1
	o.observerList[2] = conn2

	fp := util.FletaPacket{
		Command: "OBSERVER",
	}

	fp.Content = "3"
	p, err := fp.Packet()
	if err != nil {
		o.Error("%s", err)
		return
	}
	conn1.Write(p)

	fp.Content = "1"
	p, err = fp.Packet()
	if err != nil {
		o.Error("%s", err)
		return
	}
	conn2.Write(p)

}

//MakeGenesisBlock TODO
// func (o *ObserverNode) MakeGenesisBlock() *mockblock.Block {
// 	block := &mockblock.Block{
// 		MakeBlockTime: time.Now(),
// 		Addr:          o.Localhost(),
// 		Height:        0,
// 	}
// 	return block
// }

//FlanetImpl TODO
type FlanetImpl interface {
	concentrator.IRouter
}
