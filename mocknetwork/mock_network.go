package mocknetwork

import (
	"net"
	"strconv"
	"sync"
	"time"

	"git.fleta.io/fleta/mocknet"
	"git.fleta.io/fleta/mocknet/concentrator"
	"git.fleta.io/fleta/mocknet/util"

	"git.fleta.io/fleta/framework/log"
)

//IPlayer player interface
type IPlayer interface {
	Start() error
	MakeBlock()
	VisualizationData() map[string][]string
}

var pCreator func() IPlayer

//Player is set player
func Player(p func() IPlayer) {
	pCreator = p
}

// NodeInfo has node infomation type, ID, data channel
type NodeInfo struct {
	Address       string
	ConnParamChan chan ConnParam
	ft            IPlayer
}

//Addr TODO
func (n *NodeInfo) Addr() string {
	return n.Address
}

//DetectedTime TODO
func (n *NodeInfo) DetectedTime() time.Time {
	return time.Now()
}

//BlockTime TODO
func (n *NodeInfo) BlockTime() time.Time {
	return time.Now()
}

var nodeMap *sync.Map

func init() {
	nodeMap = &sync.Map{}
}

//LoadNodeMap is safe-thread map Load()
func LoadNodeMap(key string) NodeInfo {
	if value, ok := nodeMap.Load(key); ok {
		if val, ok := value.(NodeInfo); ok {
			return val
		}
	}
	return NodeInfo{}
}

//StoreNodeMap is safe-thread map Store()
func StoreNodeMap(key string, n NodeInfo) {
	nodeMap.Store(key, n)
}

//Run is start mocknetwork
func Run() {
	mockCount := simulationdata.InitNodeCount

	go func() {
		for i := 0; i < mockCount; i++ {
			AddFleta()
		}
	}()

	sender := make(chan concentrator.Visualization, 100)
	nodeAdder := make(chan concentrator.Msg, 100)
	go concentrator.VisualizationStart(sender, nodeAdder)
	go NodeAdder(nodeAdder)
	sendVisualizationData(sender)
}

var totalCount int
var addLock sync.Mutex

//AddFleta TODO
func AddFleta() {
	addLock.Lock()
	StartWithID(totalCount, func() {
		appendNode()
	})
	go func(i int) {
		StartWithID(i, func() {
			LoadNodeMap(util.Sha256HexInt(i)).ft.Start()
		})
		// go mockDataSend(i)
	}(totalCount)
	totalCount++
	// time.Sleep(time.Millisecond * 10)
	addLock.Unlock()
}

//NodeAdder TODO
func NodeAdder(sender <-chan concentrator.Msg) {
	for {
		msg := <-sender
		switch msg.Command {
		case "empty":
			log.Debug(msg.Num)
			// log.Println(msg.Num)
		case "addFormulator":
			for i := 0; i < msg.Num; i++ {
				AddFleta()
			}
		case "makeBlock":
			for i := 0; i < totalCount; i++ {
				StartWithID(i, func() {
					LoadNodeMap(util.Sha256HexInt(i)).ft.MakeBlock()
				})
			}
		case "makeBreak":
			log.Debug("makeBreak")
		}
	}
}

func sendVisualizationData(sender chan<- concentrator.Visualization) {
	for {
		time.Sleep(time.Second)

		data := make(map[string]map[string][]string)
		for i := 0; i < totalCount; i++ {

			idata := LoadNodeMap(util.Sha256HexInt(i)).ft.VisualizationData()
			if idata != nil {
				data[strconv.Itoa(i)] = idata
			}
		}
		sender <- data

	}
}

// func itoNodeType(i int) string {
// 	if i < simulationdata.FormulatorNodeStartIndex {
// 		return flanetinterface.ObserverNode
// 	} else if i < simulationdata.NormalNodeStartIndex {
// 		return flanetinterface.FormulatorNode
// 	} else {
// 		return flanetinterface.NormalNode
// 	}
// }

func appendNode() {
	i := GetSimulationID()
	appendNodeAddress(util.Sha256HexString(strconv.Itoa(i)))
}

func appendNodeAddress(addr string) {
	nodeInfo := NodeInfo{
		Address: addr,
	}
	StoreNodeMap(nodeInfo.Address, nodeInfo)
}

func childIDGen(nodeID string, index int) string {
	runes := []rune(nodeID)

	s := string(runes[index:]) + string(runes[:index])

	return util.Sha256HexString(s)
}

//ConnParam has Reader, Writer, network, address
type ConnParam struct {
	Conn        net.Conn
	NetworkType string
	Address     string
	DialHost    string
}

//RegistDial is temp store reader and writer
func RegistDial(networkType, address string, localhost string) net.Conn {
	// func RegistDial(networkType, address string, localhost string) (*io.PipeReader, *io.PipeWriter, chan bool) {
	for LoadNodeMap(address).ConnParamChan == nil {
		time.Sleep(100 * time.Millisecond)
	}

	s, c := net.Pipe()

	connParam := ConnParam{
		Conn:        s,
		NetworkType: networkType,
		Address:     address,
		DialHost:    localhost,
	}
	LoadNodeMap(address).ConnParamChan <- connParam

	// return cRead, cWrite, readyChan
	return c
}

//RegistAccept is temp store reader and writer
func RegistAccept(addr string) (node NodeInfo) {
	if LoadNodeMap(addr).Address == "" {
		appendNodeAddress(addr)
	}

	node = LoadNodeMap(addr)
	node.ConnParamChan = make(chan ConnParam, 256)
	StoreNodeMap(addr, node)

	return node
}

// //GetMainID is tracking call-stack until find "fleta.(*Fleta).Start" and return that ID
// //HARD DEFENDENCE ON MOCKNETWORK
// func GetMainID() string {
// 	buf := make([]byte, 1<<16)
// 	runtime.Stack(buf, true)
// 	strs := strings.Split(string(buf), "\n")
// 	re := regexp.MustCompile("fleta.\\(\\*Fleta\\).Start[^(]*\\(")
// 	for _, str := range strs {
// 		if strings.Contains(str, "fleta.(*Fleta).Start") {
// 			str = re.ReplaceAllLiteralString(str, "")
// 			str = strings.TrimRight(str, ")")
// 			ids := strings.Split(str, ",")
// 			if len(ids) >= 2 {
// 				var num int64
// 				num, err := strconv.ParseInt(strings.TrimPrefix(ids[1], " 0x"), 16, 32)
// 				if err != nil {
// 					log.Fatal(err)
// 				}
// 				i := int(num)
// 				return util.Sha256HexInt(i)

// 			}
// 		}
// 	}

// 	return util.Sha256HexInt(-1)
// }
