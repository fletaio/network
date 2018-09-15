package mocknetwork

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"fleta/flanetinterface"
	"fleta/mock"
	util "fleta/samutil"
	"fleta/samutil/concentrator"

	"git.fleta.io/common/log"
)

//IFleta Fleta interface
type IFleta interface {
	Start(i int, nodeType string) error
	Close()
	NewFleta() IFleta
	VisualizationData() map[string][]string
}

var fletaTest IFleta

//Fleta is set Fleta
func Fleta(_fleta IFleta) {
	fletaTest = _fleta
}

// NodeInfo has node infomation type, ID, data channel
type NodeInfo struct {
	Address       string
	NodeType      string
	ConnParamChan chan ConnParam
	ft            IFleta
}

//Addr TODO
func (n *NodeInfo) Addr() string {
	return n.Address
}

//Type TODO
func (n *NodeInfo) Type() string {
	return n.NodeType
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
	os.RemoveAll("./hardstate")

	{
		// Trace
		defer log.WithTrace().Info("time to run")

		log.Debug("debug")
		log.Info("info")
		log.Notice("notice")
		log.Warn("warn")
		// log.Panic("panic") // this will panic
		log.Alert("alert")
		// log.Fatal("fatal") // this will call os.Exit(1)

		err := errors.New("the is an error")
		// logging with fields can be used with any of the above

		// predefined global fields
		log.WithError(err).Error("error")
		log.WithError(err).WithFields(log.F("key", "value")).Info("test info")
		log.Debug("error")

		log.WithField("key", "value").Info("testing default fields")

		// or request scoped default fields
		logger := log.WithFields(
			log.F("request", "req"),
			log.F("scoped", "sco"),
		)

		logger.WithField("key", "value").Info("test")

	}
	// mockCount := simulationdata.ObserverNodeCount + simulationdata.FormulatorNodeCount + simulationdata.NormalNodeCount
	mockCount := simulationdata.InitNodeCount

	go func() {
		for i := 0; i < mockCount; i++ {
			AddFleta(itoNodeType(i))
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
func AddFleta(nodeType string) {
	addLock.Lock()
	appendNode(totalCount)
	go func(i int, nodeType string) {
		defer LoadNodeMap(util.Sha256HexInt(i)).ft.Close()
		LoadNodeMap(util.Sha256HexInt(i)).ft.Start(i, nodeType)

		go mockDataSend(i)
	}(totalCount, nodeType)
	totalCount++
	time.Sleep(time.Millisecond * 100)
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
				AddFleta(flanetinterface.FormulatorNode)
			}
		case "addNormalNode":
			for i := 0; i < msg.Num; i++ {
				AddFleta(flanetinterface.NormalNode)
			}
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

func itoNodeType(i int) string {
	if i < simulationdata.FormulatorNodeStartIndex {
		return flanetinterface.ObserverNode
	} else if i < simulationdata.NormalNodeStartIndex {
		return flanetinterface.FormulatorNode
	} else {
		return flanetinterface.NormalNode
	}
}

func appendNode(i int) {
	appendNodeAddress(itoNodeType(i), util.Sha256HexString(strconv.Itoa(i)))
}

func appendNodeAddress(nodeName string, addr string) {
	var ft IFleta
	if fletaTest != nil {
		ft = fletaTest.NewFleta()
	}
	nodeInfo := NodeInfo{
		NodeType: nodeName,
		Address:  addr,
		ft:       ft,
	}

	StoreNodeMap(nodeInfo.Address, nodeInfo)
}

func childIDGen(nodeID string, index int) string {
	runes := []rune(nodeID)

	var strBuilder strings.Builder
	strBuilder.WriteString(string(runes[index:]))
	strBuilder.WriteString(string(runes[:index]))

	return util.Sha256HexString(strBuilder.String())
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
		appendNodeAddress(addr, addr)
	}

	node = LoadNodeMap(addr)
	node.ConnParamChan = make(chan ConnParam, 256)
	StoreNodeMap(addr, node)

	return node
}

//Ping is ping to input address
func Ping(address string) {
	my := GetMainID()
	log.Debug(my)
	log.Debug(address)
	// log.Println(my)
	// log.Println(address)

	myDecoded, err := hex.DecodeString(my)
	if err != nil {
		log.Fatal(err)
	}
	addressDecoded, err := hex.DecodeString(address)
	if err != nil {
		log.Fatal(err)
	}

	data := binary.BigEndian.Uint32(myDecoded[:4]) - binary.BigEndian.Uint32(addressDecoded[:4])

	fmt.Printf("%s\n", myDecoded)
	fmt.Printf("%s\n", addressDecoded)
	fmt.Printf("%d\n", data)
}

//GetMainID is tracking call-stack until find "fleta.(*Fleta).Start" and return that ID
//HARD DEFENDENCE ON MOCKNETWORK
func GetMainID() string {
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	strs := strings.Split(string(buf), "\n")
	re := regexp.MustCompile("fleta.\\(\\*Fleta\\).Start[^(]*\\(")
	for _, str := range strs {
		if strings.Contains(str, "fleta.(*Fleta).Start") {
			str = re.ReplaceAllLiteralString(str, "")
			str = strings.TrimRight(str, ")")
			ids := strings.Split(str, ",")
			if len(ids) >= 2 {
				var num int64
				num, err := strconv.ParseInt(strings.TrimPrefix(ids[1], " 0x"), 16, 32)
				if err != nil {
					log.Fatal(err)
				}
				i := int(num)
				return util.Sha256HexInt(i)

			}
		}
	}

	return util.Sha256HexInt(-1)
}
