package minninggroup

import (
	"errors"
	"net"
	"sort"
	"sync"
	"time"

	"fleta/flanetinterface"
	"fleta/formulator"
	"fleta/mock/mocknet"
	util "fleta/samutil"
	"fleta/samutil/concentrator"
)

//minninggroup error list
var (
	ErrNotFoundCommand = errors.New("NotFoundCommand")
	ErrNotFoundConn    = errors.New("NotFoundConn")
)

//MinningGroupCount size of MinningGroup
const MinningGroupCount = 20

//MinningGroup TODO
type MinningGroup struct {
	fi        FlanetImpl
	myScore   int
	GroupLock sync.Mutex
	GroupList fList
	GroupMap  map[string]formulator.Node
	connLock  sync.Mutex
	conns     map[string]net.Conn
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
		GroupMap: make(map[string]formulator.Node),
		conns:    make(map[string]net.Conn),
		fi:       fi,
	}
	mg.Caster.Init(mg)
	mg.addProcessCommand()

	return mg
}

func (mg *MinningGroup) checkMyScore() {
	//TODO
	localhost := mg.Localhost()
	mg.myScore = mg.getScore(localhost)
	if len(mg.GroupList) >= MinningGroupCount && mg.myScore <= MinningGroupCount {
		mg.requestObserverAmIMinningGroup()
	}

}

func (mg *MinningGroup) requestObserverAmIMinningGroup() {
	mg.start()
	// mg.
}

// func (mg *MinningGroup) NewBlock(block *mockblock.Block) error {
// 	mg.GroupLock.Lock()
// 	defer mg.GroupLock.Unlock()
// 	var err error
// 	if node, ok := mg.GroupMap[block.Addr]; ok {
// 		node.Block = block.MakeBlockTime
// 		mg.reinsertSort(node)
// 		mg.checkMyScore()
// 	} else {
// 		err = mg.fi.CheckFormulator(block.Addr)
// 	}
// 	return err
// }

//NewFormulator TODO
func (mg *MinningGroup) NewFormulator(node flanetinterface.Node) {
	mg.GroupLock.Lock()
	defer mg.GroupLock.Unlock()
	if _, ok := mg.GroupMap[node.Addr()]; !ok {
		mg.insertSort(formulator.Node{
			Address: node.Addr(),
			// Detected: node.DetectedTime(),
			// Block:    node.BlockTime(),
		})
		mg.checkMyScore()
	}
}

func (mg *MinningGroup) meshNetwork() {
	localhost := mg.Localhost()
	mg.myScore = mg.getScore(localhost)
	if mg.myScore < 0 || mg.myScore >= 20 {
		return
	}

	mg.connLock.Lock()
	addrList := make([]string, MinningGroupCount)

	for i := mg.myScore + 1; i < MinningGroupCount; i++ {
		addrList[i-mg.myScore-1] = mg.GroupList[i].Addr()
	}
	mg.connLock.Unlock()

	for _, addr := range addrList {
		if _, ok := mg.conns[addr]; !ok && addr != "" {
			conn, err := mocknet.Dial("tcp", addr, mg.Localhost())
			if err == nil {
				fp := util.FletaPacket{
					Command: "MGEXLOOP",
				}
				p, err := fp.Packet()
				if err == nil {
					conn.Write(p)
					mg.setConn(conn)
				}
			}
		}
	}

}

func (mg *MinningGroup) setConn(conn net.Conn) {
	mg.connLock.Lock()
	if _conn, ok := mg.conns[conn.RemoteAddr().String()]; ok {
		_conn.Close()
	}
	mg.conns[conn.RemoteAddr().String()] = conn
	mg.connLock.Unlock()
	go mg.readPacket(conn)
}

func (mg *MinningGroup) broadCastPacket(fp util.FletaPacket) error {
	mg.connLock.Lock()
	defer mg.connLock.Unlock()
	for _, conn := range mg.conns {
		p, err := fp.Packet()
		if err != nil {
			return err
		}
		conn.Write(p)
		return nil
	}
	return ErrNotFoundConn
}

func (mg *MinningGroup) sendPacket(direct string, fp util.FletaPacket) error {
	mg.connLock.Lock()
	defer mg.connLock.Unlock()
	if conn, ok := mg.conns[direct]; ok {
		p, err := fp.Packet()
		if err != nil {
			return err
		}
		conn.Write(p)
		return nil
	}
	return ErrNotFoundConn
}

func (mg *MinningGroup) readPacket(conn net.Conn) {
	pChan, err := util.ReadFletaPacket(conn)
	for {
		fp, ok := <-pChan
		if !ok {
			mg.connLock.Lock()
			delete(mg.conns, conn.RemoteAddr().String())
			conn.Close()
			mg.connLock.Unlock()
			break
		}
		if err != nil {
			mg.Error("%s", err)
		}

		if function := mg.GetCommands(fp.Command); function != nil {
			function(conn, fp)
		} else {
			mg.Error("%s", ErrNotFoundCommand)
		}
	}

}

type fList []formulator.Node

func (a fList) Len() int      { return len(a) }
func (a fList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

//TODO add calculate block time
func (a fList) Less(i, j int) bool {
	// iCom := a[i].Block
	// if iCom.Before(a[i].Detected) {
	// 	iCom = a[i].Detected
	// }
	// jCom := a[j].Block
	// if jCom.Before(a[j].Detected) {
	// 	jCom = a[j].Detected
	// }
	// return iCom.Before(jCom)
	return false
}

//Review TODO
type Review struct {
	BlockHeight int
	NodeInfos   []NodeForReview
}

//NodeForReview TODO
type NodeForReview struct {
	Addr string
	Time time.Time
}

func (mg *MinningGroup) groupListIndex(el formulator.Node) int {
	index := sort.Search(len(mg.GroupList), func(i int) bool {
		// iCom := mg.GroupList[i].Block
		// if iCom.Before(mg.GroupList[i].Detected) {
		// 	iCom = mg.GroupList[i].Detected
		// }

		// jCom := el.Block
		// if jCom.Before(el.Detected) {
		// 	jCom = el.Detected
		// }

		// return iCom.Before(jCom)
		return false
	})

	return index
}

func (mg *MinningGroup) reinsertSort(el formulator.Node) {
	index := mg.groupListIndex(el)
	mg.GroupList = append(mg.GroupList[:index], mg.GroupList[index+1:]...)
	mg.insertSort(el)
}

func (mg *MinningGroup) insertSort(el formulator.Node) {
	index := mg.groupListIndex(el)
	mg.GroupList = append(mg.GroupList, formulator.Node{})
	copy(mg.GroupList[index+1:], mg.GroupList[index:])
	mg.GroupList[index] = el
	mg.GroupMap[el.Addr()] = el
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
		pChan, _ := util.ReadFletaPacket(conn)
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
			// iCom := node.Block
			// if iCom.Before(node.Detected) {
			// 	iCom = node.Detected
			// }

			mcs.NodeInfos = append(mcs.NodeInfos, NodeForReview{
				Addr: node.Addr(),
				// Time: iCom,
			})
		}

		fletaPacket := util.FletaPacket{
			Command: "OBNDSCOR",
			Content: util.ToJSON(mcs),
		}

		p, err := fletaPacket.Packet()
		if err == nil {
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
	mg.meshNetwork()
	if mg.myScore == 0 {
		mg.fi.MakeBlock()
		mg.Log("%s", mg.connsString())
	}
}

//connsString is handling process packet
func (mg *MinningGroup) connsString() []string {
	var connstr []string

	mg.connLock.Lock()
	defer mg.connLock.Unlock()
	for conn := range mg.conns {
		connstr = append(connstr, conn)
	}
	return connstr
}

func (mg *MinningGroup) addProcessCommand() {
	mg.AddCommand("MGEXLOOP", func(conn net.Conn, fp util.FletaPacket) (exit bool, err error) {
		mg.setConn(conn)
		return true, nil
	})
}

//GetConnList GetConnList
func (mg *MinningGroup) GetConnList() []net.Conn {
	var conns []net.Conn

	mg.connLock.Lock()
	defer mg.connLock.Unlock()
	for _, conn := range mg.conns {
		conns = append(conns, conn)
	}

	return conns
}

//VisualizationData TODO
func (mg *MinningGroup) VisualizationData() []string {
	list := []string{}
	mg.connLock.Lock()
	defer mg.connLock.Unlock()
	for _, conn := range mg.conns {
		if conn != nil {
			list = append(list, conn.RemoteAddr().String())
		}
	}
	return list
}

//RegisteredRouter TODO
func (mg *MinningGroup) RegisteredRouter() error {
	return nil
}

//Close TODO
func (mg *MinningGroup) Close() {
}

//FlanetImpl TODO
type FlanetImpl interface {
	// FormulatorList() ([]flanetinterface.Node, error)
	MakeBlock() error
	GetMakeBlockTime(addr string) (time.Time, error)
	CheckFormulator(addr string) error
	GetObserverNodeAddr() string
	GetBlockHeight() int
}
