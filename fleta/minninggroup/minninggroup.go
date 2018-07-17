package minninggroup

import (
	"net"
	"sort"
	"sync"
	"time"

	"fleta/flanetinterface"
	"fleta/mock/mockblock"
	"fleta/mock/mocknet"
	util "fleta/samutil"
	"fleta/samutil/concentrator"
)

//MinningGroupCount size of MinningGroup
const MinningGroupCount = 20

//MinningGroup TODO
type MinningGroup struct {
	fi             FlanetImpl
	imMinningGroup bool
	myScore        int
	GroupList      fList
	checkGroupLock sync.Mutex
	checkGroup     map[string]*flanetinterface.Node
	conns          []net.Conn
	concentrator.Caster
	requestBlockHeight int
}

//Location TODO
func (mg MinningGroup) Location() string {
	return "MG"
}

//New TODO
func New(fi FlanetImpl) *MinningGroup {
	mg := &MinningGroup{
		checkGroup: make(map[string]*flanetinterface.Node),
	}
	mg.fi = fi
	mg.Caster.Init(mg)

	mg.conns = make([]net.Conn, MinningGroupCount)

	mg.addProcessCommand()

	return mg
}

//RenewScore is renew formulator node score and spread to peerlist
func (mg *MinningGroup) RenewScore() {
	//TODO
}

func (mg *MinningGroup) NewBlock(block *mockblock.Block) error {
	mg.checkGroupLock.Lock()
	var err error
	if node, ok := mg.checkGroup[block.Addr]; ok {
		node.BlockTime = block.MakeBlockTime
		mg.reinsertSort(node)
	} else {
		err = mg.fi.CheckFormulator(block.Addr)
	}
	mg.checkGroupLock.Unlock()
	return err
}

func (mg *MinningGroup) NewFormulator(node *flanetinterface.Node) {
	mg.checkGroupLock.Lock()
	if _, ok := mg.checkGroup[node.Addr()]; !ok {
		if time, err := mg.fi.GetMakeBlockTime(node.Addr()); err == nil {
			node.BlockTime = time
		}
		mg.insertSort(node)
	}
	mg.checkGroupLock.Unlock()
}

func (mg *MinningGroup) meshNetwork() {
	localhost := mg.Localhost()
	mg.myScore = mg.getScore(localhost)
	if mg.myScore < 0 || mg.myScore >= 20 {
		return
	}
	mg.conns[mg.myScore] = nil
	for index := mg.myScore + 1; index < mg.myScore+(MinningGroupCount/2); index++ {
		i := (index + MinningGroupCount) % MinningGroupCount
		conn, err := mocknet.Dial("tcp", mg.GroupList[i].Addr(), mg.Localhost())
		if err == nil {
			fp := util.FletaPacket{
				Command: "MGEXLOOP",
			}
			p, err := fp.Packet()
			if err == nil {
				conn.Write(p)
				mg.SetConn(conn)
			}
		}
	}
}

//SetConn TODO
func (mg *MinningGroup) SetConn(conn net.Conn) {
	index := mg.getScore(conn.RemoteAddr().String())
	if index >= 0 && index < MinningGroupCount {
		mg.conns[index] = conn
	}

}

type fList []*flanetinterface.Node

func (a fList) Len() int      { return len(a) }
func (a fList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

//TODO add calculate block time
func (a fList) Less(i, j int) bool {
	iCom := a[i].BlockTime
	if iCom.Before(a[i].DetectedTime) {
		iCom = a[i].DetectedTime
	}
	jCom := a[j].BlockTime
	if jCom.Before(a[j].DetectedTime) {
		jCom = a[j].DetectedTime
	}
	return iCom.Before(jCom)
}

type Review struct {
	BlockHeight int
	NodeInfos   []NodeForReview
}
type NodeForReview struct {
	Addr string
	Time time.Time
}

func (mg *MinningGroup) groupListIndex(el *flanetinterface.Node) int {
	index := sort.Search(len(mg.GroupList), func(i int) bool {
		iCom := mg.GroupList[i].BlockTime
		if iCom.Before(mg.GroupList[i].DetectedTime) {
			iCom = mg.GroupList[i].DetectedTime
		}

		jCom := el.BlockTime
		if jCom.Before(el.DetectedTime) {
			jCom = el.DetectedTime
		}

		return iCom.Before(jCom)
	})

	return index
}

func (mg *MinningGroup) reinsertSort(el *flanetinterface.Node) {
	index := mg.groupListIndex(el)
	mg.GroupList = append(mg.GroupList[:index], mg.GroupList[index+1:]...)
	mg.insertSort(el)
}

func (mg *MinningGroup) insertSort(el *flanetinterface.Node) {
	index := mg.groupListIndex(el)
	mg.GroupList = append(mg.GroupList, &flanetinterface.Node{})
	copy(mg.GroupList[index+1:], mg.GroupList[index:])
	mg.GroupList[index] = el
	mg.checkGroup[el.Addr()] = el
}

//CalculateScore TODO
func (mg *MinningGroup) CalculateScore() {

	sort.Sort(fList(mg.GroupList))
	blockHeight := mg.fi.GetBlockHeight()
	mg.Log("CalculateScore %s", mg.requestBlockHeight < blockHeight)
	if len(mg.GroupList) > 20 && mg.requestBlockHeight < blockHeight {
		mg.requestBlockHeight = blockHeight
		mg.Log("score len %d %s", len(mg.GroupList), mg.GroupList[:3])

		obAddr := mg.fi.GetObserverNodeAddr()
		conn, err := mocknet.Dial("tcp", obAddr, mg.Localhost())
		if err != nil {
			mg.Error("%s", err)
			return
		}
		readyCh, pChan, _ := util.ReadFletaPacket(conn)
		go func() {
			fp := <-pChan
			//TODO
			mg.Log(fp.Command)
			conn.Close()
		}()

		minningCandidate := mg.GroupList[:20]

		mcs := Review{
			BlockHeight: blockHeight,
		}
		for _, node := range minningCandidate {
			iCom := node.BlockTime
			if iCom.Before(node.DetectedTime) {
				iCom = node.DetectedTime
			}

			mcs.NodeInfos = append(mcs.NodeInfos, NodeForReview{
				Addr: node.Addr(),
				Time: iCom,
			})
		}

		fletaPacket := util.FletaPacket{
			Command: "OBNDSCOR",
			Content: util.ToJSON(mcs),
		}

		p, err := fletaPacket.Packet()
		if err == nil {
			<-readyCh
			conn.Write(p)
		}
	}

}

func (mg *MinningGroup) getScore(addr string) int {
	gLen := len(mg.GroupList)
	for i := 0; i < gLen; i++ {
		if mg.GroupList[i].Addr() == addr {
			return i
		}
	}
	return -1
}

func (mg *MinningGroup) start() {
	for {
		mg.meshNetwork()
		if mg.myScore == 0 {
			mg.fi.MakeBlock()
			mg.imMinningGroup = false
			mg.Log("%s", mg.connsString())
		}
		time.Sleep(time.Second * 5)
	}
}

//connsString is handling process packet
func (mg *MinningGroup) connsString() []string {
	var connstr []string
	for i := 0; i < MinningGroupCount; i++ {
		if mg.conns[i] != nil {
			connstr = append(connstr, mg.conns[i].RemoteAddr().String())
		} else {
			connstr = append(connstr, "empty")
		}
	}
	return connstr
}

func (mg *MinningGroup) addProcessCommand() {
	mg.AddCommand("MGEXLOOP", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		mg.SetConn(conn)
		return true, nil
	})
}

//GetConnList GetConnList
func (mg *MinningGroup) GetConnList() []net.Conn {
	return mg.conns
}
func (mg *MinningGroup) VisualizationData() []string {
	list := []string{}
	for _, node := range mg.GroupList {
		list = append(list, node.Addr())
	}
	return list
}

//FlanetImpl TODO
type FlanetImpl interface {
	FormulatorList() ([]*flanetinterface.Node, error)
	MakeBlock() error
	GetMakeBlockTime(addr string) (time.Time, error)
	CheckFormulator(addr string) error
	GetObserverNodeAddr() string
	GetBlockHeight() int
}

//ImMinningGroup TODO
func (mg *MinningGroup) ImMinningGroup() {
	if mg.imMinningGroup == false {
		mg.imMinningGroup = true
		mg.start()
	}

}
