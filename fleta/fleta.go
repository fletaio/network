package fleta

import (
	"fleta/formulator"
	"fleta/minninggroup"
	"fleta/observernode"
	"fleta/peerlist"
	"sync"

	"fleta/flanet"
	"fleta/flanetinterface"
	"fleta/mock/mockblock"
	"fleta/mock/network"
)

//Fleta struct
type Fleta struct {
	FletaID int
	Fn      *flanet.Flanet
	sync.Mutex
}

//NewFleta NewFleta
func (f *Fleta) NewFleta() network.IFleta {
	return &Fleta{}
}

func (f *Fleta) VisualizationData() map[string][]string {
	return f.Fn.VisualizationData()
}

//Start Fleta
func (f *Fleta) Start(i int, nodeType string) error {
	fn := flanet.NewFlanet(i, nodeType)
	f.Fn = fn

	if f.Fn.GetNodeType() == flanetinterface.ObserverNode {
		o := observernode.New(fn)
		fn.RegistCaster(o)
		go o.ConnectObserver(i)
	} else {
		pl := peerlist.New(fn)
		fn.RegistCaster(pl)

		b := mockblock.New(fn)
		fn.RegistCaster(b)
		b.Start()

		fm := formulator.New(fn)
		fn.RegistCaster(fm)
		fm.Start()

		if f.Fn.GetNodeType() == flanetinterface.FormulatorNode {
			mg := minninggroup.New(fn)
			fn.RegistCaster(mg)
			go mg.RenewScore()
			go pl.RenewPeerList()
		}
		if f.Fn.GetNodeType() == flanetinterface.NormalNode {
			go pl.RenewPeerList()
		}
	}

	f.Fn.OpenFlanet()
	return nil

}
