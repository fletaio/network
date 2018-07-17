package network

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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
)

//IFleta Fleta interface
type IFleta interface {
	Start(i int, nodeType string) error
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
	NodeType      string
	NodeID        string
	ConnParamChan chan ConnParam
	ft            IFleta
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
	totalCount := simulationdata.ObserverNodeCount + simulationdata.FormulatorNodeCount + simulationdata.NormalNodeCount
	for i := 0; i < totalCount; i++ {
		appendNode(i)
	}
	for i := 0; i < totalCount; i++ {
		go func(i int) {
			err := LoadNodeMap(util.Sha256HexInt(i)).ft.Start(i, itoNodeType(i))
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}

	for j := 0; j < totalCount; j++ {
		go mockDataSend(j)
	}

	sender := make(chan concentrator.Visualization, 100)
	go sendVisualizationData(sender)
	concentrator.VisualizationStart(sender)
}

func sendVisualizationData(sender chan<- concentrator.Visualization) {
	for {
		time.Sleep(time.Second)

		totalCount := simulationdata.ObserverNodeCount + simulationdata.FormulatorNodeCount + simulationdata.NormalNodeCount
		data := make(map[string]map[string][]string)
		for i := 0; i < totalCount; i++ {
			idata := LoadNodeMap(util.Sha256HexInt(i)).ft.VisualizationData()
			data[util.Sha256HexInt(i)] = idata
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
	nodeInfo := NodeInfo{
		NodeType: nodeName,
		NodeID:   addr,
		ft:       fletaTest.NewFleta(),
	}

	StoreNodeMap(nodeInfo.NodeID, nodeInfo)
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
	Reader      *io.PipeReader
	Writer      *io.PipeWriter
	NetworkType string
	Address     string
	DialHost    string
}

//RegistDial is temp store reader and writer
func RegistDial(networkType, address string, localhost string) (_cRead *io.PipeReader, _cWrite *io.PipeWriter) {
	for LoadNodeMap(address).ConnParamChan == nil {
		time.Sleep(100 * time.Millisecond)
	}
	sRead, cWrite := io.Pipe()
	cRead, sWrite := io.Pipe()

	connParam := ConnParam{
		Reader:      sRead,
		Writer:      sWrite,
		NetworkType: networkType,
		Address:     address,
		DialHost:    localhost,
	}
	LoadNodeMap(address).ConnParamChan <- connParam

	return cRead, cWrite
}

//RegistAccept is temp store reader and writer
func RegistAccept(addr string) (node NodeInfo) {
	if LoadNodeMap(addr).NodeID == "" {
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
	log.Println(my)
	log.Println(address)

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
