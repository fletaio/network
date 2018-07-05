package fleta

import (
	"sync"

	"fleta/flanet"
	"fleta/flanetinterface"
	"fleta/formulator"
	"fleta/minninggroup"
	"fleta/mock/network"
	"fleta/peerlist"
)

//Fleta struct
type Fleta struct {
	FletaID int
	Fn      flanet.Flanet
	sync.Mutex
}

//NewFleta NewFleta
func (f *Fleta) NewFleta() network.IFleta {
	return &Fleta{}
}

//Start Fleta
func (f *Fleta) Start(i int, nodeType string) {
	fn := flanet.NewFlanet(i, nodeType)
	f.Fn = *fn

	// mg := minninggroup.MinningGroup{}
	// f.Fn.SetMinninggroup(nodecommunicator.NodeCommunicator{&mg})
	// mg.SetFlanet(&f.Fn)

	// fm := *formulator.NewFormulator(fn.CastRouter, "FM", nodeType)

	pl := *peerlist.New(fn, fn.CastRouter, nodeType)
	fn.PutNetCaster(&pl.NetCaster)

	if f.Fn.GetNodeType() == flanetinterface.MasterNode || f.Fn.GetNodeType() == flanetinterface.NormalNode {
		go pl.RenewPeerList()
	}
	if f.Fn.GetNodeType() == flanetinterface.MasterNode {
		mg := *minninggroup.New(fn, fn.CastRouter, nodeType)
		fn.PutNetCaster(&mg.NetCaster)

		fm := *formulator.New(fn, fn.CastRouter, nodeType)
		fn.PutNetCaster(&fm.NetCaster)
		go fm.RenewScore()
	}

	f.Fn.OpenFlanet()

}
