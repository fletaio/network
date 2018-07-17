package formulator

import (
	"fleta/peerlist"
	"net"

	"fleta/flanetinterface"
	util "fleta/samutil"
	"fleta/samutil/concentrator"
	"fleta/samutil/hardstate"
)

//Formulator TODO
//Absolute([Formulator Transaction 생성 블록, 마지막 보상 블록] 중 가장 마지막 블록의 Hash – 해당 블록의 100 블록 전의 Hash)
// + (현재시간 - [Formulator Transaction 생성 블록 시간, 발견 시간, 마지막 보상 블록 시간] 중 가장 마지막 값)
type Formulator struct {
	// sync.Mutex
	fi FlanetImpl
	concentrator.Caster

	hardstate hardstate.HardState
}

//Location TODO
func (fm Formulator) Location() string {
	return "FM"
}

//New TODO
func New(fi FlanetImpl) *Formulator {
	fm := &Formulator{}
	fm.fi = fi

	fm.Caster.Init(fm)

	fm.addProcessCommand()

	fm.hardstate.Init(fm)
	return fm
}

//Start TODO
func (fm *Formulator) Start() {
	fm.hardstate.StartLoop()
}

//SelfNode return self info node
func (fm *Formulator) SelfNode() *flanetinterface.Node {
	var result *flanetinterface.Node
	if node, ok := fm.hardstate.GetNode(fm.Localhost()); ok {
		result = node
	} else {
		result := &flanetinterface.Node{
			Address:  fm.Localhost(),
			NodeType: fm.GetHint(),
		}
		fm.hardstate.AddCandidateNode([]*flanetinterface.Node{result})
	}
	return result
}

//CheckFormulator TODO
func (fm *Formulator) CheckFormulator(addr string) error {
	return fm.hardstate.AddCandidateNodeAddr([]string{addr})
}

//GetConnList TODO
func (fm *Formulator) GetConnList() []net.Conn {
	conns := make([]net.Conn, 0)
	return conns
}

func (fm *Formulator) VisualizationData() []string {
	list, err := fm.hardstate.FormulatorAddrList()
	if err != nil {
		fm.Error("%s", err)
	}
	return list
}

//PeerList TODO
func (fm *Formulator) PeerList() ([]flanetinterface.Node, error) {
	return fm.fi.PeerList()
}

//NewFormulator TODO
func (fm *Formulator) NewFormulator(node *flanetinterface.Node) error {
	err := fm.fi.NewFormulator(node)
	return err
}

//SendNewFormulators TODO
func (fm *Formulator) SendNewFormulators(addr []string) error {
	fp := util.FletaPacket{
		Command:     "FMHSSEND",
		Compression: true,
		Content:     util.ToJSON(addr),
	}

	return fm.fi.BroadCastToFormulator(fp)
}

//RequestFormulatorList TODO
func (fm *Formulator) RequestFormulatorList(addr string) error {
	fp := util.FletaPacket{
		Command: "FMHSRQFL",
		Content: addr,
	}

	return fm.fi.PeerSend(addr, fp)
}

//PeerRouter TODO
func (fm *Formulator) PeerRouter(fp util.FletaPacket) error {
	return fm.ConsignmentCast(peerlist.PeerList{}.Location(), fp)
}

//FlanetImpl TODO
type FlanetImpl interface {
	PeerList() ([]flanetinterface.Node, error)
	PeerSend(string, util.FletaPacket) error
	BroadCastToFormulator(fp util.FletaPacket) error
	NewFormulator(node *flanetinterface.Node) error
}

//DetectFormulatorNode TODO
func (fm *Formulator) DetectFormulatorNode(node flanetinterface.Node) {
	fm.hardstate.AddCandidateNode([]*flanetinterface.Node{&node})
}

//NewPeer TODO
func (fm *Formulator) NewPeer(address string) {
	fm.hardstate.NewPeer(address)
}

//FormulatorList return formulator list
func (fm *Formulator) FormulatorList() ([]*flanetinterface.Node, error) {
	return fm.hardstate.FormulatorList()
}
