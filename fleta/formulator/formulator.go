package formulator

import (
	"net"
	"sort"
	"time"

	"fleta/flanetinterface"
	"fleta/mock/mockblock"
	"fleta/peerlist"
	"fleta/util"
	"fleta/util/hardstate"
	"fleta/util/netcaster"
)

//Formulator TODO
//Absolute([Formulator Transaction 생성 블록, 마지막 보상 블록] 중 가장 마지막 블록의 Hash – 해당 블록의 100 블록 전의 Hash)
// + (현재시간 - [Formulator Transaction 생성 블록 시간, 발견 시간, 마지막 보상 블록 시간] 중 가장 마지막 값)
type Formulator struct {
	// sync.Mutex
	scoreBoard []string
	fi         FormulatorImpl
	netcaster.NetCaster
	hardstate hardstate.HardState
}

//Location TODO
func Location() string {
	return "FM"
}

//Location TODO
func (fm *Formulator) Location() string {
	return Location()
}

//New TODO
func New(fi FormulatorImpl, cr *netcaster.CastRouter, hint string) *Formulator {
	fm := &Formulator{
		scoreBoard: make([]string, 0),
	}
	fm.fi = fi

	fm.NetCaster = netcaster.NetCaster{
		fm, cr, hint,
	}

	// fm.hardstate = hardstate.HardState{fm, &sync.Map{}, &sync.Map{}}
	fm.hardstate.Init(fm)
	go fm.hardstate.Start()

	return fm
}

type fList []*mockblock.BlockGen

func (a fList) Len() int      { return len(a) }
func (a fList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

//TODO add calculate block time
func (a fList) Less(i, j int) bool {
	iCom := a[i].MakeBlockTime
	if iCom == 0 {
		iCom = a[i].GenTime
	}
	jCom := a[j].MakeBlockTime
	if jCom == 0 {
		jCom = a[j].GenTime
	}
	return iCom < jCom
}

func (fm *Formulator) CalculateScore() []*mockblock.BlockGen {
	bt := mockblock.BlockGenTime
	sort.Sort(fList(bt))

	minningGroupCandidate := bt[:30]

	return minningGroupCandidate
}

//RenewScore is renew master node score and spread to peerlist
func (fm *Formulator) RenewScore() {
	for {
		time.Sleep(time.Second * 5)

		minningGroupCandidate := fm.CalculateScore()

		for _, node := range minningGroupCandidate {
			if node.Addr == fm.Localhost() {
				fm.fi.ImMinningGroup(minningGroupCandidate)
				break
			}
		}

		// fm.Log("%s", bt)
	}
}

//AddFormulatorMap keep whole MasterNode
func (fm *Formulator) SelfNode() flanetinterface.Node {
	return flanetinterface.Node{
		Address:  fm.Localhost(),
		NodeType: fm.Hint,
	}
}

//ProcessPacket is handling process packet
func (fm *Formulator) ProcessPacket(conn net.Conn, p util.FletaPacket) error {
	switch p.Command {
	case "FMHSRQFL":
		fl := fm.hardstate.FormulatorAddrList()
		fp := util.FletaPacket{
			Command:     "FMHSRPFL",
			Compression: true,
			Content:     util.ToJSON(fl),
		}

		p, err := fp.Packet()
		if err != nil {
			return err
		}
		conn.Write(p)
	case "FMHSRPFL":
		var nodes []string
		util.FromJSON(&nodes, p.Content)
		fm.hardstate.AddCandidateNodeAddr(nodes)
	default:
		cs := fm.LocalRouter(p.Command[:2])
		if cs != nil {
			return cs.ProcessPacket(conn, p)
		}
	}
	return nil
}

//GetConnList GetConnList
func (fm *Formulator) GetConnList() []net.Conn {
	conns := make([]net.Conn, 0)
	return conns
}

//PeerList TODO
func (fm *Formulator) PeerList() []flanetinterface.Node {
	return fm.fi.PeerList()
}

//Send TODO
func (fm *Formulator) Send(addr string, val []string) error {
	fp := util.FletaPacket{
		Command:     "FMHSSEND",
		Compression: true,
		Content:     util.ToJSON(val),
	}

	return fm.fi.PeerSend(addr, fp)
}

func (fm *Formulator) RequestFormulatorList(addr string) error {
	fp := util.FletaPacket{
		Command: "FMHSRQFL",
		Content: addr,
	}

	return fm.fi.PeerSend(addr, fp)
}

func (fm *Formulator) PeerRouter(fp util.FletaPacket) {
	fm.Router(peerlist.Location(), fp)
}

//FormulatorImpl TODO
type FormulatorImpl interface {
	PeerList() []flanetinterface.Node
	PeerSend(string, util.FletaPacket) error
	ImMinningGroup([]*mockblock.BlockGen)
}

//IFormulator TODO
type IFormulator interface {
	DetectMasterNode(flanetinterface.Node)
	NewPeer(address string)
	CalculateScore() []*mockblock.BlockGen
}

func (fm *Formulator) DetectMasterNode(node flanetinterface.Node) {
	fm.hardstate.AddCandidateNode([]flanetinterface.Node{node})
}
func (fm *Formulator) NewPeer(address string) {
	fm.hardstate.NewPeer(address)
}
