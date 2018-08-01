package fleta

import (
	"sync"

	"fleta/discovery"
	"fleta/message"
	"fleta/peer"

	"fleta/flanet"
	"fleta/flanet/flanetwork"
	"fleta/flanetinterface"
	"fleta/mock/mockblock"
	"fleta/mock/mocknetwork"
)

//Fleta struct
type Fleta struct {
	FletaID int
	Fn      *flanet.Flanet
	sync.Mutex
}

//NewFleta NewFleta
func (f *Fleta) NewFleta() mocknetwork.IFleta {
	return &Fleta{}
}

func (f *Fleta) VisualizationData() map[string][]string {
	if f.Fn != nil {
		return f.Fn.VisualizationData()
	}
	return nil
}

//Start i is node identify nodeType is fleta type than defined on "flanetinterface"
//	ObserverNode   = "OBSERVER_NODE"
//	FormulatorNode = "FORMULATOR_NODE"
//	NormalNode     = "NORMAL_NODE"
func (f *Fleta) Start(i int, nodeType string) error {
	fn, err := flanet.NewFlanet(i, nodeType)
	if err != nil {
		return err
	}
	f.Fn = fn

	consumer := make(chan message.Message)
	flanetConsumer := make(chan message.Message)

	//handler chaining
	h1 := flanetwork.NewMessageHandler(nil, flanetConsumer)
	h2 := discovery.NewMessageHandler(h1, consumer)
	//init with start handler
	pm := peer.NewManager(h2)

	fn.FlanetConsumer(flanetConsumer)
	fn.PeerManager(pm)

	//observernode 의 경우 seednode를 찾는다거나 peer관리를 하지 않습니다.
	//추후 만들어진 peermanager를 사용하여 observernode 간의 연결을 만들예정입니다.
	if f.Fn.GetNodeType() == flanetinterface.ObserverNode {
		// o := observernode.New(fn)
		// go o.ConnectObserver(i)
	} else {
		// pl := peerlist.New(fn)
		// fn.RegistCaster(pl)

	}
	sync := mockblock.New(fn)
	fn.Sync(sync)

	f.Fn.OpenFlanet() // Listen을 하는 부분입니다.

	return nil
}

//Close Fleta
func (f *Fleta) Close() {
	f.Fn.Close()
}
